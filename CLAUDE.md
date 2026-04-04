# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

md-tui is a Go TUI application for managing Sony NetMD MiniDisc devices. Built with Charm's bubbletea/lipgloss/bubbles libraries. Uses a vendored fork of go-netmd-lib (`internal/netmd/`) for USB communication. Supports listing tracks, uploading audio files (WAV, MP3, FLAC, etc.), renaming, deleting, moving tracks, and wiping discs.

## Build Commands

```bash
go build .                    # Build binary (produces ./md-tui)
./md-tui                      # Run with real NetMD device (requires libusb-1.0)
./md-tui --mock               # Run with mock device (no hardware needed)
./md-tui --debug              # Run with verbose USB logging
```

Requires `libusb-1.0-dev` (Linux) or `libusb` (macOS via Homebrew) for real device support. Non-WAV uploads require `ffmpeg` in PATH for conversion. LP2 uploads require `atracdenc` in PATH.

## Architecture

### Elm Architecture (Model-Update-View)

The app follows bubbletea's Elm Architecture pattern. The root model (`internal/ui/app.go`) routes messages to child models based on `ViewState`. Each view is a separate package with its own `Model`, `Update()`, and `View()`.

### Package Structure

- `internal/netmd/` — **Vendored fork of go-netmd-lib** with bug fixes. Do not replace with upstream without preserving fixes (see below).
- `internal/device/` — Device abstraction layer. `DeviceService` interface decouples the TUI from USB details.
  - `device.go` — Interface + domain types (Track, Disc, Encoding, TransferProgress)
  - `netmd.go` — Real implementation wrapping vendored netmd. Handles MP3/FLAC/etc. conversion to WAV via ffmpeg, and LP2 encoding via atracdenc.
  - `mock.go` — Mock for UI development without hardware (`--mock` flag)
- `internal/ui/` — TUI layer
  - `app.go` — Root model, view state machine, message routing
  - `theme/theme.go` — All lipgloss styles and keybinding definitions (shared by all views)
  - `messages.go` — Shared message types
  - `discview/` — Main screen: track table (bubbles/table) + disc info sidebar
  - `deviceselect/` — Device scan and selection
  - `browser/` — Yazi-style 3-pane Miller columns file browser (parent/current/preview)
  - `upload/` — Upload flow: file browser → title input → progress bar. Supports single file and batch directory upload.
  - `download/` — Track extraction with progress
  - `trackedit/` — Rename modal (track or disc title)
  - `confirm/` — Reusable yes/no confirmation dialog
  - `statusbar/` — Bottom bar with device name and status

### Key Design Decisions

- **theme package** exists to break import cycles — view packages import `theme`, not `ui`
- **Modals** (trackedit, confirm, upload, download) are overlaid via `lipgloss.Place` and consume all key input when active
- **Async USB ops** run as `tea.Cmd` functions; long operations (upload/download) use goroutine + channel pattern with `tea.Program.Send()` for progress updates
- **All mutations** (rename, delete, move, wipe, upload) trigger `refreshDisc()` to reload disc contents
- **Error banners** auto-dismiss after 5 seconds via `tea.Tick`
- **Upload pipeline**: audio file → ffmpeg (if non-WAV) → atracdenc (if LP2) → NewTrack → Send

### Vendored netmd Fixes (internal/netmd/)

The vendored library has critical fixes over upstream `github.com/enimatek-nl/go-netmd-lib`:

- **USB reset** (`netmd.go`): `dev.Reset()` called after opening — without this, control transfers time out on MZ-N505
- **Control timeout** (`netmd.go`): `dev.ControlTimeout = 2s` — google/gousb defaults to infinite, causing hangs
- **Auto-detach** (`netmd.go`): `dev.SetAutoDetach(true)` — detaches kernel usbfs driver before claiming interface
- **Config/Interface error handling** (`netmd.go`): `Config()` and `Interface()` errors are now checked instead of silently ignored with `_`; `config.Close()` no longer called inside the endpoint loop (was invalidating endpoint handles)
- **Stereo/mono fix** (`track.go:103`): Channel detection was inverted (`!= 1` instead of `== 1`), causing stereo files to be rejected by the device
- **Switched from forked gousb** to standard `github.com/google/gousb`
- **Added playback control** (`playback.go`): Play, Pause, Stop, GotoTrack, GetPosition — uses `playbackCommand()` helper to handle 0xff→0x00 check byte mismatch in responses
- **Added exploit download** (`exploit.go`, `download.go`): CachedSectorNoRamControlDownload exploit — reads ATRAC sectors via ARM code execution on device. Pre-compiled bytecode captured from MZ-N505 USB trace. Command `18 d3 ff` executes ARM code; `18 24 ff` reads firmware; sectors read in 6 chunks of 420+252 bytes = 2352 bytes per sector.
