# WireStack

WireStack is a minimal, self-hosted **WireGuard networking framework** designed for users who want full local control with zero external dependencies.  
It provides structured management for keys, peer profiles, interface orchestration, and automated configuration generation using the system’s native `wg` and `wg-quick` tools.

No cloud. No telemetry. No vendor lock-in.  
Just clean, extensible WireGuard infrastructure.

---

## Features

• Local profile storage under `~/.wirestack/`  
• Automatic key generation for servers and clients  
• Structured JSON profiles for reproducible configuration  
• Config rendering for server and client WireGuard interfaces  
• Interface orchestration via `wg-quick up` / `wg-quick down`  
• Full CLI workflow for adding servers, adding clients, exporting configs and controlling interfaces  
• Lightweight, portable, and fully open source  

---

## Build

```bash
go build ./cmd/wirestack
```

This produces the `wirestack` binary in the current directory.

To make it system-wide:

```bash
sudo mv wirestack /usr/local/bin/
```

---

## Directory Structure

| Path                        | Purpose                                        |
|----------------------------|------------------------------------------------|
| `~/.wirestack/servers`     | Stored server and client profiles (JSON)       |
| `~/.wirestack/runtime`     | Rendered `.conf` files used by `wg-quick`      |
| System `wg` / `wg-quick`   | Used for key generation and interface control  |

All operations remain fully local unless an interface is explicitly activated.

---

# Command Reference

## Version

`wirestack version`  
Displays the current CLI version.

---

## Key Management

`wirestack genkey`  
Generates a WireGuard private/public key pair using the system `wg` tool.

---

## Server Management

`wirestack add-server --name <name> --endpoint <ip:port>`  
Creates a new server profile under `~/.wirestack/servers/<name>.json`.

`wirestack list-servers`  
Lists all stored server profiles.

`wirestack delete-server <name>`  
Removes a server profile completely.

`wirestack show server <name>`  
Displays full server details including keys, peers, and metadata.

---

## Client Management

`wirestack add-client --server <name> --client <clientName>`  
Creates a new client profile and attaches it to a server.

`wirestack list-clients --server <name>`  
Lists all clients registered under a server.

`wirestack show client <server> <client>`  
Shows a client’s details.

`wirestack export-client --server <name> --client <clientName> --output <path>`  
Exports a standalone WireGuard `.conf` file without activating an interface.

---

## Interface Control

`wirestack up <server>`  
Renders and activates the server interface using `wg-quick up`.

`wirestack down <server>`  
Shuts down a running server interface.

`wirestack connect --server <name> --client <clientName>`  
Renders and activates a local client interface.

`wirestack disconnect --server <name> --client <clientName>`  
Brings down the active local client interface.

---

## Notes

• WireStack relies entirely on system `wg` and `wg-quick`.  
• No background services or daemons are used.  
• All data remains on the local machine unless explicitly exported.  

---

## License

Licensed under **OpenNET LLC**.

