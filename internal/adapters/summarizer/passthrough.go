package summarizer

import (
	"context"
	"strings"
)

// Passthrough speaks what the agent wrote, unchanged.
type Passthrough struct{}

func (Passthrough) Summarize(_ context.Context, paragraphs []string) (string, error) {
	return strings.Join(paragraphs, " "), nil
}
