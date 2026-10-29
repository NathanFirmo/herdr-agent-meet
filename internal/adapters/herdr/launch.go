package herdr

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/NathanFirmo/herdr-agent-meet/internal/meet"
)

const (
	readyTimeout  = 90 * time.Second
	readyInterval = time.Second
)

// Launch opens a workspace without taking the user's focus and runs the
// command in its shell; Herdr recognizes the agent that starts there. The
// first message, if any, is sent in the background once the agent is ready.
func (s *Source) Launch(ctx context.Context, request meet.LaunchRequest) (string, error) {
	params := map[string]any{"label": request.Name, "focus": false}
	if request.Dir != "" {
		params["cwd"] = request.Dir
	}
	var created struct {
		RootPane struct {
			PaneID string `json:"pane_id"`
		} `json:"root_pane"`
	}
	if err := s.call(ctx, "workspace.create", params, &created); err != nil {
		return "", err
	}
	paneID := created.RootPane.PaneID
	if paneID == "" {
		return "", fmt.Errorf("workspace.create returned no pane")
	}

	run := map[string]any{"pane_id": paneID, "text": request.Command, "keys": []string{"enter"}}
	var result json.RawMessage
	if err := s.call(ctx, "pane.send_input", run, &result); err != nil {
		return paneID, fmt.Errorf("run %q: %w", request.Command, err)
	}

	if prompt := strings.TrimSpace(request.Prompt); prompt != "" {
		go s.promptWhenReady(paneID, prompt)
	}
	return paneID, nil
}

func (s *Source) promptWhenReady(paneID, prompt string) {
	ctx, cancel := context.WithTimeout(context.Background(), readyTimeout)
	defer cancel()

	ticker := time.NewTicker(readyInterval)
	defer ticker.Stop()
	for {
		var result struct {
			Agent struct {
				Status string `json:"agent_status"`
			} `json:"agent"`
		}
		err := s.call(ctx, "agent.get", map[string]string{"target": paneID}, &result)
		if err == nil && (result.Agent.Status == meet.StatusIdle || result.Agent.Status == meet.StatusDone) {
			if err := s.SendMessage(ctx, paneID, prompt); err != nil {
				slog.Warn("send first message", "pane", paneID, "err", err)
			}
			return
		}
		select {
		case <-ctx.Done():
			slog.Warn("agent not ready for its first message", "pane", paneID)
			return
		case <-ticker.C:
		}
	}
}
