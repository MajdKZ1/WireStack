package core

import (
	"fmt"
	"path/filepath"

	"wirestack/internal/utils"
)

const (
	defaultConfigDir = ".wirestack"
	serversDir       = "servers"
	runtimeDir       = "runtime"
)

// ConfigRoot returns the base configuration directory (~/.wirestack) and ensures it exists.
func ConfigRoot() (string, error) {
	homePath, err := utils.ExpandPath("~/" + defaultConfigDir)
	if err != nil {
		return "", err
	}
	if err := utils.EnsureDir(homePath); err != nil {
		return "", err
	}
	return homePath, nil
}

// ServersRoot returns the directory used for storing server profiles.
func ServersRoot() (string, error) {
	root, err := ConfigRoot()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(root, serversDir)
	if err := utils.EnsureDir(dir); err != nil {
		return "", err
	}
	return dir, nil
}

// RuntimeRoot returns the directory used for generated WireGuard config files.
func RuntimeRoot() (string, error) {
	root, err := ConfigRoot()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(root, runtimeDir)
	if err := utils.EnsureDir(dir); err != nil {
		return "", err
	}
	return dir, nil
}

// ServerProfilePath returns the expected JSON path for a server profile.
func ServerProfilePath(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("server name is empty")
	}
	root, err := ServersRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, fmt.Sprintf("%s.json", name)), nil
}

// ServerRuntimeConfigPath returns the path where a server config file is rendered.
func ServerRuntimeConfigPath(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("server name is empty")
	}
	root, err := RuntimeRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, fmt.Sprintf("%s.conf", name)), nil
}

// ClientRuntimeConfigPath returns the path where a client config file is rendered.
func ClientRuntimeConfigPath(serverName, clientName string) (string, error) {
	if serverName == "" {
		return "", fmt.Errorf("server name is empty")
	}
	if clientName == "" {
		return "", fmt.Errorf("client name is empty")
	}
	root, err := RuntimeRoot()
	if err != nil {
		return "", err
	}
	file := fmt.Sprintf("client-%s-%s.conf", serverName, clientName)
	return filepath.Join(root, file), nil
}
