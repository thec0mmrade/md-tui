# Firmware Dump Research — MZ-N505

## Goal

Extract the ARM firmware code from the Sony MZ-N505 for reverse engineering (disassembly, finding new exploit entry points, supporting more devices).

## What We Tried

### 1. Factory Read Command (18 24) — Partial Success

The `18 24 ff` factory command reads 16 bytes from a 16-bit address space. We dumped the full 0x0000-0xFFFF range (64KB).

**Result**: The address space is a ~83-byte circular status buffer, not actual firmware code or registers. Contains disc controller state (sector sync patterns, track pointers, calibration counters). Useful for Sony service diagnostics, not firmware extraction.

### 2. Factory Memory Read (1820/1821) — Wrong Address Space

Used `changeMemState` (1820) + `read` (1821) with 32-bit addresses to read from the ARM code space (0x00050000+).

**Result**: All zeros. These commands can access peripherals (0x03800000+) and DRAM (0x02000000+) but NOT the firmware code space. The code ROM/flash is mapped differently for data access.

### 3. Custom ARM Code via 18 d3 — Response Size Limitation

Wrote ARM code (memcpy from source address to USB response buffer) and sent via the `18 d3` exploit command.

**Result**: ARM code executes but the `18 d3` handler returns exactly 88 bytes (the echoed payload) regardless of what the ARM code writes. Only `read_atrac_dram()` triggers the firmware's internal response size extension mechanism. Custom ARM code can write to the buffer, but the USB handler doesn't include the extra bytes in the response.

## Why Sector Reads Work But Memory Reads Don't

The `read_atrac_dram()` firmware function has a side effect that updates the USB response size. When the sector read ARM code calls this function, the firmware knows to include the extra data bytes in the USB response. A simple `memcpy` loop doesn't trigger this mechanism — the firmware's USB handler returns only the echoed payload (88 bytes).

## Viable Approach: Modified CachedSectorControlDownload

The `CachedSectorControlDownload` exploit (from netmd-exploits) works around the response size limitation by patching the USB **read handler** itself, not the `18 d3` handler:

1. **Loads persistent ARM code** into device RAM at `0x02003ce0`
2. **Patches two USB read handler addresses** (`0x000574fc`, `0x00057500`) with jump instructions to the resident code
3. **The resident code intercepts `readReply()` calls** and serves data from its own buffer
4. **Response size is controlled by the resident code**, not the `18 d3` handler

### Modified Version for Firmware Dump

Replace the resident code's `read_atrac_dram()` call with `memcpy()`:

```
Original:  read_atrac_dram(sector, 0, sectorBuffer, 2352)
Modified:  memcpy(sectorBuffer, firmwareBaseAddr + sector * 2352, 2352)
```

Each `readReply(2352)` would then return 2352 bytes of firmware from sequential addresses. Reading ~150 sectors covers ~350KB of firmware code.

### Implementation Steps

1. **Force CachedSectorControlDownload** in `scripts/download.mjs`:
   ```js
   import { CachedSectorControlDownload } from 'netmd-exploits';
   const exploit = await stateManager.require(CachedSectorControlDownload);
   ```
   Test if the loader/patcher even works on MZ-N505 (R1.400).

2. **If it works**: Modify the resident ARM code's read function to copy from a memory address instead of calling `read_atrac_dram`.

3. **Create a JS script** (`scripts/firmware-dump.mjs`) that:
   - Initializes the exploit
   - Loads modified resident code
   - Calls `readReply(2352)` in a loop to read sequential firmware blocks
   - Writes concatenated output to a binary file

### Device-Specific Constants (MZ-N505 R1.400)

From netmd-exploits source:
```
residentCodeAddress:        0x02003ce0
enabledFlagAddress:         0x02003cd4
sectorToReadAddress:        0x02003cd0
sectorBuffer:               0x03240
usbReadStandardResponse:    0x000574fc
usbReadStandardResponseNext: 0x00057500
handlerPatchValue:          0x47104a00  (THUMB: ldr r2,[pc]; bx r2)
read_atrac_dram:            0x0005edf1
onePatchAddress:            0x00057be8
```

### Risk Assessment

- **Low risk**: Only reads memory, never writes to flash
- **Worst case**: Device hang requiring USB replug (no permanent damage)
- **Firmware patches are volatile** (DRAM) — cleared on power cycle

## Status

Blocked until CachedSectorControlDownload is tested on MZ-N505. The `getBestSuited()` function returns CachedSectorNoRamControlDownload for this device — the Control variant needs to be forced manually.
