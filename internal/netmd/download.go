package netmd

import (
	"fmt"
	"log"
	"os"
	"time"
)

type DownloadProgress struct {
	Sector     int
	TotalSectors int
	Phase      string
}

// DownloadTrack reads a track's ATRAC data from the disc using the exploit mechanism.
// Returns raw sector data. The caller is responsible for writing it to a file.
func (md *NetMD) DownloadTrack(trackIndex int, totalSectors int, progress chan<- DownloadProgress) ([]byte, error) {
	if progress != nil {
		defer close(progress)
	}

	// Step 1: Fill the disc cache BEFORE factory mode.
	// The exploit reads ATRAC data from the anti-shock DRAM buffer.
	// Factory mode operations may prevent normal playback, so we fill
	// the cache first while the device is in normal operating mode.
	md.Stop()
	md.Wait()

	if md.debug {
		log.Printf("Navigating to track %d and starting playback...", trackIndex)
	}
	if progress != nil {
		progress <- DownloadProgress{Phase: "initializing"}
	}
	if err := md.GotoTrack(trackIndex); err != nil {
		return nil, fmt.Errorf("goto track %d: %w", trackIndex, err)
	}
	if err := md.SeekToStart(); err != nil {
		if md.debug {
			log.Printf("SeekToStart: %v (non-fatal)", err)
		}
	}
	if err := md.Play(); err != nil {
		if md.debug {
			log.Printf("Play: %v (non-fatal)", err)
		}
	}

	// Wait for disc to spin up and fill the anti-shock buffer with ATRAC data
	if md.debug {
		log.Println("Waiting for disc cache to fill...")
	}
	time.Sleep(5 * time.Second)

	// Stop playback before entering factory mode. The disc cache (DRAM)
	// persists after stopping — but the patched firmware crashes if
	// playback is active during exploit commands.
	md.Stop()
	md.Wait()

	// Step 2: Enter factory mode and patch firmware for exploit reads.
	if md.debug {
		log.Println("Entering factory mode...")
	}
	if err := md.EnterFactoryMode(); err != nil {
		return nil, fmt.Errorf("factory mode init: %w", err)
	}

	if md.debug {
		log.Println("Reading firmware block...")
	}
	_, err := md.ReadFirmwareBlock(0x0000, 0x0930)
	if err != nil {
		return nil, fmt.Errorf("firmware read failed: %w", err)
	}

	if md.debug {
		log.Println("Patching firmware...")
	}
	if progress != nil {
		progress <- DownloadProgress{Phase: "patching"}
	}
	if err := md.PatchFirmware(); err != nil {
		return nil, fmt.Errorf("firmware patch failed: %w", err)
	}

	// Brief settle after patching
	time.Sleep(500 * time.Millisecond)

	// Read sectors
	if md.debug {
		log.Printf("Reading %d sectors...", totalSectors)
	}

	var allData []byte
	for sector := 0; sector < totalSectors; sector++ {
		if progress != nil {
			progress <- DownloadProgress{
				Sector:       sector,
				TotalSectors: totalSectors,
				Phase:        "reading",
			}
		}

		data, err := md.ExploitReadSector(uint16(sector))
		if err != nil {
			return nil, fmt.Errorf("read sector %d: %w", sector, err)
		}
		allData = append(allData, data...)
	}

	if md.debug {
		log.Printf("Download complete: %d bytes from %d sectors", len(allData), totalSectors)
	}

	return allData, nil
}

// WriteRawFile writes raw ATRAC data to a file.
func WriteRawFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

// EstimateSectors estimates the number of sectors for a track based on duration and encoding.
// LP2: ~6 sectors per second, SP: ~18 sectors per second
func EstimateSectors(durationSec int, encoding Encoding) int {
	switch encoding {
	case EncLP2:
		return durationSec * 6
	case EncLP4:
		return durationSec * 3
	default: // SP
		return durationSec * 18
	}
}
