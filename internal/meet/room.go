package meet

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"
)

const (
	historySize  = 300
	blockedLine  = "Chefe, preciso da sua aprovação aqui!"
	testLine     = "Oi, chefe! Este é um teste de áudio. Se você está me ouvindo, está tudo certo."
	speakTimeout = 90 * time.Second
)

type Participant struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Kind     string    `json:"kind"`
	Status   string    `json:"status"`
	Voice    string    `json:"voice"`
	Activity *Activity `json:"activity,omitempty"`
	Prompt   *Prompt   `json:"prompt,omitempty"`
	Dir      string    `json:"dir,omitempty"`
	Model    string    `json:"model,omitempty"`
	// OriginalName is the name from the agent source, set when the user renamed the participant.
	OriginalName string `json:"originalName,omitempty"`
}

type Activity struct {
	Kind  string    `json:"kind"`
	Label string    `json:"label,omitempty"`
	Since time.Time `json:"since"`
}

// Speech is one chunk of what an agent says. A turn is split into sentences
// so the first can play while the rest is still being synthesized; chunks of
// a turn share TurnID and the last one is Final.
type Speech struct {
	ID            string    `json:"id"`
	TurnID        string    `json:"turnId"`
	Final         bool      `json:"final"`
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
	Messages     []ChatMessage `json:"messages,omitempty"`
	Message      *ChatMessage  `json:"message,omitempty"`
}

type Config struct {
	Source    AgentSource
	Messenger AgentMessenger
	// Controller answers prompts on the agent's screen.
	Controller AgentController
	Closer     AgentCloser
	Launcher   AgentLauncher
	// Messages is optional; agents it cannot follow are read from their screen.
	Messages   MessageSource
	Parsers    map[string]ScreenParser
	Summarizer Summarizer
	Speaker    Speaker
	Names      NameStore
	Voices     []string
}

type Room struct {
	cfg Config

	mu            sync.Mutex
	participants  map[string]*Participant
	order         []string
	seen          map[string]map[string]bool
	lastParagraph map[string]string
	speakQueues   map[string]chan speakJob
	names         map[string]string
	voiceCursor   int
	testCursor    int
	speeches      []Speech
	messages      []ChatMessage
	audio         map[string]*Audio
	audioOrder    []string
	subscribers   map[chan Event]struct{}
}

func NewRoom(cfg Config) *Room {
	names, err := cfg.Names.Load()
	if err != nil {
		slog.Warn("load participant names", "err", err)
		names = map[string]string{}
	}
	return &Room{
		cfg:           cfg,
		names:         names,
		participants:  map[string]*Participant{},
		seen:          map[string]map[string]bool{},
		lastParagraph: map[string]string{},
		speakQueues:   map[string]chan speakJob{},
		audio:         map[string]*Audio{},
		subscribers:   map[chan Event]struct{}{},
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
			r.enqueueSpeech(agent.ID, speakJob{text: blockedLine})
		}

		followed := r.readMessages(agent)

		parser, ok := r.cfg.Parsers[agent.Kind]
		if !ok {
			continue
		}
		screen, err := r.cfg.Source.ReadScreen(ctx, agent.ID)
		if err != nil {
			slog.Debug("read screen", "agent", agent.ID, "err", err)
			continue
		}
		parsed := parser.Parse(screen)
		r.updateActivity(agent, parsed.Activity)
		r.updatePrompt(agent, parsed.Prompt)
		if followed {
			continue
		}
		paragraphs := r.settled(agent, parsed.Paragraphs)
		if fresh := r.unseen(agent.ID, paragraphs); len(fresh) > 0 {
			r.enqueueSpeech(agent.ID, speakJob{paragraphs: fresh})
		}
	}

	r.removeGone(agents)
}

// readMessages speaks what the message source reports for the agent and says
// whether the source follows it, in which case its screen is not parsed for speech.
func (r *Room) readMessages(agent Agent) (followed bool) {
	if r.cfg.Messages == nil {
		return false
	}
	update, ok, err := r.cfg.Messages.NewMessages(agent)
	if err != nil {
		slog.Warn("read messages", "agent", agent.ID, "err", err)
	}
	if update.Model != "" {
		r.setModel(agent.ID, update.Model)
	}
	if ok && len(update.Messages) > 0 {
		r.enqueueSpeech(agent.ID, speakJob{paragraphs: update.Messages})
	}
	return ok
}

func (r *Room) setModel(id, model string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if p, ok := r.participants[id]; ok && p.Model != model {
		p.Model = model
		r.broadcastLocked(Event{Type: "participants", Participants: r.participantsLocked()})
	}
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

	name, original := agent.Name, ""
	if alias := r.names[agent.ID]; alias != "" {
		name, original = alias, agent.Name
	}
	changed := !exists || p.Name != name || p.OriginalName != original || p.Status != agent.Status || p.Kind != agent.Kind || p.Dir != agent.Dir
	becameBlocked = exists && p.Status != StatusBlocked && agent.Status == StatusBlocked
	p.Name, p.OriginalName, p.Status, p.Kind, p.Dir = name, original, agent.Status, agent.Kind, agent.Dir
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
	var gone []string
	for _, id := range r.order {
		if !alive[id] {
			gone = append(gone, id)
		}
	}
	for _, id := range gone {
		r.dropLocked(id)
	}
	if len(gone) > 0 {
		r.broadcastLocked(Event{Type: "participants", Participants: r.participantsLocked()})
	}
}

func (r *Room) dropLocked(id string) {
	r.order = slices.DeleteFunc(r.order, func(other string) bool { return other == id })
	delete(r.participants, id)
	delete(r.seen, id)
	delete(r.lastParagraph, id)
	if jobs, ok := r.speakQueues[id]; ok {
		close(jobs)
		delete(r.speakQueues, id)
	}
}

// updateActivity broadcasts only real changes: the start time read from a
// screen drifts by a second or two between polls.
func (r *Room) updateActivity(agent Agent, current *ScreenActivity) {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.participants[agent.ID]
	if !ok {
		return
	}
	if agent.Status != StatusWorking {
		current = nil
	}

	var next *Activity
	if current != nil {
		next = &Activity{Kind: current.Kind, Label: current.Label, Since: time.Now().Add(-current.Elapsed).Truncate(time.Second)}
	}
	if sameActivity(p.Activity, next) {
		return
	}
	p.Activity = next
	r.broadcastLocked(Event{Type: "participants", Participants: r.participantsLocked()})
}

func sameActivity(a, b *Activity) bool {
	if a == nil || b == nil {
		return a == b
	}
	drift := a.Since.Sub(b.Since).Abs()
	return a.Kind == b.Kind && a.Label == b.Label && drift < 5*time.Second
}

// settled drops the last paragraph of a working agent until it reads the same
// on two polls in a row, since it may still be streaming. A long-running tool
// call settles this way instead of staying silent until the next message.
func (r *Room) settled(agent Agent, paragraphs []string) []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if agent.Status != StatusWorking || len(paragraphs) == 0 {
		delete(r.lastParagraph, agent.ID)
		return paragraphs
	}
	last := paragraphs[len(paragraphs)-1]
	if r.lastParagraph[agent.ID] == last {
		return paragraphs
	}
	r.lastParagraph[agent.ID] = last
	return paragraphs[:len(paragraphs)-1]
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

	turnID := newID()
	chunks := splitSentences(text)
	for i, chunk := range chunks {
		speech := Speech{
			ID:            newID(),
			TurnID:        turnID,
			Final:         i == len(chunks)-1,
			ParticipantID: agentID,
			Name:          name,
			Text:          chunk,
			CreatedAt:     time.Now(),
		}
		audio, err := r.cfg.Speaker.Speak(ctx, chunk, voice)
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
	}
	slog.Info("speech", "agent", name, "chunks", len(chunks), "text", text)
}

func (r *Room) storeAudioLocked(id string, audio *Audio) {
	r.audio[id] = audio
	r.audioOrder = append(r.audioOrder, id)
	if len(r.audioOrder) > historySize {
		delete(r.audio, r.audioOrder[0])
		r.audioOrder = r.audioOrder[1:]
	}
}

// TestAudio synthesizes a fixed line, rotating through the voices so each
// call lets the listener hear a different one.
func (r *Room) TestAudio(ctx context.Context) (*Audio, error) {
	r.mu.Lock()
	voice := r.cfg.Voices[r.testCursor%len(r.cfg.Voices)]
	r.testCursor++
	r.mu.Unlock()
	return r.cfg.Speaker.Speak(ctx, testLine, voice)
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
	events <- Event{Type: "snapshot", Participants: r.participantsLocked(), Speeches: slices.Clone(r.speeches), Messages: slices.Clone(r.messages)}
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

var ErrUnknownParticipant = errors.New("unknown participant")

const maxNameLength = 40

// Rename gives a participant a name of the user's choosing; an empty name
// restores the one from the agent source.
func (r *Room) Rename(id, name string) error {
	name = strings.Join(strings.Fields(name), " ")
	if runes := []rune(name); len(runes) > maxNameLength {
		name = string(runes[:maxNameLength])
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.participants[id]
	if !ok {
		return ErrUnknownParticipant
	}
	original := p.Name
	if p.OriginalName != "" {
		original = p.OriginalName
	}
	if name == "" || name == original {
		delete(r.names, id)
		p.Name, p.OriginalName = original, ""
	} else {
		r.names[id] = name
		p.Name, p.OriginalName = name, original
	}
	if err := r.cfg.Names.Save(r.names); err != nil {
		slog.Warn("save participant names", "err", err)
	}
	r.broadcastLocked(Event{Type: "participants", Participants: r.participantsLocked()})
	return nil
}

// Remove ends the agent by closing its terminal and drops it from the call
// right away, without waiting for the next poll.
func (r *Room) Remove(ctx context.Context, id string) error {
	r.mu.Lock()
	_, ok := r.participants[id]
	r.mu.Unlock()
	if !ok {
		return ErrUnknownParticipant
	}
	if err := r.cfg.Closer.CloseAgent(ctx, id); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.dropLocked(id)
	if _, named := r.names[id]; named {
		delete(r.names, id)
		if err := r.cfg.Names.Save(r.names); err != nil {
			slog.Warn("save participant names", "err", err)
		}
	}
	r.broadcastLocked(Event{Type: "participants", Participants: r.participantsLocked()})
	return nil
}
