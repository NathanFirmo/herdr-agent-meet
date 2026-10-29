package store

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// JSONNames keeps participant names in a JSON file.
type JSONNames struct {
	path string
}

func NewJSONNames(path string) *JSONNames {
	return &JSONNames{path: path}
}

func (s *JSONNames) Load() (map[string]string, error) {
	names := map[string]string{}
	data, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return names, nil
	}
	if err != nil {
		return nil, err
	}
	return names, json.Unmarshal(data, &names)
}

func (s *JSONNames) Save(names map[string]string) error {
	data, err := json.MarshalIndent(names, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
