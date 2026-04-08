package netmd

import (
	"fmt"
	"log"
)

// DeviceFamily represents the processor type in the NetMD device.
type DeviceFamily int

const (
	FamilyUnknown DeviceFamily = iota
	FamilyR                    // Type-R (CXD2677): MZ-N505, MZ-N707
	FamilyS                    // Type-S (CXD2680): MZ-NE410, MZ-N510, MZ-N910
)

// DeviceProfile contains all exploit-relevant addresses for a specific
// firmware version. Populated from the version code returned by the
// factory 1812 command.
type DeviceProfile struct {
	VersionCode string
	Family      DeviceFamily

	// Patch peripheral
	PatchBase       uint32 // 0x03802000 for R/S
	PatchControl    uint32 // base + totalSlots*0x10
	TotalPatchSlots int    // 4 for R, 8 for S

	// USB code execution
	ExecCommand     byte     // 0xd3 for R, 0xd2 for S
	OnePatchAddress uint32   // ROM address patched for code execution
	OnePatchValue   [4]byte  // THUMB jump stub

	// NoRam exploit
	GDiscStateStruct uint32 // firmware disc state struct
	DiscStructOffset int    // offset for response size field (0x18, 0x1c, or 0x24)
	ReadAtracDram    uint32 // read_atrac_dram function address

	// CachedSectorControlDownload
	ResidentCodeAddr uint32 // SRAM: where resident code lives
	EnabledFlagAddr  uint32 // SRAM: 0=normal USB, 1=serve sectors
	SectorToReadAddr uint32 // SRAM: auto-incremented sector number
	SectorBufferAddr uint32 // SRAM buffer for sector data
	USBReadHandler1  uint32 // ROM: USB read response handler
	USBReadHandler2  uint32 // ROM: second handler instruction
	UsbDoResponse    uint32 // usb_do_response function address
	GUsbBuff         uint32 // g_usb_buff address

	// Firmware dump
	RomSize  uint32 // ROM size in bytes
	SramSize uint32 // SRAM size in bytes
}

// chipTypeToFamily maps the chipType byte from the 1812 response to a device family.
func chipTypeToFamily(chipType byte) (DeviceFamily, string) {
	switch chipType {
	case 0x20:
		return FamilyR, "R"
	case 0x21:
		return FamilyS, "S"
	default:
		return FamilyUnknown, ""
	}
}

// parseVersionCode builds a version code string from the 1812 response fields.
// chipType: 0x20=R, 0x21=S; versionByte: BCD byte (0x16 = v1.6); subversion: hex suffix.
// The version byte is BCD-encoded: high nibble = major, low nibble = minor.
func parseVersionCode(chipType byte, versionByte byte, subversion byte) string {
	family, prefix := chipTypeToFamily(chipType)
	major := (versionByte >> 4) & 0x0f
	minor := versionByte & 0x0f
	if family == FamilyUnknown {
		return fmt.Sprintf("0x%02x?%d.%d%02X", chipType, major, minor, subversion)
	}
	return fmt.Sprintf("%s%d.%d%02X", prefix, major, minor, subversion)
}

// LookupProfile returns the device profile for a version code, or nil if unknown.
func LookupProfile(versionCode string) *DeviceProfile {
	p, ok := deviceProfiles[versionCode]
	if !ok {
		log.Printf("WARNING: No device profile for version %q", versionCode)
		return nil
	}
	return p
}

// All addresses below sourced from netmd-exploits JS:
//   usb-code-execution.js — onePatchAddress, onePatchValue, command
//   core-macros.js — g_DiscStateStruct, read_atrac_dram, usb_do_response, g_usb_buff
//   cached-sector-noram-transfers.js — discStructOffset
//   cached-sector-control-transfers.js — usbReadStandardResponse, residentCodeAddress, etc.
//   firmware-dumper.js — romSize, ramSize
//   exploit-state.js — patchAmount

var deviceProfiles = map[string]*DeviceProfile{
	// ==================== Type-R (CXD2677) ====================
	"R1.000": {
		VersionCode:      "R1.000",
		Family:           FamilyR,
		PatchBase:        0x03802000,
		PatchControl:     0x03802040,
		TotalPatchSlots:  4,
		ExecCommand:      0xd3,
		OnePatchAddress:  0x00056228,
		OnePatchValue:    [4]byte{0x1a, 0x48, 0x00, 0x47},
		GDiscStateStruct: 0x02000b34,
		DiscStructOffset: 0x1c,
		ReadAtracDram:    0x0005d37d,
		ResidentCodeAddr: 0x02003ce0,
		EnabledFlagAddr:  0x02003cd4,
		SectorToReadAddr: 0x02003cd0,
		SectorBufferAddr: 0x02003240,
		USBReadHandler1:  0x00055b8c,
		USBReadHandler2:  0x00055b90,
		UsbDoResponse:    0x00058e35,
		GUsbBuff:         0x020040f0,
		RomSize:          0x70000,
		SramSize:         0x4800,
	},
	"R1.100": {
		VersionCode:      "R1.100",
		Family:           FamilyR,
		PatchBase:        0x03802000,
		PatchControl:     0x03802040,
		TotalPatchSlots:  4,
		ExecCommand:      0xd3,
		OnePatchAddress:  0x00056aac,
		OnePatchValue:    [4]byte{0x1a, 0x48, 0x00, 0x47},
		GDiscStateStruct: 0x02000b38,
		DiscStructOffset: 0x1c,
		ReadAtracDram:    0x0005dc65,
		ResidentCodeAddr: 0x02003ce0,
		EnabledFlagAddr:  0x02003cd4,
		SectorToReadAddr: 0x02003cd0,
		SectorBufferAddr: 0x02003240,
		USBReadHandler1:  0x00056410,
		USBReadHandler2:  0x00056414,
		UsbDoResponse:    0x00059715,
		GUsbBuff:         0x02004104,
		RomSize:          0x70000,
		SramSize:         0x4800,
	},
	"R1.200": {
		VersionCode:      "R1.200",
		Family:           FamilyR,
		PatchBase:        0x03802000,
		PatchControl:     0x03802040,
		TotalPatchSlots:  4,
		ExecCommand:      0xd3,
		OnePatchAddress:  0x000577f8,
		OnePatchValue:    [4]byte{0x1a, 0x48, 0x00, 0x47},
		GDiscStateStruct: 0x02000b34,
		DiscStructOffset: 0x18,
		ReadAtracDram:    0x0005e9fd,
		ResidentCodeAddr: 0x02003ce0,
		EnabledFlagAddr:  0x02003cd4,
		SectorToReadAddr: 0x02003cd0,
		SectorBufferAddr: 0x02003240,
		USBReadHandler1:  0x0005715c,
		USBReadHandler2:  0x00057160,
		UsbDoResponse:    0x0005a4bd,
		GUsbBuff:         0x02004108,
		RomSize:          0x70000,
		SramSize:         0x4800,
	},
	"R1.300": {
		VersionCode:      "R1.300",
		Family:           FamilyR,
		PatchBase:        0x03802000,
		PatchControl:     0x03802040,
		TotalPatchSlots:  4,
		ExecCommand:      0xd3,
		OnePatchAddress:  0x00057b48,
		OnePatchValue:    [4]byte{0x1a, 0x48, 0x00, 0x47},
		GDiscStateStruct: 0x02000b34,
		DiscStructOffset: 0x18,
		ReadAtracDram:    0x0005ed51,
		ResidentCodeAddr: 0x02003ce0,
		EnabledFlagAddr:  0x02003cd4,
		SectorToReadAddr: 0x02003cd0,
		SectorBufferAddr: 0x02003240,
		USBReadHandler1:  0x0005745c,
		USBReadHandler2:  0x00057460,
		UsbDoResponse:    0x0005a811,
		GUsbBuff:         0x0200410c,
		RomSize:          0x70000,
		SramSize:         0x4800,
	},
	"R1.400": {
		VersionCode:      "R1.400",
		Family:           FamilyR,
		PatchBase:        0x03802000,
		PatchControl:     0x03802040,
		TotalPatchSlots:  4,
		ExecCommand:      0xd3,
		OnePatchAddress:  0x00057be8,
		OnePatchValue:    [4]byte{0x1a, 0x48, 0x00, 0x47},
		GDiscStateStruct: 0x02000b38,
		DiscStructOffset: 0x18,
		ReadAtracDram:    0x0005edf1,
		ResidentCodeAddr: 0x02003ce0,
		EnabledFlagAddr:  0x02003cd4,
		SectorToReadAddr: 0x02003cd0,
		SectorBufferAddr: 0x02003240,
		USBReadHandler1:  0x000574fc,
		USBReadHandler2:  0x00057500,
		UsbDoResponse:    0x0005a8b1,
		GUsbBuff:         0x02004110,
		RomSize:          0x70000,
		SramSize:         0x4800,
	},

	// ==================== Type-S (CXD2680) ====================
	"S1.000": {
		VersionCode:      "S1.000",
		Family:           FamilyS,
		PatchBase:        0x03802000,
		PatchControl:     0x03802080,
		TotalPatchSlots:  8,
		ExecCommand:      0xd2,
		OnePatchAddress:  0x0000e784,
		OnePatchValue:    [4]byte{0x13, 0x48, 0x00, 0x47},
		GDiscStateStruct: 0x020001e8,
		DiscStructOffset: 0x24,
		ReadAtracDram:    0x00078365,
		ResidentCodeAddr: 0x02005f00,
		EnabledFlagAddr:  0x02005500,
		SectorToReadAddr: 0x02005600,
		SectorBufferAddr: 0x02006f00,
		USBReadHandler1:  0x0000df34,
		USBReadHandler2:  0x0000df38,
		UsbDoResponse:    0x00077ba5,
		GUsbBuff:         0x0200117c,
		RomSize:          0xa0000,
		SramSize:         0x9000,
	},
	"S1.100": {
		VersionCode:      "S1.100",
		Family:           FamilyS,
		PatchBase:        0x03802000,
		PatchControl:     0x03802080,
		TotalPatchSlots:  8,
		ExecCommand:      0xd2,
		OnePatchAddress:  0x0000d784,
		OnePatchValue:    [4]byte{0x13, 0x48, 0x00, 0x47},
		GDiscStateStruct: 0x020001cc,
		DiscStructOffset: 0x24,
		ReadAtracDram:    0x00071a35,
		ResidentCodeAddr: 0x02005f00,
		EnabledFlagAddr:  0x02005500,
		SectorToReadAddr: 0x02005600,
		SectorBufferAddr: 0x02006f00,
		USBReadHandler1:  0x0000cf5c,
		USBReadHandler2:  0x0000cf60,
		UsbDoResponse:    0x00071321,
		GUsbBuff:         0x02001014,
		RomSize:          0xa0000,
		SramSize:         0x9000,
	},
	"S1.200": {
		VersionCode:      "S1.200",
		Family:           FamilyS,
		PatchBase:        0x03802000,
		PatchControl:     0x03802080,
		TotalPatchSlots:  8,
		ExecCommand:      0xd2,
		OnePatchAddress:  0x0000d834,
		OnePatchValue:    [4]byte{0x13, 0x48, 0x00, 0x47},
		GDiscStateStruct: 0x020001cc,
		DiscStructOffset: 0x24,
		ReadAtracDram:    0x000723bd,
		ResidentCodeAddr: 0x02005f00,
		EnabledFlagAddr:  0x02005500,
		SectorToReadAddr: 0x02005600,
		SectorBufferAddr: 0x02006f00,
		USBReadHandler1:  0x0000cfe8,
		USBReadHandler2:  0x0000cfec,
		UsbDoResponse:    0x00071ca9,
		GUsbBuff:         0x02001018,
		RomSize:          0xa0000,
		SramSize:         0x9000,
	},
	"S1.300": {
		VersionCode:      "S1.300",
		Family:           FamilyS,
		PatchBase:        0x03802000,
		PatchControl:     0x03802080,
		TotalPatchSlots:  8,
		ExecCommand:      0xd2,
		OnePatchAddress:  0x0000daa8,
		OnePatchValue:    [4]byte{0x13, 0x48, 0x00, 0x47},
		GDiscStateStruct: 0x020001d8,
		DiscStructOffset: 0x24,
		ReadAtracDram:    0x00073a71,
		ResidentCodeAddr: 0x02005f00,
		EnabledFlagAddr:  0x02005500,
		SectorToReadAddr: 0x02005600,
		SectorBufferAddr: 0x02006f00,
		USBReadHandler1:  0x0000d25c,
		USBReadHandler2:  0x0000d260,
		UsbDoResponse:    0x000732c1,
		GUsbBuff:         0x0200102c,
		RomSize:          0xa0000,
		SramSize:         0x9000,
	},
	"S1.400": {
		VersionCode:      "S1.400",
		Family:           FamilyS,
		PatchBase:        0x03802000,
		PatchControl:     0x03802080,
		TotalPatchSlots:  8,
		ExecCommand:      0xd2,
		OnePatchAddress:  0x0000e4c4,
		OnePatchValue:    [4]byte{0x13, 0x48, 0x00, 0x47},
		GDiscStateStruct: 0x020001e8,
		DiscStructOffset: 0x24,
		ReadAtracDram:    0x00077131,
		ResidentCodeAddr: 0x02005f00,
		EnabledFlagAddr:  0x02005500,
		SectorToReadAddr: 0x02005600,
		SectorBufferAddr: 0x02006f00,
		USBReadHandler1:  0x0000dca4,
		USBReadHandler2:  0x0000dca8,
		UsbDoResponse:    0x00076971,
		GUsbBuff:         0x0200113c,
		RomSize:          0xa0000,
		SramSize:         0x9000,
	},
	"S1.500": {
		VersionCode:      "S1.500",
		Family:           FamilyS,
		PatchBase:        0x03802000,
		PatchControl:     0x03802080,
		TotalPatchSlots:  8,
		ExecCommand:      0xd2,
		OnePatchAddress:  0x0000e538,
		OnePatchValue:    [4]byte{0x13, 0x48, 0x00, 0x47},
		GDiscStateStruct: 0x020001e8,
		DiscStructOffset: 0x24,
		ReadAtracDram:    0x00077805,
		ResidentCodeAddr: 0x02005f00,
		EnabledFlagAddr:  0x02005500,
		SectorToReadAddr: 0x02005600,
		SectorBufferAddr: 0x02006f00,
		USBReadHandler1:  0x0000dcec,
		USBReadHandler2:  0x0000dcf0,
		UsbDoResponse:    0x00077045,
		GUsbBuff:         0x02001158,
		RomSize:          0xa0000,
		SramSize:         0x9000,
	},
	"S1.600": {
		VersionCode:      "S1.600",
		Family:           FamilyS,
		PatchBase:        0x03802000,
		PatchControl:     0x03802080,
		TotalPatchSlots:  8,
		ExecCommand:      0xd2,
		OnePatchAddress:  0x0000e69c,
		OnePatchValue:    [4]byte{0x13, 0x48, 0x00, 0x47},
		GDiscStateStruct: 0x020001e8,
		DiscStructOffset: 0x24,
		ReadAtracDram:    0x000781fd,
		ResidentCodeAddr: 0x02005f00,
		EnabledFlagAddr:  0x02005500,
		SectorToReadAddr: 0x02005600,
		SectorBufferAddr: 0x02006f00,
		USBReadHandler1:  0x0000de4c,
		USBReadHandler2:  0x0000de50,
		UsbDoResponse:    0x00077a3d,
		GUsbBuff:         0x02001170,
		RomSize:          0xa0000,
		SramSize:         0x9000,
	},
}
