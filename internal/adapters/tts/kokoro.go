package tts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/NathanFirmo/herdr-agent-meet/internal/meet"
)

// Kokoro uses the OpenAI-compatible speech endpoint of Kokoro-FastAPI.
type Kokoro struct {
	baseURL string
	speed   float64
	http    *http.Client
}

func NewKokoro(baseURL string, speed float64) *Kokoro {
	return &Kokoro{
		baseURL: strings.TrimRight(baseURL, "/"),
		speed:   speed,
		http:    &http.Client{Timeout: 60 * time.Second},
	}
}

func (k *Kokoro) Speak(ctx context.Context, text, voice string) (*meet.Audio, error) {
	body, err := json.Marshal(map[string]any{
		"model":           "kokoro",
		"input":           text,
		"voice":           voice,
		"response_format": "mp3",
		"speed":           k.speed,
		"lang_code":       "p",
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, k.baseURL+"/v1/audio/speech", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := k.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("kokoro: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("kokoro: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kokoro: %s: %s", resp.Status, bytes.TrimSpace(data))
	}
	return &meet.Audio{Data: data, ContentType: "audio/mpeg"}, nil
}
