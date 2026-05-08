package upload

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type StorageBackend interface {
	Save(filename string, data []byte) (string, error)
	Delete(path string) error
}

type LocalStorage struct {
	BasePath string
	BaseURL  string
}

func NewLocalStorage(basePath, baseURL string) *LocalStorage {
	os.MkdirAll(basePath, 0755)
	return &LocalStorage{BasePath: basePath, BaseURL: baseURL}
}

func (s *LocalStorage) Save(filename string, data []byte) (string, error) {
	// filename can include subdirectories: comics/4/0001.webp
	// Sanitize only the final filename component
	dir := filepath.Dir(filename)
	base := sanitizeFilename(filepath.Base(filename))
	safePath := filepath.Join(dir, base)
	dst := filepath.Join(s.BasePath, safePath)

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return "", fmt.Errorf("create dir: %w", err)
	}

	if err := os.WriteFile(dst, data, 0644); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	// Return URL path relative to base
	rel := strings.TrimPrefix(safePath, "/")
	return s.BaseURL + "/" + rel, nil
}

func (s *LocalStorage) Delete(path string) error {
	return os.Remove(path)
}

func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "..", "")
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	return name
}
