# Listnr - Terminal Music Player

A modern, modular terminal-based music player written in Go.

## Features

- 🎵 Support for MP3, WAV, FLAC, OGG, M4A formats
- 📁 Directory-based music library browsing
- ⚡ Real-time playback controls
- 🎛️ Volume control with visual feedback
- ⌨️ Vim-inspired keyboard shortcuts
- 🎨 Clean, responsive TUI interface

## Architecture

```
listnr/
├── cmd/listnr/          # Application entry point
├── internal/
│   ├── audio/           # Audio engine (decoding, playback)
│   ├── library/         # Music library management
│   ├── config/          # Configuration handling
│   ├── ui/              # Terminal user interface
│   │   └── components/  # Reusable UI components
│   └── events/          # Event system for component communication
```

## Installation

```bash
git clone https://github.com/sammwyy/listnr
cd listnr
go build ./cmd/listnr
./listnr
```

## Usage

### Navigation
- `ESC`: Toggle between library explorer and playback controls
- `←/→`: Navigate between sidebar and song list (explorer mode)
- `←/→`: Seek backward/forward 5 seconds (controls mode)
- `↑/↓`: Volume up/down (controls mode)

### Playback
- `SPACE`: Play/pause
- `A`: Previous song
- `D`: Next song
- `Q`: Quit
- `R`: Toggle repeat mode
- `N`: Toggle autoplay mode

### Configuration

Configuration file is automatically created at `~/.config/listnr.json`:

```json
{
  "music_routes": ["/home/user/Music"],
  "volume": 0.5,
  "last_path": "",
  "autoplay_enabled": true,
  "repeat_mode": false
}
```