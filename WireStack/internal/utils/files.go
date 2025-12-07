package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ExpandPath replaces a leading ~ with the current user's home directory.
func ExpandPath(path string) (string, error) {
	if len(path) == 0 {
		return "", fmt.Errorf("path is empty")
	}
	if path[0] != '~' {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to resolve home directory: %w", err)
	}
	return filepath.Join(home, path[1:]), nil
}

// EnsureDir creates the directory path if it does not already exist.
func EnsureDir(path string) error {
	if path == "" {
		return fmt.Errorf("directory path is empty")
	}
	if err := os.MkdirAll(path, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}
	return nil
}

// WriteFile writes data to the given path creating parent directories as needed.
func WriteFile(path string, data []byte, perm os.FileMode) error {
	if path == "" {
		return fmt.Errorf("file path is empty")
	}
	if err := EnsureDir(filepath.Dir(path)); err != nil {
		return err
	}
	if err := os.WriteFile(path, data, perm); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}
	return nil
}

// ReadFile reads the contents of a file.
func ReadFile(path string) ([]byte, error) {
	if path == "" {
		return nil, fmt.Errorf("file path is empty")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return data, nil
}

// WriteJSON marshals the value as indented JSON and writes it to the provided path.
func WriteJSON(path string, v any, perm os.FileMode) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	if err := WriteFile(path, data, perm); err != nil {
		return err
	}
	return nil
}

// ReadJSON reads JSON from the provided path into the supplied destination.
func ReadJSON(path string, v any) error {
	data, err := ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to parse JSON in %s: %w", path, err)
	}
	return nil
}
