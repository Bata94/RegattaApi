package utils

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/bata94/RegattaApi/internal/config"
)

func GetFilenames(dir string) ([]string, error) {
	var files []string
	path := filepath.Join(config.C.Paths.FilesDir, dir)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(fmt.Sprint(path), os.ModePerm); err != nil {
			return nil, err
		}

	}

	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, info.Name())
		}
		return nil
	})
	if err != nil {
		return files, err
	}
	return files, nil
}

func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true // File exists
	}
	if errors.Is(err, os.ErrNotExist) {
		return false // File does not exist
	}
	// Post-check: file might exist but we have permission issues or other errors
	slog.Warn("stat file failed (not os.ErrNotExist)", "filename", filename, "err", err)
	return false
}
