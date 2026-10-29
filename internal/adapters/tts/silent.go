package tts

import (
	"context"

	"github.com/NathanFirmo/herdr-agent-meet/internal/meet"
)

// Silent keeps the call caption-only.
type Silent struct{}

func (Silent) Speak(context.Context, string, string) (*meet.Audio, error) {
	return nil, nil
}
