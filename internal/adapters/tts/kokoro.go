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
	baseURL  string
	speed    float64
	langCode string
	http     *http.Client
}

// NewKokoro speaks language, a BCP 47 tag. Languages Kokoro does not know are
// left to the voice, whose first letter is its language in Kokoro.
func NewKokoro(baseURL string, speed float64, language string) *Kokoro {
	return &Kokoro{
		baseURL:  strings.TrimRight(baseURL, "/"),
		speed:    speed,
		langCode: kokoroLangCode(language),
		http:     &http.Client{Timeout: 60 * time.Second},
	}
}

var kokoroVoices = map[string][]string{
	"a": {"am_michael", "af_heart", "am_fenrir"},
	"b": {"bm_george", "bf_emma", "bm_lewis"},
	"e": {"em_alex", "ef_dora", "em_santa"},
	"f": {"ff_siwis"},
	"h": {"hm_omega", "hf_alpha", "hm_psi"},
	"i": {"im_nicola", "if_sara"},
	"j": {"jm_kumo", "jf_alpha", "jf_nezumi"},
	"p": {"pm_alex", "pf_dora", "pm_santa"},
	"z": {"zm_yunxi", "zf_xiaobei", "zm_yunjian"},
}

// KokoroVoices are the default voices for language, American English ones
// when Kokoro does not speak it.
func KokoroVoices(language string) []string {
	if voices, ok := kokoroVoices[kokoroLangCode(language)]; ok {
		return voices
	}
	return kokoroVoices["a"]
}

func kokoroLangCode(language string) string {
	switch base := meet.BaseLanguage(language); base {
	case "en":
		if strings.EqualFold(strings.ReplaceAll(language, "_", "-"), "en-GB") {
			return "b"
		}
		return "a"
	case "es", "fr", "hi", "it", "ja", "pt":
		return base[:1]
	case "zh":
		return "z"
	default:
		return ""
	}
}

func (k *Kokoro) Speak(ctx context.Context, text, voice string) (*meet.Audio, error) {
	request := map[string]any{
		"model":           "kokoro",
		"input":           text,
		"voice":           voice,
		"response_format": "mp3",
		"speed":           k.speed,
	}
	if k.langCode != "" {
		request["lang_code"] = k.langCode
	}
	body, err := json.Marshal(request)
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
