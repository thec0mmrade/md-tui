package mdstore

import (
	"fmt"
	"os"
)

// GenerateCalibrationWAV creates a test WAV where each 192-byte LP2 frame
// has a known pattern, making it easy to find frames in raw sector data.
//
// Frame N: byte 0 = 0x01 (data type), bytes 1-2 = N (LE), bytes 3-191 = byte(N & 0xFF)
func GenerateCalibrationWAV(outputPath string, numFrames int) error {
	if numFrames < 2 {
		numFrames = 64
	}

	frameData := make([]byte, numFrames*lp2FrameSize)
	for n := 0; n < numFrames; n++ {
		offset := n * lp2FrameSize
		frameData[offset] = frameTypeData
		frameData[offset+1] = byte(n & 0xFF)
		frameData[offset+2] = byte((n >> 8) & 0xFF)
		// Fill remaining bytes with recognizable pattern
		fill := byte(n & 0xFF)
		for i := 3; i < lp2FrameSize; i++ {
			frameData[offset+i] = fill
		}
	}

	wav := BuildATRAC3WAV(frameData)
	if err := os.WriteFile(outputPath, wav, 0644); err != nil {
		return fmt.Errorf("write calibration WAV: %w", err)
	}

	fmt.Printf("Generated calibration WAV: %s (%d frames, %d bytes)\n",
		outputPath, numFrames, len(wav))
	return nil
}

// AnalyzeCalibration reads downloaded raw sector data and searches for
// calibration frame patterns to determine the LP2 sector layout.
func AnalyzeCalibration(rawPath string) (string, error) {
	data, err := os.ReadFile(rawPath)
	if err != nil {
		return "", fmt.Errorf("read raw file: %w", err)
	}

	sectorSize := 2352
	numSectors := len(data) / sectorSize
	report := fmt.Sprintf("Raw file: %d bytes, %d sectors\n\n", len(data), numSectors)

	// Dump first few sectors' hex for manual inspection
	maxDump := 4
	if numSectors < maxDump {
		maxDump = numSectors
	}

	for s := 0; s < maxDump; s++ {
		sectorStart := s * sectorSize
		sector := data[sectorStart : sectorStart+sectorSize]
		report += fmt.Sprintf("=== Sector %d (offset 0x%x) ===\n", s, sectorStart)
		report += fmt.Sprintf("  Header (first 32 bytes): % x\n", sector[:32])

		// Search for calibration patterns within this sector
		for frameN := 0; frameN < 256; frameN++ {
			pattern := byte(frameN & 0xFF)
			// Look for runs of 8+ identical bytes matching a frame fill pattern
			for i := 0; i < len(sector)-8; i++ {
				match := true
				for j := 0; j < 8; j++ {
					if sector[i+j] != pattern {
						match = false
						break
					}
				}
				if match && pattern != 0x00 { // skip zero runs (too common)
					report += fmt.Sprintf("  Found frame %d pattern (0x%02x×8) at sector offset %d (0x%x)\n",
						frameN, pattern, i, i)
					break
				}
			}
		}
	}

	// Global search: find all calibration frame starts across all data
	report += "\n=== Global frame detection ===\n"
	found := 0
	for i := 0; i < len(data)-lp2FrameSize; i++ {
		// Look for frame type byte + sequence number + fill pattern
		if data[i] == frameTypeData && i+lp2FrameSize <= len(data) {
			seq := int(data[i+1]) | int(data[i+2])<<8
			fill := byte(seq & 0xFF)
			if fill == 0 {
				continue // skip ambiguous zeros
			}
			// Check if bytes 3-10 all match the expected fill
			allMatch := true
			for j := 3; j < 11; j++ {
				if data[i+j] != fill {
					allMatch = false
					break
				}
			}
			if allMatch {
				sector := i / sectorSize
				sectorOff := i % sectorSize
				report += fmt.Sprintf("  Frame seq=%d (fill=0x%02x) at byte %d (sector %d, offset %d)\n",
					seq, fill, i, sector, sectorOff)
				found++
				if found >= 32 {
					report += "  ... (truncated)\n"
					break
				}
			}
		}
	}

	if found == 0 {
		report += "  No calibration frames found. The sector layout may require different parsing.\n"
		report += "  Try examining raw hex: hexdump -C <rawfile> | head -100\n"
	}

	return report, nil
}
