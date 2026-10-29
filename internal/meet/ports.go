package meet

import (
	"context"
	"time"
)

const (
	StatusIdle    = "idle"
	StatusWorking = "working"
	StatusBlocked = "blocked"
	StatusDone    = "done"
)

type Agent struct {
	ID        string
	Name      string
	Kind      string
	Status    string
	SessionID string
}

// AgentSource lists the agents taking part in the call and reads what they
// currently show on screen.
type AgentSource interface {
	ListAgents(ctx context.Context) ([]Agent, error)
	ReadScreen(ctx context.Context, agentID string) (string, error)
}

const (
	ActivityTool     = "tool"
	ActivityThinking = "thinking"
)

// ScreenActivity is what the agent is busy with right now, as read from its
// screen, and for how long.
type ScreenActivity struct {
	Kind    string
	Label   string
	Elapsed time.Duration
}

type ParsedScreen struct {
	Paragraphs []string
	Activity   *ScreenActivity
	Prompt     *Prompt
}

// Prompt is a question or approval request an agent shows on screen, with
// numbered options navigated by arrow keys.
type Prompt struct {
	Question string         `json:"question"`
	Context  string         `json:"context,omitempty"`
	Options  []PromptOption `json:"options"`
	// Selected is the index of the highlighted option.
	Selected int `json:"-"`
}

type PromptOption struct {
	Number      int    `json:"number"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
	// FreeText options take text typed by the user.
	FreeText bool `json:"freeText,omitempty"`
}

// AgentController answers prompts by driving the agent's terminal.
type AgentController interface {
	SendKeys(ctx context.Context, agentID string, keys ...string) error
	SendText(ctx context.Context, agentID, text string) error
}

// AgentMessenger delivers a chat message from the user to an agent's session.
type AgentMessenger interface {
	SendMessage(ctx context.Context, agentID, text string) error
}

// MessageSource reads what an agent wrote from somewhere better than its
// screen, such as a session transcript. ok is false when it cannot follow the
// agent, and the room falls back to parsing the screen.
type MessageSource interface {
	NewMessages(agent Agent) (messages []string, ok bool, err error)
}

// AgentCloser ends an agent for good, closing the terminal it runs in.
type AgentCloser interface {
	CloseAgent(ctx context.Context, agentID string) error
}

// NameStore persists the names the user gives to participants, by agent ID.
type NameStore interface {
	Load() (map[string]string, error)
	Save(names map[string]string) error
}

// ScreenParser extracts what an agent wrote and what it is doing from its
// screen. Each agent kind has its own layout, so parsers are registered per kind.
type ScreenParser interface {
	Parse(screen string) ParsedScreen
}

// Summarizer turns what an agent wrote into what it says in the call.
type Summarizer interface {
	Summarize(ctx context.Context, paragraphs []string) (string, error)
}

type Audio struct {
	Data        []byte
	ContentType string
}

// Speaker synthesizes speech. A nil Audio without error means the speech is
// shown only as captions.
type Speaker interface {
	Speak(ctx context.Context, text, voice string) (*Audio, error)
}
