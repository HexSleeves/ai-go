package io

import (
	"encoding/json"
	"fmt"
	"os"
)

func LoadData[T any](path string) (T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return *new(T), fmt.Errorf("failed to read config file: %w", err)
	}

	var config T
	if err := json.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

func SaveData[T any](path string, saveData T) error {
	data, err := json.MarshalIndent(saveData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
