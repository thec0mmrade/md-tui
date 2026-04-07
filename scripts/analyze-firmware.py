#!/usr/bin/env python3
"""
MZ-N505 Firmware Analyzer
Analyzes the R1.400 firmware ROM dump from the Sony MZ-N505 (CXD2677, ARM7TDMI).
Produces a comprehensive report for reverse engineering.

Usage: python3 scripts/analyze-firmware.py firmware_go.bin [firmware_go.sram]
"""

import sys
import struct
from collections import defaultdict

try:
    from capstone import Cs, CS_ARCH_ARM, CS_MODE_ARM, CS_MODE_THUMB, CS_MODE_LITTLE_ENDIAN
except ImportError:
    print("ERROR: capstone required. Install: pip3 install capstone")
    sys.exit(1)

# Known addresses from netmd-exploits and our exploit work
KNOWN_FUNCTIONS = {
    0x00057be8: ("onePatchAddress", "USB code execution hook point (THUMB)"),
    0x000574fc: ("usbReadStandardResponse", "USB read handler 1 — CachedSectorControlDownload target"),
    0x00057500: ("usbReadStandardResponseNext", "USB read handler 2 — CachedSectorControlDownload target"),
    0x0005edf1: ("read_atrac_dram", "Read ATRAC sector from disc cache (THUMB)"),
    0x0005a8b1: ("usb_do_response", "USB response function (THUMB)"),
    0x0005dce5: ("tron_set_flg", "TRON OS: set flag (THUMB)"),
    0x0005de51: ("tron_clr_flg", "TRON OS: clear flag (THUMB)"),
    0x0005da15: ("tron_twai_flg", "TRON OS: wait for flag (THUMB)"),
    0x00060355: ("ChangeIRQmask", "Change IRQ mask (THUMB)"),
}

KNOWN_DATA = {
    0x02000b38: ("g_DiscStateStruct", "Disc state structure in SRAM"),
    0x02004110: ("g_usb_buff", "USB command buffer in SRAM"),
    0x02003ce0: ("residentCodeAddr", "CachedSectorControlDownload resident code target"),
    0x02003cd4: ("enabledFlagAddr", "Sector read enabled flag"),
    0x02003cd0: ("sectorToReadAddr", "Current sector number"),
    0x03802000: ("patchPeripheralBase", "Hardware patch peripheral"),
    0x03802040: ("patchControlReg", "Patch control register"),
}

ROM_SIZE = 0x70000  # 448KB
SRAM_BASE = 0x02000000
SRAM_SIZE = 0x4800  # 18KB


def load_rom(path):
    with open(path, "rb") as f:
        data = f.read()
    if len(data) != ROM_SIZE:
        print(f"WARNING: Expected {ROM_SIZE} bytes, got {len(data)}")
    return data


def load_sram(path):
    with open(path, "rb") as f:
        return f.read()


# ─── Phase 1: Basic Structure ───────────────────────────────────────

def analyze_vectors(rom):
    """Parse the ARM exception vector table."""
    print("=" * 70)
    print("PHASE 1: BASIC STRUCTURE")
    print("=" * 70)

    print("\n### ARM Exception Vector Table")
    vector_names = [
        "Reset", "Undefined Instruction", "Software Interrupt (SWI)",
        "Prefetch Abort", "Data Abort", "Reserved", "IRQ", "FIQ"
    ]

    md = Cs(CS_ARCH_ARM, CS_MODE_ARM | CS_MODE_LITTLE_ENDIAN)
    for i, name in enumerate(vector_names):
        addr = i * 4
        insn_bytes = rom[addr:addr + 4]
        insns = list(md.disasm(insn_bytes, addr))
        if insns:
            ins = insns[0]
            print(f"  0x{addr:08x}: {ins.mnemonic:8s} {ins.op_str:30s} ; {name}")
        else:
            print(f"  0x{addr:08x}: {insn_bytes.hex():20s} ; {name} (invalid)")

    # Read vector targets (LDR pc, [pc, #offset] loads from address table)
    print("\n### Vector Targets")
    for i in range(8):
        addr = i * 4
        word = struct.unpack_from("<I", rom, addr)[0]
        # LDR pc, [pc, #imm] = 0xe59ff000 + imm
        if (word & 0xfffff000) == 0xe59ff000:
            offset = word & 0xfff
            target_addr = addr + 8 + offset  # PC + 8 + offset
            if target_addr + 4 <= len(rom):
                target = struct.unpack_from("<I", rom, target_addr)[0]
                print(f"  {vector_names[i]:30s} → 0x{target:08x} (via [0x{target_addr:08x}])")


def find_strings(rom, min_len=6):
    """Find ASCII strings in the ROM."""
    print("\n### ASCII Strings (length >= %d)" % min_len)
    strings = []
    current = []
    start = 0
    for i, b in enumerate(rom):
        if 0x20 <= b <= 0x7e:
            if not current:
                start = i
            current.append(chr(b))
        else:
            if len(current) >= min_len:
                strings.append((start, "".join(current)))
            current = []
    if len(current) >= min_len:
        strings.append((start, "".join(current)))

    for offset, s in strings[:100]:  # limit output
        print(f"  0x{offset:06x}: \"{s}\"")
    if len(strings) > 100:
        print(f"  ... ({len(strings) - 100} more strings)")
    print(f"  Total: {len(strings)} strings found")
    return strings


def find_functions(rom):
    """Find THUMB function prologues (PUSH {... lr})."""
    print("\n### Function Prologues (THUMB PUSH {... lr})")
    functions = []
    for i in range(0, len(rom) - 2, 2):
        hw = struct.unpack_from("<H", rom, i)[0]
        # THUMB PUSH with LR: 1011 0101 xxxx xxxx (0xb5xx)
        if (hw & 0xff00) == 0xb500:
            functions.append(i)

    print(f"  Found {len(functions)} THUMB function prologues")
    # Show first/last few and any near known addresses
    if functions:
        print(f"  First: 0x{functions[0]:06x}")
        print(f"  Last:  0x{functions[-1]:06x}")

    return functions


def find_code_regions(rom):
    """Identify code vs data regions."""
    print("\n### Code/Data Regions")
    region_size = 0x1000  # 4KB blocks
    regions = []
    for offset in range(0, len(rom), region_size):
        block = rom[offset:offset + region_size]
        # Count THUMB-like halfwords (common instruction patterns)
        thumb_count = 0
        zero_count = 0
        for j in range(0, len(block) - 1, 2):
            hw = struct.unpack_from("<H", block, j)[0]
            if hw == 0:
                zero_count += 1
            # Common THUMB patterns: PUSH, POP, BL, MOV, LDR, STR, ADD, SUB, CMP, B
            elif (hw & 0xff00) in (0xb500, 0xbd00, 0xf000, 0xf800, 0x4600,
                                     0x4800, 0x6800, 0x6000, 0x3000, 0x2000,
                                     0x2800, 0xe000, 0xd000):
                thumb_count += 1
        total = region_size // 2
        if zero_count > total * 0.9:
            rtype = "EMPTY"
        elif thumb_count > total * 0.15:
            rtype = "CODE"
        else:
            rtype = "DATA"
        regions.append((offset, rtype, thumb_count, zero_count))

    # Summarize
    current_type = None
    current_start = 0
    for offset, rtype, tc, zc in regions:
        if rtype != current_type:
            if current_type is not None:
                print(f"  0x{current_start:06x}-0x{offset:06x}: {current_type}")
            current_type = rtype
            current_start = offset
    if current_type:
        print(f"  0x{current_start:06x}-0x{len(rom):06x}: {current_type}")

    return regions


# ─── Phase 2: Known Address Analysis ────────────────────────────────

def disassemble_at(rom, addr, count=60, thumb=True):
    """Disassemble instructions at a given ROM offset."""
    if addr >= len(rom):
        return []

    mode = CS_MODE_THUMB if thumb else CS_MODE_ARM
    md = Cs(CS_ARCH_ARM, mode | CS_MODE_LITTLE_ENDIAN)
    md.detail = False

    # Read enough bytes
    end = min(addr + count * 4, len(rom))
    code = rom[addr:end]

    instructions = []
    for ins in md.disasm(code, addr):
        instructions.append(ins)
        if len(instructions) >= count:
            break
    return instructions


def analyze_known_functions(rom):
    """Disassemble and analyze all known function addresses."""
    print("\n" + "=" * 70)
    print("PHASE 2: KNOWN ADDRESS ANALYSIS")
    print("=" * 70)

    for addr, (name, desc) in sorted(KNOWN_FUNCTIONS.items()):
        if addr >= len(rom):
            print(f"\n### 0x{addr:08x}: {name} — OUTSIDE ROM")
            continue

        # Determine if THUMB (odd address = THUMB)
        thumb = (addr & 1) == 1
        real_addr = addr & ~1  # Clear THUMB bit

        print(f"\n### 0x{addr:08x}: {name}")
        print(f"    {desc}")
        print(f"    Mode: {'THUMB' if thumb else 'ARM'}")
        print()

        insns = disassemble_at(rom, real_addr, count=40, thumb=thumb)
        for ins in insns:
            # Mark known addresses in operands
            annotation = ""
            for known_addr, (known_name, _) in {**KNOWN_FUNCTIONS, **KNOWN_DATA}.items():
                hex_str = f"0x{known_addr:x}"
                if hex_str in ins.op_str or hex_str.replace("0x", "#0x") in ins.op_str:
                    annotation = f"  ; → {known_name}"
            print(f"    0x{ins.address:08x}: {ins.mnemonic:8s} {ins.op_str}{annotation}")


# ─── Phase 3: USB Command Handler Mapping ───────────────────────────

def find_usb_handlers(rom):
    """Search for USB command dispatch patterns."""
    print("\n" + "=" * 70)
    print("PHASE 3: USB COMMAND HANDLER MAPPING")
    print("=" * 70)

    # Search for common USB command bytes in immediate values
    # Commands: 0x18xx where xx = 09, 12, 20, 21, 22, 24, d3, c3, c5, 50
    usb_commands = {
        0x09: "status/query",
        0x12: "factory enter",
        0x20: "changeMemoryState",
        0x21: "factory read",
        0x22: "factory write",
        0x24: "firmware read",
        0xd3: "code execution (R-series)",
        0xd2: "code execution (S-series)",
        0xc3: "play/pause",
        0xc5: "stop",
        0x50: "goto/seek",
    }

    print("\n### References to USB command bytes")
    # Search for these bytes in the ROM as immediates or data
    for cmd_byte, cmd_name in usb_commands.items():
        # Search for 0x18 followed by cmd_byte nearby
        refs = []
        for i in range(len(rom) - 1):
            if rom[i] == 0x18 and rom[i + 1] == cmd_byte:
                # Check if it's in a plausible code/data context
                refs.append(i)
        if refs:
            shown = refs[:5]
            extra = f" (+{len(refs) - 5} more)" if len(refs) > 5 else ""
            addrs = ", ".join(f"0x{r:06x}" for r in shown)
            print(f"  18 {cmd_byte:02x} ({cmd_name:25s}): {addrs}{extra}")

    # Look for dispatch table patterns near the patch address
    print("\n### Disassembly around USB handler area (0x574f0-0x57c00)")
    insns = disassemble_at(rom, 0x574f0, count=80, thumb=True)
    for ins in insns:
        annotation = ""
        if ins.address in KNOWN_FUNCTIONS:
            annotation = f"  ; ← {KNOWN_FUNCTIONS[ins.address][0]}"
        elif ins.address + 1 in KNOWN_FUNCTIONS:
            annotation = f"  ; ← {KNOWN_FUNCTIONS[ins.address + 1][0]}"
        print(f"    0x{ins.address:08x}: {ins.mnemonic:8s} {ins.op_str}{annotation}")


# ─── Phase 4: EEPROM Access ─────────────────────────────────────────

def find_eeprom_access(rom):
    """Search for EEPROM register access patterns."""
    print("\n" + "=" * 70)
    print("PHASE 4: EEPROM ACCESS")
    print("=" * 70)

    # Known EEPROM registers: 0x61, 0x62, 0x63
    # Search for these values as immediates in CMP or MOV instructions
    print("\n### References to EEPROM register numbers (0x61, 0x62, 0x63)")
    for reg_num in [0x61, 0x62, 0x63]:
        refs = []
        for i in range(0, len(rom) - 1, 2):
            hw = struct.unpack_from("<H", rom, i)[0]
            # THUMB: CMP Rn, #imm8 = 0010 1nnn iiiiiiii
            if (hw >> 8) == (0x28 | (hw >> 8 & 7)) and (hw & 0xff) == reg_num:
                refs.append(i)
            # THUMB: MOV Rd, #imm8 = 0010 0ddd iiiiiiii
            if (hw >> 11) == 0x4 and (hw & 0xff) == reg_num:
                refs.append(i)
        if refs:
            addrs = ", ".join(f"0x{r:06x}" for r in refs[:10])
            print(f"  Register 0x{reg_num:02x}: found at {addrs}")

    # Search for EEPROM-related strings
    print("\n### EEPROM-related strings")
    eeprom_strings = []
    for i in range(len(rom) - 5):
        # Common EEPROM-related strings
        chunk = rom[i:i + 6]
        try:
            s = chunk.decode("ascii")
            if "eep" in s.lower() or "rom" in s.lower() or "nvm" in s.lower():
                eeprom_strings.append((i, s))
        except (UnicodeDecodeError, ValueError):
            pass


# ─── Phase 5: SRAM Analysis ─────────────────────────────────────────

def analyze_sram(sram):
    """Analyze the SRAM dump."""
    print("\n" + "=" * 70)
    print("PHASE 5: SRAM ANALYSIS")
    print("=" * 70)

    if not sram:
        print("  No SRAM dump provided")
        return

    print(f"\n### SRAM Overview ({len(sram)} bytes at 0x{SRAM_BASE:08x})")

    # Find non-zero regions
    print("\n### Non-zero regions")
    in_region = False
    region_start = 0
    for i in range(0, len(sram), 16):
        block = sram[i:i + 16]
        has_data = any(b != 0 for b in block)
        if has_data and not in_region:
            region_start = i
            in_region = True
        elif not has_data and in_region:
            abs_start = SRAM_BASE + region_start
            abs_end = SRAM_BASE + i
            print(f"  0x{abs_start:08x}-0x{abs_end:08x} ({i - region_start} bytes)")
            in_region = False
    if in_region:
        abs_start = SRAM_BASE + region_start
        abs_end = SRAM_BASE + len(sram)
        print(f"  0x{abs_start:08x}-0x{abs_end:08x} ({len(sram) - region_start} bytes)")

    # Show known SRAM locations
    print("\n### Known SRAM Locations")
    for addr, (name, desc) in sorted(KNOWN_DATA.items()):
        if SRAM_BASE <= addr < SRAM_BASE + len(sram):
            offset = addr - SRAM_BASE
            hex_data = " ".join(f"{b:02x}" for b in sram[offset:offset + 16])
            print(f"  0x{addr:08x} ({name:25s}): {hex_data}")

    # Find the USB buffer content
    print("\n### USB Buffer Area (0x02004110+)")
    usb_offset = 0x4110
    if usb_offset + 64 <= len(sram):
        for i in range(0, 64, 16):
            hex_data = " ".join(f"{b:02x}" for b in sram[usb_offset + i:usb_offset + i + 16])
            print(f"  0x{SRAM_BASE + usb_offset + i:08x}: {hex_data}")


# ─── Phase 6: Summary Report ────────────────────────────────────────

def print_summary(rom, functions, strings):
    """Print a summary of findings."""
    print("\n" + "=" * 70)
    print("SUMMARY")
    print("=" * 70)

    print(f"\n  ROM size:          {len(rom)} bytes ({len(rom) // 1024} KB)")
    print(f"  Functions found:   {len(functions)} (THUMB prologues)")
    print(f"  Strings found:     {len(strings)}")
    print(f"  Known addresses:   {len(KNOWN_FUNCTIONS)} functions, {len(KNOWN_DATA)} data")

    print("\n### Key Addresses for CachedSectorControlDownload")
    print(f"  USB read handler 1: 0x{0x574fc:08x} (patch target)")
    print(f"  USB read handler 2: 0x{0x57500:08x} (patch target)")
    print(f"  Resident code addr: 0x{0x02003ce0:08x} (SRAM)")
    print(f"  Enabled flag addr:  0x{0x02003cd4:08x} (SRAM)")
    print(f"  Sector counter:     0x{0x02003cd0:08x} (SRAM)")

    print("\n### Key Addresses for EEPROM Feature Unlock")
    print("  Register 0x61: Sound settings, line-out, title display")
    print("  Register 0x62: Program play, DPC, menu items")
    print("  MZ-N505: 0x61=0x80→0xFE, 0x62=0x10→0x7B")


# ─── Main ────────────────────────────────────────────────────────────

def main():
    if len(sys.argv) < 2:
        print(f"Usage: {sys.argv[0]} <firmware.bin> [firmware.sram]")
        sys.exit(1)

    rom_path = sys.argv[1]
    sram_path = sys.argv[2] if len(sys.argv) > 2 else None

    rom = load_rom(rom_path)
    sram = load_sram(sram_path) if sram_path else None

    print(f"MZ-N505 R1.400 Firmware Analysis")
    print(f"ROM: {rom_path} ({len(rom)} bytes)")
    if sram:
        print(f"SRAM: {sram_path} ({len(sram)} bytes)")
    print()

    # Phase 1
    analyze_vectors(rom)
    strings = find_strings(rom)
    functions = find_functions(rom)
    find_code_regions(rom)

    # Phase 2
    analyze_known_functions(rom)

    # Phase 3
    find_usb_handlers(rom)

    # Phase 4
    find_eeprom_access(rom)

    # Phase 5
    if sram:
        analyze_sram(sram)

    # Summary
    print_summary(rom, functions, strings)


if __name__ == "__main__":
    main()
