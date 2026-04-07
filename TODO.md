# TODO

## Features

- [ ] **LP4 support** — Quarter capacity vs SP, if atracdenc supports ATRAC3 LP4 encoding
- [x] ~~Download output format~~ — Downloads as MP3 via native exploit + ATRAC3 extraction + ffmpeg conversion
- [ ] **Large file storage download** — NoRam exploit reads from fixed cache positions (~76 sectors). Options: (a) modify JS script to use lower-level sector reading API for raw sectors, (b) implement CachedSectorControlDownload in Go. JS bridge's `downloadTrack()` reformats data as ATRAC3 WAV, stripping our frame structure.
- [ ] **More device support** — Exploit constants are MZ-N505-specific; other Type-R/S/Hi-MD devices need different firmware addresses from netmd-exploits device tables
- [ ] **Exploit cleanup/unpatch** — Add firmware unpatch sequence so device recovers without battery pull after failed downloads
- [ ] **Disc spinning animation** — Animated spinning disc in the disc info panel (needs better ASCII art)
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
