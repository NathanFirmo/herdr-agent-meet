package meet

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"slices"
	"sync"
	"time"
)

const (
	historySize  = 100
	blockedLine  = "Chefe, preciso da sua aprovação aqui!"
	speakTimeout = 90 * time.Second
)

type Participant struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Kind   string `json:"kind"`
	Status string `json:"status"`
	Voice  string `json:"voice"`
}

type Speech struct {
	ID            string    `json:"id"`
	ParticipantID string    `json:"participantId"`
	Name          string    `json:"name"`
	Text          string    `json:"text"`
	AudioURL      string    `json:"audioUrl,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
}

type Event struct {
	Type         string        `json:"type"`
	Participants []Participant `json:"participants,omitempty"`
	Speeches     []Speech      `json:"speeches,omitempty"`
	Speech       *Speech       `json:"speech,omitempty"`
}

type Config struct {
	Source     AgentSource
	Parsers    map[string]ScreenParser
	Summarizer Summarizer
	Speaker    Speaker
	Voices     []string
}

type Room struct {
	cfg Config

	mu           sync.Mutex
	participants map[string]*Participant
	order        []string
	seen         map[string]map[string]bool
	voiceCursor  int
	speeches     []Speech
	audio        map[string]*Audio
	audioOrder   []string
	subscribers  map[chan Event]struct{}
}

func NewRoom(cfg Config) *Room {
	return &Room{
		cfg:          cfg,
		participants: map[string]*Participant{},
		seen:         map[string]map[string]bool{},
		audio:        map[string]*Audio{},
		subscribers:  map[chan Event]struct{}{},
	}
}

func (r *Room) Run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		r.poll(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (r *Room) poll(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	agents, err := r.cfg.Source.ListAgents(ctx)
	if err != nil {
		slog.Warn("list agents", "err", err)
		return
	}

	for _, agent := range agents {
		if r.upsertParticipant(agent) {
			go r.speak(agent.ID, nil, blockedLine)
		}

		parser, ok := r.cfg.Parsers[agent.Kind]
		if !ok {
			continue
		}
		screen, err := r.cfg.Source.ReadScreen(ctx, agent.ID)
		if err != nil {
			slog.Debug("read screen", "agent", agent.ID, "err", err)
			continue
		}
		paragraphs := parser.Paragraphs(screen)
		// While the agent works, its last paragraph may still be streaming.
		if agent.Status == StatusWorking && len(paragraphs) > 0 {
			paragraphs = paragraphs[:len(paragraphs)-1]
		}
		if fresh := r.unseen(agent.ID, paragraphs); len(fresh) > 0 {
			go r.speak(agent.ID, fresh, "")
		}
	}

	r.removeGone(agents)
}

func (r *Room) upsertParticipant(agent Agent) (becameBlocked bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, exists := r.participants[agent.ID]
	if !exists {
		p = &Participant{ID: agent.ID, Voice: r.cfg.Voices[r.voiceCursor%len(r.cfg.Voices)]}
		r.voiceCursor++
		r.participants[agent.ID] = p
		r.order = append(r.order, agent.ID)
	}

	changed := !exists || p.Name != agent.Name || p.Status != agent.Status || p.Kind != agent.Kind
	becameBlocked = exists && p.Status != StatusBlocked && agent.Status == StatusBlocked
	p.Name, p.Status, p.Kind = agent.Name, agent.Status, agent.Kind
	if changed {
		r.broadcastLocked(Event{Type: "participants", Participants: r.participantsLocked()})
	}
	return becameBlocked
}

func (r *Room) removeGone(agents []Agent) {
	r.mu.Lock()
	defer r.mu.Unlock()

	alive := map[string]bool{}
	for _, agent := range agents {
		alive[agent.ID] = true
	}
	before := len(r.order)
	r.order = slices.DeleteFunc(r.order, func(id string) bool {
		if alive[id] {
			return false
		}
		delete(r.participants, id)
		delete(r.seen, id)
		return true
	})
	if len(r.order) != before {
		r.broadcastLocked(Event{Type: "participants", Participants: r.participantsLocked()})
	}
}

// unseen returns paragraphs not spoken yet. The first read of an agent only
// records what is already on screen, so joining does not replay its history.
func (r *Room) unseen(agentID string, paragraphs []string) []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	seen, known := r.seen[agentID]
	if !known {
		seen = map[string]bool{}
		r.seen[agentID] = seen
	}
	var fresh []string
	for _, paragraph := range paragraphs {
		if !seen[paragraph] {
			seen[paragraph] = true
			fresh = append(fresh, paragraph)
		}
	}
	if !known {
		return nil
	}
	return fresh
}

func (r *Room) speak(agentID string, paragraphs []string, text string) {
	ctx, cancel := context.WithTimeout(context.Background(), speakTimeout)
	defer cancel()

	if text == "" {
		summary, err := r.cfg.Summarizer.Summarize(ctx, paragraphs)
		if err != nil {
			slog.Warn("summarize", "agent", agentID, "err", err)
			return
		}
		text = summary
	}

	r.mu.Lock()
	p, ok := r.participants[agentID]
	var name, voice string
	if ok {
		name, voice = p.Name, p.Voice
	}
	r.mu.Unlock()
	if !ok {
		return
	}

	speech := Speech{ID: newID(), ParticipantID: agentID, Name: name, Text: text, CreatedAt: time.Now()}
	audio, err := r.cfg.Speaker.Speak(ctx, text, voice)
	if err != nil {
		slog.Warn("speak", "agent", agentID, "err", err)
	}

	r.mu.Lock()
	if audio != nil {
		speech.AudioURL = "/api/audio/" + speech.ID
		r.storeAudioLocked(speech.ID, audio)
	}
	r.speeches = append(r.speeches, speech)
	if len(r.speeches) > historySize {
		r.speeches = r.speeches[len(r.speeches)-historySize:]
	}
	r.broadcastLocked(Event{Type: "speech", Speech: &speech})
	r.mu.Unlock()
	slog.Info("speech", "agent", name, "text", text)
}

func (r *Room) storeAudioLocked(id string, audio *Audio) {
	r.audio[id] = audio
	r.audioOrder = append(r.audioOrder, id)
	if len(r.audioOrder) > historySize {
		delete(r.audio, r.audioOrder[0])
		r.audioOrder = r.audioOrder[1:]
	}
}

func (r *Room) Audio(id string) (*Audio, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	audio, ok := r.audio[id]
	return audio, ok
}

func (r *Room) Subscribe() (chan Event, func()) {
	events := make(chan Event, 32)
	r.mu.Lock()
	r.subscribers[events] = struct{}{}
	events <- Event{Type: "snapshot", Participants: r.participantsLocked(), Speeches: slices.Clone(r.speeches)}
	r.mu.Unlock()

	return events, func() {
		r.mu.Lock()
		delete(r.subscribers, events)
		r.mu.Unlock()
	}
}

func (r *Room) participantsLocked() []Participant {
	participants := make([]Participant, 0, len(r.order))
	for _, id := range r.order {
		participants = append(participants, *r.participants[id])
	}
	return participants
}

func (r *Room) broadcastLocked(event Event) {
	for subscriber := range r.subscribers {
		select {
		case subscriber <- event:
		default:
			slog.Warn("dropping event for slow subscriber", "type", event.Type)
		}
	}
}

func newID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
