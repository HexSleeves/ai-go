package io

import (
	"log/slog"
	"os"
	"path/filepath"
)

// FindRepoRoot finds the root of the git repository by checking for a .git directory
func FindRepoRoot(startPath string) (string, error) {
	currentPath := startPath

	for {
		gitPath := filepath.Join(currentPath, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return currentPath, nil
		}

		parent := filepath.Dir(currentPath)
		if parent == currentPath {
			return "", os.ErrNotExist
		}
		currentPath = parent
	}
}

// GetConfigDir returns the user's config directory
func GetConfigDir(local bool) (string, error) {
	var configPath string
	if local {
		root, err := FindRepoRoot(".")
		if err != nil {
			return "", err
		}

		configPath = filepath.Join(root, "assets", "config")
	} else {
		root, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}

		configPath = filepath.Join(root, "echos-in-the-dark")
	}

	if err := os.MkdirAll(configPath, 0755); err != nil {
		slog.Error("Failed to create config directory", "error", err)
		os.Exit(1)
	}

	return configPath, nil
}

func GetSaveDir(local bool) (string, error) {
	var savePath string
	if local {
		root, err := FindRepoRoot(".")
		if err != nil {
			return "", err
		}
		savePath = filepath.Join(root, "assets", "saves")
	} else {
		root, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}

		return filepath.Join(root, "echos-in-the-dark", "saves"), nil
	}

	if err := os.MkdirAll(savePath, 0755); err != nil {
		slog.Error("Failed to create save directory", "error", err)
		os.Exit(1)
	}

	return savePath, nil
}

func GetLogDir(local bool) (string, error) {
	var logPath string
	if local {
		root, err := FindRepoRoot(".")
		if err != nil {
			return "", err
		}
		logPath = filepath.Join(root, "assets", "logs")
	} else {
		root, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}

		return filepath.Join(root, "echos-in-the-dark", "logs"), nil
	}

	if err := os.MkdirAll(logPath, 0755); err != nil {
		slog.Error("Failed to create log directory", "error", err)
		os.Exit(1)
	}

	return logPath, nil
}
