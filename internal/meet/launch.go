package meet

import (
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"time"
)

const (
	defaultCommand = "claude"
	// The first message waits for the user to answer startup dialogs, such
	// as the folder trust check, so it is only dropped after a long time.
	firstMessageTimeout = 30 * time.Minute
	readyInterval       = time.Second
	// Agents report idle a moment before they take input, right after a
	// startup dialog closes, so readiness must hold for a few checks.
	readyChecks = 2
)

var ErrNameRequired = errors.New("the new agent needs a name")

// Launch starts a new agent and names it in the call; it joins as a
// participant on the next poll. Its first message, which tells it that its
// replies are spoken, is sent in the background once it is ready.
func (r *Room) Launch(ctx context.Context, request LaunchRequest) (string, error) {
	request.Name = strings.Join(strings.Fields(request.Name), " ")
	if request.Name == "" {
		return "", ErrNameRequired
	}
	if runes := []rune(request.Name); len(runes) > maxNameLength {
		request.Name = string(runes[:maxNameLength])
	}
	request.Command = strings.TrimSpace(request.Command)
	if request.Command == "" {
		request.Command = defaultCommand
	}

	id, err := r.cfg.Launcher.Launch(ctx, request)
	if id != "" {
		r.mu.Lock()
		r.names[id] = request.Name
		if saveErr := r.cfg.Names.Save(r.names); saveErr != nil {
			slog.Warn("save participant names", "err", saveErr)
		}
		r.mu.Unlock()
	}
	if err == nil {
		go r.sendWhenReady(id, r.withVoiceBriefing(request.Prompt))
	}
	return id, err
}

func (r *Room) sendWhenReady(agentID, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), firstMessageTimeout)
	defer cancel()

	ticker := time.NewTicker(readyInterval)
	defer ticker.Stop()
	ready, seen := 0, false
	for {
		agent, found, err := r.findAgent(ctx, agentID)
		switch {
		case err != nil:
		case found:
			seen = true
		case seen:
			slog.Info("agent left before its first message", "agent", agentID)
			return
		}

		if err == nil && found && r.takesInput(ctx, agent) {
			ready++
		} else {
			ready = 0
		}
		if ready >= readyChecks {
			if err := r.cfg.Messenger.SendMessage(ctx, agentID, message); err != nil {
				slog.Warn("send first message", "agent", agentID, "err", err)
			}
			return
		}

		select {
		case <-ctx.Done():
			slog.Warn("agent not ready for its first message", "agent", agentID)
			return
		case <-ticker.C:
		}
	}
}

func (r *Room) findAgent(ctx context.Context, agentID string) (Agent, bool, error) {
	agents, err := r.cfg.Source.ListAgents(ctx)
	if err != nil {
		return Agent{}, false, err
	}
	i := slices.IndexFunc(agents, func(agent Agent) bool { return agent.ID == agentID })
	if i < 0 {
		return Agent{}, false, nil
	}
	return agents[i], true, nil
}

// takesInput tells whether the agent waits at its input box rather than
// working or showing a dialog, which would swallow a message typed into it.
func (r *Room) takesInput(ctx context.Context, agent Agent) bool {
	if agent.Status != StatusIdle && agent.Status != StatusDone {
		return false
	}
	parser, ok := r.cfg.Parsers[agent.Kind]
	if !ok {
		return true
	}
	screen, err := r.cfg.Source.ReadScreen(ctx, agent.ID)
	return err == nil && parser.Parse(screen).Prompt == nil
}

func (r *Room) withVoiceBriefing(prompt string) string {
	locale := r.cfg.Locale
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return locale.VoiceBriefing + " " + locale.Introduction
	}
	return locale.VoiceBriefing + "\n\n" + prompt
}
