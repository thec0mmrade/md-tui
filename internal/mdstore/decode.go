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

	// Find metadata frame and collect data frames with sequence numbers
	var metaFrame []byte
	var dataFrames []indexedFrame

	for _, f := range allFrames {
		switch f[0] {
		case frameTypeMetadata:
			if metaFrame == nil {
				metaFrame = f
			}
		case frameTypeData:
			seq := binary.LittleEndian.Uint16(f[1:3])
			dataFrames = append(dataFrames, indexedFrame{seq: seq, data: f})
		}
	}

	if metaFrame == nil {
		return "", fmt.Errorf("no metadata frame found in %d total frames", len(allFrames))
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
