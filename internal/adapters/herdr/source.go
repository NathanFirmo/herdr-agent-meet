package herdr

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/NathanFirmo/herdr-agent-meet/internal/meet"
)

// Source reads agents from a Herdr session through its socket API, which
// speaks newline-delimited JSON requests and responses.
type Source struct {
	socketPath string
	nextID     atomic.Uint64
}

func NewSource(socketPath string) *Source {
	return &Source{socketPath: socketPath}
}

type agent struct {
	PaneID        string `json:"pane_id"`
	Kind          string `json:"agent"`
	Status        string `json:"agent_status"`
	Name          string `json:"name"`
	TerminalTitle string `json:"terminal_title_stripped"`
	Session       struct {
		Value string `json:"value"`
	} `json:"agent_session"`
}

func (a agent) displayName() string {
	switch {
	case a.Name != "":
		return a.Name
	case a.TerminalTitle != "":
		return a.TerminalTitle
	default:
		return a.PaneID
	}
}

func (s *Source) ListAgents(ctx context.Context) ([]meet.Agent, error) {
	var result struct {
		Agents []agent `json:"agents"`
	}
	if err := s.call(ctx, "agent.list", struct{}{}, &result); err != nil {
		return nil, err
	}
	agents := make([]meet.Agent, 0, len(result.Agents))
	for _, a := range result.Agents {
		agents = append(agents, meet.Agent{ID: a.PaneID, Name: a.displayName(), Kind: a.Kind, Status: a.Status, SessionID: a.Session.Value})
	}
	return agents, nil
}

func (s *Source) ReadScreen(ctx context.Context, agentID string) (string, error) {
	params := map[string]string{"target": agentID, "source": "visible"}
	var result struct {
		Read struct {
			Text string `json:"text"`
		} `json:"read"`
	}
	if err := s.call(ctx, "agent.read", params, &result); err != nil {
		return "", err
	}
	return result.Read.Text, nil
}

func (s *Source) SendMessage(ctx context.Context, agentID, text string) error {
	params := map[string]string{"target": agentID, "text": text}
	var result json.RawMessage
	return s.call(ctx, "agent.prompt", params, &result)
}

func (s *Source) SendKeys(ctx context.Context, agentID string, keys ...string) error {
	params := map[string]any{"target": agentID, "keys": keys}
	var result json.RawMessage
	return s.call(ctx, "agent.send_keys", params, &result)
}

// SendText types text into the agent's pane; agent IDs are pane IDs.
func (s *Source) SendText(ctx context.Context, agentID, text string) error {
	params := map[string]string{"pane_id": agentID, "text": text}
	var result json.RawMessage
	return s.call(ctx, "pane.send_text", params, &result)
}

// CloseAgent closes the agent's pane, which ends the agent running in it.
func (s *Source) CloseAgent(ctx context.Context, agentID string) error {
	var result json.RawMessage
	return s.call(ctx, "pane.close", map[string]string{"pane_id": agentID}, &result)
}

type request struct {
	ID     string `json:"id"`
	Method string `json:"method"`
	Params any    `json:"params"`
}

type response struct {
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (s *Source) call(ctx context.Context, method string, params, result any) error {
	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "unix", s.socketPath)
	if err != nil {
		return fmt.Errorf("connect to herdr: %w", err)
	}
	defer conn.Close()
	if deadline, ok := ctx.Deadline(); ok {
		conn.SetDeadline(deadline)
	}

	req := request{ID: strconv.FormatUint(s.nextID.Add(1), 10), Method: method, Params: params}
	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return fmt.Errorf("send %s: %w", method, err)
	}

	line, err := bufio.NewReaderSize(conn, 1<<20).ReadBytes('\n')
	if err != nil {
		return fmt.Errorf("read %s: %w", method, err)
	}

	var resp response
	if err := json.Unmarshal(line, &resp); err != nil {
		return fmt.Errorf("decode %s: %w", method, err)
	}
	if resp.Error != nil {
		return fmt.Errorf("%s: %s: %s", method, resp.Error.Code, resp.Error.Message)
	}
	return json.Unmarshal(resp.Result, result)
}
