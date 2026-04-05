# TODO

## Features

- [ ] **Track download (native exploit)** — Root cause found: must bypass poll() and read with explicit length via 0x81. Also need proper `ApplyFirmwarePatch()` using hardware patch peripheral at 0x03802000 (R* = 4 slots). Patch writes currently rejected (0x0a f3) — may need different `changeMemoryState` format or the write command format needs adjustment (checksum vs no checksum). Node.js bridge remains as fallback.

- [x] ~~Stop batch after current track~~
- [x] ~~Auto-set disc title from folder name~~ — verified working
- [ ] **LP4 support** — Quarter capacity vs SP, if atracdenc supports ATRAC3 LP4 encoding
- [x] ~~MiniDisc logo in UI~~ — dithered half-block character rendering from PNG

## Completed

- [x] ~~LP2 upload support via atracdenc~~
- [x] ~~File browser in upload dialog~~
- [x] ~~Batch upload from folder~~
- [x] ~~Track download (Node.js bridge)~~
- [x] ~~Show device model name instead of "NetMD Device 0"~~
- [x] ~~Responsive layout improvements for small terminals~~
- [x] ~~Error banner auto-dismiss after timeout~~
