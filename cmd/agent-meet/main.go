package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/NathanFirmo/herdr-agent-meet/internal/adapters/herdr"
	"github.com/NathanFirmo/herdr-agent-meet/internal/adapters/parser"
	"github.com/NathanFirmo/herdr-agent-meet/internal/adapters/store"
	"github.com/NathanFirmo/herdr-agent-meet/internal/adapters/summarizer"
	"github.com/NathanFirmo/herdr-agent-meet/internal/adapters/transcript"
	"github.com/NathanFirmo/herdr-agent-meet/internal/adapters/tts"
	"github.com/NathanFirmo/herdr-agent-meet/internal/meet"
	"github.com/NathanFirmo/herdr-agent-meet/internal/server"
)

func main() {
	if err := run(); err != nil {
		slog.Error("agent meet", "err", err)
		os.Exit(1)
	}
}

func run() error {
	home, _ := os.UserHomeDir()
	socketPath := env("HERDR_SOCKET_PATH", filepath.Join(home, ".config", "herdr", "herdr.sock"))
	listenAddr := env("LISTEN_ADDR", ":8765")
	staticDir := env("STATIC_DIR", "web/dist")
	uploadDir := env("UPLOAD_DIR", filepath.Join(os.TempDir(), "agent-meet"))
	locale := meet.NewLocale(env("MEET_LANGUAGE", meet.DefaultLanguage))
	interval, err := time.ParseDuration(env("POLL_INTERVAL", "3s"))
	if err != nil {
		return fmt.Errorf("POLL_INTERVAL: %w", err)
	}

	summary, err := newSummarizer(locale.Language)
	if err != nil {
		return err
	}
	speaker, err := newSpeaker(locale.Language)
	if err != nil {
		return err
	}

	source := herdr.NewSource(socketPath)
	room := meet.NewRoom(meet.Config{
		Source:     source,
		Messenger:  source,
		Controller: source,
		Closer:     source,
		Launcher:   source,
		Messages:   newMessageSource(home),
		Parsers:    map[string]meet.ScreenParser{"claude": parser.ClaudeCode{}},
		Summarizer: summary,
		Speaker:    speaker,
		Names:      store.NewJSONNames(filepath.Join(uploadDir, "names.json")),
		Voices:     voices(locale.Language),
		Locale:     locale,
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go room.Run(ctx, interval)

	srv := &http.Server{Addr: listenAddr, Handler: server.New(room, staticDir, uploadDir, env("HOST_HOME", home))}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdown)
	}()

	slog.Info("agent meet listening",
		"addr", listenAddr,
		"language", locale.Language,
		"herdr_socket", socketPath,
		"upload_dir", uploadDir,
		"transcripts", env("CLAUDE_TRANSCRIPT_DIRS", filepath.Join(home, ".claude*")),
		"summarizer", fmt.Sprintf("%T", summary),
		"speaker", fmt.Sprintf("%T", speaker),
	)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// newMessageSource follows Claude Code transcripts in the config directories
// listed in CLAUDE_TRANSCRIPT_DIRS (colon-separated globs); "none" turns it off.
func newMessageSource(home string) meet.MessageSource {
	dirs := env("CLAUDE_TRANSCRIPT_DIRS", filepath.Join(home, ".claude*"))
	if dirs == "none" {
		return nil
	}
	return transcript.NewClaude(strings.Split(dirs, ":"))
}

func newSummarizer(language string) (meet.Summarizer, error) {
	command := strings.TrimSpace(os.Getenv("SUMMARY_COMMAND"))
	defaultKind := "none"
	if command != "" {
		defaultKind = "command"
	}

	switch kind := env("SUMMARIZER", defaultKind); kind {
	case "command":
		if command == "" {
			return nil, errors.New("SUMMARIZER=command requires SUMMARY_COMMAND")
		}
		return summarizer.NewCommand(command, language), nil
	case "none":
		return summarizer.Passthrough{}, nil
	default:
		return nil, fmt.Errorf("unknown SUMMARIZER %q (use command or none)", kind)
	}
}

func newSpeaker(language string) (meet.Speaker, error) {
	switch kind := env("TTS", "kokoro"); kind {
	case "kokoro":
		speed, err := strconv.ParseFloat(env("SPEECH_SPEED", "1.1"), 64)
		if err != nil {
			return nil, fmt.Errorf("SPEECH_SPEED: %w", err)
		}
		return tts.NewKokoro(env("KOKORO_URL", "http://localhost:8880"), speed, language), nil
	case "none":
		return tts.Silent{}, nil
	default:
		return nil, fmt.Errorf("unknown TTS %q (use kokoro or none)", kind)
	}
}

// voices are the ones listed in VOICES, or the default Kokoro voices for the language.
func voices(language string) []string {
	var listed []string
	for _, voice := range strings.Split(os.Getenv("VOICES"), ",") {
		if voice = strings.TrimSpace(voice); voice != "" {
			listed = append(listed, voice)
		}
	}
	if len(listed) == 0 {
		return tts.KokoroVoices(language)
	}
	return listed
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
