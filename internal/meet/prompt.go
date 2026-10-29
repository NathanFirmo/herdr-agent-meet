package meet

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"slices"
)

var (
	ErrNoPrompt      = errors.New("the agent is not asking anything")
	ErrPromptChanged = errors.New("the question on screen changed, reload and try again")
	ErrUnknownOption = errors.New("unknown option")
)

// updatePrompt shows what a blocked agent is asking, and clears it once the
// agent moves on.
func (r *Room) updatePrompt(agent Agent, prompt *Prompt) {
	if agent.Status != StatusBlocked {
		prompt = nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	p, ok := r.participants[agent.ID]
	if !ok || reflect.DeepEqual(p.Prompt, prompt) {
		return
	}
	p.Prompt = prompt
	r.broadcastLocked(Event{Type: "participants", Participants: r.participantsLocked()})
}

// Answer picks an option of the agent's prompt by moving the highlight with
// the arrow keys and pressing Enter; free-text options get the text typed
// first. Option 0 cancels with Esc. The screen is read again first, so the
// keys match what the agent shows right now.
func (r *Room) Answer(ctx context.Context, agentID string, number int, text string) error {
	r.mu.Lock()
	p, ok := r.participants[agentID]
	var known *Prompt
	var kind string
	if ok {
		known, kind = p.Prompt, p.Kind
	}
	r.mu.Unlock()
	if !ok {
		return ErrUnknownParticipant
	}
	if known == nil {
		return ErrNoPrompt
	}

	parser, ok := r.cfg.Parsers[kind]
	if !ok {
		return ErrNoPrompt
	}
	screen, err := r.cfg.Source.ReadScreen(ctx, agentID)
	if err != nil {
		return fmt.Errorf("read screen: %w", err)
	}
	current := parser.Parse(screen).Prompt
	if current == nil || current.Question != known.Question {
		return ErrPromptChanged
	}

	if number == 0 {
		return r.cfg.Controller.SendKeys(ctx, agentID, "esc")
	}
	index := slices.IndexFunc(current.Options, func(option PromptOption) bool { return option.Number == number })
	if index < 0 {
		return ErrUnknownOption
	}

	var keys []string
	for step := current.Selected; step < index; step++ {
		keys = append(keys, "down")
	}
	for step := current.Selected; step > index; step-- {
		keys = append(keys, "up")
	}
	if current.Options[index].FreeText {
		if len(keys) > 0 {
			if err := r.cfg.Controller.SendKeys(ctx, agentID, keys...); err != nil {
				return err
			}
		}
		if err := r.cfg.Controller.SendText(ctx, agentID, text); err != nil {
			return err
		}
		keys = nil
	}
	return r.cfg.Controller.SendKeys(ctx, agentID, append(keys, "enter")...)
}
