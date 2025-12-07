# OpenNET

OpenNET is a self-hosted WireGuard manager. It generates and stores profiles locally (`~/.opennet/`), exports configs for clients, and can bring interfaces up/down using the system WireGuard tools. No UI yet, no Tor, and no evasion features—just standard WireGuard management.

## Build

```bash
go build ./cmd/opennet
```

## Commands

- `opennet version` – show the CLI version.
- `opennet genkey` – generate a WireGuard private/public key pair using `wg`.
- `opennet add-server --name <name> --endpoint <ip:port>` – create a server profile under `~/.opennet/servers/<name>.json`.
- `opennet list-servers` – list stored server profiles.
- `opennet delete-server <name>` – remove a server profile.
- `opennet add-client --server <name> --client <clientName>` – generate client keys and attach the client to the server profile.
- `opennet list-clients --server <name>` – list clients registered to a server.
- `opennet export-client --server <name> --client <clientName> --output <path>` – write a WireGuard client `.conf` without bringing up interfaces.
- `opennet show server <name>` – display server details.
- `opennet show client <server> <client>` – display client details.
- `opennet up <server>` – render server config and call `wg-quick up` on it.
- `opennet down <server>` – call `wg-quick down` on the rendered config.
- `opennet connect --server <name> --client <clientName> [--ninja]` – render the client config and bring it up on the current machine.
- `opennet disconnect --server <name> --client <clientName> [--ninja]` – bring down the locally running client interface.

Server and client data live in `~/.opennet/servers`, runtime configs in `~/.opennet/runtime`, and all keys/configs are produced through the system `wg`/`wg-quick` tooling.

## Terminal UI (Python wrapper)

There is a lightweight menu-driven helper in `scripts/tui.py` that wraps the CLI commands:

```bash
python scripts/tui.py   # uses ./opennet by default or $OPENNET_BIN
```

It offers numbered options for listing/adding servers and clients, exporting configs, and bringing interfaces up/down.

## Ninja (Tor) mode

The client `connect`/`disconnect` commands accept `--ninja`, which wraps `wg-quick` with `torsocks`/`torify` so you can attempt to launch through a Tor-aware wrapper. Make sure Tor is running locally and that one of those wrappers is available in `PATH`. WireGuard is UDP-based, so success depends on how your environment handles UDP over Tor.
