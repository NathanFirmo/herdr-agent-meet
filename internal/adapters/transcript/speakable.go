package transcript

import (
	"regexp"
	"strings"
)

var (
	codeFence    = regexp.MustCompile("(?s)```.*?```")
	inlineCode   = regexp.MustCompile("`([^`]*)`")
	link         = regexp.MustCompile(`\[([^\]]+)\]\([^)]*\)`)
	url          = regexp.MustCompile(`https?://\S+`)
	strong       = regexp.MustCompile(`\*\*|__`)
	italicOpen   = regexp.MustCompile(`(^|\s)\*(\S)`)
	italicClose  = regexp.MustCompile(`(\S)\*(\s|$|[.,;:!?])`)
	heading      = regexp.MustCompile(`^#{1,6}\s+`)
	bullet       = regexp.MustCompile(`^(?:[-*•+]|>)\s+`)
	tableDivider = regexp.MustCompile(`^\|?[\s:|-]+\|?$`)
	sentenceEnd  = regexp.MustCompile(`[.!?:;,…]$`)
)

// Speakable turns Markdown into plain text for speech synthesis: code blocks
// and URLs are dropped, formatting marks removed, and each line or list item
// becomes its own sentence.
func Speakable(markdown string) string {
	text := codeFence.ReplaceAllString(markdown, "\n")
	text = inlineCode.ReplaceAllString(text, "$1")
	text = link.ReplaceAllString(text, "$1")
	text = url.ReplaceAllString(text, "")
	text = strong.ReplaceAllString(text, "")
	text = italicOpen.ReplaceAllString(text, "$1$2")
	text = italicClose.ReplaceAllString(text, "$1$2")

	var sentences []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || tableDivider.MatchString(line) {
			continue
		}
		line = heading.ReplaceAllString(line, "")
		line = bullet.ReplaceAllString(line, "")
		if strings.HasPrefix(line, "|") {
			cells := strings.FieldsFunc(line, func(r rune) bool { return r == '|' })
			for i := range cells {
				cells[i] = strings.TrimSpace(cells[i])
			}
			line = strings.Join(cells, ", ")
		}
		line = strings.Join(strings.Fields(line), " ")
		if line == "" {
			continue
		}
		if !sentenceEnd.MatchString(line) {
			line += "."
		}
		sentences = append(sentences, line)
	}
	return strings.Join(sentences, " ")
}
