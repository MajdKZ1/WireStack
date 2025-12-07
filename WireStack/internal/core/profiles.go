package core

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"wirestack/internal/utils"
)

// ClientProfile captures a client and its WireGuard parameters.
type ClientProfile struct {
	Name        string   `json:"name"`
	PrivateKey  string   `json:"private_key"`
	PublicKey   string   `json:"public_key"`
	Address     string   `json:"address"`
	AllowedIPs  []string `json:"allowed_ips"`
	Description string   `json:"description,omitempty"`
}

// ServerProfile describes a WireGuard server and connected clients.
type ServerProfile struct {
	Name             string          `json:"name"`
	Endpoint         string          `json:"endpoint"`
	Address          string          `json:"address"`
	DNS              []string        `json:"dns"`
	ServerPrivateKey string          `json:"server_private_key"`
	ServerPublicKey  string          `json:"server_public_key"`
	Clients          []ClientProfile `json:"clients"`
}

// SaveServerProfile writes the server profile JSON to disk with restrictive permissions.
func SaveServerProfile(profile *ServerProfile) error {
	if profile == nil {
		return fmt.Errorf("profile is nil")
	}
	path, err := ServerProfilePath(profile.Name)
	if err != nil {
		return err
	}
	if err := utils.WriteJSON(path, profile, 0o600); err != nil {
		return err
	}
	return nil
}

// LoadServerProfile reads a server profile from disk.
func LoadServerProfile(name string) (*ServerProfile, error) {
	path, err := ServerProfilePath(name)
	if err != nil {
		return nil, err
	}
	var profile ServerProfile
	if err := utils.ReadJSON(path, &profile); err != nil {
		return nil, err
	}
	return &profile, nil
}

// ListServerProfiles returns the names of all stored server profiles.
func ListServerProfiles() ([]string, error) {
	root, err := ServersRoot()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("failed to read servers directory: %w", err)
	}
	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		names = append(names, entry.Name()[:len(entry.Name())-len(".json")])
	}
	return names, nil
}

// DeleteServerProfile removes the stored server profile JSON.
func DeleteServerProfile(name string) error {
	path, err := ServerProfilePath(name)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete server profile %s: %w", name, err)
	}
	runtimePath, err := ServerRuntimeConfigPath(name)
	if err == nil {
		_ = os.Remove(runtimePath)
	}
	return nil
}

// ProfileExists reports whether a server profile already exists.
func ProfileExists(name string) (bool, error) {
	path, err := ServerProfilePath(name)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// NextClientAddress computes the next available client address in the 10.0.0.0/24 range.
func NextClientAddress(profile *ServerProfile) (string, error) {
	_, network, err := net.ParseCIDR("10.0.0.0/24")
	if err != nil {
		return "", fmt.Errorf("failed to parse default client network: %w", err)
	}
	// Start at .2 to leave .1 for the server address.
	nextHost := 2 + len(profile.Clients)
	if nextHost >= 255 {
		return "", fmt.Errorf("client capacity exceeded for network %s", network.String())
	}
	ip := network.IP.To4()
	if ip == nil {
		return "", fmt.Errorf("default network is not IPv4")
	}
	ip[3] = byte(nextHost)
	return fmt.Sprintf("%s/32", ip.String()), nil
}

// FindClient returns the client from the profile matching the provided name.
func FindClient(profile *ServerProfile, clientName string) (*ClientProfile, error) {
	for idx := range profile.Clients {
		if profile.Clients[idx].Name == clientName {
			return &profile.Clients[idx], nil
		}
	}
	return nil, fmt.Errorf("client %s not found", clientName)
}

// DefaultServerProfile builds a base server profile with generated keys and defaults.
func DefaultServerProfile(name, endpoint, privateKey, publicKey string) *ServerProfile {
	return &ServerProfile{
		Name:             name,
		Endpoint:         endpoint,
		Address:          "10.0.0.1/24",
		DNS:              []string{"1.1.1.1", "9.9.9.9"},
		ServerPrivateKey: privateKey,
		ServerPublicKey:  publicKey,
		Clients:          []ClientProfile{},
	}
}

// ClientAllowedIPs returns default allowed IPs for clients.
func ClientAllowedIPs() []string {
	return []string{"0.0.0.0/0", "::/0"}
}
