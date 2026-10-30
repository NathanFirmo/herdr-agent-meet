package transcript

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/NathanFirmo/herdr-agent-meet/internal/meet"
)

// Claude follows Claude Code session transcripts, the JSONL files kept under
// <config dir>/projects/<project>/<session id>.jsonl, so whole replies are
// read even after they scrolled off the agent's screen.
type Claude struct {
	configDirs []string
	speakTools bool

	mu       sync.Mutex
	sessions map[string]*session
	// waiting holds agents seen before their transcript existed; Claude Code
	// creates it on the first message, so it is read from the start.
	waiting map[string]bool
}

type session struct {
	path   string
	offset int64
	// lastTool skips a tool the agent runs again with the same description.
	lastTool string
	model    string
}

// NewClaude takes config directories as glob patterns, e.g. "~/.claude*".
// With speakTools, the descriptions of the tools agents run are spoken too.
func NewClaude(configDirs []string, speakTools bool) *Claude {
	return &Claude{configDirs: configDirs, speakTools: speakTools, sessions: map[string]*session{}, waiting: map[string]bool{}}
}

// NewMessages starts at the end of a transcript that already existed when the
// agent was first seen, so history is not replayed, and then returns what was
// appended since the previous call.
func (c *Claude) NewMessages(agent meet.Agent) (meet.MessageUpdate, bool, error) {
	if agent.Kind != "claude" {
		return meet.MessageUpdate{}, false, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	s, known := c.sessions[agent.SessionID]
	if !known {
		path := ""
		if agent.SessionID != "" {
			path = c.find(agent.SessionID)
		}
		if path == "" {
			c.waiting[agent.ID] = true
			return meet.MessageUpdate{}, false, nil
		}
		s = &session{path: path}
		c.sessions[agent.SessionID] = s
		if !c.waiting[agent.ID] {
			info, err := os.Stat(path)
			if err != nil {
				return meet.MessageUpdate{}, false, err
			}
			s.offset = info.Size()
			s.model = lastModel(path, info.Size())
			return meet.MessageUpdate{Model: s.model}, true, nil
		}
		delete(c.waiting, agent.ID)
	}

	messages, err := s.readNew(c.speakTools)
	return meet.MessageUpdate{Messages: messages, Model: s.model}, true, err
}

func (c *Claude) find(sessionID string) string {
	for _, pattern := range c.configDirs {
		dirs, _ := filepath.Glob(pattern)
		for _, dir := range dirs {
			matches, _ := filepath.Glob(filepath.Join(dir, "projects", "*", sessionID+".jsonl"))
			if len(matches) > 0 {
				return matches[0]
			}
		}
	}
	return ""
}

func (s *session) readNew(speakTools bool) ([]string, error) {
	file, err := os.Open(s.path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if info.Size() < s.offset {
		s.offset = info.Size()
		return nil, nil
	}
	if _, err := file.Seek(s.offset, io.SeekStart); err != nil {
		return nil, err
	}

	var messages []string
	reader := bufio.NewReaderSize(file, 1<<20)
	for {
		line, err := reader.ReadBytes('\n')
		if errors.Is(err, io.EOF) {
			// A partial line is still being written; read it next time.
			return messages, nil
		}
		if err != nil {
			return messages, err
		}
		s.offset += int64(len(line))
		messages = append(messages, s.entryMessages(line, speakTools)...)
	}
}

type entry struct {
	Type        string `json:"type"`
	IsSidechain bool   `json:"isSidechain"`
	Message     struct {
		Model   string          `json:"model"`
		Content json.RawMessage `json:"content"`
	} `json:"message"`
}

type block struct {
	Type  string `json:"type"`
	Text  string `json:"text"`
	Input struct {
		Description string `json:"description"`
	} `json:"input"`
}

// entryMessages keeps the agent's replies and, with speakTools, the
// descriptions of the tools it runs; thinking, tool output and user turns
// are not spoken.
func (s *session) entryMessages(line []byte, speakTools bool) []string {
	var e entry
	if json.Unmarshal(line, &e) != nil || e.Type != "assistant" || e.IsSidechain {
		return nil
	}
	if realModel(e.Message.Model) {
		s.model = e.Message.Model
	}
	var blocks []block
	if json.Unmarshal(e.Message.Content, &blocks) != nil {
		return nil
	}

	var messages []string
	for _, b := range blocks {
		var text string
		switch b.Type {
		case "text":
			text = Speakable(b.Text)
			s.lastTool = ""
		case "tool_use":
			if !speakTools {
				continue
			}
			text = Speakable(b.Input.Description)
			if text == s.lastTool {
				continue
			}
			s.lastTool = text
		}
		if text != "" {
			messages = append(messages, text)
		}
	}
	return messages
}

// tailSize bounds how much of an existing transcript is scanned for the model.
const tailSize = 512 << 10

// lastModel finds the model of the latest reply in an existing transcript by
// scanning its tail, so the call shows it before the agent answers again.
func lastModel(path string, size int64) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	start := max(size-tailSize, 0)
	if _, err := file.Seek(start, io.SeekStart); err != nil {
		return ""
	}
	reader := bufio.NewReaderSize(file, 1<<20)
	if start > 0 {
		reader.ReadBytes('\n')
	}
	model := ""
	for {
		line, err := reader.ReadBytes('\n')
		var e entry
		if json.Unmarshal(line, &e) == nil && e.Type == "assistant" && !e.IsSidechain && realModel(e.Message.Model) {
			model = e.Message.Model
		}
		if err != nil {
			return model
		}
	}
}

// realModel skips placeholders Claude Code writes for local messages, such as "<synthetic>".
func realModel(model string) bool {
	return model != "" && !strings.HasPrefix(model, "<")
}
