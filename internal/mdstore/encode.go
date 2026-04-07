package mdstore

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

const (
	framePayloadSize = 189 // usable bytes per frame (192 - 3 header bytes)
	encodingVersion  = 1

	frameTypeMetadata = 0x00
	frameTypeData     = 0x01
	frameTypePadding  = 0xFF
)

// EncodeFile encodes an arbitrary file into an ATRAC3 WAV container for LP2 upload.
// The WAV can be uploaded as an LP2 track — the device stores the data verbatim.
func EncodeFile(inputPath, wavOutputPath string) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("read input file: %w", err)
	}

	dataFrameCount := (len(data) + framePayloadSize - 1) / framePayloadSize
	if dataFrameCount > 65534 { // uint16 max minus metadata frame
		return fmt.Errorf("file too large for encoding: %d bytes (%d frames, max 65534)",
			len(data), dataFrameCount)
	}

	frames := encodeToFrames(filepath.Base(inputPath), data)

	// Pad to frame boundary
	frameData := make([]byte, len(frames)*lp2FrameSize)
	for i, f := range frames {
		copy(frameData[i*lp2FrameSize:], f)
	}

	wav := BuildATRAC3WAV(frameData)
	if err := os.WriteFile(wavOutputPath, wav, 0644); err != nil {
		return fmt.Errorf("write WAV: %w", err)
	}

	fmt.Printf("Encoded %s (%d bytes) → %s (%d frames, %d bytes WAV)\n",
		filepath.Base(inputPath), len(data), wavOutputPath, len(frames), len(wav))
	return nil
}

func encodeToFrames(filename string, data []byte) [][]byte {
	checksum := sha256.Sum256(data)

	// Calculate total data frames needed
	dataFrames := (len(data) + framePayloadSize - 1) / framePayloadSize
	totalFrames := 1 + dataFrames + 1 // metadata + data + padding

	frames := make([][]byte, 0, totalFrames)

	// Frame 0: metadata
	meta := make([]byte, lp2FrameSize)
	meta[0] = frameTypeMetadata
	binary.LittleEndian.PutUint16(meta[1:3], 0) // sequence 0

	// Metadata payload (bytes 3-191)
	payload := meta[3:]
	// Filename (up to 128 bytes, null-terminated)
	nameBytes := []byte(filename)
	if len(nameBytes) > 128 {
		nameBytes = nameBytes[:128]
	}
	copy(payload[0:], nameBytes)
	// File size at offset 128
	binary.LittleEndian.PutUint64(payload[128:136], uint64(len(data)))
	// SHA-256 at offset 136
	copy(payload[136:168], checksum[:])
	// Version at offset 168
	payload[168] = encodingVersion

	frames = append(frames, meta)

	// Data frames
	seq := uint16(1)
	offset := 0
	for offset < len(data) {
		frame := make([]byte, lp2FrameSize)
		frame[0] = frameTypeData
		binary.LittleEndian.PutUint16(frame[1:3], seq)

		end := offset + framePayloadSize
		if end > len(data) {
			end = len(data)
		}
		copy(frame[3:], data[offset:end])

		frames = append(frames, frame)
		offset = end
		seq++
	}

	// Padding frame
	pad := make([]byte, lp2FrameSize)
	pad[0] = frameTypePadding
	binary.LittleEndian.PutUint16(pad[1:3], seq)
	frames = append(frames, pad)

	return frames
}
