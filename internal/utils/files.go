package utils

import (
	"os"
	"path/filepath"
)

func GetFilenames(dir string) ([]string, error) {
	var files []string
	path := filepath.Join("./files", dir)
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
