package meet

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"
)

const (
	SelfID          = "self"
	selfName        = "Você"
	deliveryTimeout = 10 * time.Second
	// MentionAll is how the UI addresses everyone; it expands to all agents client-side.
	MentionAll = "@todos"
)

type ChatMessage struct {
	ID          string            `json:"id"`
	AuthorID    string            `json:"authorId"`
	AuthorName  string            `json:"authorName"`
	Text        string            `json:"text"`
	Mentions    []string          `json:"mentions"`
	Attachments []Attachment      `json:"attachments,omitempty"`
	Failed      map[string]string `json:"failed,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
}

// Attachment is a file saved where the agents can read it: Path goes to the
// agents, URL to the browser.
type Attachment struct {
	Path string `json:"-"`
	URL  string `json:"url"`
}

var ErrEmptyMessage = errors.New("message is empty")

// Post records a message from the user and delivers it to every mentioned
// agent, without the mentions themselves and followed by the attachment
// paths. Failed deliveries are reported on the message instead of failing the
// whole post.
func (r *Room) Post(ctx context.Context, text string, mentions []string, attachments []Attachment) (ChatMessage, error) {
	text = strings.TrimSpace(text)
	if text == "" && len(attachments) == 0 {
		return ChatMessage{}, ErrEmptyMessage
	}

	r.mu.Lock()
	names := map[string]string{}
	for _, id := range mentions {
		if p, ok := r.participants[id]; ok {
			names[id] = p.Name
		}
	}
	r.mu.Unlock()

	body := strings.ReplaceAll(text, MentionAll, "")
	for _, name := range names {
		body = strings.ReplaceAll(body, "@"+name, "")
	}
	words := strings.Fields(body)
	for _, attachment := range attachments {
		words = append(words, attachment.Path)
	}
	body = strings.Join(words, " ")

	message := ChatMessage{
		ID:          newID(),
		AuthorID:    SelfID,
		AuthorName:  selfName,
		Text:        text,
		Mentions:    []string{},
		Attachments: attachments,
		CreatedAt:   time.Now(),
	}
	for id := range names {
		message.Mentions = append(message.Mentions, id)
	}
	slices.Sort(message.Mentions)

	if body != "" {
		for _, id := range message.Mentions {
			deliverCtx, cancel := context.WithTimeout(ctx, deliveryTimeout)
			err := r.cfg.Messenger.SendMessage(deliverCtx, id, body)
			cancel()
			if err != nil {
				if message.Failed == nil {
					message.Failed = map[string]string{}
				}
				message.Failed[id] = err.Error()
			}
		}
	}

	r.mu.Lock()
	r.messages = append(r.messages, message)
	if len(r.messages) > historySize {
		r.messages = r.messages[len(r.messages)-historySize:]
	}
	r.broadcastLocked(Event{Type: "message", Message: &message})
	r.mu.Unlock()
	return message, nil
}
