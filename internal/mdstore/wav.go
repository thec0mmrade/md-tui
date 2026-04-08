package mdstore

import "encoding/binary"

// ATRAC3 WAV format constants matching what track.go expects.
const (
	atrac3FormatTag  = 624   // 0x0270
	atrac3Channels   = 2     // stereo
	atrac3SampleRate = 44100
	atrac3BlockAlign = 384   // nBlockAlign — checked by track.go:93
	lp2FrameSize     = 192   // bytes per LP2 frame
)

// BuildATRAC3WAV wraps raw frame data in a valid ATRAC3 WAV container.
// Uses WAVEFORMATEX with 14-byte extradata required by ffmpeg's ATRAC3 decoder.
// The data must be pre-padded to a multiple of lp2FrameSize.
func BuildATRAC3WAV(frameData []byte) []byte {
	// WAVEFORMATEX: 18 bytes base + 14 bytes ATRAC3 extradata = 32 bytes
	extraDataSize := 14
	fmtChunkSize := 18 + extraDataSize // WAVEFORMATEX with cbSize + extradata
	dataChunkSize := len(frameData)
	riffSize := 4 + (8 + fmtChunkSize) + (8 + dataChunkSize)

	buf := make([]byte, 12+8+fmtChunkSize+8+dataChunkSize)
	w := buf

	// RIFF header
	copy(w[0:4], "RIFF")
	binary.LittleEndian.PutUint32(w[4:8], uint32(riffSize))
	copy(w[8:12], "WAVE")

	// fmt chunk (WAVEFORMATEX)
	copy(w[12:16], "fmt ")
	binary.LittleEndian.PutUint32(w[16:20], uint32(fmtChunkSize))
	binary.LittleEndian.PutUint16(w[20:22], atrac3FormatTag)       // wFormatTag = 0x0270
	binary.LittleEndian.PutUint16(w[22:24], atrac3Channels)        // nChannels = 2
	binary.LittleEndian.PutUint32(w[24:28], atrac3SampleRate)      // nSamplesPerSec = 44100
	binary.LittleEndian.PutUint32(w[28:32], 16537)                 // nAvgBytesPerSec (132296 bps / 8)
	binary.LittleEndian.PutUint16(w[32:34], atrac3BlockAlign)      // nBlockAlign = 384
	binary.LittleEndian.PutUint16(w[34:36], 0)                     // wBitsPerSample
	binary.LittleEndian.PutUint16(w[36:38], uint16(extraDataSize)) // cbSize = 14

	// ATRAC3 extradata (14 bytes):
	// Matches Sony OMA/AT3 format for joint stereo LP2
	w[38] = 1    // unknown (always 1)
	w[39] = 0    // unknown
	binary.LittleEndian.PutUint32(w[40:44], 0x0800) // samples per block = 2048
	binary.LittleEndian.PutUint16(w[44:46], 0)      // unknown
	binary.LittleEndian.PutUint16(w[46:48], 1)      // coding mode: 1 = joint stereo
	binary.LittleEndian.PutUint16(w[48:50], 1)      // coding mode duplicate
	binary.LittleEndian.PutUint16(w[50:52], 0)      // padding

	// data chunk
	copy(w[52:56], "data")
	binary.LittleEndian.PutUint32(w[56:60], uint32(dataChunkSize))
	copy(w[60:], frameData)

	return buf
}
