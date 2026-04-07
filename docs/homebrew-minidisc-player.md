# Homebrew MiniDisc Reader/Writer

Research document for building an open-source MiniDisc reader/writer using a Heltec V3 (ESP32-S3) development board, common components, and salvaged MiniDisc hardware.

## Architecture Overview

```
┌─────────────────────────────────────────────────────┐
│              Salvaged MiniDisc Mechanism             │
│  ┌──────────┐  ┌──────────┐  ┌──────────────────┐  │
│  │ Spindle  │  │  Sled    │  │ Optical Pickup   │  │
│  │  Motor   │  │  Motor   │  │ (KMS-260A etc.)  │  │
│  └────┬─────┘  └────┬─────┘  └────┬─────────────┘  │
│       │              │             │                 │
│  ┌────┴──────────────┴─────────────┴──────────────┐ │
│  │     Signal Processing IC (CXD2677 / CXD2662)   │ │
│  │  RF amp → EFM demod → ACIRC ECC → ATRAC data  │ │
│  └────────────────────┬───────────────────────────┘ │
└───────────────────────┼─────────────────────────────┘
                        │ decoded ATRAC sectors
                        │
┌───────────────────────┼─────────────────────────────┐
│              Heltec V3 (ESP32-S3)                   │
│                       │                              │
│  ┌────────────────────┴──────────────────────────┐  │
│  │              System Controller                 │  │
│  │  • Motor control (GPIO + H-bridge drivers)    │  │
│  │  • UTOC parsing (track table of contents)     │  │
│  │  • ATRAC1/3 decode (fixed-point)              │  │
│  │  • Track/file management                      │  │
│  │  • USB/WiFi host communication                │  │
│  └──────┬────────────────────────┬───────────────┘  │
│         │                        │                   │
│    ┌────┴─────┐            ┌─────┴──────┐           │
│    │  OLED    │            │ I2S → DAC  │           │
│    │ Display  │            │ (PCM5102A) │           │
│    └──────────┘            └─────┬──────┘           │
└──────────────────────────────────┼───────────────────┘
                                   │
                              Audio Output
```

**Critical design decision**: Keep the salvaged signal processing IC (CXD2677 or similar). The RF → EFM → ACIRC → sector pipeline is extremely well-tuned analog/digital mixed-signal processing that took Sony years to optimize. Replace only the system controller (ESP32-S3 instead of the original MCU).

## MiniDisc Optical System

### Disc Types
- **Pre-mastered (MD-ROM)**: Physical pits like CD. Read-only. Photodetector reads intensity changes from pit/land transitions.
- **Recordable (MD-MO)**: Magneto-optical. Recording layer is terbium-iron-cobalt (TbFeCo) amorphous alloy. Data stored as magnetic polarization direction, read via the Kerr effect (polarized laser light reflected from the MO layer has its polarization plane rotated ~0.5-1 degree depending on magnetization direction).

### Optical Specifications
| Parameter | Value |
|-----------|-------|
| Laser wavelength | 780 nm (near-infrared) |
| Track pitch | 1.6 µm |
| Disc speed | CLV at 1.2 m/s |
| Disc diameter | 64 mm (in 68×72mm cartridge) |
| Spindle speed | ~400-900 RPM (varies with radius) |
| Channel bit rate | 4.3218 MHz (1.4112 Mbit/s) |
| User data rate | ~292 kbit/s (SP), ~132 kbit/s (LP2), ~66 kbit/s (LP4) |
| Read laser power | ~0.5 mW |
| Write laser power | ~4-6 mW |
| Minimum pit/land | 3T (694 ns, ~0.83 µm) |
| Maximum pit/land | 11T (2546 ns, ~3.05 µm) |

### Kerr Effect Detection
The reflected beam passes through a Wollaston prism (polarizing beam splitter) that splits it into two orthogonal polarization components. Two photodiodes measure these — the difference signal gives the data, the sum gives total reflected intensity for servo signals. The Kerr rotation is very small, so the RF amplifier stage is critical (~60-80dB gain needed).

## Signal Processing Chain

```
Optical Pickup (photodetector)
    │
    ├── RF signal (HF): (A+C)-(B+D) or Kerr differential
    │   Amplitude: ~100-500 µV, bandwidth: DC to ~5 MHz
    │       │
    │       ▼
    │   RF Amplifier (~60-80dB gain)
    │       │
    │       ▼
    │   Equalizer (boost HF for optical MTF compensation)
    │       │
    │       ▼
    │   Comparator/Slicer (asymmetry correction)
    │       │
    │       ▼
    │   PLL (locks to 4.3218 MHz channel bit rate)
    │       │
    │       ▼
    │   EFM Decoder (14-bit → 8-bit lookup table)
    │       │
    │       ▼
    │   ACIRC Error Correction
    │   ├── C1: RS(32,28) — corrects 2 byte errors/block
    │   └── C2: RS(28,24) — corrects 2 byte errors/block
    │       │   Interleave: 108 frames (burst correction ~2.4mm)
    │       ▼
    │   2352-byte ATRAC sectors
    │
    ├── Focus Error (FE): astigmatic method, ~10-50mV
    │   → Focus servo loop (~1-5 kHz bandwidth)
    │   → Focus voice coil actuator
    │
    └── Tracking Error (TE): push-pull method, ~5-20mV
        → Tracking servo loop (~500 Hz-1 kHz bandwidth)
        → Tracking voice coil actuator + sled motor
```

### EFM (Eight-to-Fourteen Modulation)
Identical to CD. Each 8-bit byte maps to a 14-bit channel pattern, with 3 merge bits between symbols (17 channel bits per byte). Sync patterns (unique 24+3 bit sequences) provide frame synchronization. One frame = 1 sync + 1 subcode + 24 data + 8 parity = 33 bytes. 98 frames = 1 sector.

### ACIRC Error Correction
Enhanced version of CD's CIRC. Longer interleave depth (108 frames vs 28) gives better burst error correction — essential because MO recording has higher raw error rates than pressed CDs.

## MiniDisc Disc Format

### Structure
| Unit | Size | Contents |
|------|------|----------|
| Sector | 2,352 bytes | 11 sound groups (SP) |
| Sound group | 212 bytes | 12-byte header + 200 bytes ATRAC data |
| Cluster | 36 sectors | 32 data + 4 link sectors |
| 80-min disc | ~86,400 sectors | ~193 MB raw capacity |

### UTOC (User Table of Contents)
Stored in sectors 0-2 of the lead-in area (magneto-optically recorded even on pre-mastered discs).

- **Sector 0**: Track start/end addresses (cluster:sector:soundgroup format), track mode (encoding type, stereo/mono). Supports up to 255 tracks with non-contiguous fragments (linked-list structure).
- **Sector 1**: Track titles (ASCII, ~1,700 characters total).
- **Sector 2**: Timestamps (recording date/time per track).

### ADIP (Address In Pregroove)
The recording groove wobbles at 22.05 kHz with frequency-modulated address data. Provides absolute disc position for CLV speed reference during seek, write position accuracy, and UTOC address references.

## Motor Control and Servo Systems

### Spindle Motor (CLV)
- Maintains constant linear velocity of 1.2 m/s
- Speed varies: ~510 RPM (inner) to ~228 RPM (outer)
- CLV control: EFM PLL frequency error drives spindle speed correction
- During seek: ADIP wobble frequency (22.05 kHz) used as speed reference
- Typical motor: 3-phase brushless DC

### Sled Motor
- Moves optical pickup across disc radius
- Coarse positioning via DC motor with worm gear
- Fine tracking via voice coil actuator on pickup lens assembly (~±300 µm range)
- Sled engages when fine tracking actuator reaches range limit

### Focus Servo
- Bandwidth: ~1-5 kHz
- Astigmatic focus error signal (S-curve, linear region ~±1 µm)
- Must handle disc runout up to ~0.3mm at rotation frequency
- Initial focus: ramp actuator until S-curve zero crossing detected, then close loop

### Tracking Servo
- Bandwidth: ~500 Hz-1 kHz
- Push-pull method for recordable discs, three-beam for pre-mastered
- Track jump: open loop, apply kick pulse, re-acquire

## ATRAC Codec

### ATRAC1 (SP Mode)
- Bitrate: 292 kbit/s (fixed)
- Transform: Hybrid subband/MDCT — 3 subbands via QMF filter, each transformed with MDCT (128 or 32 sample windows)
- Psychoacoustic model: simultaneous and temporal masking
- Sound group: 424 bytes = 512 PCM samples (11.6ms at 44.1kHz)

### ATRAC3 (LP2/LP4)
- LP2: 132 kbit/s, 192-byte frames
- LP4: 66 kbit/s, 96-byte frames
- Additional features: joint stereo, gain control, tonal component extraction

### Open-Source Implementations
- **ffmpeg**: ATRAC1 and ATRAC3 decoders (floating-point)
- **atracdenc**: ATRAC1/3 encoder and ATRAC1 decoder (used in our Go TUI for LP2/LP4 uploads)
- For ESP32: would need fixed-point port of the decoder (~50-100 MIPS for real-time ATRAC1)

## MO Recording (Writing)

### Magneto-Optical Write Process
1. Laser heats spot on MO layer past Curie temperature (~180-200°C for TbFeCo)
2. Material's coercivity drops to near zero
3. External magnetic field (from head on opposite side of disc) sets magnetization direction
4. Spot cools below Curie temperature, magnetization is frozen

### Magnetic Field Modulation
MiniDisc uses magnetic field modulation (MFM): laser power is constant during writing, data is encoded in the magnetic head's polarity switching at the channel bit rate (~1.4 MHz). This allows direct overwrite without a separate erase pass.

### Magnetic Head Requirements
- Field strength: ~80-300 Oe
- Modulation rate: ~1.4 MHz (channel bit rate)
- Head inductance: ~1-5 µH (must be low for MHz switching)
- Driver: current-mode H-bridge with fast switching transistors

## Optical Pickup Signals

Typical MiniDisc optical pickup (KMS-260A or similar):

| Signal | Description | Level | Amplification |
|--------|-------------|-------|---------------|
| RF+/RF- | Differential data (Kerr effect) | ~100-500 µV | 60-80 dB (dedicated RF amp IC) |
| FE | Focus error (astigmatic) | ~10-50 mV | ~40 dB (op-amp) |
| TE | Tracking error (push-pull) | ~5-20 mV | ~40 dB (op-amp) |
| LD+/LD- | Laser diode drive | — | Constant current driver |
| PD | Monitor photodiode (APC) | µA | Transimpedance amp |
| FCA/FCB | Focus coil | — | H-bridge, ~50-200 mA |
| TCA/TCB | Tracking coil | — | H-bridge, ~50-200 mA |

## ESP32-S3 (Heltec V3) Capabilities

| Feature | Spec | Suitability |
|---------|------|-------------|
| CPU | Dual Xtensa LX7, 240 MHz | Sufficient for ATRAC decode + system control |
| SRAM | 512 KB | Enough for buffers and stack |
| PSRAM | 8 MB (with PSRAM module) | Plenty for sector/audio buffering |
| ADC | 12-bit, ~83 kSPS | **Too slow for RF** — need external ADC or dedicated IC |
| I2S | 2 interfaces, up to 24-bit | Perfect for audio DAC output |
| GPIO | 45 pins | Sufficient for all control signals |
| SPI | Up to 80 MHz | Fast enough for external ADC/FPGA interface |
| WiFi | 802.11 b/g/n | Wireless track management |
| Bluetooth | BLE 5.0 | Wireless audio streaming |
| OLED | 128×64 built-in (Heltec V3) | Track info display |
| LoRa | Built-in (Heltec V3) | Not needed, but available |

**Verdict**: ESP32-S3 is suitable as the system controller but **cannot handle raw optical signal processing** (EFM demod needs ~10 MSPS ADC, ACIRC needs dedicated RS decoder). Keep the salvaged signal processing IC for that.

## Bill of Materials

### Salvaged from Dead MiniDisc Player
| Part | Source | Notes |
|------|--------|-------|
| Complete optical mechanism | MDS-JE330/530 (desktop, easier to work with) or any portable | Includes spindle motor, sled, pickup, magnetic head |
| Signal processing IC | Same unit (CXD2662 for desktop, CXD2677 for portable) | Handles RF → EFM → ACIRC → sector data |
| RF amplifier IC | Same unit (CXA1081, BA7765, or integrated) | May be part of the signal processing IC |

### New Components
| Part | Example | Cost (approx) | Purpose |
|------|---------|---------------|---------|
| Heltec V3 | ESP32-S3 + OLED + LoRa | $15-20 | System controller |
| I2S DAC | PCM5102A breakout | $2-5 | Audio output |
| Motor driver (spindle) | DRV8313 or TB6612FNG | $3-5 | 3-phase BLDC / DC motor |
| Motor driver (sled) | TB6612FNG or DRV8871 | $2-3 | Sled H-bridge |
| Laser driver | iC-HG or ADN2830 | $5-10 | Constant current + APC |
| Op-amps | OPA2134 or NE5532 (×2) | $2-4 | Servo signal conditioning |
| Power supply | 3.3V + 5V regulator | $2-3 | System power |
| Connectors, PCB, passives | Various | $10-20 | Assembly |
| **Total (new parts)** | | **~$40-70** | |

### For Writing (Additional)
| Part | Example | Cost | Purpose |
|------|---------|------|---------|
| Magnetic head driver | Discrete transistor H-bridge | $5 | MHz-rate field modulation |
| Write laser power control | Enhanced APC circuit | $5 | Higher power for write mode |

## Key Technical Challenges

### 1. Signal Processing IC Interface (HARDEST)
The salvaged CXD2677/CXD2662 communicates with the system controller via a proprietary bus protocol. Understanding this protocol is essential. Our firmware dump (448KB ROM from MZ-N505) contains the code that drives this interface — reverse engineering it would reveal the command set.

**Approach**: Analyze the firmware dump to find the servo/signal-processing control functions. Look for sequences of register writes to I/O port addresses. The CXD2677 likely uses memory-mapped I/O in the 0x03000000+ address range (peripheral space).

### 2. Servo Tuning
Focus and tracking servos must be tuned for the specific mechanism. Original players have factory-calibrated servo parameters in EEPROM. Getting stable focus lock and tracking on a homebrew controller will require:
- Implementing PID loops with adjustable gains
- Focus search algorithm (ramp + zero-crossing detection)
- Anti-shock / error recovery logic

### 3. ATRAC Real-Time Decode on ESP32
The ESP32-S3 can likely handle ATRAC1 decode in real-time, but it needs:
- Fixed-point port of the decoder (no hardware FPU for double-precision)
- Optimized MDCT using lookup tables
- Double-buffered audio output via I2S DMA
- Estimated ~50-100 MIPS, ESP32-S3 has ~400-600 MIPS available

### 4. UTOC Read/Write
The UTOC must be parsed to find track locations and written to update after recording. The format is documented but intricate — fragment linked lists, address encoding, title storage.

### 5. CLV Speed Control
Maintaining constant linear velocity requires a control loop that adjusts spindle speed based on EFM PLL error or ADIP wobble frequency. The signal processing IC may handle this internally, or the system controller may need to participate.

### 6. MO Write Timing
For recording, the magnetic head must switch polarity in sync with the laser position. The timing relationship between the write clock, magnetic field, and disc position must be precise to within ~100ns.

## Development Roadmap

### Phase 1: Read-Only Player (3-6 months)
1. Salvage mechanism from desktop MiniDisc unit (MDS-JE330 recommended — larger, easier to probe)
2. Identify the signal processing IC's data output bus (logic analyzer on the bus between signal IC and original MCU)
3. Wire ESP32-S3 to the data bus
4. Implement basic spindle/sled motor control
5. Read raw ATRAC sectors from the signal processing IC
6. Implement UTOC parser to find tracks
7. Port ATRAC1 decoder to ESP32 (fixed-point)
8. Audio output via I2S + PCM5102A
9. Basic OLED UI (track number, title, time)

### Phase 2: Full Player Features (3-6 months)
1. Track navigation (skip, search, program play)
2. LP2/LP4 ATRAC3 decode
3. Anti-shock buffer (read-ahead into PSRAM)
4. WiFi-based track management (web UI)
5. Battery power management
6. Compact PCB design

### Phase 3: Recording (6-12 months)
1. Reverse-engineer write mode protocol from firmware dump
2. Implement magnetic head driver
3. Implement laser power switching (read → write)
4. Write ATRAC sectors to disc
5. UTOC update after recording
6. USB audio input → ATRAC encode → write pipeline

### Phase 4: Custom Hardware (ongoing)
1. Design a custom PCB combining ESP32-S3 + drivers + DAC
2. 3D-printed enclosure
3. Open-source hardware release (KiCad schematics + Gerbers)
4. Community documentation and build guide

## How Our Firmware Dump Helps

The 448KB ROM dump from the MZ-N505 contains all the code that controls the CXD2677 signal processing IC. By reverse-engineering specific sections:

- **Servo initialization** (around reset vector 0x3c): how the firmware configures focus/tracking/spindle parameters
- **I/O port writes** (0x03000000+ peripheral space): the command protocol for the signal processing subsystem
- **UTOC read/write** (near the disc access functions): exact sector addresses and data format
- **EEPROM access** (registers 0x61-0x63): servo calibration values
- **Error recovery**: how the firmware handles disc read errors, loss of focus, track loss

This is the single most valuable resource for the homebrew project — it's a complete reference implementation of how to control MiniDisc hardware.

## Existing Resources

### Software
- [linux-minidisc](https://github.com/glaubitz/linux-minidisc) — Python NetMD USB tools
- [netmd-js](https://github.com/niclas-porzin/netmd-js) — JavaScript NetMD implementation
- [netmd-exploits](https://github.com/asivery/netmd-exploits) — Firmware exploits (ARM code execution)
- [atracdenc](https://github.com/dcherednik/atracdenc) — Open-source ATRAC encoder/decoder
- [ffmpeg](https://ffmpeg.org/) — ATRAC1/3 decoder (reference for ESP32 port)
- [md-tui](https://github.com/thec0mmrade/md-tui) — This project: Go TUI with native exploit download, firmware dump, file storage

### Hardware References
- [MiniDisc.org Technical Pages](https://www.minidisc.org/tech_specs.html) — Disc format documentation
- [MiniDisc Wiki](https://www.minidisc.wiki/) — Community knowledge base
- [CD/MiniDisc Service Manuals](https://www.minidisc.org/manuals/) — Schematics for various models

### No Known Homebrew MiniDisc Hardware Projects
This would be a genuinely novel open-source hardware project. CD player homebrew projects exist and share some fundamentals (EFM, servo systems) but none target MiniDisc specifically.
