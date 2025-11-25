# OpenNET
By Majd Alzadjali

**OpenNET** is a minimal, self-hosted WireGuard manager designed for users who want full local control with zero external dependencies.
It handles key generation, profile management, config rendering, and interface control using the system’s native `wg` / `wg-quick` tools.

No GUI. No cloud. No analytics.
Just clean, local WireGuard management.

---

## **Features**

* Local profile storage under `~/.opennet/`
* Server & client managers with automatic key generation
* Config rendering for WireGuard servers and clients
* Interface control using `wg-quick up/down`
* Optional “Ninja Mode” wrapper via `torsocks`/`torify`
* Lightweight Python TUI (optional) for easier interaction

---

## **Build**

```bash
go build ./cmd/opennet
```

This produces the `opennet` binary in the current working directory.

---

## **Command Reference**

### **Version**

* `opennet version`
  Displays the current CLI version.

---

### **Key Management**

* `opennet genkey`
  Generates a WireGuard private/public key pair using the system `wg` tool.

---

### **Server Management**

* `opennet add-server --name <name> --endpoint <ip:port>`
  Creates a new server profile under `~/.opennet/servers/<name>.json`.

* `opennet list-servers`
  Lists all stored server profiles.

* `opennet delete-server <name>`
  Deletes the specified server profile.

* `opennet show server <name>`
  Displays full details for a server (keys, peers, endpoint, etc.).

---

### **Client Management**

* `opennet add-client --server <name> --client <clientName>`
  Generates new client keys and attaches the client to the specified server.

* `opennet list-clients --server <name>`
  Lists all clients associated with the server.

* `opennet show client <server> <client>`
  Displays full detail for the specified client.

* `opennet export-client --server <name> --client <clientName> --output <path>`
  Writes a standalone `.conf` file without bringing any interfaces up.

---

### **Interface Control**

* `opennet up <server>`
  Renders the WireGuard server config and calls `wg-quick up`.

* `opennet down <server>`
  Shuts down the server interface via `wg-quick down`.

* `opennet connect --server <name> --client <clientName> [--ninja]`
  Renders the client config and brings the interface up on the local machine.

* `opennet disconnect --server <name> --client <clientName> [--ninja]`
  Brings down the locally-running client interface.

---

## **Directory Structure**

| Path                     | Purpose                                        |
| ------------------------ | ---------------------------------------------- |
| `~/.opennet/servers`     | Stored server and client profiles (JSON).      |
| `~/.opennet/runtime`     | Rendered `.conf` files used by `wg-quick`.     |
| System `wg` / `wg-quick` | Used for key generation and interface control. |

All data stays local — no external network operations unless you explicitly bring interfaces up.

---

## **Python Terminal UI**

A lightweight helper interface is available at:

```
scripts/tui.py
```

Run it with:

```bash
python3 scripts/tui.py
```

It wraps core OpenNET commands with numbered menus for:

* Adding/listing servers
* Adding/listing clients
* Exporting configs
* Bringing interfaces up/down

Set `$OPENNET_BIN` to override the binary path if needed.

---

## **Ninja Mode (Tor Wrapper)**

Some client commands support:

```
--ninja
```

This attempts to route `wg-quick` through `torsocks` or `torify`.

**Notes:**

* Tor must be running on the system.
* One of `torsocks` or `torify` must be available in `$PATH`.
* WireGuard uses UDP — most Tor setups will block or break UDP.
  Ninja mode is experimental and may not work depending on your environment.

---

