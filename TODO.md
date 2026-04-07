# TODO

## Features

- [ ] **LP4 support** — Quarter capacity vs SP, if atracdenc supports ATRAC3 LP4 encoding
- [x] ~~Download output format~~ — Downloads as MP3 via native exploit + ATRAC3 extraction + ffmpeg conversion
- [ ] **Large file storage download** — NoRam exploit reads from fixed cache positions (~76 sectors). Options: (a) modify JS script to use lower-level sector reading API for raw sectors, (b) implement CachedSectorControlDownload in Go. JS bridge's `downloadTrack()` reformats data as ATRAC3 WAV, stripping our frame structure.
- [ ] **More device support** — Exploit constants are MZ-N505-specific; other Type-R/S/Hi-MD devices need different firmware addresses from netmd-exploits device tables
- [x] ~~Exploit cleanup/unpatch~~ — Resolved by CRC16, g_DiscStateStruct, and no-poll fixes. PatchFirmware is idempotent; USB replug sufficient for recovery.
- [ ] **Disc spinning animation** — Animated spinning disc in the disc info panel (needs better ASCII art)
- [ ] **File storage: TUI integration** — Store/retrieve files from within the TUI (currently CLI-only via --store)

## Bugs

- [x] ~~WAV header bounds check~~ — added length check before header parsing
- [x] ~~Secure session bounds check~~ — added `len(r) < 15` check before slicing
- [x] ~~Sequence number overflow~~ — reject files >65534 frames (~12.4MB) at encode time
- [x] ~~Division by zero in download progress~~ — already guarded in both upload and download
- [x] ~~Ignored errors in factory mode~~ — EnterFactoryMode now checks all submit/receive errors
- [x] ~~USB poll error not checked~~ — poll() now checks Control() error and logs in debug mode
- [x] ~~factoryReceive poll error not checked~~ — returns immediately on USB error instead of spinning
- [x] ~~findDownloadScript symlink resolution~~ — uses filepath.EvalSymlinks() to resolve /proc/self/exe

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
