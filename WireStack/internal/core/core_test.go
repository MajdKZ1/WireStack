package core

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func setupTempHome(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	return tmp
}

func TestProfileCRUDAndConfigRendering(t *testing.T) {
	setupTempHome(t)

	profile := DefaultServerProfile("test-srv", "203.0.113.1:51820", "server-priv", "server-pub")
	client := ClientProfile{
		Name:       "alice",
		PrivateKey: "client-priv",
		PublicKey:  "client-pub",
		Address:    "10.0.0.2/32",
		AllowedIPs: []string{"10.0.0.2/32"},
	}
	profile.Clients = append(profile.Clients, client)

	if err := SaveServerProfile(profile); err != nil {
		t.Fatalf("SaveServerProfile: %v", err)
	}

	loaded, err := LoadServerProfile("test-srv")
	if err != nil {
		t.Fatalf("LoadServerProfile: %v", err)
	}
	if len(loaded.Clients) != 1 || loaded.Clients[0].Name != "alice" {
		t.Fatalf("expected one client named alice, got %+v", loaded.Clients)
	}

	names, err := ListServerProfiles()
	if err != nil {
		t.Fatalf("ListServerProfiles: %v", err)
	}
	if len(names) != 1 || names[0] != "test-srv" {
		t.Fatalf("unexpected server names: %v", names)
	}

	clientCfg, err := BuildClientConfig(loaded, client)
	if err != nil {
		t.Fatalf("BuildClientConfig: %v", err)
	}
	if !strings.Contains(clientCfg, "AllowedIPs = 10.0.0.2/32") {
		t.Fatalf("client AllowedIPs missing: %s", clientCfg)
	}

	serverCfg, err := BuildServerConfig(loaded)
	if err != nil {
		t.Fatalf("BuildServerConfig: %v", err)
	}
	if !strings.Contains(serverCfg, "AllowedIPs = 10.0.0.2/32") {
		t.Fatalf("server peer AllowedIPs missing: %s", serverCfg)
	}

	serverPath, err := WriteServerConfig(loaded)
	if err != nil {
		t.Fatalf("WriteServerConfig: %v", err)
	}
	if err := expectFilePerm(serverPath, 0o600); err != nil {
		t.Fatalf("server config perms: %v", err)
	}

	clientPath, err := WriteClientConfig(loaded, client)
	if err != nil {
		t.Fatalf("WriteClientConfig: %v", err)
	}
	if err := expectFilePerm(clientPath, 0o600); err != nil {
		t.Fatalf("client config perms: %v", err)
	}
}

func TestConfigDirectoriesAreLockedDown(t *testing.T) {
	setupTempHome(t)

	root, err := ConfigRoot()
	if err != nil {
		t.Fatalf("ConfigRoot: %v", err)
	}
	if err := expectDirPerm(root, 0o700); err != nil {
		t.Fatalf("ConfigRoot perms: %v", err)
	}

	servers, err := ServersRoot()
	if err != nil {
		t.Fatalf("ServersRoot: %v", err)
	}
	if err := expectDirPerm(servers, 0o700); err != nil {
		t.Fatalf("ServersRoot perms: %v", err)
	}

	runtime, err := RuntimeRoot()
	if err != nil {
		t.Fatalf("RuntimeRoot: %v", err)
	}
	if err := expectDirPerm(runtime, 0o700); err != nil {
		t.Fatalf("RuntimeRoot perms: %v", err)
	}
}

func expectFilePerm(path string, perm os.FileMode) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.Mode().Perm() != perm {
		return fmt.Errorf("got %v, want %v", info.Mode().Perm(), perm)
	}
	return nil
}

func expectDirPerm(path string, perm os.FileMode) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}
	if info.Mode().Perm() != perm {
		return fmt.Errorf("%s perms %v, want %v", path, info.Mode().Perm(), perm)
	}
	return nil
}
