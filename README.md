**Languages:** [English](README.md) · [Polski](README/README.pl.md) · [Deutsch](README/README.de.md) · [Français](README/README.fr.md) · [Español](README/README.es.md) · [Italiano](README/README.it.md) · [Português](README/README.pt.md) · [Nederlands](README/README.nl.md) · [Svenska](README/README.sv.md) · [Norsk](README/README.no.md) · [Dansk](README/README.da.md) · [Suomi](README/README.fi.md) · [Čeština](README/README.cs.md) · [Slovenčina](README/README.sk.md) · [Magyar](README/README.hu.md) · [Română](README/README.ro.md) · [Türkçe](README/README.tr.md) · [Bahasa Indonesia](README/README.id.md) · [Bahasa Melayu](README/README.ms.md) · [Tiếng Việt](README/README.vi.md) · [Русский](README/README.ru.md) · [Українська](README/README.uk.md) · [Български](README/README.bg.md) · [Ελληνικά](README/README.el.md) · [العربية](README/README.ar.md) · [עברית](README/README.he.md) · [हिन्दी](README/README.hi.md) · [中文](README/README.zh.md) · [日本語](README/README.ja.md) · [한국어](README/README.ko.md)

# LAN Transfer

[![CI](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml/badge.svg)](https://github.com/snakex21/LAN-Transfer/actions/workflows/ci.yml) [![Release](https://img.shields.io/github/v/release/snakex21/LAN-Transfer)](https://github.com/snakex21/LAN-Transfer/releases/latest) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE) [![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go)](go.mod) [![Platform](https://img.shields.io/badge/platform-Windows%2010%2F11-0078D6)](#download-and-run)

A small portable tool for sending files and folders directly between two computers on the same local network. No cloud, no account and no relay server: the data goes straight from one PC to the other over an encrypted connection. Go core, a WebView2 window and a command-line version for scripts.

![LAN Transfer main window: a paired receiver and three items ready to send](docs/screenshots/main.png)

## Contents

- [Features](#features)
- [Interface language](#interface-language)
- [Screenshots](#screenshots)
- [Download and run](#download-and-run)
- [How to pair](#how-to-pair)
- [Sending files](#sending-files)
- [Command line](#command-line)
- [Ports and firewall](#ports-and-firewall)
- [Security notes](#security-notes)
- [Environment variables](#environment-variables)
- [Building and tests](#building-and-tests)
- [License](#license)

## Features

- Automatic discovery of other computers running LAN Transfer (UDP broadcast, with a throttled /24 scan as a fallback).
- Pairing with a 6-digit code shown on the receiver. Without pairing nothing can be sent and no folders can be browsed.
- TLS 1.3 encryption. Each run creates a new self-signed certificate; its fingerprint is pinned during pairing and checked on every connection.
- Files and whole folders, with the folder structure preserved and up to 4 files sent in parallel.
- Resume of interrupted transfers and a SHA-256 check of every file.
- Unfinished data is kept in a hidden `.lantransfer-partial` folder, so a truncated file never appears under its final name.
- Name collisions become `file (1).txt` or `Folder (1)`; existing files are never overwritten.
- Dangerous names are rejected (`..`, absolute paths, `file:stream` alternate data streams, reserved names such as `CON`, `NUL`, `COM1`). Free disk space is checked before receiving.
- Drag and drop from Explorer, a queue with per-file progress, speed, time remaining, cancel, resume and a history of sent and received items.
- Portable: a single EXE, settings and history in a `data` folder next to it.
- The window, its title and the command-line output in 30 languages (the same as this README), chosen automatically from the Windows display language; right-to-left layout for Arabic and Hebrew.

## Interface language

LAN Transfer speaks the same 30 languages as this README. On the first start it uses the Windows display language (the first supported language in the Windows language list); if Windows uses a language that is not on the list, the interface is in English.

To change it, click the gear button (**Settings**) in the top-right corner and pick a language. The window, its title and all messages switch immediately, and the choice is saved in `data\settings.json`. **Automatic (Windows)** follows the Windows display language again. Numbers, sizes, speeds, dates and plural forms follow the rules of the chosen language; Arabic and Hebrew use a mirrored right-to-left layout, while paths, IP addresses, codes and fingerprints stay left-to-right. Messages from the other computer (for example "not enough space on the receiver's disk") are shown in your language, not in the receiver's.

`LANTransferCLI.exe` uses the same language. `LANTRANSFER_LANG` (for example `LANTRANSFER_LANG=de`) overrides the saved choice for both the window and the command line.

| English | Deutsch | العربية (RTL) |
|---|---|---|
| ![LAN Transfer in English](docs/screenshots/lang-en.png) | ![LAN Transfer auf Deutsch](docs/screenshots/lang-de.png) | ![LAN Transfer in Arabic, right-to-left layout](docs/screenshots/lang-ar.png) |

## Screenshots

| Pairing | Sending with progress |
|---|---|
| ![Pairing dialog with the 6-digit code and the certificate fingerprint](docs/screenshots/pairing.png) | ![Transfer queue with progress, speed and time remaining](docs/screenshots/sending.png) |
| **Resume after cancel** | **Receiver: history** |
| ![A cancelled transfer resumed from the received offset, SHA-256 verified](docs/screenshots/resume.png) | ![Receiver window with received items in the history](docs/screenshots/history.png) |

## Download and run

1. Download `LANTransfer-2.1.0-windows-amd64.zip` from [Releases](https://github.com/snakex21/LAN-Transfer/releases/latest) and check it against the published `.sha256` file.
2. Extract it to any folder on both computers.
3. Start `LANTransfer.exe`. On the first start Windows asks (UAC) for permission to add two firewall rules; see [Ports and firewall](#ports-and-firewall).

Requirements: Windows 10 or 11 (64-bit) with the WebView2 runtime, which is part of current Windows. Both computers must run version 2. Version 2 cannot talk to version 1.

## How to pair

1. Start LAN Transfer on both computers. The receiver shows its pairing code and certificate fingerprint in the top-right corner.
2. On the sending computer, click the receiver in the **Receiver** list. If it does not appear, use **Connect manually (IP address)** and enter its IP address.
3. Type the 6-digit code from the receiver's screen and compare the certificate fingerprint shown in the dialog with the one on the receiver. If they differ, do not continue.
4. Click **Pair**. The device now shows **✓ paired**.

After 5 wrong codes from one computer the receiver blocks further attempts from it. **New code** creates a new code and cancels all existing pairings. The receiver creates a new code and certificate every time it starts, so after the receiver restarts, pair again.

## Sending files

1. Pair with the receiver.
2. Drag files or folders onto the window, use **Choose files** / **Choose folder**, or paste a path.
3. Choose the destination: the receiver's default folder (`Downloads\LAN Transfer`) or another existing folder on the receiver.
4. Click **Send**. Each item shows progress, speed and time remaining and can be cancelled. A cancelled or interrupted transfer can be resumed; only the missing part is sent.

## Command line

`LANTransferCLI.exe` is the same program without a window, for scripts and automation.

```powershell
LANTransferCLI.exe serve                          # receive; prints the pairing code
LANTransferCLI.exe devices                        # list devices (JSON)
LANTransferCLI.exe send file.zip folder --to 192.168.0.15 --code 123456 [--dest D:\Inbox] [--parallel 4] [--json]
LANTransferCLI.exe version
```

The code can also be passed in `LANTRANSFER_CODE`. A receiver without a window can use a fixed code from `LANTRANSFER_PIN`. Running the same `send` command again resumes an interrupted transfer.

## Ports and firewall

- UDP 45454: discovery. TCP 45455: transfer (HTTPS).
- On the first start the app asks once (UAC) to add two inbound rules: `LAN Transfer 2 Discovery (UDP)` and `LAN Transfer 2 Transfer (TCP)`.
- The rules accept connections only from the local subnet (`remoteip=LocalSubnet`) and only on the Private and Domain network profiles (`profile=private,domain`). On a network that Windows marks as Public (café, hotel, airport) other computers cannot connect to LAN Transfer.
- If the receiver is not visible on your home network, check that the network is set to Private in the Windows network settings.
- Version 1 created rules named `LAN Transfer Discovery` and `LAN Transfer Transfer` for all profiles, Public included. If version 2 finds them, it shows a banner offering to replace them. Nothing is changed until you click **Update rules** and confirm the Windows prompt.

## Security notes

- Discovery over UDP is not authenticated; it is only a hint. Trust comes from pairing: the code plus the pinned certificate fingerprint.
- During the first pairing an active attacker on the same network (man in the middle) could pretend to be the receiver. Compare the certificate fingerprint on both screens before pairing.
- A paired sender can browse the receiver's folders and save files into any existing folder there (the custom destination option). Pair only computers you trust, and use **New code** to revoke all pairings.
- The firewall rules do not open the ports on networks marked Public.

## Environment variables

These are meant for tests and automation; normal use needs none of them.

| Variable | Meaning |
|---|---|
| `LANTRANSFER_PORT`, `LANTRANSFER_UDP_PORT` | other ports |
| `LANTRANSFER_RECEIVE_DIR`, `LANTRANSFER_DATA_DIR` | receive folder, program data folder |
| `LANTRANSFER_NAME` | computer name shown to others |
| `LANTRANSFER_LANG` | interface and CLI language (`en`, `de`, `pl`, …); overrides the saved choice |
| `LANTRANSFER_LOOPBACK=1` | 127.0.0.1 only: no broadcasts, no LAN scan, no firewall rules |
| `LANTRANSFER_PEER_UDP_PORTS` | UDP ports of other instances in loopback mode |
| `LANTRANSFER_NO_FIREWALL=1` | do not create or check firewall rules |
| `LANTRANSFER_PIN`, `LANTRANSFER_CODE` | fixed receiver code, code used by the sender |

## Building and tests

Go 1.23 or newer on Windows 10/11.

```powershell
build.bat
go test ./...
python tools/check_i18n.py
```

`build.bat` runs the tests and builds `LANTransfer.exe` (window) and `LANTransferCLI.exe` (console) with the version from `VERSION` (default `2.1.0`). The tests include end-to-end transfers over 127.0.0.1: pairing and lockout, rejected unauthenticated requests, files from 0 B to 96 MB, Unicode folders and paths over 260 characters, name collisions, hostile names, a wrong checksum, cancel and resume, a full disk and certificate pinning.

The translations live in `locales/ui.<code>.json` (Polish is the source, English the fallback). `tools/check_i18n.py` checks that all 30 files have the same keys, placeholders and plural forms and contain no leftover Polish; a Go test checks that every key used by the program exists in every language.

Release notes: [docs/releases/v2.1.0.md](docs/releases/v2.1.0.md) · [v2.0.0](docs/releases/v2.0.0.md).

## License

MIT, see [LICENSE](LICENSE).
