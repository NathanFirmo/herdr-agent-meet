package parser

import "strings"

const (
	messageMarker    = "●"
	toolOutputMarker = "⎿"
	minLength        = 10
)

// ClaudeCode reads Claude Code screens: each message or tool description
// starts with "●" and continues on indented rows.
type ClaudeCode struct{}

func (ClaudeCode) Paragraphs(screen string) []string {
	rows := strings.Split(screen, "\n")
	var paragraphs []string
	for start, row := range rows {
		trimmed := strings.TrimSpace(row)
		if !strings.HasPrefix(trimmed, messageMarker) {
			continue
		}
		parts := []string{strings.TrimPrefix(trimmed, messageMarker)}
		for _, next := range rows[start+1:] {
			nextTrimmed := strings.TrimSpace(next)
			if !strings.HasPrefix(next, "  ") || nextTrimmed == "" || strings.HasPrefix(nextTrimmed, toolOutputMarker) {
				break
			}
			parts = append(parts, nextTrimmed)
		}
		paragraph := strings.Join(strings.Fields(strings.Join(parts, " ")), " ")
		if len([]rune(paragraph)) > minLength {
			paragraphs = append(paragraphs, paragraph)
		}
	}
	return paragraphs
}
