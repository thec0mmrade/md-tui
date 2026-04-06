# TODO

## Features

- [ ] **LP4 support** — Quarter capacity vs SP, if atracdenc supports ATRAC3 LP4 encoding
- [ ] **Download output format** — Convert downloaded raw ATRAC data to MP3 via ffmpeg
- [ ] **More device support** — Exploit constants are MZ-N505-specific; other Type-R/S/Hi-MD devices need different firmware addresses from netmd-exploits device tables
- [ ] **Exploit cleanup/unpatch** — Add firmware unpatch sequence so device recovers without battery pull after failed downloads
- [x] ~~UI themes~~ — 7 built-in themes (Default, OneDark Pro, Tokyo Night, Catppuccin, Gruvbox, Dracula, Nord), cycle with t/T
- [ ] **Disc spinning animation** — Animated spinning disc in the disc info panel (needs better ASCII art)
- [x] ~~Animations (partial)~~ — Progress bar dot animation, dim/brighten view transitions, modal slide-in
- [x] ~~Arbitrary file storage~~ — Encode any file as LP2 track, decode back after download. CLI: `--store encode/decode/calibrate/analyze`. Tested up to 175KB on MZ-N505 (anti-shock cache limit). Larger files need chunked download (future).
- [ ] **File storage: chunked download** — Read sectors while disc plays to support files larger than 175KB cache limit
- [ ] **File storage: TUI integration** — Store/retrieve files from within the TUI (currently CLI-only via --store)

## Completed

- [x] ~~Track download (native exploit)~~ — CRC16 checksums, g_DiscStateStruct param, no-poll reads, cache pre-fill via Play()
- [x] ~~Stop batch after current track~~
- [x] ~~Auto-set disc title from folder name~~
- [x] ~~MiniDisc logo in UI~~ — dithered half-block character rendering from PNG
- [x] ~~LP2 upload support via atracdenc~~
- [x] ~~File browser in upload dialog~~
- [x] ~~Batch upload from folder~~
- [x] ~~Track download (Node.js bridge)~~
- [x] ~~Show device model name instead of "NetMD Device 0"~~
- [x] ~~Responsive layout improvements for small terminals~~
- [x] ~~Error banner auto-dismiss after timeout~~
