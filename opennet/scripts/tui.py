#!/usr/bin/env python3
"""
Simple terminal UI wrapper for the OpenNET CLI.
Provides a numbered menu that calls the compiled Go binary for common tasks.
"""

from __future__ import annotations

import os
import shlex
import subprocess
from dataclasses import dataclass
from pathlib import Path


DEFAULT_BINARY = Path(__file__).resolve().parent.parent / "opennet"


@dataclass
class CommandResult:
    """Captures the output of a command invocation."""

    command: str
    stdout: str
    stderr: str
    returncode: int


def run(cmd: list[str]) -> CommandResult:
    """Run a command and return structured output."""
    proc = subprocess.run(cmd, text=True, capture_output=True)
    return CommandResult(" ".join(shlex.quote(str(x)) for x in cmd), proc.stdout.strip(), proc.stderr.strip(), proc.returncode)


def find_binary() -> Path:
    """Locate the OpenNET binary, defaulting to the repo root build output."""
    env_path = os.environ.get("OPENNET_BIN")
    if env_path:
        return Path(env_path)
    return DEFAULT_BINARY


def ensure_binary(binary: Path) -> Path:
    """Ensure the binary exists; prompt to build if it doesn't."""
    if not binary.exists():
        raise FileNotFoundError(f"OpenNET binary not found at {binary}. Build it with 'go build ./cmd/opennet' or set $OPENNET_BIN.")
    return binary


def prompt(msg: str) -> str:
    """Read input from stdin with a prompt."""
    return input(msg).strip()


def ask_yes_no(msg: str) -> bool:
    """Prompt the user for a yes/no answer."""
    choice = prompt(msg).lower()
    return choice.startswith("y")


def print_result(result: CommandResult) -> None:
    """Pretty-print a command result."""
    print(f"\n$ {result.command}")
    if result.stdout:
        print(result.stdout)
    if result.stderr:
        print(f"[stderr] {result.stderr}")
    print()


def menu(binary: Path) -> None:
    """Interactive loop showing banner, consent, and role selection."""
    try:
        binary = ensure_binary(binary)
    except FileNotFoundError as exc:
        print(exc)
        return

    print_banner()
    if not ask_yes_no("Press Y to continue or N to quit: "):
        print("Declined. Bye.")
        return

    while True:
        print("\nChoose mode:")
        print("1) Host (run/manage server)")
        print("2) Run on current device (client)")
        print("q) Quit")
        choice = prompt("Select option: ").lower()
        if choice == "1":
            host_menu(binary)
        elif choice == "2":
            device_menu(binary)
        elif choice == "q":
            print("Bye.")
            return
        else:
            print("Invalid choice.\n")


def advanced_menu(binary: Path) -> None:
    """Show advanced actions."""
    actions = {
        "1": ("List servers", lambda: run([str(binary), "list-servers"])),
        "2": ("Show server details", lambda: show_server(binary)),
        "3": ("Add server", lambda: add_server(binary)),
        "4": ("Delete server", lambda: delete_server(binary)),
        "5": ("List clients for server", lambda: list_clients(binary)),
        "6": ("Add client", lambda: add_client(binary)),
        "7": ("Generate key pair", lambda: run([str(binary), "genkey"])),
        "b": ("Back", None),
    }
    while True:
        print("\nAdvanced")
        for key, (label, _) in actions.items():
            print(f"{key}) {label}")
        choice = prompt("Select option: ")
        if choice.lower() == "b":
            print()
            return
        action = actions.get(choice)
        if not action:
            print("Invalid choice.\n")
            continue
        label, func = action
        if func is None:
            return
        print(f"== {label} ==")
        try:
            result = func()
            if isinstance(result, CommandResult):
                print_result(result)
        except KeyboardInterrupt:
            print("\nCancelled.\n")
        except Exception as exc:
            print(f"Error: {exc}\n")


def host_menu(binary: Path) -> None:
    """Menu for hosting/managing the server."""
    actions = {
        "1": ("Start server (up)", lambda: start_server(binary)),
        "2": ("Stop server (down)", lambda: stop_server(binary)),
        "3": ("Reload server (down/up)", lambda: reload_server(binary)),
        "4": ("Export client config", lambda: export_client(binary)),
        "5": ("Advanced options", None),
        "b": ("Back", None),
    }

    while True:
        print("\nHost mode")
        for key, (label, _) in actions.items():
            print(f"{key}) {label}")
        choice = prompt("Select option: ")
        if choice.lower() == "b":
            return
        if choice == "5":
            advanced_menu(binary)
            continue
        action = actions.get(choice)
        if not action:
            print("Invalid choice.\n")
            continue
        label, func = action
        if func is None:
            return
        print(f"== {label} ==")
        try:
            result = func()
            if isinstance(result, CommandResult):
                print_result(result)
        except KeyboardInterrupt:
            print("\nCancelled.\n")
        except Exception as exc:
            print(f"Error: {exc}\n")


def device_menu(binary: Path) -> None:
    """Menu for running the VPN on the current device."""
    actions = {
        "1": ("Connect as client", lambda: connect_client(binary)),
        "2": ("Disconnect client", lambda: disconnect_client(binary)),
        "3": ("Start Tor (ninja mode)", lambda: tor_service("start")),
        "4": ("Stop Tor", lambda: tor_service("stop")),
        "5": ("Reload Tor", lambda: tor_service("reload")),
        "6": ("Tor status", lambda: tor_service("status")),
        "b": ("Back", None),
    }

    while True:
        print("\nRun on current device")
        for key, (label, _) in actions.items():
            print(f"{key}) {label}")
        choice = prompt("Select option: ")
        if choice.lower() == "b":
            return
        action = actions.get(choice)
        if not action:
            print("Invalid choice.\n")
            continue
        label, func = action
        if func is None:
            return
        print(f"== {label} ==")
        try:
            result = func()
            if isinstance(result, CommandResult):
                print_result(result)
        except KeyboardInterrupt:
            print("\nCancelled.\n")
        except Exception as exc:
            print(f"Error: {exc}\n")


def show_server(binary: Path) -> CommandResult:
    """Show details for a server."""
    name = prompt("Server name: ")
    return run([str(binary), "show", "server", name])


def add_server(binary: Path) -> CommandResult:
    """Add a new server profile."""
    name = prompt("Server name: ")
    endpoint = prompt("Endpoint (ip:port): ")
    return run([str(binary), "add-server", "--name", name, "--endpoint", endpoint])


def delete_server(binary: Path) -> CommandResult:
    """Delete a server profile."""
    name = prompt("Server name to delete: ")
    return run([str(binary), "delete-server", name])


def list_clients(binary: Path) -> CommandResult:
    """List clients attached to a server."""
    name = prompt("Server name: ")
    return run([str(binary), "list-clients", "--server", name])


def add_client(binary: Path) -> CommandResult:
    """Add a client to a server."""
    server = prompt("Server name: ")
    client = prompt("Client name: ")
    return run([str(binary), "add-client", "--server", server, "--client", client])


def export_client(binary: Path) -> CommandResult:
    """Export a client configuration file."""
    server = prompt("Server name: ")
    client = prompt("Client name: ")
    output = prompt("Output path (e.g., ~/client.conf): ")
    return run([str(binary), "export-client", "--server", server, "--client", client, "--output", output])


def start_server(binary: Path) -> CommandResult:
    """Bring up a server interface using wg-quick."""
    server = prompt("Server name: ")
    return run([str(binary), "up", server])


def stop_server(binary: Path) -> CommandResult:
    """Bring down a server interface using wg-quick."""
    server = prompt("Server name: ")
    return run([str(binary), "down", server])


def reload_server(binary: Path) -> CommandResult:
    """Reload a server interface by calling down then up."""
    server = prompt("Server name: ")
    down_result = run([str(binary), "down", server])
    if down_result.returncode != 0:
        return down_result
    return run([str(binary), "up", server])


def connect_client(binary: Path) -> CommandResult:
    """Bring up a client interface on this machine."""
    server = prompt("Server name: ")
    client = prompt("Client name: ")
    use_ninja = ask_yes_no("Use ninja mode (Tor)? [y/N]: ")
    cmd = [str(binary), "connect", "--server", server, "--client", client]
    if use_ninja:
        cmd.append("--ninja")
    return run(cmd)


def disconnect_client(binary: Path) -> CommandResult:
    """Bring down a client interface on this machine."""
    server = prompt("Server name: ")
    client = prompt("Client name: ")
    use_ninja = ask_yes_no("Use ninja mode (Tor)? [y/N]: ")
    cmd = [str(binary), "disconnect", "--server", server, "--client", client]
    if use_ninja:
        cmd.append("--ninja")
    return run(cmd)


def tor_service(action: str) -> CommandResult:
    """Control the Tor service via systemctl (Linux)."""
    if action not in {"start", "stop", "reload", "status"}:
        return CommandResult("", "", f"Unsupported action {action}", 1)

    service_names = ["tor", "tor.service", "tor@default", "tor@default.service"]
    last_result: CommandResult | None = None
    for name in service_names:
        result = run(["systemctl", action, name])
        last_result = result
        if result.returncode == 0:
            return result
    # If all attempts failed, surface the last error and hint about sudo.
    stderr = last_result.stderr if last_result else ""
    hint = " (Tor may not be installed or systemctl may require sudo)"
    return CommandResult("", last_result.stdout if last_result else "", stderr + hint, last_result.returncode if last_result else 1)


def print_banner() -> None:
    """Display a simple OpenNET banner."""
    art = r"""
  ____                  _   _   _ ______ _______ 
 / ___|  ___ _ ____   _| \ | | | |  _ \ \_   _/ \
 \___ \ / _ \ '__\ \ / /  \| | | | |_) | | | |/  /
  ___) |  __/ |   \ V /| |\  | |_| |  _ <  | | /\ \
 |____/ \___|_|    \_/ |_| \_|\___/|_| \_\ |_| \/ /
    OpenNET by Majd Alzadjali
"""
    print(art)


if __name__ == "__main__":
    menu(find_binary())
