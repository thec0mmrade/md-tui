# TODO

## Features

- [x] ~~LP2 upload support via atracdenc~~
- [x] ~~File browser in upload dialog~~
- [x] ~~Batch upload from folder~~
- [ ] **Track download (native exploit)** — Native Go exploit infrastructure built (factory mode, firmware read, DRAM patching all work) but sector reads return empty data. Need to deep-read `netmd-exploits` TypeScript source to understand what ARM code expects from DRAM patch addresses and whether patch values are firmware-version-specific. pcaps at `/tmp/netmd-download2.pcap` and `/tmp/netmd-full-init.pcap`.
- [x] ~~Track download (Node.js bridge)~~ — Downloads tracks via Node.js helper using netmd-exploits. Outputs ATRAC3 WAV, convertible to PCM via ffmpeg.

- [ ] **Stop batch after current track** — Press Esc during batch upload to stop after the current track finishes, keeping tracks already written
- [ ] **Group management** — Create, rename, and delete track groups (protocol library has Root/Group support)
- [ ] **Auto-set disc title from folder name** — When batch uploading, offer to set the disc title to the folder name
- [ ] **LP4 support** — Quarter capacity vs SP, if atracdenc supports ATRAC3 LP4 encoding

## Completed

- [x] ~~Show device model name instead of "NetMD Device 0"~~
- [x] ~~Responsive layout improvements for small terminals~~
- [x] ~~Error banner auto-dismiss after timeout~~
