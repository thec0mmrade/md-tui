package netmd

import (
	"fmt"
	"log"
)

type TransferType string

type Transfer struct {
	Type        TransferType
	Track       int
	Transferred int
	Error       error
}

const (
	TtSetup TransferType = "setup"
	TtSend  TransferType = "send"
	TtPoll  TransferType = "poll"
	TtTrack TransferType = "track"
)

// Send will transmit the Track data encrypted (concurrent) to the NetMD
// the Transfer channel will receive different TransferType's so the process can be followed closely from a different process
func (md *NetMD) Send(trk *Track, c chan Transfer) {
	if c == nil {
		return
	}
	defer close(c)

	// housekeeping — errors ignored as some commands are device-specific
	md.leaveSecureSession()
	md.acquire() // Sharp-specific, may be rejected on Sony
	md.trackProtection(0x01)

	c <- Transfer{
		Type: TtSetup,
	}

	md.enterSecureSession()
	md.Wait()

	md.sendKeyData()
	md.Wait()

	md.sessionKeyExchange()
	sessionKey, _ := md.ekb.RetailMAC()
	md.kekExchange(sessionKey)
	md.Wait()

	err := md.startSecureSend(trk.Format, trk.DiscFormat, trk.Frames, trk.TotalBytes())
	if err != nil {
		c <- Transfer{
			Error: fmt.Errorf("startSecureSend: %w", err),
		}
		return
	}

	key, _ := DESDecrypt(trk.key, md.ekb.kek)
	dataCounter := 0

	c <- Transfer{
		Type:        TtSend,
		Transferred: dataCounter,
	}

	for i, p := range trk.Packets {
		s := make([]byte, 0)
		if p.first {
			s = append(s, intToHex64(int64(trk.Frames*FrameSize[trk.Format]))...)
			s = append(s, key...)
			s = append(s, md.ekb.iv...)
		}
		s = append(s, p.data...)
		t, err := md.out.Write(s)
		if err != nil {
			c <- Transfer{
				Error: err,
			}
			close(c)
			return
		}
		dataCounter += t
		c <- Transfer{
			Type:        TtSend,
			Transferred: dataCounter,
		}
		if md.debug {
			log.Printf("Packet %d / %d Transmitted: %d bytes", i, len(trk.Packets), dataCounter)
		}
	}

	if md.debug {
		log.Println("Going to wait for MD to finish data write...")
	}

	r, err := md.finishSecureSend(c)
	if err != nil {
		c <- Transfer{
			Error: fmt.Errorf("finishSecureSend: %w", err),
		}
		return
	}

	j := hexToInt16(r[17:19])
	trackNr := int(j)
	if md.debug {
		log.Printf("track %d committed", trackNr)
	}

	c <- Transfer{
		Type:  TtTrack,
		Track: trackNr,
	}

	md.Wait()

	err = md.cacheTOC()
	if err != nil {
		c <- Transfer{
			Error: fmt.Errorf("cacheTOC: %w", err),
		}
		return
	}

	err = md.SetTrackTitle(trackNr, trk.Title, true)
	if err != nil {
		c <- Transfer{
			Error: fmt.Errorf("setTrackTitle: %w", err),
		}
		return
	}

	err = md.syncTOC()
	if err != nil {
		c <- Transfer{
			Error: fmt.Errorf("syncTOC: %w", err),
		}
		return
	}

	md.Wait()

	err = md.commitTrack(trackNr, sessionKey)
	if err != nil {
		c <- Transfer{
			Error: fmt.Errorf("commitTrack: %w", err),
		}
		return
	}

	md.forgetSecureKey()
	md.leaveSecureSession()
	md.release()

	return
}
