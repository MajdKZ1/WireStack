package core

import (
	"fmt"
	"net"
	"path/filepath"
	"strings"

	"wirestack/internal/utils"
)

// GenerateKeyPair uses the system WireGuard tools to produce a key pair.
func GenerateKeyPair() (string, string, error) {
	privateKey, err := utils.RunCommand("wg", "genkey")
	if err != nil {
		return "", "", err
	}
	publicKey, err := utils.RunCommandWithInput(privateKey, "wg", "pubkey")
	if err != nil {
		return "", "", err
	}
	return privateKey, publicKey, nil
}

// BuildClientConfig renders a WireGuard client configuration for the provided client.
func BuildClientConfig(profile *ServerProfile, client ClientProfile) (string, error) {
	if profile == nil {
		return "", fmt.Errorf("server profile is nil")
	}
	if client.Name == "" {
		return "", fmt.Errorf("client name is empty")
	}

	builder := &strings.Builder{}
	fmt.Fprintf(builder, "[Interface]\n")
	fmt.Fprintf(builder, "PrivateKey = %s\n", client.PrivateKey)
	fmt.Fprintf(builder, "Address = %s\n", client.Address)
	if len(profile.DNS) > 0 {
		fmt.Fprintf(builder, "DNS = %s\n", strings.Join(profile.DNS, ", "))
	}
	fmt.Fprintf(builder, "\n")
	fmt.Fprintf(builder, "[Peer]\n")
	fmt.Fprintf(builder, "PublicKey = %s\n", profile.ServerPublicKey)
	fmt.Fprintf(builder, "AllowedIPs = %s\n", strings.Join(client.AllowedIPs, ", "))
	fmt.Fprintf(builder, "Endpoint = %s\n", profile.Endpoint)
	fmt.Fprintf(builder, "PersistentKeepalive = 25\n")
	return builder.String(), nil
}

// BuildServerConfig renders a WireGuard server configuration including peers.
func BuildServerConfig(profile *ServerProfile) (string, error) {
	if profile == nil {
		return "", fmt.Errorf("server profile is nil")
	}
	host, port, err := net.SplitHostPort(profile.Endpoint)
	if err != nil {
		return "", fmt.Errorf("invalid endpoint %s: %w", profile.Endpoint, err)
	}
	if host == "" || port == "" {
		return "", fmt.Errorf("endpoint must include host and port")
	}

	builder := &strings.Builder{}
	fmt.Fprintf(builder, "[Interface]\n")
	fmt.Fprintf(builder, "Address = %s\n", profile.Address)
	fmt.Fprintf(builder, "PrivateKey = %s\n", profile.ServerPrivateKey)
	fmt.Fprintf(builder, "ListenPort = %s\n", port)
	fmt.Fprintf(builder, "SaveConfig = false\n")
	fmt.Fprintf(builder, "\n")
	for _, client := range profile.Clients {
		fmt.Fprintf(builder, "[Peer]\n")
		fmt.Fprintf(builder, "PublicKey = %s\n", client.PublicKey)
		allowed := client.AllowedIPs
		if len(allowed) == 0 {
			allowed = []string{client.Address}
		}
		fmt.Fprintf(builder, "AllowedIPs = %s\n", strings.Join(allowed, ", "))
		fmt.Fprintf(builder, "\n")
	}
	return builder.String(), nil
}

// WriteServerConfig materializes the server config to the runtime directory.
func WriteServerConfig(profile *ServerProfile) (string, error) {
	config, err := BuildServerConfig(profile)
	if err != nil {
		return "", err
	}
	path, err := ServerRuntimeConfigPath(profile.Name)
	if err != nil {
		return "", err
	}
	if err := utils.WriteFile(path, []byte(config), 0o600); err != nil {
		return "", err
	}
	return filepath.Clean(path), nil
}

// WriteClientConfig materializes the client config to the runtime directory.
func WriteClientConfig(profile *ServerProfile, client ClientProfile) (string, error) {
	config, err := BuildClientConfig(profile, client)
	if err != nil {
		return "", err
	}
	path, err := ClientRuntimeConfigPath(profile.Name, client.Name)
	if err != nil {
		return "", err
	}
	if err := utils.WriteFile(path, []byte(config), 0o600); err != nil {
		return "", err
	}
	return filepath.Clean(path), nil
}
