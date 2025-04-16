package config

import (
	"log"
	"os"
	"path/filepath"
)

var (
	tempDir string
	appName = "fastbin"
)

func init() {
	dir := os.TempDir()

	err := os.MkdirAll(filepath.Join(dir, appName), os.ModePerm)
	if err != nil {
		log.Fatal(err) // TODO: report proper error
	}

	tempDir = filepath.Join(dir, appName)
}

func TempDir() string {
	return tempDir
}
