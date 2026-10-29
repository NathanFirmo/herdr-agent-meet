package server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/NathanFirmo/herdr-agent-meet/internal/meet"
)

const (
	maxImageSize   = 10 << 20
	maxAttachments = 5
)

var (
	imageExtensions = map[string]string{
		"image/png":  ".png",
		"image/jpeg": ".jpg",
		"image/gif":  ".gif",
		"image/webp": ".webp",
	}
	uploadName = regexp.MustCompile(`^[0-9a-f]{16}\.(png|jpg|gif|webp)$`)
)

// uploads keeps chat images in a directory the agents can read, so a message
// can hand them the file path.
type uploads struct {
	dir string
}

func (u uploads) save(header *multipart.FileHeader) (meet.Attachment, error) {
	if header.Size > maxImageSize {
		return meet.Attachment{}, fmt.Errorf("%s is larger than 10 MB", header.Filename)
	}
	file, err := header.Open()
	if err != nil {
		return meet.Attachment{}, err
	}
	defer file.Close()

	sniff := make([]byte, 512)
	n, _ := io.ReadFull(file, sniff)
	ext, ok := imageExtensions[http.DetectContentType(sniff[:n])]
	if !ok {
		return meet.Attachment{}, fmt.Errorf("%s is not a PNG, JPEG, GIF or WebP image", header.Filename)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return meet.Attachment{}, err
	}

	if err := os.MkdirAll(u.dir, 0o755); err != nil {
		return meet.Attachment{}, err
	}
	name := randomHex() + ext
	path := filepath.Join(u.dir, name)
	out, err := os.Create(path)
	if err != nil {
		return meet.Attachment{}, err
	}
	defer out.Close()
	if _, err := io.Copy(out, io.LimitReader(file, maxImageSize)); err != nil {
		return meet.Attachment{}, err
	}
	return meet.Attachment{Path: path, URL: "/api/uploads/" + name}, nil
}

func (u uploads) serve(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if !uploadName.MatchString(name) {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "private, max-age=86400")
	http.ServeFile(w, r, filepath.Join(u.dir, name))
}

func randomHex() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
