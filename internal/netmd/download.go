package netmd

import (
	"fmt"
	"log"
	"os"
)

type DownloadProgress struct {
	Sector       int
	TotalSectors int
	Phase        string
	Pass         int // -1 for single-pass
	TotalPasses  int // 0 for single-pass
}

// DownloadTrack reads a track's ATRAC data from the disc using the exploit mechanism.
//
// Sequence matches Web MiniDisc Pro pcap:
// 1. Stop → Factory mode → Firmware reads
// 2. GotoTrack → SeekToStart (disc starts spinning, fills cache)
// 3. PatchFirmware (cache fills during patching)
// 4. Read sectors immediately — no delays, disc keeps spinning
func (md *NetMD) DownloadTrack(trackIndex int, totalSectors int, encoding Encoding, progress chan<- DownloadProgress) ([]byte, error) {
	if progress != nil {
		defer close(progress)
	}

	// NoRam exploit (cache-limited to ~76 sectors for now)
	// CachedSectorControlDownload is not yet working — the USB read handler
	// at 0x574fc triggers sending but doesn't control response content.
	// See docs/firmware-dump-research.md for details.
	md.Stop()
	md.Wait()

	// Enter factory mode first (before playback)
	if md.debug {
		log.Println("Entering factory mode...")
	}
	if progress != nil {
		progress <- DownloadProgress{Phase: "initializing", Pass: -1}
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

	// GotoTrack starts the disc spinning and caching
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

	// Patch firmware while disc fills the cache
	if md.debug {
		log.Println("Patching firmware...")
	}
	if progress != nil {
		progress <- DownloadProgress{Phase: "patching", Pass: -1}
	}
	if err := md.PatchFirmware(); err != nil {
		return nil, fmt.Errorf("firmware patch failed: %w", err)
	}

	// Read sectors from the cache
	if md.debug {
		log.Printf("Reading %d sectors...", totalSectors)
	}

	var allData []byte
	emptySectors := 0
	for sector := 0; sector < totalSectors; sector++ {
		if progress != nil {
			progress <- DownloadProgress{
				Sector:       sector,
				TotalSectors: totalSectors,
				Phase:        "reading",
				Pass:         -1,
			}
		}

		data, err := md.ExploitReadSector(uint16(sector))
		if err != nil {
			return nil, fmt.Errorf("read sector %d: %w", sector, err)
		}

		if isEmptySector(data) {
			emptySectors++
			if emptySectors >= 3 {
				if md.debug {
					log.Printf("3 consecutive empty sectors at %d, stopping early", sector)
				}
				break
			}
		} else {
			emptySectors = 0
		}

		allData = append(allData, data...)
	}

	if md.debug {
		log.Printf("Download complete: %d bytes from %d sectors", len(allData), len(allData)/sectorSize)
	}

	return allData, nil
}

// downloadControlTransfer uses CachedSectorControlDownload for full-speed
// sequential sector reads. No cache window limits.
func (md *NetMD) downloadControlTransfer(trackIndex int, totalSectors int, progress chan<- DownloadProgress) ([]byte, error) {
	md.Stop()
	md.Wait()

	if md.debug {
		log.Println("Entering factory mode...")
	}
	if progress != nil {
		progress <- DownloadProgress{Phase: "initializing", Pass: -1}
	}
	if err := md.EnterFactoryMode(); err != nil {
		return nil, fmt.Errorf("factory mode: %w", err)
	}

	// Navigate to track — disc starts caching
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

	// Install resident code and patch USB handlers
	if md.debug {
		log.Println("Setting up CachedSectorControlDownload...")
	}
	if progress != nil {
		progress <- DownloadProgress{Phase: "patching", Pass: -1}
	}
	if err := md.SetupControlDownload(); err != nil {
		return nil, fmt.Errorf("setup control download: %w", err)
	}

	// Start playback so disc feeds the cache
	if err := md.Play(); err != nil {
		if md.debug {
			log.Printf("Play: %v (non-fatal)", err)
		}
	}

	// Enable sector reading from sector 0
	md.EnableSectorReading(0)

	// Read sectors sequentially
	if md.debug {
		log.Printf("Reading %d sectors via control transfer...", totalSectors)
	}

	var allData []byte
	emptySectors := 0
	for sector := 0; sector < totalSectors; sector++ {
		if progress != nil {
			progress <- DownloadProgress{
				Sector:       sector,
				TotalSectors: totalSectors,
				Phase:        "reading",
				Pass:         -1,
			}
		}

		data, err := md.ReadSectorControl()
		if err != nil {
			if md.debug {
				log.Printf("Sector %d read error: %v", sector, err)
			}
			// Try to continue
			allData = append(allData, make([]byte, sectorSize)...)
			continue
		}

		if isEmptySector(data) {
			emptySectors++
			if emptySectors >= 3 {
				if md.debug {
					log.Printf("3 consecutive empty sectors at %d, stopping", sector)
				}
				break
			}
		} else {
			emptySectors = 0
		}

		allData = append(allData, data...)
	}

	// Disable sector reading (restore normal USB)
	md.DisableSectorReading()

	if md.debug {
		log.Printf("Download complete: %d bytes from %d sectors", len(allData), len(allData)/sectorSize)
	}

	return allData, nil
}

func fillRatePerSec(enc Encoding) float64 {
	switch enc {
	case EncLP2:
		return 9.0
	case EncLP4:
		return 5.0
	default:
		return 18.0
	}
}

func isEmptySector(data []byte) bool {
	for _, b := range data {
		if b != 0 {
			return false
		}
	}
	return true
}

// WriteRawFile writes raw ATRAC data to a file.
func WriteRawFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

// EstimateSectors estimates the number of sectors for a track based on duration and encoding.
func EstimateSectors(durationSec int, encoding Encoding) int {
	var sectors int
	switch encoding {
	case EncLP2:
		sectors = durationSec * 9
	case EncLP4:
		sectors = durationSec * 5
	default:
		sectors = durationSec * 18
	}
	if sectors < 1 {
		sectors = 1
	}
	return sectors
}

