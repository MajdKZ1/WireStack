package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"wirestack/internal/core"
	"wirestack/internal/utils"
)

const version = "0.1.0"

// main runs the CLI entrypoint.
func main() {
	if err := newRootCommand().Execute(); err != nil {
		log.Fatal(err)
	}
}

// newRootCommand constructs the root Cobra command.
func newRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wirestack",
		Short: "Wirestack controls local WireGuard configurations",
	}

	cmd.AddCommand(
		versionCommand(),
		genKeyCommand(),
		addServerCommand(),
		listServersCommand(),
		deleteServerCommand(),
		addClientCommand(),
		listClientsCommand(),
		exportClientCommand(),
		showServerCommand(),
		showClientCommand(),
		upCommand(),
		downCommand(),
		connectCommand(),
		disconnectCommand(),
	)

	return cmd
}

// versionCommand prints the CLI version.
func versionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show the Wirestack version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(version)
			return nil
		},
	}
}

// genKeyCommand generates a WireGuard private/public key pair using system tools.
func genKeyCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "genkey",
		Short: "Generate a WireGuard key pair",
		RunE: func(cmd *cobra.Command, args []string) error {
			privateKey, publicKey, err := core.GenerateKeyPair()
			if err != nil {
				return err
			}
			fmt.Printf("PrivateKey: %s\nPublicKey: %s\n", privateKey, publicKey)
			return nil
		},
	}
}

// addServerCommand registers a new server profile.
func addServerCommand() *cobra.Command {
	var name string
	var endpoint string

	cmd := &cobra.Command{
		Use:   "add-server",
		Short: "Create a server profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" || endpoint == "" {
				return fmt.Errorf("both --name and --endpoint are required")
			}

			exists, err := core.ProfileExists(name)
			if err != nil {
				return err
			}
			if exists {
				return fmt.Errorf("server %s already exists", name)
			}

			privateKey, publicKey, err := core.GenerateKeyPair()
			if err != nil {
				return err
			}

			profile := core.DefaultServerProfile(name, endpoint, privateKey, publicKey)
			if err := core.SaveServerProfile(profile); err != nil {
				return err
			}

			fmt.Printf("Server %s created at %s\n", name, mustPath(core.ServerProfilePath(name)))
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Server name")
	cmd.Flags().StringVar(&endpoint, "endpoint", "", "Endpoint in the form ip:port")
	return cmd
}

// listServersCommand prints all configured server profiles.
func listServersCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list-servers",
		Short: "List server profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			names, err := core.ListServerProfiles()
			if err != nil {
				return err
			}
			if len(names) == 0 {
				fmt.Println("no servers found")
				return nil
			}
			for _, name := range names {
				fmt.Println(name)
			}
			return nil
		},
	}
}

// deleteServerCommand removes a server profile by name.
func deleteServerCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete-server <name>",
		Short: "Delete a server profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if name == "" {
				return fmt.Errorf("server name is required")
			}
			return core.DeleteServerProfile(name)
		},
	}
}

// addClientCommand appends a new client to an existing server profile.
func addClientCommand() *cobra.Command {
	var serverName string
	var clientName string

	cmd := &cobra.Command{
		Use:   "add-client",
		Short: "Add a client to a server profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			if serverName == "" || clientName == "" {
				return fmt.Errorf("both --server and --client are required")
			}

			profile, err := core.LoadServerProfile(serverName)
			if err != nil {
				return err
			}

			if _, err := core.FindClient(profile, clientName); err == nil {
				return fmt.Errorf("client %s already exists on server %s", clientName, serverName)
			}

			privateKey, publicKey, err := core.GenerateKeyPair()
			if err != nil {
				return err
			}

			address, err := core.NextClientAddress(profile)
			if err != nil {
				return err
			}

			client := core.ClientProfile{
				Name:       clientName,
				PrivateKey: privateKey,
				PublicKey:  publicKey,
				Address:    address,
				AllowedIPs: core.ClientAllowedIPs(),
			}

			profile.Clients = append(profile.Clients, client)

			if err := core.SaveServerProfile(profile); err != nil {
				return err
			}

			fmt.Printf("Client %s added to server %s\n", clientName, serverName)
			return nil
		},
	}

	cmd.Flags().StringVar(&serverName, "server", "", "Server name")
	cmd.Flags().StringVar(&clientName, "client", "", "Client name")
	return cmd
}

// listClientsCommand prints clients for a specific server.
func listClientsCommand() *cobra.Command {
	var serverName string

	cmd := &cobra.Command{
		Use:   "list-clients",
		Short: "List clients for a server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if serverName == "" {
				return fmt.Errorf("--server is required")
			}
			profile, err := core.LoadServerProfile(serverName)
			if err != nil {
				return err
			}
			if len(profile.Clients) == 0 {
				fmt.Println("no clients found")
				return nil
			}
			for _, client := range profile.Clients {
				fmt.Printf("%s\t%s\n", client.Name, client.Address)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&serverName, "server", "", "Server name")
	return cmd
}

// exportClientCommand writes a WireGuard client configuration to a given path.
func exportClientCommand() *cobra.Command {
	var serverName string
	var clientName string
	var outputPath string

	cmd := &cobra.Command{
		Use:   "export-client",
		Short: "Export a WireGuard client configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if serverName == "" || clientName == "" || outputPath == "" {
				return fmt.Errorf("--server, --client, and --output are required")
			}

			profile, err := core.LoadServerProfile(serverName)
			if err != nil {
				return err
			}

			client, err := core.FindClient(profile, clientName)
			if err != nil {
				return err
			}

			config, err := core.BuildClientConfig(profile, *client)
			if err != nil {
				return err
			}

			resolvedPath, err := utils.ExpandPath(outputPath)
			if err != nil {
				return err
			}

			if err := utils.WriteFile(resolvedPath, []byte(config), 0o600); err != nil {
				return err
			}

			fmt.Printf("Client configuration written to %s\n", resolvedPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&serverName, "server", "", "Server name")
	cmd.Flags().StringVar(&clientName, "client", "", "Client name")
	cmd.Flags().StringVar(&outputPath, "output", "", "Path to write the client configuration")
	return cmd
}

// showServerCommand displays the stored server profile.
func showServerCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show server <name>",
		Short: "Show server profile details",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] != "server" {
				return fmt.Errorf("usage: wirestack show server <name>")
			}
			name := args[1]
			profile, err := core.LoadServerProfile(name)
			if err != nil {
				return err
			}
			fmt.Printf("Name: %s\nEndpoint: %s\nAddress: %s\nClients: %d\n", profile.Name, profile.Endpoint, profile.Address, len(profile.Clients))
			for _, client := range profile.Clients {
				fmt.Printf("- %s (%s)\n", client.Name, client.Address)
			}
			return nil
		},
	}
}

// showClientCommand displays client details from a server.
func showClientCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show client <server> <client>",
		Short: "Show client details",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] != "client" {
				return fmt.Errorf("usage: wirestack show client <server> <client>")
			}
			serverName := args[1]
			clientName := args[2]
			profile, err := core.LoadServerProfile(serverName)
			if err != nil {
				return err
			}
			client, err := core.FindClient(profile, clientName)
			if err != nil {
				return err
			}
			fmt.Printf("Server: %s\nClient: %s\nAddress: %s\nPublicKey: %s\nAllowedIPs: %s\n", serverName, client.Name, client.Address, client.PublicKey, strings.Join(client.AllowedIPs, ", "))
			return nil
		},
	}
}

// upCommand generates and brings up a WireGuard interface for a server profile.
func upCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "up <server>",
		Short: "Bring up the WireGuard interface for a server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverName := args[0]
			profile, err := core.LoadServerProfile(serverName)
			if err != nil {
				return err
			}
			configPath, err := core.WriteServerConfig(profile)
			if err != nil {
				return err
			}
			output, err := utils.RunCommand("wg-quick", "up", configPath)
			if err != nil {
				return err
			}
			if output != "" {
				fmt.Println(output)
			}
			return nil
		},
	}
}

// downCommand brings down a WireGuard interface for a server profile.
func downCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "down <server>",
		Short: "Bring down the WireGuard interface for a server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			serverName := args[0]
			configPath, err := core.ServerRuntimeConfigPath(serverName)
			if err != nil {
				return err
			}
			output, err := utils.RunCommand("wg-quick", "down", configPath)
			if err != nil {
				return err
			}
			if output != "" {
				fmt.Println(output)
			}
			_ = os.Remove(configPath)
			return nil
		},
	}
}

// connectCommand brings up a client interface on the local machine.
func connectCommand() *cobra.Command {
	var serverName string
	var clientName string

	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Bring up a WireGuard client interface on this machine",
		RunE: func(cmd *cobra.Command, args []string) error {
			if serverName == "" || clientName == "" {
				return fmt.Errorf("--server and --client are required")
			}
			profile, err := core.LoadServerProfile(serverName)
			if err != nil {
				return err
			}
			client, err := core.FindClient(profile, clientName)
			if err != nil {
				return err
			}

			configPath, err := core.WriteClientConfig(profile, *client)
			if err != nil {
				return err
			}

			output, err := utils.RunCommand("wg-quick", "up", configPath)
			if err != nil {
				return err
			}
			if output != "" {
				fmt.Println(output)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&serverName, "server", "", "Server name")
	cmd.Flags().StringVar(&clientName, "client", "", "Client name to connect with")
	return cmd
}

// disconnectCommand brings down a client interface on the local machine.
func disconnectCommand() *cobra.Command {
	var serverName string
	var clientName string

	cmd := &cobra.Command{
		Use:   "disconnect",
		Short: "Bring down a WireGuard client interface on this machine",
		RunE: func(cmd *cobra.Command, args []string) error {
			if serverName == "" || clientName == "" {
				return fmt.Errorf("--server and --client are required")
			}

			profile, err := core.LoadServerProfile(serverName)
			if err != nil {
				return err
			}
			client, err := core.FindClient(profile, clientName)
			if err != nil {
				return err
			}

			configPath, err := core.WriteClientConfig(profile, *client)
			if err != nil {
				return err
			}

			output, err := utils.RunCommand("wg-quick", "down", configPath)
			if err != nil {
				return err
			}
			if output != "" {
				fmt.Println(output)
			}
			_ = os.Remove(configPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&serverName, "server", "", "Server name")
	cmd.Flags().StringVar(&clientName, "client", "", "Client name to disconnect")
	return cmd
}

// mustPath resolves a path helper while ignoring errors that have already been validated.
func mustPath(path string, err error) string {
	if err != nil {
		fmt.Fprintf(os.Stderr, "unexpected path error: %v\n", err)
		os.Exit(1)
	}
	return path
}
