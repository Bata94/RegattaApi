package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func GetFilenames(dir string) ([]string, error) {
	var files []string
	path := filepath.Join("./files", dir)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(fmt.Sprint(path), os.ModePerm)

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
	fmt.Printf("Error, Error is not os.ErrNotExist. So maybe we have permission issues or other errors. Error: %v\n", err)
	return false
}
