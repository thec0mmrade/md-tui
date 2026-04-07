# md-tui

A terminal UI for managing Sony NetMD MiniDisc devices.

<!-- TODO: add screenshot -->
<!-- ![md-tui screenshot](screenshot.png) -->

## Features

- **List tracks** — title, duration, format (SP/LP2/LP4), channel info
- **Upload audio** — MP3, FLAC, OGG, AAC, WAV (non-WAV files auto-converted via ffmpeg)
- **SP, LP2, and LP4 formats** — LP2 doubles, LP4 quadruples disc capacity via atracdenc encoding
- **Batch upload** — select a folder to upload all audio files sequentially, auto-sets disc title
- **Stop batch** — press Esc during batch upload to stop after current track
- **3-pane file browser** — yazi-style Miller columns for navigating to files
- **Rename** tracks and disc titles
- **Delete** individual tracks
- **Move** tracks to reorder
- **Wipe** entire disc
- **Download/rip tracks** — extract audio from disc as MP3 via native exploit (requires ffmpeg for MP3 conversion)
- **Disc info** — used/free/total time, write-protection status
- **Color themes** — 7 built-in themes (Default, OneDark Pro, Tokyo Night, Catppuccin, Gruvbox, Dracula, Nord)
- **File storage** — store arbitrary files (images, docs, etc.) on MiniDisc via CLI

## Requirements

- **libusb 1.0**
  - Arch: `pacman -S libusb`
  - Debian/Ubuntu: `apt install libusb-1.0-0-dev`
  - macOS: `brew install libusb`
- **ffmpeg** (for uploading non-WAV formats and downloading as MP3)
- **atracdenc** (optional, for LP2 uploads — [github.com/dcherednik/atracdenc](https://github.com/dcherednik/atracdenc))
- **Node.js 18+** (optional fallback for track download — run `npm install` in `scripts/`; not needed if native exploit works)
- **Go 1.21+** (to build from source)

### Linux udev rules

To access the device without root, create `/etc/udev/rules.d/60-netmd.rules`:

```
SUBSYSTEM=="usb", ATTR{idVendor}=="054c", MODE="0666"
SUBSYSTEM=="usb", ATTR{idVendor}=="04dd", MODE="0666"
```

Then reload: `sudo udevadm control --reload-rules && sudo udevadm trigger`

## Install

```bash
git clone https://github.com/thec0mmrade/md-tui.git
cd md-tui
go build .
```

## Usage

```
./md-tui              # connect to NetMD device
./md-tui --mock       # demo mode (no device needed)
./md-tui --debug      # verbose USB logging
```

### Keybindings

| Key | Action |
|-----|--------|
| `u` | Upload a track or folder |
| `r` | Rename selected track |
| `R` | Rename disc |
| `d` | Delete selected track |
| `m` | Move selected track |
| `W` | Wipe disc |
| `x` | Extract/download track |
| `t/T` | Next/previous color theme |
| `j/k` `↑/↓` | Navigate |
| `?` | Toggle help |
| `q` | Quit |

## Tested Devices

- Sony MZ-N505

Other NetMD devices (MZ-N1, MZ-N707, MZ-N710, MZ-RH1, Sharp IM-DR series, etc.) should work but are untested. Please open an issue if you encounter problems with your device.

## Track Download

Track downloading uses a native implementation of the CachedSectorNoRamControlDownload exploit, which reads ATRAC audio data directly from the device's anti-shock DRAM buffer via ARM code execution.

Downloaded tracks default to MP3 format. The exploit reads raw ATRAC sectors, extracts ATRAC3 frames, wraps them in a WAV container, and converts to MP3 via ffmpeg. Use `.raw` extension to save raw sector data instead.

If the native exploit fails, md-tui falls back to a Node.js bridge (`scripts/download.mjs`). The fallback requires Node.js 18+ and `npm install` in the `scripts/` directory.

Currently verified on the Sony MZ-N505 (R1.400 firmware). Other Type-R NetMD devices may work but need device-specific firmware constants.

## File Storage

md-tui can store arbitrary files (images, documents, etc.) on MiniDisc by encoding them as LP2 tracks. The device stores LP2 data verbatim without lossy re-encoding, enabling lossless round-trip storage.

```bash
# Encode a file into an LP2 WAV
./md-tui --store encode photo.jpg photo.wav

# Upload photo.wav as LP2 via the TUI, then download the track

# Decode the downloaded raw data back to the original file
./md-tui --store decode photo.raw ./output/
```

Each file is split into 192-byte frames with a metadata header containing the original filename and SHA-256 checksum. Currently limited to ~175KB per track (the device's anti-shock DRAM cache size).

## Acknowledgments

- [go-netmd-lib](https://github.com/enimatek-nl/go-netmd-lib) — Go NetMD protocol implementation (vendored with fixes)
- [Charm](https://github.com/charmbracelet) — bubbletea, lipgloss, bubbles TUI libraries
- [Web MiniDisc Pro](https://github.com/asivery/webminidisc) — reference for NetMD protocol behavior
- [netmd-exploits](https://github.com/asivery/netmd-exploits) — reference for CachedSectorNoRamControlDownload exploit protocol

## License

MIT
