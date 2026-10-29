package meet

import "context"

const (
	StatusIdle    = "idle"
	StatusWorking = "working"
	StatusBlocked = "blocked"
	StatusDone    = "done"
)

type Agent struct {
	ID     string
	Name   string
	Kind   string
	Status string
}

// AgentSource lists the agents taking part in the call and reads what they
// currently show on screen.
type AgentSource interface {
	ListAgents(ctx context.Context) ([]Agent, error)
	ReadScreen(ctx context.Context, agentID string) (string, error)
}

// ScreenParser extracts what an agent wrote from its screen. Each agent kind
// has its own layout, so parsers are registered per kind.
type ScreenParser interface {
	Paragraphs(screen string) []string
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
