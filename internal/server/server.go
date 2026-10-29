package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/NathanFirmo/herdr-agent-meet/internal/meet"
)

func New(room *meet.Room, staticDir string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/events", events(room))
	mux.HandleFunc("GET /api/audio/{id}", audio(room))
	mux.Handle("GET /", spa(staticDir))
	return mux
}

func events(room *meet.Room) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming unsupported", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		subscription, unsubscribe := room.Subscribe()
		defer unsubscribe()
		heartbeat := time.NewTicker(15 * time.Second)
		defer heartbeat.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case <-heartbeat.C:
				fmt.Fprint(w, ": heartbeat\n\n")
			case event := <-subscription:
				data, err := json.Marshal(event)
				if err != nil {
					continue
				}
				fmt.Fprintf(w, "data: %s\n\n", data)
			}
			flusher.Flush()
		}
	}
}

func audio(room *meet.Room) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clip, ok := room.Audio(r.PathValue("id"))
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", clip.ContentType)
		w.Header().Set("Cache-Control", "private, max-age=3600")
		w.Write(clip.Data)
	}
}

func spa(staticDir string) http.Handler {
	files := http.FileServer(http.Dir(staticDir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat(filepath.Join(staticDir, filepath.Clean(r.URL.Path))); err != nil {
			http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
			return
		}
		files.ServeHTTP(w, r)
	})
}
