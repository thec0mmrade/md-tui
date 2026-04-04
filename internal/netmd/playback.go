package netmd

import (
	"fmt"

	"github.com/google/gousb"
)

// Playback control commands — standard NetMD AV/C protocol

func (md *NetMD) Play() error {
	return md.playbackCommand([]byte{0x18, 0xc3}, 0xff, []byte{0x75, 0x00, 0x00, 0x00}, 0x00)
}

func (md *NetMD) Pause() error {
	return md.playbackCommand([]byte{0x18, 0xc3}, 0xff, []byte{0x7d, 0x00, 0x00, 0x00}, 0x00)
}

func (md *NetMD) Stop() error {
	return md.playbackCommand([]byte{0x18, 0xc5}, 0xff, []byte{0x00, 0x00, 0x00, 0x00}, 0x00)
}

// playbackCommand sends a command with one check byte value but expects
// a different value in the response (e.g. send 0xff, response has 0x00).
func (md *NetMD) playbackCommand(prefix []byte, sendByte byte, payload []byte, recvByte byte) error {
	cmd := []byte{0x00}
	cmd = append(cmd, prefix...)
	cmd = append(cmd, sendByte)
	cmd = append(cmd, payload...)
	md.poll()
	if _, err := md.devs[md.index].Control(gousb.ControlOut|gousb.ControlVendor|gousb.ControlInterface, 0x80, 0, 0, cmd); err != nil {
		return err
	}
	check := make([]byte, len(prefix)+1)
	copy(check, prefix)
	check[len(prefix)] = recvByte
	_, err := md.receive(ControlAccepted, check, nil)
	return err
}

func (md *NetMD) GotoTrack(track int) error {
	// Send with 0xff, but device responds with 0x00 in that byte
	cmd := []byte{0x00, 0x18, 0x50, 0xff, 0x01, 0x00, 0x00, 0x00, 0x00}
	cmd = append(cmd, intToHex16(int16(track))...)
	md.poll()
	if _, err := md.devs[md.index].Control(gousb.ControlOut|gousb.ControlVendor|gousb.ControlInterface, 0x80, 0, 0, cmd); err != nil {
		return err
	}
	_, err := md.receive(ControlAccepted, []byte{0x18, 0x50, 0x00}, nil)
	return err
}

// SeekToStart seeks to the beginning of the current track.
// Sends the 15-byte seek command: 18 50 ff 00 00 00 00 00 00 00 00 00 00 00
func (md *NetMD) SeekToStart() error {
	cmd := []byte{0x00, 0x18, 0x50, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	md.poll()
	if _, err := md.devs[md.index].Control(gousb.ControlOut|gousb.ControlVendor|gousb.ControlInterface, 0x80, 0, 0, cmd); err != nil {
		return err
	}
	_, err := md.receive(ControlAccepted, []byte{0x18, 0x50, 0x00}, nil)
	return err
}

// GetPosition returns the current playback position.
func (md *NetMD) GetPosition() (track, hours, minutes, seconds, frames int, err error) {
	r, err := md.submit(ControlAccepted,
		[]byte{0x18, 0x09, 0x80, 0x01},
		[]byte{0x04, 0x30, 0x88, 0x02, 0x00, 0x30, 0x88, 0x05, 0x00, 0x30, 0x00, 0x03, 0x00, 0x30, 0x00, 0x02, 0x00, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00})
	if err != nil {
		return
	}
	// Response structure (from end):
	// Last 7 bytes: [track_h][track_l][hours][minutes][seconds][frames_h][frames_l]
	n := len(r)
	if n < 7 {
		err = fmt.Errorf("position response too short: %d bytes", n)
		return
	}
	track = int(hexToInt16(r[n-7 : n-5]))
	hours = int(r[n-5])
	minutes = int(r[n-4])
	seconds = int(r[n-3])
	frames = int(hexToInt16(r[n-2 : n]))
	return
}

// GetDeviceFirmware returns the firmware/codec version string from the device.
// This is used to determine exploit compatibility.
func (md *NetMD) GetDeviceFirmware() (string, error) {
	r, err := md.submit(ControlAccepted,
		[]byte{0x18, 0x09, 0x80, 0x01, 0x03, 0x30},
		[]byte{0x88, 0x01, 0x00, 0x30, 0x88, 0x05, 0x00, 0x30, 0x88, 0x07, 0x00, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00})
	if err != nil {
		return "", err
	}
	if len(r) < 36 {
		return "", fmt.Errorf("firmware response too short: %d bytes", len(r))
	}
	return fmt.Sprintf("enc=0x%02x ch=0x%02x", r[34], r[35]), nil
}
