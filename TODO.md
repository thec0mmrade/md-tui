# TODO

## Features

- [x] ~~LP2 upload support via atracdenc~~
- [x] ~~File browser in upload dialog~~
- [x] ~~Batch upload from folder~~
- [ ] **Track download/ripping** — Phases 1-3 done: playback control, factory mode, firmware read, DRAM patching (37 commands) all work. Phase 4 blocked: sector read command `18 d3 ff` executes but returns 88 bytes (header echo only, no audio data). ARM code runs but doesn't populate output buffer. May need exact seek/stop sequencing from pcap, or sector parameters are subtly wrong. pcaps at `/tmp/netmd-download2.pcap` and `/tmp/netmd-full-init.pcap`.

- [ ] **Stop batch after current track** — Press Esc during batch upload to stop after the current track finishes, keeping tracks already written
- [ ] **Group management** — Create, rename, and delete track groups (protocol library has Root/Group support)
- [ ] **Auto-set disc title from folder name** — When batch uploading, offer to set the disc title to the folder name
- [ ] **LP4 support** — Quarter capacity vs SP, if atracdenc supports ATRAC3 LP4 encoding

## Completed

- [x] ~~Show device model name instead of "NetMD Device 0"~~
- [x] ~~Responsive layout improvements for small terminals~~
- [x] ~~Error banner auto-dismiss after timeout~~
