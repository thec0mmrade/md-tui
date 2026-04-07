# Changelog

## 0.5.0 — 2026-04-06

### Added
- LP4 upload support — quadruples disc capacity (80min → 320min) via atracdenc `--bitrate 64`
  - Format selector in upload dialog now shows SP / LP2 / LP4
  - LP4 uses 96-byte ATRAC3 frames at 64kbps
- MP3 download output — tracks download directly as playable MP3 files
  - Native exploit reads raw ATRAC sectors, extracts ATRAC3 frames from sound groups
  - Frames wrapped in ATRAC3 WAV container, converted to MP3 via ffmpeg
  - Default output extension changed from `.raw` to `.mp3`
  - Use `.raw` extension to save raw sector data (for file storage decode)

- Firmware dump — `--store firmware <output.bin>` extracts full 448KB ROM + 18KB SRAM from device
  - Bypasses factory read boundary check by patching SRAM boundary table
  - Verified byte-identical to netmd-exploits JS FirmwareDumper output
- Firmware analysis script (`scripts/analyze-firmware.py`) — automated RE with capstone
  - 2,536 functions identified, 1,570 strings, USB command handler mapping
  - Known exploit addresses disassembled and cross-referenced
- Homebrew MiniDisc reader/writer research document (`docs/homebrew-minidisc-player.md`)

### Fixed
- WAV header bounds check: truncated WAV files now error instead of panicking
- Secure session bounds check: short device responses now error instead of panicking
- Sequence number overflow: files >65534 frames (~12.4MB) rejected at encode time
- EnterFactoryMode: all USB submit/receive errors now checked (were silently ignored)
- poll(): USB Control() errors now checked and logged (was returning -1 silently)
- factoryReceive: fails fast on USB error instead of spinning for 30 seconds
- findDownloadScript: resolves /proc/self/exe symlinks via filepath.EvalSymlinks()

## 0.4.0 — 2026-04-05

### Added
- Arbitrary file storage on MiniDisc — store any file (images, documents, etc.) as LP2 tracks
  - CLI commands: `--store encode`, `--store decode`, `--store calibrate`, `--store analyze`
  - Files encoded into LP2 ATRAC3 WAV containers with metadata frame (filename, SHA-256 checksum)
  - Device stores data verbatim via LP2 upload path (no lossy re-encoding)
  - Decoder handles rotated sector order from circular disc cache using frame sequence numbers and raw byte scanning
  - Tested up to 175KB per track on MZ-N505 (anti-shock DRAM cache limit)
- Upload now detects pre-encoded ATRAC3 WAV files and skips atracdenc conversion
- Animated progress dots on upload/download phase text
- Dim/brighten fade transition on view switch, modal slide-in animation

### Changed
- Download flow simplified to match Web MiniDisc Pro pcap: Stop → Factory → GotoTrack → Patch → Read (no Play/Pause cycling)
- Removed 50ms per-chunk delay from exploit reads (~2.3x faster downloads)

### Fixed
- LP2 sector rate estimate increased from 6 to 9 sectors/second (was underestimating download size)
- Short tracks (< 1 second) now download at least 1 sector instead of 0

## 0.3.0 — 2026-04-05

### Added
- 7 built-in color themes: Default, OneDark Pro, Tokyo Night, Catppuccin, Gruvbox, Dracula, Nord
  - Press `t` to cycle forward, `T` to cycle backward
  - Theme name shown in status bar on switch
  - All views update immediately (colors, borders, selections, logo)

## 0.2.0 — 2026-04-05

### Added
- Native exploit-based track download — no longer requires Node.js/npm
  - CachedSectorNoRamControlDownload exploit implemented in pure Go
  - ARM code execution on device reads ATRAC sectors from anti-shock DRAM buffer
  - Automatic fallback to Node.js bridge if native exploit fails
- Factory write commands now include CRC16-CCITT checksums (required by device for patch peripheral writes)

### Fixed
- Firmware patch writes to hardware patch peripheral were silently rejected (missing CRC16 checksums)
- Exploit ARM code crashed device due to missing `g_DiscStateStruct` parameter (4 DWORDs required after bytecode, not 3)
- Exploit response reads failed when using poll — hooked firmware handler bypasses normal poll mechanism; now reads directly via USB request 0x81
- Exploit activation command (`18 d3`) moved after DRAM patches so it executes via the installed hook instead of being echoed
- Disc cache now pre-filled via Play() before entering factory mode (cache must be populated for exploit reads to return audio data)
- `crc16ccitt()` had incorrect algorithm (replaced crc instead of XOR-ing with temp)

## 0.1.0 — 2026-04-04

Initial release.

### Added
- TUI with track table, disc info sidebar, and keybind-driven navigation
- Device auto-discovery and connection (scans USB for NetMD devices)
- Track listing with title, duration, format (SP/LP2/LP4), and channel info
- Upload audio files to MiniDisc (WAV native, MP3/FLAC/OGG/AAC via ffmpeg conversion)
- Rename tracks and disc titles
- Delete individual tracks with confirmation
- Move/reorder tracks
- Wipe entire disc with confirmation
- Track download/ripping via Node.js exploit bridge (requires Node.js + npm install in scripts/)
- Mock device mode (`--mock`) for UI development without hardware
- Debug mode (`--debug`) for verbose USB logging
- Device select view with auto-connect for single device

### Changed
- Device model name (e.g. "Sony MZ-N505") now shown in title bar and status bar instead of "NetMD Device 0"
- Upload dialog replaced with yazi-style 3-pane Miller columns file browser (parent/current/preview)
- Navigation: h/j/k/l or arrow keys, Enter/u to upload file or batch-upload directory
- Batch upload: selecting a folder queues all audio files for sequential upload with track counter
- Batch upload uses natural sort order (1, 2, 10 instead of 1, 10, 2)
- LP2 upload support via atracdenc — doubles disc capacity (80min → 160min)
- Format selector (SP/LP2) in single upload title screen and batch confirm screen
- Error banners auto-dismiss after 5 seconds
- Responsive layout: terminal resize updates all views including file browser panes
- Playback control commands: play, pause, stop, goto track, get position
- Track download exploit: factory mode init, firmware read, DRAM patching sequence decoded from USB pcap
- Factory mode commands: `EnterFactoryMode()`, `ReadFirmware()`, `factoryReceive()` with adaptive request routing
- DRAM patching: 37-command firmware patch sequence (`PatchFirmware()`) enables exploit sector reading
- 16-bit poll size fix: `poll()` now reads `buf[2] | buf[3]<<8` for responses >255 bytes
- Track download via Node.js bridge: extracts ATRAC audio from disc using netmd-exploits exploit engine
- Download output: ATRAC3 WAV file (convertible to PCM WAV via ffmpeg)
- Auto-set disc title to folder name after batch upload
- MiniDisc logo rendered as dithered half-block art in disc info panel
- Auto-set disc title to folder name after batch upload
- Fixed download dialog hanging after completion (Node.js exploit cleanup was blocking process exit)
- Fixed double-close panic in send.go during upload errors
- Fixed time display showing raw seconds as minutes (now shows h:mm:ss)
- Fixed concurrent USB commands during batch rename + refresh

### Fixed (vendored go-netmd-lib)
- USB device reset required before communication (MZ-N505 control transfers time out otherwise)
- Control transfer timeout set to 2s (google/gousb defaults to infinite)
- Auto-detach kernel driver before claiming USB interface
- Config/Interface error handling (was silently ignored, causing nil pointer panics)
- `config.Close()` removed from inside endpoint loop (was invalidating endpoint handles)
- Stereo/mono channel detection inverted (`!= 1` instead of `== 1`)
- Switched from forked `enimatek-nl/gousb` to standard `google/gousb`
- Upload error handling: errors now close the upload modal and display in error banner (was stuck on progress screen)
- Added `Wait()` calls between secure session commands in send flow
- Better error context in transfer errors (prefixed with failing step name)
