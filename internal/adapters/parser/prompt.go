package parser

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/NathanFirmo/herdr-agent-meet/internal/meet"
)

const (
	selectedMarker = "❯"
	separatorRune  = '─'
	maxContext     = 400
)

var (
	optionRow  = regexp.MustCompile(`^(❯\s*)?(\d+)\.\s+(.+)$`)
	promptHint = regexp.MustCompile(`Enter to (select|confirm)|Esc to cancel|Do you want to`)
	freeText   = regexp.MustCompile(`(?i)^type something`)
)

// parsePrompt reads the selection dialog Claude Code shows when it asks a
// question or needs approval: a question, then numbered options, one of them
// highlighted with "❯", optionally with indented descriptions.
func parsePrompt(rows []string) *meet.Prompt {
	if !promptHint.MatchString(strings.Join(rows, "\n")) {
		return nil
	}

	highlighted := -1
	for i := len(rows) - 1; i >= 0; i-- {
		if match := optionRow.FindStringSubmatch(promptText(rows[i])); match != nil && match[1] != "" {
			highlighted = i
			break
		}
	}
	if highlighted < 0 {
		return parseUnnumberedPrompt(rows)
	}

	first := highlighted
	for i := highlighted - 1; i >= 0; i-- {
		text := promptText(rows[i])
		if optionRow.MatchString(text) {
			first = i
			continue
		}
		if text == "" || !isIndented(rows[i]) {
			break
		}
	}

	prompt := &meet.Prompt{}
	questionRow := first - 1
	for questionRow >= 0 && promptText(rows[questionRow]) == "" {
		questionRow--
	}
	if questionRow >= 0 {
		prompt.Question = promptText(rows[questionRow])
		var context []string
		for i := questionRow - 1; i >= 0 && !isSeparator(rows[i]); i-- {
			if text := promptText(rows[i]); text != "" {
				context = append([]string{text}, context...)
			}
		}
		prompt.Context = truncate(strings.Join(context, " "), maxContext)
	}

	for i := first; i < len(rows); i++ {
		text := promptText(rows[i])
		if text == "" || isSeparator(rows[i]) {
			continue
		}
		if promptHint.MatchString(text) && !optionRow.MatchString(text) {
			break
		}
		if match := optionRow.FindStringSubmatch(text); match != nil {
			number, _ := strconv.Atoi(match[2])
			if match[1] != "" {
				prompt.Selected = len(prompt.Options)
			}
			prompt.Options = append(prompt.Options, meet.PromptOption{
				Number:   number,
				Label:    match[3],
				FreeText: freeText.MatchString(match[3]),
			})
			continue
		}
		if len(prompt.Options) == 0 || !isIndented(rows[i]) {
			break
		}
		last := &prompt.Options[len(prompt.Options)-1]
		last.Description = strings.TrimSpace(last.Description + " " + text)
	}

	if len(prompt.Options) == 0 {
		return nil
	}
	return prompt
}

// promptText strips the box borders dialogs are drawn with.
func promptText(row string) string {
	return strings.TrimSpace(strings.Trim(strings.TrimSpace(row), "│"))
}

func isIndented(row string) bool {
	return strings.HasPrefix(row, "  ")
}

func isSeparator(row string) bool {
	trimmed := strings.TrimSpace(row)
	return trimmed != "" && strings.Trim(trimmed, string(separatorRune)) == ""
}

func truncate(text string, limit int) string {
	if runes := []rune(text); len(runes) > limit {
		return string(runes[:limit]) + "…"
	}
	return text
}

// parseUnnumberedPrompt reads dialogs whose options carry no numbers, such as
// the folder trust check: the highlighted option starts with "❯" and its
// siblings sit indented on the rows around it. Options are numbered in order.
func parseUnnumberedPrompt(rows []string) *meet.Prompt {
	highlighted := -1
	for i := len(rows) - 1; i >= 0; i-- {
		if strings.HasPrefix(promptText(rows[i]), selectedMarker+" ") {
			highlighted = i
			break
		}
	}
	if highlighted < 0 {
		return nil
	}

	isOption := func(i int) bool {
		text := promptText(rows[i])
		return text != "" && !promptHint.MatchString(text) && (i == highlighted || strings.HasPrefix(rows[i], "   "))
	}
	first, last := highlighted, highlighted
	for first > 0 && isOption(first-1) {
		first--
	}
	for last < len(rows)-1 && isOption(last+1) {
		last++
	}

	prompt := &meet.Prompt{}
	for i := first; i <= last; i++ {
		label := strings.TrimSpace(strings.TrimPrefix(promptText(rows[i]), selectedMarker))
		if i == highlighted {
			prompt.Selected = len(prompt.Options)
		}
		prompt.Options = append(prompt.Options, meet.PromptOption{Number: len(prompt.Options) + 1, Label: label})
	}

	var question []string
	for i := first - 1; i >= 0 && !isSeparator(rows[i]); i-- {
		if text := promptText(rows[i]); text != "" {
			question = append([]string{text}, question...)
		}
	}
	prompt.Question = truncate(strings.Join(question, " "), maxContext)
	return prompt
}
