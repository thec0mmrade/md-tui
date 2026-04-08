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
./md-tui --store encode <file> <out.wav>   # Encode file for MiniDisc storage
./md-tui --store decode <raw> <outdir>     # Decode downloaded raw data to file
./md-tui --store calibrate <out.wav> [N]   # Generate calibration WAV
./md-tui --store analyze <raw>             # Analyze raw sector layout
./md-tui --store firmware <out.bin>        # Dump firmware ROM + SRAM (~10 min, size varies by device)
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
- `internal/mdstore/` — Arbitrary file storage on MiniDisc
  - `encode.go` — Encodes files into LP2 ATRAC3 WAV containers (192-byte frames with metadata + data + padding)
  - `decode.go` — Decodes downloaded raw sector data back to files. Handles rotated sector order via frame sequence numbers. LP2 sector layout: 20-byte header + 11 × (12-byte SG header + 192-byte frame + 8-byte padding) = 2352 bytes.
  - `wav.go` — ATRAC3 WAV container builder (format tag 624, nBlockAlign 384). Also used by the MP3 download pipeline to wrap extracted ATRAC3 frames for ffmpeg conversion.
  - `calibrate.go` — Generates calibration WAVs and analyzes raw sector data to map sector layout
- `scripts/` — Helper scripts
  - `download.mjs` — Node.js track download fallback. Requires `npm install`.
  - `firmware-dump.mjs` — JS firmware dump using netmd-exploits FirmwareDumper
  - `analyze-firmware.py` — Automated firmware analysis using capstone (ARM disassembler)
- `docs/` — Research and analysis documents
  - `firmware-analysis.md` — MZ-N505 R1.400 firmware RE report (2536 functions, USB handlers, EEPROM refs)
  - `firmware-dump-research.md` — Firmware extraction techniques and boundary bypass details
  - `homebrew-minidisc-player.md` — Open-source MiniDisc reader/writer hardware research

### Key Design Decisions

- **theme package** exists to break import cycles — view packages import `theme`, not `ui`
- **Modals** (trackedit, confirm, upload, download) are overlaid via `lipgloss.Place` and consume all key input when active
- **Async USB ops** run as `tea.Cmd` functions; long operations (upload/download) use goroutine + channel pattern with `tea.Program.Send()` for progress updates
- **All mutations** (rename, delete, move, wipe, upload) trigger `refreshDisc()` to reload disc contents
- **Error banners** auto-dismiss after 5 seconds via `tea.Tick`
- **Upload pipeline**: audio file → ffmpeg (if non-WAV) → atracdenc (if LP2/LP4, skipped if already ATRAC3) → NewTrack → Send. LP4 uses `--bitrate 64` flag.
- **Download pipeline**: R-series native exploit reads raw sectors → if MP3: extract ATRAC3 frames from SG structure → wrap in ATRAC3 WAV via `mdstore.BuildATRAC3WAV()` → ffmpeg converts to MP3. If `.raw`: save raw sectors directly. S-series downloads use JS bridge fallback (netmd-exploits) — `.raw` works, `.mp3` conversion not yet working (needs encoding-aware WAV container for SP vs LP2 vs LP4).
- **Theme system**: 7 built-in palettes defined in `theme.go` as `Palette` structs. `Apply()` reassigns all color vars and recomputes all style vars. Views read `theme.*` on each `View()` call so changes take effect immediately. `CycleTheme()` cycles through palettes via `t`/`T` keybindings.
- **File storage**: LP2 upload path stores data verbatim (no re-encoding). `track.go:152` does `break` for WfLP2 — no byte transformation. Files encoded as 192-byte frames: 3-byte header (type + sequence) + 189-byte payload. Metadata frame stores filename, size, SHA-256. Decoder handles circular cache rotation via sequence number sorting and deduplication.
- **Download limitation**: The NoRam exploit reads from fixed DRAM cache positions (~76 sectors). Files >175KB and audio tracks >8s (LP2) may have incomplete data. The CachedSectorControlDownload exploit variant is needed for full-size downloads — it patches the firmware USB handler to serve sectors sequentially (like Web MiniDisc Pro).
- **Multi-device support**: Device type detected at runtime via factory `1812` command (BCD version byte). R-series (MZ-N505) uses native NoRam exploit for downloads. S-series (MZ-NE410 etc.) falls back to Node.js bridge — native S-series exploit not yet working (hardware patches apply but `readReply` interception fails). Device-specific address tables in `device_profile.go` cover R1.000-R1.400 and S1.000-S1.600.

### Vendored netmd Fixes (internal/netmd/)

The vendored library has critical fixes over upstream `github.com/enimatek-nl/go-netmd-lib`:

- **USB reset** (`netmd.go`): `dev.Reset()` called after opening — without this, control transfers time out on MZ-N505
- **Control timeout** (`netmd.go`): `dev.ControlTimeout = 2s` — google/gousb defaults to infinite, causing hangs
- **Auto-detach** (`netmd.go`): `dev.SetAutoDetach(true)` — detaches kernel usbfs driver before claiming interface
- **Config/Interface error handling** (`netmd.go`): `Config()` and `Interface()` errors are now checked instead of silently ignored with `_`; `config.Close()` no longer called inside the endpoint loop (was invalidating endpoint handles)
- **Stereo/mono fix** (`track.go:103`): Channel detection was inverted (`!= 1` instead of `== 1`), causing stereo files to be rejected by the device
- **Switched from forked gousb** to standard `github.com/google/gousb`
- **Added playback control** (`playback.go`): Play, Pause, Stop, GotoTrack, GetPosition — uses `playbackCommand()` helper to handle 0xff→0x00 check byte mismatch in responses
- **Multi-device exploit support** (`device_profile.go`, `exploit.go`, `exploit_setup.go`, `exploit_control.go`, `download.go`): Runtime device detection and dynamic firmware patching for R-series (R1.000-R1.400) and S-series (S1.000-S1.600) devices.
  - `device_profile.go`: `DeviceProfile` struct with all exploit-relevant addresses per firmware version. Address tables sourced from netmd-exploits JS library. `LookupProfile()` maps version code to profile. `parseVersionCode()` builds version code from factory `1812` response (BCD format).
  - `exploit.go`: `EnterFactoryMode()` parses `1812` response to detect chipType (0x20=R, 0x21=S) and firmware version, stores profile on `NetMD` struct. `buildNoRamARMCode()` patches `discStructOffset` and `read_atrac_dram` into ARM code template. `ExploitReadSectorChunk` uses profile for exec command (`0xd3` R-series, `0xd2` S-series), `g_DiscStateStruct`, and ARM code. `DumpFullFirmware` uses profile ROM/SRAM sizes. S-series poll-based reads for diagnostics.
  - `exploit_setup.go`: `PatchFirmware()` uses dynamic `ApplyFirmwarePatch()` with profile addresses (slot, onePatchAddress, onePatchValue). Falls back to `patchFirmwareLegacy()` (37 pre-captured MZ-N505 commands) when no profile detected.
  - `exploit_control.go`: `buildResidentCode()` patches constant pool with device-specific addresses. `SetupControlDownload()` uses `applyPatchViaARM()` for S-series (ARM code execution via `18 d2` for USB handler patching). `EnableSectorReading`/`DisableSectorReading` use profile addresses.
  - `download.go`: R-series uses NoRam exploit (native). S-series skipped for native (not yet working), falls back to Node.js bridge. Device detection happens in `downloadNative` via device name check to avoid entering factory mode before JS bridge needs USB access.
