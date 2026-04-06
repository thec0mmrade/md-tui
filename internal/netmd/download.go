package netmd

import (
	"fmt"
	"log"
	"math"
	"os"
	"time"
)

// Cache and chunking constants (empirically determined on MZ-N505).
const (
	cacheCapacitySectors = 76 // ~175KB, tested limit
	safeSectorsPerPass   = 64 // 85% of cache, margin for timing imprecision
)

type DownloadProgress struct {
	Sector       int
	TotalSectors int
	Phase        string
	Pass         int // current pass (0-based), -1 for single-pass
	TotalPasses  int // 0 for single-pass
}

// DownloadTrack reads a track's ATRAC data from the disc using the exploit mechanism.
// For small tracks (≤76 sectors), uses single-pass cache fill.
// For large tracks, uses multi-pass chunked reading.
func (md *NetMD) DownloadTrack(trackIndex int, totalSectors int, encoding Encoding, progress chan<- DownloadProgress) ([]byte, error) {
	if progress != nil {
		defer close(progress)
	}

	if totalSectors > cacheCapacitySectors {
		return md.downloadChunked(trackIndex, totalSectors, encoding, progress)
	}
	return md.downloadSinglePass(trackIndex, totalSectors, progress)
}

// downloadSinglePass is the original download flow for small tracks.
func (md *NetMD) downloadSinglePass(trackIndex int, totalSectors int, progress chan<- DownloadProgress) ([]byte, error) {
	md.Stop()
	md.Wait()

	if md.debug {
		log.Printf("Navigating to track %d and starting playback...", trackIndex)
	}
	if progress != nil {
		progress <- DownloadProgress{Phase: "initializing", Pass: -1}
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

	if md.debug {
		log.Println("Waiting for disc cache to fill...")
	}
	time.Sleep(5 * time.Second)

	md.Stop()
	md.Wait()

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
		progress <- DownloadProgress{Phase: "patching", Pass: -1}
	}
	if err := md.PatchFirmware(); err != nil {
		return nil, fmt.Errorf("firmware patch failed: %w", err)
	}

	time.Sleep(500 * time.Millisecond)

	return md.readSectors(0, totalSectors, totalSectors, -1, 0, progress)
}

// downloadChunked reads large tracks in multiple passes, sliding the cache
// window forward each time.
//
// Flow:
// 1. GotoTrack → SeekToStart → Play → wait (fill initial cache) → Pause
// 2. EnterFactoryMode → PatchFirmware (once)
// 3. Read first batch of sectors from cache
// 4. Loop: Play (resume) → wait (advance cache) → Pause → read next batch
func (md *NetMD) downloadChunked(trackIndex int, totalSectors int, encoding Encoding, progress chan<- DownloadProgress) ([]byte, error) {
	fillRate := fillRatePerSec(encoding)
	totalPasses := int(math.Ceil(float64(totalSectors) / float64(safeSectorsPerPass)))

	if md.debug {
		log.Printf("Chunked download: %d sectors, %d passes, fill rate %.1f sectors/sec",
			totalSectors, totalPasses, fillRate)
	}

	// Step 1: Navigate and start playback
	md.Stop()
	md.Wait()

	if progress != nil {
		progress <- DownloadProgress{Phase: "initializing", Pass: 0, TotalPasses: totalPasses}
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
		return nil, fmt.Errorf("play: %w", err)
	}

	// Wait briefly for disc spin-up and initial caching.
	// Don't wait for full cache fill — for long tracks the disc reads ahead
	// fast and early sectors would be evicted from the circular cache.
	// 3 seconds is enough for spin-up + ~27 sectors at LP2 rate.
	if md.debug {
		log.Println("Waiting for disc spin-up...")
	}
	time.Sleep(3 * time.Second)

	// Pause to freeze cache state (maintains disc position)
	if err := md.Pause(); err != nil {
		if md.debug {
			log.Printf("Pause: %v (non-fatal, trying Stop)", err)
		}
		md.Stop()
		md.Wait()
	}

	// Step 2: Enter factory mode and patch firmware (once)
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
		progress <- DownloadProgress{Phase: "patching", Pass: 0, TotalPasses: totalPasses}
	}
	if err := md.PatchFirmware(); err != nil {
		return nil, fmt.Errorf("firmware patch failed: %w", err)
	}
	time.Sleep(500 * time.Millisecond)

	// Step 3: Read sectors in passes using Pause/Play resume.
	// Works well for medium files (~250KB). For very large files,
	// cumulative drift causes missed sectors — use CachedSectorControlDownload
	// exploit variant for full-speed sequential reads (TODO).
	var allData []byte
	sector := 0
	pass := 0
	emptySectors := 0

	for sector < totalSectors {
		batchEnd := sector + safeSectorsPerPass
		if batchEnd > totalSectors {
			batchEnd = totalSectors
		}

		if md.debug {
			log.Printf("Pass %d/%d: reading sectors %d-%d", pass+1, totalPasses, sector, batchEnd-1)
		}

		// Read this batch
		for s := sector; s < batchEnd; s++ {
			if progress != nil {
				progress <- DownloadProgress{
					Sector:       s,
					TotalSectors: totalSectors,
					Phase:        "reading",
					Pass:         pass,
					TotalPasses:  totalPasses,
				}
			}

			data, err := md.ExploitReadSector(uint16(s))
			if err != nil {
				return nil, fmt.Errorf("read sector %d: %w", s, err)
			}

			if isEmptySector(data) {
				emptySectors++
				if emptySectors >= 3 {
					if md.debug {
						log.Printf("3 consecutive empty sectors at %d, stopping early", s)
					}
					goto done
				}
			} else {
				emptySectors = 0
			}

			allData = append(allData, data...)
		}

		sector = batchEnd
		pass++

		// If more sectors remain, resume playback to advance cache window
		if sector < totalSectors {
			if md.debug {
				log.Printf("Advancing cache: Play → wait → Pause...")
			}

			if err := md.Play(); err != nil {
				if md.debug {
					log.Printf("Play: %v (non-fatal)", err)
				}
			}

			fillTime := float64(safeSectorsPerPass)/fillRate + 1.0
			time.Sleep(time.Duration(fillTime*1000) * time.Millisecond)

			if err := md.Pause(); err != nil {
				if md.debug {
					log.Printf("Pause: %v (non-fatal)", err)
				}
			}
			time.Sleep(500 * time.Millisecond)
		}
	}

done:
	if md.debug {
		log.Printf("Download complete: %d bytes from %d sectors (%d passes)",
			len(allData), sector, pass)
	}

	return allData, nil
}

// readSectors reads a range of sectors and reports progress.
func (md *NetMD) readSectors(start, end, totalSectors, pass, totalPasses int, progress chan<- DownloadProgress) ([]byte, error) {
	if md.debug {
		log.Printf("Reading %d sectors...", end-start)
	}

	var allData []byte
	for s := start; s < end; s++ {
		if progress != nil {
			progress <- DownloadProgress{
				Sector:       s,
				TotalSectors: totalSectors,
				Phase:        "reading",
				Pass:         pass,
				TotalPasses:  totalPasses,
			}
		}

		data, err := md.ExploitReadSector(uint16(s))
		if err != nil {
			return nil, fmt.Errorf("read sector %d: %w", s, err)
		}
		allData = append(allData, data...)
	}

	if md.debug {
		log.Printf("Download complete: %d bytes from %d sectors", len(allData), end-start)
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
