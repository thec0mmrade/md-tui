# md-tui

A terminal UI for managing Sony NetMD MiniDisc devices.

<!-- TODO: add screenshot -->
<!-- ![md-tui screenshot](screenshot.png) -->

## Features

- **List tracks** — title, duration, format (SP/LP2/LP4), channel info
- **Upload audio** — MP3, FLAC, OGG, AAC, WAV (non-WAV files auto-converted via ffmpeg)
- **SP and LP2 formats** — LP2 doubles disc capacity via atracdenc encoding
- **Batch upload** — select a folder to upload all audio files sequentially, auto-sets disc title
- **Stop batch** — press Esc during batch upload to stop after current track
- **3-pane file browser** — yazi-style Miller columns for navigating to files
- **Rename** tracks and disc titles
- **Delete** individual tracks
- **Move** tracks to reorder
- **Wipe** entire disc
- **Download/rip tracks** — extract audio from disc via native exploit-based download (no external dependencies)
- **Disc info** — used/free/total time, write-protection status

## Requirements

- **libusb 1.0**
  - Arch: `pacman -S libusb`
  - Debian/Ubuntu: `apt install libusb-1.0-0-dev`
  - macOS: `brew install libusb`
- **ffmpeg** (for uploading MP3, FLAC, and other non-WAV formats)
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
| `j/k` `↑/↓` | Navigate |
| `?` | Toggle help |
| `q` | Quit |

## Tested Devices

- Sony MZ-N505

Other NetMD devices (MZ-N1, MZ-N707, MZ-N710, MZ-RH1, Sharp IM-DR series, etc.) should work but are untested. Please open an issue if you encounter problems with your device.

## Track Download

Track downloading uses a native implementation of the CachedSectorNoRamControlDownload exploit, which reads ATRAC audio data directly from the device's anti-shock DRAM buffer via ARM code execution. This works without any external dependencies beyond libusb.

If the native exploit fails, md-tui automatically falls back to a Node.js bridge (`scripts/download.mjs`). The fallback requires Node.js 18+ and `npm install` in the `scripts/` directory.

Currently verified on the Sony MZ-N505 (R1.400 firmware). Other Type-R NetMD devices may work but need device-specific firmware constants.

## Acknowledgments

- [go-netmd-lib](https://github.com/enimatek-nl/go-netmd-lib) — Go NetMD protocol implementation (vendored with fixes)
- [Charm](https://github.com/charmbracelet) — bubbletea, lipgloss, bubbles TUI libraries
- [Web MiniDisc Pro](https://github.com/asivery/webminidisc) — reference for NetMD protocol behavior
- [netmd-exploits](https://github.com/asivery/netmd-exploits) — reference for CachedSectorNoRamControlDownload exploit protocol

## License

MIT
