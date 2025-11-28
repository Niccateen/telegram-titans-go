# TDL GUI

A graphical user interface wrapper for the `tdl` command-line tool.

## Overview

This GUI is a simple wrapper that executes `tdl` commands and displays output. It does **NOT** implement any Telegram logic itself - all functionality is provided by the `tdl` binary.

## Features

- **Login**: Multiple authentication methods (Desktop auto-detect, passcode, QR code, phone+code)
- **Download**: Download files from Telegram URLs with custom save directory
- **Upload**: Upload files to Saved Messages or specific chats
- **Forward**: Forward messages between chats
- **Backup/Recover**: Backup and restore account data
- **List Chats**: View all your Telegram chats
- **Settings**: Configure account namespace and proxy settings

## Prerequisites

### 1. Install tdl

Make sure `tdl` is installed and available in your PATH:

```bash
# Check if tdl is installed
tdl version
```

If not installed, follow the [official installation guide](https://docs.iyear.me/tdl/getting-started/installation/).

### 2. Install System Dependencies (Linux)

Fyne requires OpenGL and X11 development libraries:

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y libgl1-mesa-dev xorg-dev libxxf86vm-dev libxcursor-dev libxrandr-dev libxinerama-dev libxi-dev

# Fedora
sudo dnf install mesa-libGL-devel libXcursor-devel libXrandr-devel libXinerama-devel libXi-devel libXxf86vm-devel

# Arch Linux
sudo pacman -S libgl libxcursor libxrandr libxinerama libxi
```

### 3. Install Go

Go 1.23+ is required:

```bash
go version  # Should show 1.23.x or later
```

## Building

```bash
cd gui

# Download dependencies
go mod tidy

# Build the application
go build -o tdl-gui .

# Or build with optimizations
go build -ldflags="-s -w" -o tdl-gui .
```

## Running

```bash
./tdl-gui
```

## Architecture

```
┌─────────────────────────────────────────┐
│             TDL GUI (Fyne)              │
│                                         │
│  ┌─────────────────────────────────┐    │
│  │         User Interface           │    │
│  │  (Buttons, Text Areas, etc.)    │    │
│  └─────────────────────────────────┘    │
│                  │                       │
│                  ▼                       │
│  ┌─────────────────────────────────┐    │
│  │      exec.Command("tdl", ...)   │    │
│  │     (subprocess execution)       │    │
│  └─────────────────────────────────┘    │
└─────────────────────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────┐
│              tdl binary                  │
│  (handles all Telegram functionality)   │
└─────────────────────────────────────────┘
```

## Command Mapping

| GUI Button | tdl Command |
|------------|-------------|
| LOGIN (Desktop) | `tdl login` |
| LOGIN (Passcode) | `tdl login -p <passcode>` |
| LOGIN (QR) | `tdl login -T qr` |
| LOGIN (Phone) | `tdl login -T code` |
| DOWNLOAD | `tdl dl -u <url> [-d <dir>]` |
| UPLOAD | `tdl up -p <file> [-c <chat>]` |
| FORWARD | `tdl forward --from <url> [--to <chat>]` |
| BACKUP | `tdl backup [-d <file>]` |
| RECOVER | `tdl recover -f <file>` |
| LIST CHATS | `tdl chat ls` |
| VERSION | `tdl version` |

## Global Settings

- **Account Name**: Uses `-n <namespace>` flag for multi-account support
- **Proxy**: Uses `--proxy <url>` flag for proxy connections

## Troubleshooting

### Build errors about missing libraries

Make sure you've installed all system dependencies (see Prerequisites above).

### "tdl: command not found"

Ensure `tdl` is installed and in your PATH:

```bash
which tdl  # Should show the path to tdl
```

### Login not working

Some login methods are interactive and require terminal input. Use QR code login for best GUI experience.

## License

Same license as the main tdl project.

