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
// The data must be pre-padded to a multiple of lp2FrameSize.
func BuildATRAC3WAV(frameData []byte) []byte {
	fmtChunkSize := 16
	dataChunkSize := len(frameData)
	riffSize := 4 + (8 + fmtChunkSize) + (8 + dataChunkSize) // "WAVE" + fmt chunk + data chunk

	buf := make([]byte, 12+8+fmtChunkSize+8+dataChunkSize)
	w := buf

	// RIFF header
	copy(w[0:4], "RIFF")
	binary.LittleEndian.PutUint32(w[4:8], uint32(riffSize))
	copy(w[8:12], "WAVE")

	// fmt chunk
	copy(w[12:16], "fmt ")
	binary.LittleEndian.PutUint32(w[16:20], uint32(fmtChunkSize))
	binary.LittleEndian.PutUint16(w[20:22], atrac3FormatTag)       // wFormatTag
	binary.LittleEndian.PutUint16(w[22:24], atrac3Channels)        // nChannels
	binary.LittleEndian.PutUint32(w[24:28], atrac3SampleRate)      // nSamplesPerSec
	binary.LittleEndian.PutUint32(w[28:32], atrac3SampleRate*4)    // nAvgBytesPerSec (approximate)
	binary.LittleEndian.PutUint16(w[32:34], atrac3BlockAlign)      // nBlockAlign (track.go checks this == 384)
	binary.LittleEndian.PutUint16(w[34:36], 0)                     // wBitsPerSample

	// data chunk
	copy(w[36:40], "data")
	binary.LittleEndian.PutUint32(w[40:44], uint32(dataChunkSize))
	copy(w[44:], frameData)

	return buf
}
