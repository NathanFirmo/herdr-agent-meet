package parser

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/NathanFirmo/herdr-agent-meet/internal/meet"
)

const (
	messageMarker    = "●"
	toolOutputMarker = "⎿"
	minLength        = 10
)

var (
	// Running tools end their line with a live timer such as "· 4m 55s".
	toolTimer = regexp.MustCompile(`\s+·\s+((?:\d+h\s+)?(?:\d+m\s+)?\d+s)$`)
	// While a turn runs, a spinner line shows its elapsed time, e.g. "✽ Pondering… (11m 30s · ↓ 8.8k tokens)".
	spinner  = regexp.MustCompile(`^\S\s+\S+…\s+\(((?:\d+h\s+)?(?:\d+m\s+)?\d+s)`)
	duration = regexp.MustCompile(`(\d+)([hms])`)
	// UI hints such as "high · /effort" also start with "●" but are not messages.
	slashHint = regexp.MustCompile(`·\s*/[a-z][a-z-]*$`)
	// Bullet markers mean nothing out loud; numbered items keep their numbers.
	listBullet = regexp.MustCompile(`^[-*•]\s+`)
)

// ClaudeCode reads Claude Code screens: each message or tool description
// starts with "●" and continues on indented rows, across blank lines, until
// tool output ("⎿"), the next message or the UI below it. Each block between
// blank lines is its own paragraph, so a long message scrolling off screen is
// not spoken twice. Tool descriptions are only spoken with SpeakTools.
type ClaudeCode struct {
	SpeakTools bool
}

func (c ClaudeCode) Parse(screen string) meet.ParsedScreen {
	var parsed meet.ParsedScreen
	var runningTool *meet.ScreenActivity
	var turn time.Duration
	var thinking bool

	rows := strings.Split(screen, "\n")
	for i := 0; i < len(rows); i++ {
		trimmed := strings.TrimSpace(rows[i])
		if match := spinner.FindStringSubmatch(trimmed); match != nil {
			thinking, turn = true, parseDuration(match[1])
			continue
		}
		if !strings.HasPrefix(trimmed, messageMarker) {
			continue
		}

		blocks := [][]string{{strings.TrimPrefix(trimmed, messageMarker)}}
		next := i + 1
		for ; next < len(rows); next++ {
			row := rows[next]
			rowTrimmed := strings.TrimSpace(row)
			if rowTrimmed == "" {
				if len(blocks[len(blocks)-1]) > 0 {
					blocks = append(blocks, nil)
				}
				continue
			}
			if !strings.HasPrefix(row, "  ") || strings.HasPrefix(rowTrimmed, toolOutputMarker) || strings.HasPrefix(rowTrimmed, messageMarker) {
				break
			}
			blocks[len(blocks)-1] = append(blocks[len(blocks)-1], listBullet.ReplaceAllString(rowTrimmed, ""))
		}
		i = next - 1
		tool := next < len(rows) && strings.HasPrefix(strings.TrimSpace(rows[next]), toolOutputMarker)

		runningTool = nil
		for index, block := range blocks {
			paragraph := clean(strings.Join(block, " "))
			if index == 0 {
				if match := toolTimer.FindStringSubmatch(paragraph); match != nil {
					paragraph = strings.TrimSuffix(paragraph, match[0])
					runningTool = &meet.ScreenActivity{Kind: meet.ActivityTool, Label: paragraph, Elapsed: parseDuration(match[1])}
					tool = true
				}
			}
			if tool && !c.SpeakTools {
				continue
			}
			if len([]rune(paragraph)) > minLength && !slashHint.MatchString(paragraph) {
				parsed.Paragraphs = append(parsed.Paragraphs, paragraph)
			}
		}
	}

	parsed.Prompt = parsePrompt(rows)

	switch {
	case runningTool != nil:
		parsed.Activity = runningTool
	case thinking:
		parsed.Activity = &meet.ScreenActivity{Kind: meet.ActivityThinking, Elapsed: turn}
	}
	return parsed
}

func parseDuration(text string) time.Duration {
	var total time.Duration
	for _, match := range duration.FindAllStringSubmatch(text, -1) {
		value, _ := strconv.Atoi(match[1])
		switch match[2] {
		case "h":
			total += time.Duration(value) * time.Hour
		case "m":
			total += time.Duration(value) * time.Minute
		case "s":
			total += time.Duration(value) * time.Second
		}
	}
	return total
}

// clean drops table borders and collapses whitespace.
func clean(text string) string {
	text = strings.Map(func(r rune) rune {
		if r >= 0x2500 && r <= 0x257F {
			return ' '
		}
		return r
	}, text)
	return strings.Join(strings.Fields(text), " ")
}
