package mdstore

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// LP2 sector layout constants (determined via calibration on MZ-N505).
const (
	sectorSize      = 2352
	sectorHeaderLen = 20
	sgHeaderLen     = 12 // sound group header: ac 00 00 00 00 00 00 00 00 00 00 00
	sgTotalLen      = 212 // 12 header + 192 frame + 8 padding
	sgPerSector     = 11
	sgFrameLen      = 192 // LP2 frame within each sound group
)

type indexedFrame struct {
	seq  uint16
	data []byte
}

// DecodeFile extracts the original file from downloaded raw sector data.
// Handles rotated sector order (circular disc cache) by using frame
// sequence numbers for reassembly.
func DecodeFile(rawPath, outputDir string) (string, error) {
	data, err := os.ReadFile(rawPath)
	if err != nil {
		return "", fmt.Errorf("read raw file: %w", err)
	}

	allFrames := extractFramesFromSectors(data)
	if len(allFrames) == 0 {
		return "", fmt.Errorf("no frames found in raw data")
	}

	// Find metadata frame and collect data frames with sequence numbers.
	// Validate metadata by checking for a plausible filename (printable ASCII)
	// and non-zero file size, since stale cache data may have type byte 0x00.
	var metaFrame []byte
	var dataFrames []indexedFrame

	for _, f := range allFrames {
		switch f[0] {
		case frameTypeMetadata:
			if metaFrame == nil && isValidMetadata(f) {
				metaFrame = f
			}
		case frameTypeData:
			seq := binary.LittleEndian.Uint16(f[1:3])
			dataFrames = append(dataFrames, indexedFrame{seq: seq, data: f})
		}
	}

	// Fallback: scan raw data directly for metadata frame pattern.
	// The SG-based extraction can miss the metadata when cache wrapping
	// shifts frame alignment relative to sound group boundaries.
	if metaFrame == nil {
		metaFrame = scanRawForMetadata(data)
		// Data frames from SG extraction are still valid — keep them
	}

	if metaFrame == nil {
		return "", fmt.Errorf("no valid metadata frame found in raw data (%d bytes)", len(data))
	}

	// Sort data frames by sequence number
	sort.Slice(dataFrames, func(i, j int) bool {
		return dataFrames[i].seq < dataFrames[j].seq
	})

	// Parse metadata
	payload := metaFrame[3:]
	nameEnd := 0
	for nameEnd < 128 && payload[nameEnd] != 0 {
		nameEnd++
	}
	filename := string(payload[:nameEnd])
	if filename == "" {
		filename = "recovered_file"
	}
	fileSize := binary.LittleEndian.Uint64(payload[128:136])
	var expectedHash [32]byte
	copy(expectedHash[:], payload[136:168])

	expectedFrames := (int(fileSize) + framePayloadSize - 1) / framePayloadSize
	fmt.Printf("Found metadata: %s (%d bytes, need %d data frames, found %d)\n",
		filename, fileSize, expectedFrames, len(dataFrames))

	// Deduplicate frames by sequence number (keep first occurrence,
	// which is more likely valid since cache fills from the start).
	seen := make(map[uint16]bool)
	dedupFrames := make([]indexedFrame, 0, len(dataFrames))
	for _, df := range dataFrames {
		if !seen[df.seq] {
			seen[df.seq] = true
			dedupFrames = append(dedupFrames, df)
		}
	}
	dataFrames = dedupFrames

	// Sort by sequence number and take only expected count
	sort.Slice(dataFrames, func(i, j int) bool {
		return dataFrames[i].seq < dataFrames[j].seq
	})
	if len(dataFrames) > expectedFrames {
		dataFrames = dataFrames[:expectedFrames]
	}
	if len(dataFrames) < expectedFrames {
		fmt.Printf("WARNING: only %d of %d expected data frames found\n", len(dataFrames), expectedFrames)
	}

	// Reassemble file data in sequence order
	var fileData []byte
	for _, df := range dataFrames {
		fileData = append(fileData, df.data[3:]...)
	}

	// Trim to original size
	if uint64(len(fileData)) < fileSize {
		return "", fmt.Errorf("insufficient data: got %d bytes, expected %d", len(fileData), fileSize)
	}
	fileData = fileData[:fileSize]

	// Verify checksum
	actualHash := sha256.Sum256(fileData)
	if actualHash != expectedHash {
		fmt.Printf("WARNING: SHA-256 mismatch!\n  Expected: %x\n  Got:      %x\n", expectedHash, actualHash)
	} else {
		fmt.Printf("SHA-256 checksum verified OK\n")
	}

	// Write output file
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("create output dir: %w", err)
	}
	outputPath := filepath.Join(outputDir, filename)
	if err := os.WriteFile(outputPath, fileData, 0644); err != nil {
		return "", fmt.Errorf("write output: %w", err)
	}

	fmt.Printf("Decoded %s (%d bytes)\n", outputPath, fileSize)
	return outputPath, nil
}

// scanRawForMetadata searches the raw data byte-by-byte for a valid metadata
// frame. Used when SG-based extraction misses it due to cache alignment shifts.
func scanRawForMetadata(rawData []byte) []byte {
	for i := 0; i <= len(rawData)-sgFrameLen; i++ {
		if rawData[i] == frameTypeMetadata && rawData[i+1] == 0x00 && rawData[i+2] == 0x00 {
			candidate := rawData[i : i+sgFrameLen]
			if isValidMetadata(candidate) {
				frame := make([]byte, sgFrameLen)
				copy(frame, candidate)
				fmt.Printf("Found metadata via raw scan at byte offset %d\n", i)
				return frame
			}
		}
	}
	return nil
}

// isValidMetadata checks if a frame is a genuine metadata frame by verifying
// it has a printable filename, non-zero file size, and correct version byte.
func isValidMetadata(f []byte) bool {
	if len(f) < sgFrameLen {
		return false
	}
	if f[0] != frameTypeMetadata || f[1] != 0 || f[2] != 0 {
		return false // must be type=0x00, seq=0x0000
	}
	payload := f[3:]
	// Check filename starts with printable ASCII
	if payload[0] < 0x20 || payload[0] > 0x7e {
		return false
	}
	// Check file size is non-zero
	fileSize := binary.LittleEndian.Uint64(payload[128:136])
	if fileSize == 0 || fileSize > 1<<32 { // sanity: max 4GB
		return false
	}
	// Check version byte
	if payload[168] != encodingVersion {
		return false
	}
	return true
}

// extractFramesFromSectors parses raw sector data into 192-byte LP2 frames.
//
// LP2 sector layout (2352 bytes per sector):
//   - 20-byte sector header
//   - 11 sound groups × 212 bytes:
//     - 12-byte SG header (ac 00 ...)
//     - 192-byte LP2 frame (our data)
//     - 8-byte padding
func extractFramesFromSectors(rawData []byte) [][]byte {
	numSectors := len(rawData) / sectorSize
	frames := make([][]byte, 0, numSectors*sgPerSector)

	for s := 0; s < numSectors; s++ {
		sectorStart := s * sectorSize

		for sg := 0; sg < sgPerSector; sg++ {
			sgStart := sectorStart + sectorHeaderLen + sg*sgTotalLen
			frameStart := sgStart + sgHeaderLen

			if frameStart+sgFrameLen > len(rawData) {
				break
			}

			frame := make([]byte, sgFrameLen)
			copy(frame, rawData[frameStart:frameStart+sgFrameLen])
			frames = append(frames, frame)
		}
	}

	return frames
}
