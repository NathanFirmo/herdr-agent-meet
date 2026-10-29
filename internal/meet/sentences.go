package meet

import (
	"strings"
	"unicode"
)

// minChunkRunes merges short sentences into the next one, so speech is not
// cut into choppy pieces.
const minChunkRunes = 25

// splitSentences breaks text after sentence-ending punctuation followed by a
// space, keeping short sentences together.
func splitSentences(text string) []string {
	var sentences []string
	var current strings.Builder
	runes := []rune(strings.TrimSpace(text))
	for i, r := range runes {
		current.WriteRune(r)
		atBoundary := strings.ContainsRune(".!?…", r) && i+1 < len(runes) && unicode.IsSpace(runes[i+1])
		if atBoundary && len([]rune(strings.TrimSpace(current.String()))) >= minChunkRunes {
			sentences = append(sentences, strings.TrimSpace(current.String()))
			current.Reset()
		}
	}
	if rest := strings.TrimSpace(current.String()); rest != "" {
		sentences = append(sentences, rest)
	}
	if len(sentences) == 0 {
		return []string{text}
	}
	return sentences
}
