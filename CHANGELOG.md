# Changelog

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
- Download view (UI present, protocol not yet supported by library)
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
