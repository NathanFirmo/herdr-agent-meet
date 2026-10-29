package server

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const maxDirSuggestions = 20

// dirs completes a directory path the way a shell does: "/a/b" suggests
// entries of /a starting with "b", and "/a/b/" lists /a/b. "~" is the user's
// home; hidden directories show only once their dot is typed.
func dirs(home string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		typed := r.URL.Query().Get("path")
		expanded := typed
		if expanded == "~" || strings.HasPrefix(expanded, "~/") {
			expanded = home + strings.TrimPrefix(expanded, "~")
		}

		parent, prefix := filepath.Dir(expanded), filepath.Base(expanded)
		if expanded == "" || strings.HasSuffix(expanded, "/") {
			parent, prefix = expanded, ""
		}
		if parent == "" {
			parent = home
		}

		entries, _ := os.ReadDir(parent)
		suggestions := []string{}
		for _, entry := range entries {
			name := entry.Name()
			isDir := entry.IsDir()
			if entry.Type()&os.ModeSymlink != 0 {
				info, err := os.Stat(filepath.Join(parent, name))
				isDir = err == nil && info.IsDir()
			}
			if !isDir || (strings.HasPrefix(name, ".") && !strings.HasPrefix(prefix, ".")) {
				continue
			}
			if !strings.HasPrefix(strings.ToLower(name), strings.ToLower(prefix)) {
				continue
			}
			suggestions = append(suggestions, filepath.Join(parent, name)+"/")
		}
		slices.Sort(suggestions)
		if len(suggestions) > maxDirSuggestions {
			suggestions = suggestions[:maxDirSuggestions]
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(suggestions)
	}
}
