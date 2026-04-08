# TODO

## Features

- [x] ~~LP4 support~~ — LP4 upload via atracdenc `--bitrate 64`, quadruples disc capacity vs SP
- [ ] **Large file storage download** — NoRam exploit reads from fixed cache positions (~76 sectors). CachedSectorControlDownload resident code assembled and verified but USB response interception point is wrong — handler at 0x574fc triggers send AFTER data is in buffer, need to intercept where response data is PREPARED. See `exploit_control.go` for working ARM code.
- [ ] **More device support** — Dynamic exploit patching with device detection (R1.000-R1.400, S1.000-S1.600). Runtime firmware version detection via factory 1812 command. S-series downloads work via JS bridge fallback. S-series native exploit blocked — hardware patches apply but don't intercept `readReply`. R-series native download works with dynamic patching.
- [ ] **S-series MP3 download** — JS bridge downloads raw ATRAC data but MP3 conversion fails. `BuildATRAC3WAV` assumes LP2 (ATRAC3, nBlockAlign=384) but SP tracks use ATRAC1 (nBlockAlign=2048, different codec params). Need encoding-aware WAV container: detect track encoding (SP/LP2/LP4) and build appropriate WAV header before ffmpeg conversion.
- [ ] **Disc spinning animation** — Animated spinning disc in the disc info panel (needs better ASCII art)
- [ ] **File storage: TUI integration** — Store/retrieve files from within the TUI (currently CLI-only via --store)

## Completed

- [x] ~~Firmware dump~~ — 448KB ROM + 18KB SRAM via boundary check bypass (Go native + JS)
- [x] ~~Firmware analysis~~ — Automated analysis script, 2536 functions, USB handler mapping, EEPROM references
- [x] ~~LP4 support~~ — LP4 upload via atracdenc `--bitrate 64`
- [x] ~~Bug fixes (0.5.0)~~ — WAV/secure bounds checks, sequence overflow protection, factory mode error handling, USB poll/receive error checking, symlink resolution
- [x] ~~Download output format~~ — Downloads as MP3 via native exploit + ATRAC3 extraction + ffmpeg conversion
- [x] ~~Exploit cleanup/unpatch~~ — Resolved by CRC16, g_DiscStateStruct, and no-poll fixes
- [x] ~~Arbitrary file storage~~ — Encode any file as LP2 track, decode back after download (≤175KB)
- [x] ~~UI themes~~ — 7 built-in themes, cycle with t/T
- [x] ~~Animations~~ — Progress bar dots, dim/brighten transitions, modal slide-in
- [x] ~~Track download (native exploit)~~ — CRC16 checksums, g_DiscStateStruct param, no-poll reads
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
