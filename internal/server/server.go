package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/NathanFirmo/herdr-agent-meet/internal/meet"
)

func New(room *meet.Room, staticDir, uploadDir string) http.Handler {
	files := uploads{dir: uploadDir}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/events", events(room))
	mux.HandleFunc("GET /api/audio/{id}", audio(room))
	mux.HandleFunc("GET /api/audio-test", audioTest(room))
	mux.HandleFunc("POST /api/chat", chat(room, files))
	mux.HandleFunc("GET /api/uploads/{name}", files.serve)
	mux.HandleFunc("PUT /api/participants/{id}/name", rename(room))
	mux.HandleFunc("POST /api/participants/{id}/answer", answer(room))
	mux.HandleFunc("DELETE /api/participants/{id}", remove(room))
	mux.HandleFunc("POST /api/agents", launch(room))
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
		http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(clip.Data))
	}
}

func audioTest(room *meet.Room) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clip, err := room.TestAudio(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		if clip == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", clip.ContentType)
		w.Header().Set("Cache-Control", "no-store")
		w.Write(clip.Data)
	}
}

func chat(room *meet.Room, files uploads) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxAttachments*maxImageSize+1<<20)
		if err := r.ParseMultipartForm(8 << 20); err != nil {
			http.Error(w, "invalid message", http.StatusBadRequest)
			return
		}
		defer r.MultipartForm.RemoveAll()

		images := r.MultipartForm.File["images"]
		if len(images) > maxAttachments {
			http.Error(w, fmt.Sprintf("at most %d images per message", maxAttachments), http.StatusBadRequest)
			return
		}
		var attachments []meet.Attachment
		for _, image := range images {
			attachment, err := files.save(image)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			attachments = append(attachments, attachment)
		}

		message, err := room.Post(r.Context(), r.FormValue("text"), r.MultipartForm.Value["mentions"], attachments)
		if errors.Is(err, meet.ErrEmptyMessage) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(message)
	}
}

func rename(room *meet.Room) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&input); err != nil {
			http.Error(w, "invalid name", http.StatusBadRequest)
			return
		}
		if err := room.Rename(r.PathValue("id"), input.Name); errors.Is(err, meet.ErrUnknownParticipant) {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func answer(room *meet.Room) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Option int    `json:"option"`
			Text   string `json:"text"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&input); err != nil {
			http.Error(w, "invalid answer", http.StatusBadRequest)
			return
		}
		err := room.Answer(r.Context(), r.PathValue("id"), input.Option, input.Text)
		switch {
		case err == nil:
			w.WriteHeader(http.StatusNoContent)
		case errors.Is(err, meet.ErrUnknownParticipant):
			http.NotFound(w, r)
		case errors.Is(err, meet.ErrNoPrompt), errors.Is(err, meet.ErrPromptChanged), errors.Is(err, meet.ErrUnknownOption):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusBadGateway)
		}
	}
}

func remove(room *meet.Room) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := room.Remove(r.Context(), r.PathValue("id"))
		switch {
		case err == nil:
			w.WriteHeader(http.StatusNoContent)
		case errors.Is(err, meet.ErrUnknownParticipant):
			http.NotFound(w, r)
		default:
			http.Error(w, err.Error(), http.StatusBadGateway)
		}
	}
}

func launch(room *meet.Room) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Name    string `json:"name"`
			Command string `json:"command"`
			Dir     string `json:"dir"`
			Prompt  string `json:"prompt"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&input); err != nil {
			http.Error(w, "invalid agent", http.StatusBadRequest)
			return
		}
		id, err := room.Launch(r.Context(), meet.LaunchRequest{Name: input.Name, Command: input.Command, Dir: input.Dir, Prompt: input.Prompt})
		switch {
		case errors.Is(err, meet.ErrNameRequired):
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		case err != nil:
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"id": id})
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
