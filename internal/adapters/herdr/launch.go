package herdr

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/NathanFirmo/herdr-agent-meet/internal/meet"
)

// Launch opens a workspace without taking the user's focus and runs the
// command in its shell; Herdr recognizes the agent that starts there.
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

	return paneID, nil
}
