package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const configFileName = ".humble-cli-key"

// GetConfigFileName returns the full path to the config file
func GetConfigFileName() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return filepath.Join(homeDir, configFileName), nil
}

// GetConfig reads the session key from the config file
func GetConfig() (string, error) {
	configPath, err := GetConfigFileName()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("config file not found. Use 'humble-cli auth <SESSION-KEY>' to set it")
		}
		return "", fmt.Errorf("failed to read config file: %w", err)
	}

	sessionKey := strings.TrimSpace(string(data))
	if sessionKey == "" {
		return "", fmt.Errorf("config file is empty")
	}

	return sessionKey, nil
}

// SetConfig writes the session key to the config file
func SetConfig(sessionKey string) error {
	configPath, err := GetConfigFileName()
	if err != nil {
		return err
	}

	sessionKey = strings.TrimSpace(sessionKey)
	if sessionKey == "" {
		return fmt.Errorf("session key cannot be empty")
	}

	err = os.WriteFile(configPath, []byte(sessionKey), 0600)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
