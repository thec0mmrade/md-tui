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

	// Stop any current playback
	md.Stop()
	md.Wait()

	// Enter factory mode (required for exploit commands)
	if md.debug {
		log.Println("Entering factory mode...")
	}
	if progress != nil {
		progress <- DownloadProgress{Phase: "initializing"}
	}
	if err := md.EnterFactoryMode(); err != nil {
		return nil, fmt.Errorf("factory mode init: %w", err)
	}

	// Read firmware block (initializes exploit state on device)
	if md.debug {
		log.Println("Reading firmware block...")
	}
	_, err := md.ReadFirmwareBlock(0x0000, 0x0930)
	if err != nil {
		return nil, fmt.Errorf("firmware read failed: %w", err)
	}

	// Navigate to track — must happen before firmware patching
	if md.debug {
		log.Printf("Navigating to track %d...", trackIndex)
	}
	if err := md.GotoTrack(trackIndex); err != nil {
		return nil, fmt.Errorf("goto track %d: %w", trackIndex, err)
	}
	if err := md.SeekToStart(); err != nil {
		if md.debug {
			log.Printf("SeekToStart: %v (non-fatal)", err)
		}
	}

	// Patch firmware to enable exploit-based sector reading
	if md.debug {
		log.Println("Patching firmware...")
	}
	if progress != nil {
		progress <- DownloadProgress{Phase: "patching"}
	}
	if err := md.PatchFirmware(); err != nil {
		return nil, fmt.Errorf("firmware patch failed: %w", err)
	}

	// Wait for disc to spin up and cache sectors
	time.Sleep(2 * time.Second)

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
