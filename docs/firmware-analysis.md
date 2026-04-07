MZ-N505 R1.400 Firmware Analysis
ROM: firmware_go.bin (458752 bytes)
SRAM: firmware_go.sram (18432 bytes)

======================================================================
PHASE 1: BASIC STRUCTURE
======================================================================

### ARM Exception Vector Table
  0x00000000: ldr      pc, [pc, #0x18]                ; Reset
  0x00000004: ldr      pc, [pc, #0x18]                ; Undefined Instruction
  0x00000008: ldr      pc, [pc, #0x18]                ; Software Interrupt (SWI)
  0x0000000c: ldr      pc, [pc, #0x18]                ; Prefetch Abort
  0x00000010: ldr      pc, [pc, #0x18]                ; Data Abort
  0x00000014: mov      r0, r0                         ; Reserved
  0x00000018: ldr      pc, [pc, #0x14]                ; IRQ
  0x0000001c: ldr      pc, [pc, #0x14]                ; FIQ

### Vector Targets
  Reset                          → 0x0000003c (via [0x00000020])
  Undefined Instruction          → 0x000000c8 (via [0x00000024])
  Software Interrupt (SWI)       → 0x000000c0 (via [0x00000028])
  Prefetch Abort                 → 0x000000c0 (via [0x0000002c])
  Data Abort                     → 0x000000c0 (via [0x00000030])
  IRQ                            → 0x00000140 (via [0x00000034])
  FIQ                            → 0x000001bc (via [0x00000038])

### ASCII Strings (length >= 6)
  0x00052a: "?H?I*J?K"
  0x000a33: "(2M3N4"
  0x000f5b: "C0p0x@"
  0x0014d7: "!!pap8"
  0x0015a9: " `pbx!x8"
  0x0017a0: "fuftQH"
  0x001d00: "6H@xxr"
  0x002067: "#0zGIXCG"
  0x003293: "pe Xpr "
  0x003e65: "C p x@"
  0x0049c1: " xCvIE"
  0x004ba8: "Push  "
  0x004c77: "H@x&"zC"
  0x005140: "P.D.M."
  0x005148: "PERSONAL"
  0x00515c: "GP ON "
  0x005164: "GROUP ON "
  0x005170: "GP OFF"
  0x005178: "GROUP OFF"
  0x0053e8: " (O)  "
  0x0053f0: "( O ) "
  0x0053f8: "  ((O))  "
  0x005404: " (( O )) "
  0x005513: " xCDID"
  0x00562c: "   ERROR "
  0x005638: " NO DISC "
  0x005644: "  PC>>MD "
  0x005650: "ERROR    "
  0x005768: "GP -- "
  0x005770: "GP __ "
  0x005778: "Group -- "
  0x005784: "Group __ "
  0x005974: "\(6"9E\X.<9#["
  0x005ab1: " pC0I@"
  0x0060e5: "pG XpP "
  0x00619b: "pG XpP "
  0x006669: "pG XpP "
  0x00688d: "pG XpP "
  0x0068bb: "pG XpP "
  0x0068e9: "pG XpP "
  0x0069b5: "pr Xpo "
  0x0069f7: "pr Xpo "
  0x006ae9: "pP Xp: "
  0x00767c: "SOUND 1  "
  0x007688: "SOUND 2  "
  0x007828: "SPEED    "
  0x007834: "CONTROL  "
  0x007d8a: "&O8h&I"
  0x008833: "$7H@x-"
  0x008ffc: "008p xU"
  0x009008: ")0xp`xU"
  0x00905a: " &>p~p"
  0x0090e0: "008q yU"
  0x009148: ")0`q  "
  0x009313: "p& pCeI@"
  0x009350: "&!qCVJ"
  0x009368: "& pCPI@"
  0x0094c9: "p& pC9IE"
  0x009508: "&!qC)J"
  0x009754: "ILhFax'"
  0x0098eb: "pP Xp: "
  0x009ecc: "$+5552"
  0x00a03c: "      "
  0x00a144: "111111"
  0x00a14c: "000000"
  0x00a654: "         "
  0x00a8d8: "Adj NG"
  0x00ad60: "RMC OK"
  0x00ad6c: "SET OK"
  0x00ae9f: "s[h}k+C{c"
  0x00b000: "Assy**"
  0x00b364: "Ofset!"
  0x00b36c: "CLOSE!"
  0x00b374: "MO RUN"
  0x00b38c: "SetCD!"
  0x00b394: "CD RUN"
  0x00b3a4: "NoTmp!"
  0x00b3ac: "Ofst**"
  0x00b538: "    NG"
  0x00b54c: "CD OK "
  0x00b560: "MO OK "
  0x00b56c: "OfstOK"
  0x00b574: "ADJ OK"
  0x00b754: "    NG"
  0x00b764: "ADJ OK"
  0x00b86c: "PUSH  "
  0x00b874: "JOG+OK"
  0x00b87c: "JOG+ 3"
  0x00b884: "JOG+ 2"
  0x00b88c: "JOG+ 1"
  0x00b894: "JOG OK"
  0x00b89c: "JOG- 3"
  0x00b8a4: "JOG- 2"
  0x00b8ac: "JOG- 1"
  0x00b9f8: "ClrOK?"
  0x00ba00: "ErrCLR"
  0x00ba0c: "RecT 0"
  0x00bba0: ""sUNUM"
  0x00bbba: " {PKQJ"
  0x00bfec: "000011"
  ... (1470 more strings)
  Total: 1570 strings found

### Function Prologues (THUMB PUSH {... lr})
  Found 2536 THUMB function prologues
  First: 0x000528
  Last:  0x068208

### Code/Data Regions
  0x000000-0x016000: CODE
  0x016000-0x017000: DATA
  0x017000-0x05f000: CODE
  0x05f000-0x060000: DATA
  0x060000-0x062000: CODE
  0x062000-0x069000: DATA
  0x069000-0x070000: EMPTY

======================================================================
PHASE 2: KNOWN ADDRESS ANALYSIS
======================================================================

### 0x000574fc: usbReadStandardResponse
    USB read handler 1 — CachedSectorControlDownload target
    Mode: ARM

    0x000574fc: stmdami  r3, {r8, sl, ip, sp, pc}
    0x00057500: stmdami  r3, {r0, r7, r8, fp, sp, lr}

### 0x00057500: usbReadStandardResponseNext
    USB read handler 2 — CachedSectorControlDownload target
    Mode: ARM

    0x00057500: stmdami  r3, {r0, r7, r8, fp, sp, lr}

### 0x00057be8: onePatchAddress
    USB code execution hook point (THUMB)
    Mode: ARM


### 0x0005a8b1: usb_do_response
    USB response function (THUMB)
    Mode: THUMB

    0x0005a8b0: push     {r7, lr}
    0x0005a8b2: sub      sp, #8
    0x0005a8b4: ldr      r7, [pc, #0x38]
    0x0005a8b6: movs     r2, #0
    0x0005a8b8: strh     r2, [r7, #0x2c]
    0x0005a8ba: strh     r1, [r7, #0x24]
    0x0005a8bc: str      r0, [r7, #0x20]
    0x0005a8be: strh     r2, [r7, #0x26]
    0x0005a8c0: movs     r1, #0x40
    0x0005a8c2: mvns     r1, r1
    0x0005a8c4: movs     r0, #0xc
    0x0005a8c6: bl       #0x5de50
    0x0005a8ca: bl       #0x5a8f4
    0x0005a8ce: movs     r0, #2
    0x0005a8d0: strb     r0, [r7]
    0x0005a8d2: subs     r2, r0, #3
    0x0005a8d4: str      r2, [sp]
    0x0005a8d6: movs     r2, #0x5f
    0x0005a8d8: movs     r3, #2
    0x0005a8da: movs     r1, #0xc
    0x0005a8dc: add      r0, sp, #4
    0x0005a8de: bl       #0x5da14
    0x0005a8e2: ldr      r1, [sp, #4]
    0x0005a8e4: bl       #0x5af08
    0x0005a8e8: add      sp, #8
    0x0005a8ea: pop      {r7}
    0x0005a8ec: pop      {r3}
    0x0005a8ee: bx       r3
    0x0005a8f0: lsrs     r4, r5, #0xe
    0x0005a8f2: lsls     r0, r0, #8
    0x0005a8f4: push     {r4, r5, r7, lr}
    0x0005a8f6: ldr      r2, [pc, #0x40]
    0x0005a8f8: ldrh     r7, [r2, #0x24]
    0x0005a8fa: cmp      r7, #0
    0x0005a8fc: beq      #0x5a928
    0x0005a8fe: ldr      r4, [pc, #0x3c]
    0x0005a900: ldr      r3, [r2, #0x20]
    0x0005a902: movs     r0, #0
    0x0005a904: ldrh     r1, [r2, #0x26]
    0x0005a906: cmp      r1, r7

### 0x0005da15: tron_twai_flg
    TRON OS: wait for flag (THUMB)
    Mode: THUMB

    0x0005da14: push     {r0, r1, r2, r3, r4, r5, r6, r7, lr}
    0x0005da16: ldr      r5, [sp, #0x24]
    0x0005da18: cmp      r5, #0
    0x0005da1a: beq      #0x5da30
    0x0005da1c: ldr      r0, [pc, #0x180]
    0x0005da1e: ldr      r0, [r0]
    0x0005da20: lsrs     r0, r0, #3
    0x0005da22: blo      #0x5da30
    0x0005da24: movs     r0, #0x44
    0x0005da26: mvns     r0, r0
    0x0005da28: add      sp, #0x10
    0x0005da2a: pop      {r4, r5, r6, r7}
    0x0005da2c: pop      {r3}
    0x0005da2e: bx       r3
    0x0005da30: movs     r3, #1
    0x0005da32: cmn      r5, r3
    0x0005da34: bge      #0x5da3c
    0x0005da36: movs     r0, #0x20
    0x0005da38: mvns     r0, r0
    0x0005da3a: b        #0x5da28
    0x0005da3c: ldr      r1, [sp, #4]
    0x0005da3e: cmp      r1, #0
    0x0005da40: ble      #0x5da4c
    0x0005da42: ldr      r0, [pc, #0x160]
    0x0005da44: ldr      r0, [r0]
    0x0005da46: ldr      r1, [sp, #4]
    0x0005da48: cmp      r0, r1
    0x0005da4a: bhs      #0x5da52
    0x0005da4c: movs     r0, #0x22
    0x0005da4e: mvns     r0, r0
    0x0005da50: b        #0x5da28
    0x0005da52: ldr      r3, [sp, #0xc]
    0x0005da54: cmp      r3, #3
    0x0005da56: bhi      #0x5da5e
    0x0005da58: ldr      r2, [sp, #8]
    0x0005da5a: cmp      r2, #0
    0x0005da5c: bne      #0x5da64
    0x0005da5e: movs     r0, #0x20
    0x0005da60: mvns     r0, r0
    0x0005da62: b        #0x5da28

### 0x0005dce5: tron_set_flg
    TRON OS: set flag (THUMB)
    Mode: THUMB

    0x0005dce4: push     {r0, r1, r4, r5, r6, r7, lr}
    0x0005dce6: sub      sp, #0xc
    0x0005dce8: ldr      r0, [sp, #0xc]
    0x0005dcea: cmp      r0, #0
    0x0005dcec: ble      #0x5dcf8
    0x0005dcee: ldr      r0, [pc, #0x150]
    0x0005dcf0: ldr      r0, [r0]
    0x0005dcf2: ldr      r1, [sp, #0xc]
    0x0005dcf4: cmp      r0, r1
    0x0005dcf6: bhs      #0x5dd04
    0x0005dcf8: movs     r0, #0x22
    0x0005dcfa: mvns     r0, r0
    0x0005dcfc: add      sp, #0x14
    0x0005dcfe: pop      {r4, r5, r6, r7}
    0x0005dd00: pop      {r3}
    0x0005dd02: bx       r3
    0x0005dd04: movs     r1, #0
    0x0005dd06: movs     r0, #1
    0x0005dd08: bl       #0x60354
    0x0005dd0c: str      r0, [sp, #8]
    0x0005dd0e: ldr      r0, [sp, #0xc]
    0x0005dd10: lsls     r0, r0, #3
    0x0005dd12: ldr      r1, [pc, #0x130]
    0x0005dd14: adds     r0, r0, r1
    0x0005dd16: subs     r4, r0, #7
    0x0005dd18: subs     r4, #1
    0x0005dd1a: ldr      r0, [r4, #4]
    0x0005dd1c: ldr      r1, [sp, #0x10]
    0x0005dd1e: orrs     r0, r1
    0x0005dd20: str      r0, [r4, #4]
    0x0005dd22: ldr      r5, [r4]
    0x0005dd24: ldr      r0, [r4]
    0x0005dd26: str      r0, [sp, #4]
    0x0005dd28: cmp      r5, #0
    0x0005dd2a: bne      #0x5dd2e
    0x0005dd2c: b        #0x5ddfe
    0x0005dd2e: ldr      r6, [r5, #0x18]
    0x0005dd30: ldr      r0, [r5, #0x1c]
    0x0005dd32: lsls     r0, r0, #0x1e
    0x0005dd34: lsrs     r0, r0, #0x1e

### 0x0005de51: tron_clr_flg
    TRON OS: clear flag (THUMB)
    Mode: THUMB

    0x0005de50: push     {r4, r5, r6, r7, lr}
    0x0005de52: adds     r4, r0, #0
    0x0005de54: adds     r5, r1, #0
    0x0005de56: cmp      r4, #0
    0x0005de58: ble      #0x5de62
    0x0005de5a: ldr      r0, [pc, #0x40]
    0x0005de5c: ldr      r0, [r0]
    0x0005de5e: cmp      r0, r4
    0x0005de60: bhs      #0x5de6c
    0x0005de62: movs     r0, #0x22
    0x0005de64: mvns     r0, r0
    0x0005de66: pop      {r4, r5, r6, r7}
    0x0005de68: pop      {r3}
    0x0005de6a: bx       r3
    0x0005de6c: movs     r1, #0
    0x0005de6e: movs     r0, #1
    0x0005de70: bl       #0x60354
    0x0005de74: adds     r7, r0, #0
    0x0005de76: lsls     r0, r4, #3
    0x0005de78: ldr      r1, [pc, #0x24]
    0x0005de7a: adds     r0, r0, r1
    0x0005de7c: subs     r6, r0, #7
    0x0005de7e: subs     r6, #1
    0x0005de80: ldr      r0, [r6, #4]
    0x0005de82: ands     r0, r5
    0x0005de84: str      r0, [r6, #4]
    0x0005de86: lsrs     r0, r7, #8
    0x0005de88: bhs      #0x5de92
    0x0005de8a: movs     r0, #0
    0x0005de8c: adds     r1, r7, #0
    0x0005de8e: bl       #0x60354
    0x0005de92: movs     r0, #0
    0x0005de94: b        #0x5de66
    0x0005de96: movs     r0, #0
    0x0005de98: b        #0x5de66
    0x0005de9a: movs     r0, r0
    0x0005de9c: lsrs     r0, r7
    0x0005de9e: movs     r6, r0
    0x0005dea0: movs     r4, #0xb8
    0x0005dea2: lsls     r0, r0, #8

### 0x0005edf1: read_atrac_dram
    Read ATRAC sector from disc cache (THUMB)
    Mode: THUMB

    0x0005edf0: push     {r4, r5, r6, r7, lr}
    0x0005edf2: adds     r4, r1, #0
    0x0005edf4: adds     r5, r0, #0
    0x0005edf6: adds     r7, r2, #0
    0x0005edf8: mov      r8, r3
    0x0005edfa: movs     r1, #0
    0x0005edfc: movs     r0, #1
    0x0005edfe: bl       #0x60354
    0x0005ee02: mov      sb, r0
    0x0005ee04: ldr      r0, [pc, #0xc0]
    0x0005ee06: ldrb     r1, [r0]
    0x0005ee08: movs     r2, #1
    0x0005ee0a: orrs     r1, r2
    0x0005ee0c: movs     r2, #0x10
    0x0005ee0e: bics     r1, r2
    0x0005ee10: strb     r1, [r0]
    0x0005ee12: movs     r0, #0xbc
    0x0005ee14: bl       #0x5eb2e
    0x0005ee18: sub      sp, #4
    0x0005ee1a: str      r5, [sp]
    0x0005ee1c: movs     r0, #0xac
    0x0005ee1e: mov      r1, sp
    0x0005ee20: movs     r2, #2
    0x0005ee22: bl       #0x5ead0
    0x0005ee26: str      r4, [sp]
    0x0005ee28: movs     r0, #0xad
    0x0005ee2a: mov      r1, sp
    0x0005ee2c: movs     r2, #2
    0x0005ee2e: bl       #0x5ead0
    0x0005ee32: add      sp, #4
    0x0005ee34: movs     r0, #0
    0x0005ee36: ldr      r1, [pc, #0x94]
    0x0005ee38: strb     r0, [r1]
    0x0005ee3a: ldr      r1, [pc, #0x94]
    0x0005ee3c: strb     r0, [r1]
    0x0005ee3e: movs     r0, #0xd0
    0x0005ee40: ldr      r1, [pc, #0x90]
    0x0005ee42: strb     r0, [r1]
    0x0005ee44: movs     r0, #0x38
    0x0005ee46: ldr      r2, [pc, #0x90]

### 0x00060355: ChangeIRQmask
    Change IRQ mask (THUMB)
    Mode: THUMB

    0x00060354: push     {r2}
    0x00060356: adr      r2, #4
    0x00060358: bx       r2
    0x0006035a: movs     r0, r0
    0x0006035c: movs     r0, r0
    0x0006035e: b        #0x60a04
    0x00060360: asrs     r0, r4, #0x20
    0x00060362: asrs     r1, r0, #0xf
    0x00060364: and      r1, r1, #0x290029
    0x00060368: movs     r7, r0
    0x0006036a: subs     r0, r0, r0
    0x0006036c: movs     r0, r0
    0x0006036e: b        #0x60a12
    0x00060370: movs     r0, #0
    0x00060372: b        #0x60594
    0x00060374: movs     r2, r0
    0x00060376: b        #0x606ba
    0x00060378: movs     r0, #0x80
    0x0006037a: asrs     r2, r0, #0xe
    0x0006037c: movs     r0, #0x20
    0x0006037e: asrs     r2, r0, #0xf
    0x00060380: movs     r0, #0xa0
    0x00060382: lsls     r2, r0, #0xf

======================================================================
PHASE 3: USB COMMAND HANDLER MAPPING
======================================================================

### References to USB command bytes
  18 09 (status/query             ): 0x004027, 0x00415f, 0x026201, 0x031aa4, 0x032990 (+29 more)
  18 12 (factory enter            ): 0x0049c7, 0x01aacb, 0x026b21, 0x031817, 0x05957f (+1 more)
  18 20 (changeMemoryState        ): 0x00aaa5, 0x010169, 0x012d7d, 0x012da1, 0x0130ab (+31 more)
  18 21 (factory read             ): 0x000d2e, 0x000e72, 0x00150a, 0x001e36, 0x009359 (+9 more)
  18 22 (factory write            ): 0x00bb9f, 0x013114, 0x013168, 0x013248, 0x01324e (+6 more)
  18 24 (firmware read            ): 0x00174c
  18 d3 (code execution (R-series)): 0x000f3c, 0x004c83, 0x00d7cb, 0x00f5fe, 0x014574 (+7 more)
  18 d2 (code execution (S-series)): 0x00f352, 0x011cbb, 0x011ec9, 0x018a6c, 0x018db6 (+7 more)
  18 c3 (play/pause               ): 0x01a1b9
  18 c5 (stop                     ): 0x00931b, 0x0154cd, 0x05dedd, 0x05df61, 0x05e155
  18 50 (goto/seek                ): 0x029733, 0x045b79

### Disassembly around USB handler area (0x574f0-0x57c00)
    0x000574f0: bx       r3
    0x000574f2: movs     r0, r0
    0x000574f4: lsrs     r0, r7, #0xc
    0x000574f6: lsls     r0, r0, #8
    0x000574f8: asrs     r0, r2
    0x000574fa: lsls     r0, r0, #8
    0x000574fc: push     {lr}  ; ← usbReadStandardResponse
    0x000574fe: ldr      r0, [pc, #0xc]
    0x00057500: ldr      r1, [r0, #0x18]  ; ← usbReadStandardResponseNext
    0x00057502: ldr      r0, [pc, #0xc]
    0x00057504: bl       #0x5a8b0
    0x00057508: pop      {r3}
    0x0005750a: bx       r3
    0x0005750c: lsrs     r0, r7, #0xc
    0x0005750e: lsls     r0, r0, #8
    0x00057510: asrs     r0, r2
    0x00057512: lsls     r0, r0, #8
    0x00057514: push     {r7, lr}
    0x00057516: sub      sp, #8
    0x00057518: ldr      r0, [pc, #0x104]
    0x0005751a: ldrb     r0, [r0, #0xd]
    0x0005751c: cmp      r0, #7
    0x0005751e: bhs      #0x575e4
    0x00057520: adr      r3, #4
    0x00057522: ldrb     r3, [r3, r0]
    0x00057524: lsls     r3, r3, #1
    0x00057526: add      pc, r3
    0x00057528: lsls     r3, r5, #0xd
    0x0005752a: cmp      r6, #0x14
    0x0005752c: ldr      r7, [r7, #0x30]
    0x0005752e: lsls     r0, r2, #1
    0x00057530: movs     r2, #0
    0x00057532: str      r2, [sp]
    0x00057534: movs     r3, #2
    0x00057536: lsls     r2, r3, #0x11
    0x00057538: movs     r1, #0xc
    0x0005753a: add      r0, sp, #4
    0x0005753c: bl       #0x5da14
    0x00057540: cmp      r0, #0
    0x00057542: bne      #0x57550
    0x00057544: bl       #0x59228
    0x00057548: ldr      r1, [pc, #0xd8]
    0x0005754a: movs     r0, #0xc
    0x0005754c: bl       #0x5de50
    0x00057550: b        #0x57600
    0x00057552: movs     r2, #0
    0x00057554: str      r2, [sp]
    0x00057556: movs     r3, #2
    0x00057558: lsls     r2, r3, #0xf
    0x0005755a: movs     r1, #0xc
    0x0005755c: add      r0, sp, #4
    0x0005755e: bl       #0x5da14
    0x00057562: cmp      r0, #0
    0x00057564: bne      #0x5756c
    0x00057566: bl       #0x59474
    0x0005756a: b        #0x57584
    0x0005756c: movs     r2, #0
    0x0005756e: str      r2, [sp]
    0x00057570: movs     r3, #2
    0x00057572: lsls     r2, r3, #0x10
    0x00057574: movs     r1, #0xc
    0x00057576: add      r0, sp, #4
    0x00057578: bl       #0x5da14
    0x0005757c: cmp      r0, #0
    0x0005757e: bne      #0x57584
    0x00057580: bl       #0x594c8
    0x00057584: b        #0x57600
    0x00057586: movs     r2, #0
    0x00057588: str      r2, [sp]
    0x0005758a: movs     r3, #2
    0x0005758c: lsls     r2, r3, #0x12
    0x0005758e: movs     r1, #0xc
    0x00057590: add      r0, sp, #4
    0x00057592: bl       #0x5da14
    0x00057596: cmp      r0, #0
    0x00057598: bne      #0x575a6
    0x0005759a: ldr      r1, [pc, #0x8c]
    0x0005759c: movs     r0, #0xc
    0x0005759e: bl       #0x5de50
    0x000575a2: bl       #0x59564

======================================================================
PHASE 4: EEPROM ACCESS
======================================================================

### References to EEPROM register numbers (0x61, 0x62, 0x63)
  Register 0x61: found at 0x0029b2, 0x005810, 0x00a188, 0x012606, 0x023bea, 0x02e4cc, 0x030112, 0x0340b4, 0x0417f4, 0x04915e
  Register 0x62: found at 0x005814, 0x005c56, 0x005c9e, 0x005e3e, 0x006c1e, 0x006c66, 0x006dd6, 0x00a438, 0x00ac84, 0x013f42
  Register 0x63: found at 0x004a12, 0x005510, 0x0055a0, 0x0071ae, 0x00724a, 0x009276, 0x00c912, 0x013724, 0x013734, 0x0242de

### EEPROM-related strings

======================================================================
PHASE 5: SRAM ANALYSIS
======================================================================

### SRAM Overview (18432 bytes at 0x02000000)

### Non-zero regions
  0x02000000-0x02000020 (32 bytes)
  0x02000050-0x02000060 (16 bytes)
  0x02000080-0x02000090 (16 bytes)
  0x020000c0-0x02000110 (80 bytes)
  0x02000120-0x02000140 (32 bytes)
  0x02000160-0x02000170 (16 bytes)
  0x02000180-0x020001f0 (112 bytes)
  0x02000200-0x02000270 (112 bytes)
  0x020002d0-0x02000380 (176 bytes)
  0x020003a0-0x020003c0 (32 bytes)
  0x020003d0-0x02000420 (80 bytes)
  0x02000430-0x02000440 (16 bytes)
  0x02000470-0x020005d0 (352 bytes)
  0x020005e0-0x02000660 (128 bytes)
  0x02000670-0x02000680 (16 bytes)
  0x02000690-0x02000800 (368 bytes)
  0x02000810-0x02000890 (128 bytes)
  0x020008b0-0x02000900 (80 bytes)
  0x02000910-0x02000920 (16 bytes)
  0x02000940-0x020009a0 (96 bytes)
  0x020009b0-0x020009c0 (16 bytes)
  0x020009d0-0x02000a00 (48 bytes)
  0x02000a10-0x02000a50 (64 bytes)
  0x02000a80-0x02000af0 (112 bytes)
  0x02000b00-0x02000b60 (96 bytes)
  0x02000ba0-0x02000bf0 (80 bytes)
  0x02000c00-0x02000c40 (64 bytes)
  0x02000c50-0x02000c70 (32 bytes)
  0x02000c90-0x02000cc0 (48 bytes)
  0x02000d20-0x02000d30 (16 bytes)
  0x02000d60-0x02000de0 (128 bytes)
  0x02000e00-0x02000e20 (32 bytes)
  0x02000e50-0x02000e60 (16 bytes)
  0x02000fe0-0x020010e0 (256 bytes)
  0x02001290-0x02001310 (128 bytes)
  0x02001410-0x02001470 (96 bytes)
  0x02001480-0x020014b0 (48 bytes)
  0x02001510-0x02001650 (320 bytes)
  0x02001680-0x020016a0 (32 bytes)
  0x02001780-0x020020b0 (2352 bytes)
  0x020022c0-0x020022d0 (16 bytes)
  0x020022e0-0x020023f0 (272 bytes)
  0x02002410-0x02002470 (96 bytes)
  0x02002490-0x020024c0 (48 bytes)
  0x020024e0-0x02002520 (64 bytes)
  0x02002750-0x020027d0 (128 bytes)
  0x020029a0-0x02002a90 (240 bytes)
  0x02002b20-0x02002ce0 (448 bytes)
  0x02002d60-0x02002dd0 (112 bytes)
  0x02002e10-0x02002ed0 (192 bytes)
  0x02002f20-0x02003050 (304 bytes)
  0x02003120-0x02003230 (272 bytes)
  0x020037d0-0x020037f0 (32 bytes)
  0x02003b20-0x02003b30 (16 bytes)
  0x02004110-0x02004150 (64 bytes)
  0x02004580-0x020045c0 (64 bytes)
  0x02004610-0x02004630 (32 bytes)
  0x02004720-0x02004800 (224 bytes)

### Known SRAM Locations
  0x02000b38 (g_DiscStateStruct        ): 00 ff ff ff ff ff 70 a2 ff ff ff ff ff 00 00 00
  0x02003cd0 (sectorToReadAddr         ): 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
  0x02003cd4 (enabledFlagAddr          ): 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
  0x02003ce0 (residentCodeAddr         ): 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00
  0x02004110 (g_usb_buff               ): 09 00 10 00 00 36 00 00 00 34 00 02 00 02 fe 75

### USB Buffer Area (0x02004110+)
  0x02004110: 09 00 10 00 00 36 00 00 00 34 00 02 00 02 fe 75
  0x02004120: 10 00 18 00 00 0d 00 0b 01 10 01 03 01 20 32 41
  0x02004130: 02 10 00 00 19 4e 65 74 4d 44 10 ff 00 00 33 77
  0x02004140: 4e 65 74 20 4d 44 20 57 61 6c 6b 6d 61 6e 43 53

======================================================================
SUMMARY
======================================================================

  ROM size:          458752 bytes (448 KB)
  Functions found:   2536 (THUMB prologues)
  Strings found:     1570
  Known addresses:   9 functions, 7 data

### Key Addresses for CachedSectorControlDownload
  USB read handler 1: 0x000574fc (patch target)
  USB read handler 2: 0x00057500 (patch target)
  Resident code addr: 0x02003ce0 (SRAM)
  Enabled flag addr:  0x02003cd4 (SRAM)
  Sector counter:     0x02003cd0 (SRAM)

### Key Addresses for EEPROM Feature Unlock
  Register 0x61: Sound settings, line-out, title display
  Register 0x62: Program play, DPC, menu items
  MZ-N505: 0x61=0x80→0xFE, 0x62=0x10→0x7B
