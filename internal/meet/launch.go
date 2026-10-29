package meet

import (
	"context"
	"errors"
	"log/slog"
	"strings"
)

const defaultCommand = "claude"

var ErrNameRequired = errors.New("the new agent needs a name")

// Launch starts a new agent and names it in the call; it joins as a
// participant on the next poll.
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
	return id, err
}
