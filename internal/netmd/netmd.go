package netmd

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/gousb"
)

type NetMD struct {
	debug bool
	index int
	devs  []*gousb.Device
	ctx   *gousb.Context
	out   *gousb.OutEndpoint
	ekb   *EKB
}

type Encoding byte

type Channels byte

type Control byte

type TrackProt byte

const (
	EncSP  Encoding = 0x90
	EncLP2 Encoding = 0x92
	EncLP4 Encoding = 0x93

	ChanStereo Channels = 0x00
	ChanMono   Channels = 0x01

	ControlRejected Control = 0x0a
	ControlAccepted Control = 0x09
	ControlInterim  Control = 0x0f
	ControlStub     Control = 0x08

	TrackProtected   TrackProt = 0x03
	TrackUnprotected TrackProt = 0x00
)

var (
	ByteArr16 = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
)

func NewNetMD(index int, debug bool) (md *NetMD, err error) {
	md = &NetMD{
		index: index,
		debug: debug,
		ekb:   NewEKB(),
	}

	md.ctx = gousb.NewContext()

	// OpenDevices may return devices AND an error (e.g. permission errors on
	// other USB devices). We only care if we got enough devices.
	md.devs, err = md.ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		for _, d := range Devices {
			if d.deviceId == desc.Product && d.vendorId == desc.Vendor {
				if md.debug {
					log.Printf("Found %s", d.Name)
				}
				return true
			}
		}
		return false
	})

	if len(md.devs) == 0 || len(md.devs) <= md.index {
		if err == nil {
			err = errors.New("no compatible netmd device found or incorrect index")
		}
		md.ctx.Close()
		return nil, err
	}
	// Clear partial errors — we have the device we need
	err = nil

	// Claim the device's configuration and interface, find the OUT endpoint.
	dev := md.devs[md.index]

	// Reset the device to put it in a known state before communication.
	// Without this, control transfers time out on many devices (e.g. MZ-N505).
	if err = dev.Reset(); err != nil {
		if debug {
			log.Printf("Reset failed (non-fatal): %v", err)
		}
	}
	time.Sleep(200 * time.Millisecond)

	// Set a control transfer timeout — google/gousb defaults to 0 (infinite).
	// poll() is called in tight loops so this must be short.
	dev.ControlTimeout = 2 * time.Second

	// Detach any kernel driver so we can claim the interface.
	if err = dev.SetAutoDetach(true); err != nil {
		if debug {
			log.Printf("SetAutoDetach failed (non-fatal): %v", err)
		}
	}
	var foundOut bool
	for num := range dev.Desc.Configs {
		config, cfgErr := dev.Config(num)
		if cfgErr != nil {
			if debug {
				log.Printf("Config(%d) failed: %v", num, cfgErr)
			}
			continue
		}
		for _, desc := range config.Desc.Interfaces {
			intf, intfErr := config.Interface(desc.Number, 0)
			if intfErr != nil {
				if debug {
					log.Printf("Interface(%d, 0) failed: %v", desc.Number, intfErr)
				}
				continue
			}
			for _, endpointDesc := range intf.Setting.Endpoints {
				if endpointDesc.Direction == gousb.EndpointDirectionOut {
					md.out, err = intf.OutEndpoint(endpointDesc.Number)
					if err != nil {
						if debug {
							log.Printf("OutEndpoint(%d) failed: %v", endpointDesc.Number, err)
						}
						continue
					}
					if md.debug {
						log.Printf("OUT endpoint: %s", endpointDesc)
					}
					foundOut = true
				}
			}
			// Do NOT close config here — it invalidates the endpoint handles
		}
		if foundOut {
			break
		}
		config.Close()
	}

	if !foundOut || md.out == nil {
		md.Close()
		return nil, fmt.Errorf("could not find OUT endpoint on device")
	}

	return
}

// DeviceName returns the model name of the connected device.
func (md *NetMD) DeviceName() string {
	if md == nil || len(md.devs) <= md.index {
		return "Unknown"
	}
	desc := md.devs[md.index].Desc
	for _, d := range Devices {
		if d.deviceId == desc.Product && d.vendorId == desc.Vendor {
			return d.Name
		}
	}
	return fmt.Sprintf("NetMD %04x:%04x", desc.Vendor, desc.Product)
}

func (md *NetMD) Close() {
	if md == nil {
		return
	}
	for _, d := range md.devs {
		if d != nil {
			d.Close()
		}
	}
	if md.ctx != nil {
		md.ctx.Close()
	}
}

// Wait makes sure the device is truly finished, needed to prevent crashes on the SHARP IM-DR410/IM-DR420 and the Sony MZ-N420D
func (md *NetMD) Wait() error {
	buf := make([]byte, 4)
	for i := 0; i < 10; i++ {
		c, err := md.devs[md.index].Control(gousb.ControlIn|gousb.ControlVendor|gousb.ControlInterface, 0x01, 0, 0, buf)
		if err != nil {
			return err
		}
		if c != 4 {
			if md.debug {
				log.Println("sync response != 4 bytes")
			}
		} else {
			if bytes.Equal(buf, []byte{0x00, 0x00, 0x00, 0x00}) {
				return nil
			}
		}
		time.Sleep(time.Millisecond * 100)
	}
	return errors.New("no sync response")
}

// RequestDiscCapacity returns the totals in seconds
func (md *NetMD) RequestDiscCapacity() (recorded uint64, total uint64, available uint64, err error) {
	r, err := md.submit(ControlAccepted, []byte{0x18, 0x06, 0x02, 0x10, 0x10, 0x00}, []byte{0x30, 0x80, 0x03, 0x00, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00})
	if err != nil {
		return
	}
	recorded = (hexToInt(r[29]) * 3600) + (hexToInt(r[30]) * 60) + hexToInt(r[31])
	total = (hexToInt(r[35]) * 3600) + (hexToInt(r[36]) * 60) + hexToInt(r[37])
	available = (hexToInt(r[42]) * 3600) + (hexToInt(r[43]) * 60) + hexToInt(r[44])
	return
}

// SetDiscHeader will write  a raw title to the disc
func (md *NetMD) SetDiscHeader(t string) error {
	md.poll()
	o, err := md.RequestDiscHeader()
	if err != nil {
		return err
	}
	j := len(o) // length of old title
	h := len(t) // length of new title
	c := []byte{0x00, 0x00, 0x30, 0x00, 0x0a, 0x00, 0x50, 0x00}
	c = append(c, intToHex16(int16(h))...)
	c = append(c, 0x00, 0x00)
	c = append(c, intToHex16(int16(j))...)
	c = append(c, []byte(t)...)
	_, err = md.submit(ControlAccepted, []byte{0x18, 0x08, 0x10, 0x18, 0x01, 0x01}, []byte{0x00})
	_, err = md.submit(ControlAccepted, []byte{0x18, 0x08, 0x10, 0x18, 0x01, 0x00}, []byte{0x00})
	_, err = md.submit(ControlAccepted, []byte{0x18, 0x08, 0x10, 0x18, 0x01, 0x03}, []byte{0x00})
	_, err = md.submit(ControlAccepted, []byte{0x18, 0x07, 0x02, 0x20, 0x18, 0x01}, c) // actual call
	_, err = md.submit(ControlAccepted, []byte{0x18, 0x08, 0x10, 0x18, 0x01, 0x00}, []byte{0x00})
	if err != nil {
		return err
	}
	return nil
}

// RequestDiscHeader returns the raw title of the disc
func (md *NetMD) RequestDiscHeader() (string, error) {
	r, err := md.submit(ControlAccepted, []byte{0x18, 0x06, 0x02, 0x20, 0x18, 0x01}, []byte{0x00, 0x00, 0x30, 0x00, 0x0a, 0x00, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00})
	if err != nil {
		return "", err
	}
	return string(r[25:]), nil
}

// RecordingParameters current default recording parameters set on the NetMD
func (md *NetMD) RecordingParameters() (encoding Encoding, channels Channels, err error) {
	r, err := md.submit(ControlAccepted, []byte{0x18, 0x09, 0x80, 0x01, 0x03, 0x30}, []byte{0x88, 0x01, 0x00, 0x30, 0x88, 0x05, 0x00, 0x30, 0x88, 0x07, 0x00, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00})
	if err != nil {
		return
	}
	encoding = Encoding(r[34])
	channels = Channels(r[35])
	return
}

// RequestStatus returns known status flags
func (md *NetMD) RequestStatus() (disk bool, err error) {
	//_, err = md.rawCall([]byte{0x00, 0x18, 0x08, 0x80, 0x00, 0x01}, []byte{0x00})
	r, err := md.submit(ControlAccepted, []byte{0x18, 0x09, 0x80, 0x01, 0x02, 0x30}, []byte{0x88, 0x00, 0x00, 0x30, 0x88, 0x04, 0x00, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00})
	if err != nil {
		return
	}
	disk = r[26] == 0x40 // 0x80 no disk
	return
}

func (md *NetMD) RequestTrackCount() (c int, err error) {
	_, err = md.submit(ControlAccepted, []byte{0x18, 0x08, 0x10, 0x10, 0x01, 0x01}, []byte{0x00})
	r, err := md.submit(ControlAccepted, []byte{0x18, 0x06, 0x02, 0x10, 0x10, 0x01}, []byte{0x30, 0x00, 0x10, 0x00, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00})
	if err != nil {
		return
	}
	c = int(hexToInt16(r[23:]))
	return
}

// RequestTrackTitle returns the raw title of the trk number starting from 0
func (md *NetMD) RequestTrackTitle(trk int) (t string, err error) {
	r, err := md.submit(ControlAccepted, []byte{0x18, 0x06, 0x02, 0x20, 0x18, byte(2) & 0xff}, []byte{0x00, byte(trk) & 0xff, 0x30, 0x00, 0x0a, 0x00, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00})
	if err != nil {
		return
	}
	t = string(r[25:])
	return
}

// SetTrackTitle set the title of the trk number starting from 0, isNew can be be true if it's a newadded track
func (md *NetMD) SetTrackTitle(trk int, t string, isNew bool) (err error) {
	j := 0
	if !isNew {
		o, err := md.RequestTrackTitle(trk)
		if err != nil {
			return err
		}
		j = len(o) // length of old title
	}
	h := len(t) // length of new title
	s := []byte{0x00, byte(trk) & 0xff, 0x30, 0x00, 0x0a, 0x00, 0x50, 0x00}
	s = append(s, intToHex16(int16(h))...)
	s = append(s, 0x00, 0x00)
	s = append(s, intToHex16(int16(j))...)
	s = append(s, []byte(t)...)

	if !isNew {
		_, err = md.submit(ControlAccepted, []byte{0x18, 0x08, 0x10, 0x18, 0x02, 0x00}, []byte{0x00})
		_, err = md.submit(ControlAccepted, []byte{0x18, 0x08, 0x10, 0x18, 0x02, 0x03}, []byte{0x00})
	}

	_, err = md.submit(ControlAccepted, []byte{0x18, 0x07, 0x02, 0x20, 0x18, byte(2) & 0xff}, s)

	if !isNew {
		_, err = md.submit(ControlAccepted, []byte{0x18, 0x08, 0x10, 0x18, 0x02, 0x00}, []byte{0x00})
	}

	if err != nil {
		return
	}
	return
}

//
func (md *NetMD) RequestTrackFlag(trk int) (flag TrackProt, err error) {
	s := []byte{0x01, 0x20, 0x10, 0x01}
	s = append(s, intToHex16(int16(trk))...)
	s = append(s, 0xff, 0x00, 0x00, 0x01, 0x00, 0x08)
	d, err := md.submit(ControlAccepted, []byte{0x18, 0x06}, s)
	if err != nil {
		return
	}
	return TrackProt(d[15]), nil
}

// EraseTrack will erase the trk number starting from 0
func (md *NetMD) EraseTrack(trk int) error {
	s := []byte{0xff, 0x01, 0x00, 0x20, 0x10, 0x01}
	s = append(s, intToHex16(int16(trk))...)
	_, err := md.submit(ControlAccepted, []byte{0x18, 0x40}, s)
	if err != nil {
		return err
	}
	return nil
}

// MoveTrack will move the trk number to a new position
func (md *NetMD) MoveTrack(trk, to int) error {
	s := []byte{0xff, 0x00, 0x00, 0x20, 0x10, 0x01}
	s = append(s, intToHex16(int16(trk))...)
	s = append(s, 0x20, 0x10, 0x01)
	s = append(s, intToHex16(int16(to))...)
	_, err := md.submit(ControlAccepted, []byte{0x18, 0x08, 0x10, 0x10, 0x01, 0x00}, []byte{0x00})
	_, err = md.submit(ControlAccepted, []byte{0x18, 0x43}, s)
	if err != nil {
		return err
	}
	return nil
}

// RequestTrackLength returns the duration in seconds of the trk starting from 0
func (md *NetMD) RequestTrackLength(trk int) (duration uint64, err error) {
	s := []byte{0x02, 0x20, 0x10, 0x01}
	s = append(s, intToHex16(int16(trk))...)
	s = append(s, 0x30, 0x00, 0x01, 0x00, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00)
	r, err := md.submit(ControlAccepted, []byte{0x18, 0x06}, s)
	if err != nil {
		return
	}
	duration = (hexToInt(r[27]) * 3600) + (hexToInt(r[28]) * 60) + hexToInt(r[29])
	return
}

// RequestTrackEncoding returns the Encoding of the trk starting from 0
func (md *NetMD) RequestTrackEncoding(trk int) (encoding Encoding, err error) {
	s := append(intToHex16(int16(trk)), 0x30, 0x80, 0x07, 0x00, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00)
	r, err := md.submit(ControlAccepted, []byte{0x18, 0x06, 0x02, 0x20, 0x10, 0x01}, s)
	if err != nil {
		return
	}
	return Encoding(r[len(r)-2]), nil
}

// submit will submit the `check + payload` wait for replies matching the `check` and `control`
func (md *NetMD) submit(control Control, check []byte, payload []byte) ([]byte, error) {
	i := []byte{0x00}
	i = append(i, check...)
	i = append(i, payload...)
	md.poll()
	if md.debug {
		log.Printf("<- sending data: % x", i)
	}
	if _, err := md.devs[md.index].Control(gousb.ControlOut|gousb.ControlVendor|gousb.ControlInterface, 0x80, 0, 0, i); err != nil {
		return nil, err
	}
	return md.receive(control, check, nil)
}

func (md *NetMD) receive(control Control, check []byte, c chan Transfer) ([]byte, error) {
	for tries := 0; tries < 300; tries++ {
		if c != nil {
			c <- Transfer{
				Type: TtPoll,
			}
		}
		if h := md.poll(); h != -1 {
			recv := make([]byte, h)
			if _, err := md.devs[md.index].Control(gousb.ControlIn|gousb.ControlVendor|gousb.ControlInterface, 0x81, 0, 0, recv); err != nil {
				return nil, err
			}
			chkLen := len(check) + 1
			if bytes.Equal(recv[1:len(check)+1], check) {
				ctrl := Control(recv[0])
				if md.debug {
					log.Printf("-> incoming data matched check: % x", recv[1:chkLen])
					if ctrl == ControlAccepted || ctrl == ControlInterim {
						log.Printf("-> payload: % x", recv[chkLen:])
					}
				}
				switch ctrl {
				case ControlAccepted:
					if ctrl == control {
						return recv, nil
					} else if md.debug {
						log.Printf("!! skipped accepted call: % x", recv[chkLen:])
					}
				case ControlInterim:
					if ctrl == control {
						return recv, nil
					} else if md.debug {
						log.Printf("?? skipped interim call: % x", recv[chkLen:])
					}
				case ControlRejected:
					return nil, errors.New("!! submit was rejected")
				case ControlStub:
					if md.debug {
						log.Printf("?? not implemented: % x", recv[chkLen:])
					}
					return recv, nil
				}
			} else {
				if md.debug {
					log.Printf("-> !! incoming data: % x did not match check: % x", recv[1:chkLen], check)
				}
			}
		}
		time.Sleep(time.Millisecond * 100)
	}
	return nil, errors.New("no data matched check, timed out")
}

func (md *NetMD) poll() int {
	buf := make([]byte, 4)
	md.devs[md.index].Control(gousb.ControlIn|gousb.ControlVendor|gousb.ControlInterface, 0x01, 0, 0, buf)
	if buf[0] == 0x01 {
		h := int(buf[2]) | int(buf[3])<<8
		if md.debug && h > 255 {
			log.Printf("poll: buf=%x size=%d (16-bit)", buf, h)
		}
		return h
	}
	return -1
}
