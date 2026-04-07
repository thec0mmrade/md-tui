package device

import (
    "bufio"
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "time"

    "github.com/c0mmrade/md-tui/internal/mdstore"
    netmd "github.com/c0mmrade/md-tui/internal/netmd"
)

type NetMDService struct {
    debug     bool
    md        *netmd.NetMD
    connected bool
    name      string
    // Hold the scanned device so we don't close+reopen
    pendingMD    *netmd.NetMD
    pendingIndex int
}

func NewNetMDService(debug bool) *NetMDService {
    return &NetMDService{debug: debug}
}

func (s *NetMDService) Scan() (devices []DeviceInfo, err error) {
    // go-netmd-lib / gousb can panic if libusb is missing or USB fails
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("USB initialization failed (is libusb-1.0 installed?): %v", r)
            devices = nil
        }
    }()

    // Close any previously held pending connection
    if s.pendingMD != nil {
        s.pendingMD.Close()
        s.pendingMD = nil
    }

    // Try to open device at index 0. NewNetMD opens all matching devices
    // internally, so if this succeeds there's at least one device.
    // We keep it open to avoid close+reopen issues with gousb.
    md, e := netmd.NewNetMD(0, s.debug)
    if e != nil {
        // No devices found
        return nil, nil
    }

    s.pendingMD = md
    s.pendingIndex = 0
    devices = append(devices, DeviceInfo{
        Index: 0,
        Name:  md.DeviceName(),
    })
    return devices, nil
}

func (s *NetMDService) Connect(index int) (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("USB connection failed: %v", r)
        }
    }()

    var md *netmd.NetMD

    // Reuse the pending connection from Scan if indices match
    if s.pendingMD != nil && s.pendingIndex == index {
        md = s.pendingMD
        s.pendingMD = nil
    } else {
        // Close pending if different index
        if s.pendingMD != nil {
            s.pendingMD.Close()
            s.pendingMD = nil
        }
        var e error
        md, e = netmd.NewNetMD(index, s.debug)
        if e != nil {
            return fmt.Errorf("failed to connect to device %d: %w", index, e)
        }
    }

    // Ensure device is ready before issuing commands
    _ = md.Wait()

    // Check disc is present
    if diskPresent, e := md.RequestStatus(); e == nil && !diskPresent {
        md.Close()
        return fmt.Errorf("no disc in device")
    }

    s.md = md
    s.connected = true
    s.name = s.md.DeviceName()
    return nil
}

func (s *NetMDService) Close() {
    if s.md != nil {
        s.md.Close()
        s.md = nil
    }
    if s.pendingMD != nil {
        s.pendingMD.Close()
        s.pendingMD = nil
    }
    s.connected = false
}

func (s *NetMDService) DeviceName() string {
    return s.name
}

func (s *NetMDService) IsConnected() bool {
    return s.connected
}

func (s *NetMDService) ListContent() (*Disc, error) {
    if s.md == nil {
        return nil, fmt.Errorf("not connected")
    }

    trackCount, err := s.md.RequestTrackCount()
    if err != nil {
        return nil, fmt.Errorf("failed to get track count: %w", err)
    }

    disc := &Disc{}

    header, err := s.md.RequestDiscHeader()
    if err == nil {
        disc.Title = header
    }

    recorded, total, available, err := s.md.RequestDiscCapacity()
    if err == nil {
        disc.UsedSeconds = int(recorded)
        disc.TotalSeconds = int(total)
        disc.FreeSeconds = int(available)
    }

    for i := 0; i < trackCount; i++ {
        track := Track{Index: i, Channels: 2}

        title, err := s.md.RequestTrackTitle(i)
        if err == nil {
            track.Title = title
        }

        length, err := s.md.RequestTrackLength(i)
        if err == nil {
            track.Duration = time.Duration(length) * time.Second
        }

        enc, err := s.md.RequestTrackEncoding(i)
        if err == nil {
            switch enc {
            case netmd.EncSP:
                track.Encoding = EncodingSP
            case netmd.EncLP2:
                track.Encoding = EncodingLP2
            case netmd.EncLP4:
                track.Encoding = EncodingLP4
            }
        }

        disc.Tracks = append(disc.Tracks, track)
    }

    return disc, nil
}

func (s *NetMDService) Upload(filePath, title string, format UploadFormat, progress chan<- TransferProgress) error {
    if s.md == nil {
        return fmt.Errorf("not connected")
    }
    defer close(progress)

    // Convert non-WAV files to WAV via ffmpeg
    wavPath := filePath
    if needsConversion(filePath) {
        progress <- TransferProgress{Phase: "converting"}
        tmp, err := convertToWAV(filePath)
        if err != nil {
            return err
        }
        defer os.Remove(tmp)
        wavPath = tmp
    }

    // Encode to ATRAC3 for LP2/LP4 uploads (skip if already ATRAC3-encoded)
    if (format == FormatLP2 || format == FormatLP4) && !isATRAC3WAV(wavPath) {
        progress <- TransferProgress{Phase: fmt.Sprintf("encoding %s", format)}
        atracPath, err := convertToATRAC3(wavPath, format)
        if err != nil {
            return err
        }
        defer os.Remove(atracPath)
        wavPath = atracPath
    }

    track, err := s.md.NewTrack(title, wavPath)
    if err != nil {
        return fmt.Errorf("failed to create track: %w", err)
    }

    totalBytes := track.TotalBytes()

    ch := make(chan netmd.Transfer)
    go s.md.Send(track, ch)

    for t := range ch {
        if t.Error != nil {
            return fmt.Errorf("transfer error: %w", t.Error)
        }
        switch t.Type {
        case netmd.TtSetup:
            progress <- TransferProgress{Phase: "setup", TotalBytes: int64(totalBytes)}
        case netmd.TtSend:
            progress <- TransferProgress{
                BytesSent:  int64(t.Transferred),
                TotalBytes: int64(totalBytes),
                Phase:      "sending",
            }
        case netmd.TtTrack:
            progress <- TransferProgress{
                BytesSent:  int64(totalBytes),
                TotalBytes: int64(totalBytes),
                Phase:      "finalizing",
            }
        }
    }

    return nil
}

func (s *NetMDService) Download(trackIndex int, destPath string, progress chan<- TransferProgress) error {
    defer close(progress)

    if s.md == nil {
        return fmt.Errorf("not connected")
    }

    // Try native exploit download
    err := s.downloadNative(trackIndex, destPath, progress)
    if err == nil {
        return nil
    }

    // Native failed — fall back to Node.js bridge
    s.md.Close()
    s.md = nil
    s.connected = false
    time.Sleep(2 * time.Second)

    return s.downloadJS(trackIndex, destPath, progress)
}

func (s *NetMDService) downloadNative(trackIndex int, destPath string, progress chan<- TransferProgress) error {
    length, err := s.md.RequestTrackLength(trackIndex)
    if err != nil {
        return fmt.Errorf("failed to get track length: %w", err)
    }
    enc, err := s.md.RequestTrackEncoding(trackIndex)
    if err != nil {
        return fmt.Errorf("failed to get track encoding: %w", err)
    }

    totalSectors := netmd.EstimateSectors(int(length), enc)

    progress <- TransferProgress{Phase: "initializing"}

    dlProgress := make(chan netmd.DownloadProgress, 100)
    go func() {
        for p := range dlProgress {
            progress <- TransferProgress{
                BytesSent:  int64(p.Sector),
                TotalBytes: int64(totalSectors),
                Phase:      p.Phase,
            }
        }
    }()

    data, err := s.md.DownloadTrack(trackIndex, totalSectors, enc, dlProgress)
    if err != nil {
        return err
    }

    if wantsMP3(destPath) {
        progress <- TransferProgress{Phase: "converting"}
        return convertSectorsToMP3(data, destPath)
    }

    return netmd.WriteRawFile(destPath, data)
}

func (s *NetMDService) downloadJS(trackIndex int, destPath string, progress chan<- TransferProgress) error {
    progress <- TransferProgress{Phase: "downloading"}

    // Find the download helper script
    scriptPath, err := findDownloadScript()
    if err != nil {
        return err
    }

    // If MP3 requested, download to a temp WAV first then convert
    jsOutputPath := destPath
    mp3Convert := wantsMP3(destPath)
    if mp3Convert {
        tmp, err := os.CreateTemp("", "md-tui-dl-*.wav")
        if err != nil {
            return fmt.Errorf("failed to create temp file: %w", err)
        }
        tmp.Close()
        jsOutputPath = tmp.Name()
        defer os.Remove(jsOutputPath)
    }

    // Run: node scripts/download.mjs <trackIndex> <outputPath>
    cmd := exec.Command("node", scriptPath,
        fmt.Sprintf("%d", trackIndex), jsOutputPath)

    stderr, err := cmd.StderrPipe()
    if err != nil {
        return fmt.Errorf("failed to create pipe: %w", err)
    }

    // Capture ALL stderr for error reporting
    var stderrBuf strings.Builder
    if err := cmd.Start(); err != nil {
        return fmt.Errorf("failed to start download helper: %w", err)
    }

    // Parse progress from stderr
    var lastError string
    scanner := bufio.NewScanner(stderr)
    for scanner.Scan() {
        line := scanner.Text()
        stderrBuf.WriteString(line + "\n")
        if strings.HasPrefix(line, "PROGRESS: reading ") {
            pctStr := strings.TrimPrefix(line, "PROGRESS: reading ")
            var pct int
            fmt.Sscanf(pctStr, "%d", &pct)
            progress <- TransferProgress{
                BytesSent:  int64(pct),
                TotalBytes: 100,
                Phase:      "reading",
            }
        } else if strings.HasPrefix(line, "PROGRESS: ") {
            phase := strings.TrimPrefix(line, "PROGRESS: ")
            progress <- TransferProgress{Phase: phase}
        } else if strings.HasPrefix(line, "ERROR: ") {
            lastError = strings.TrimPrefix(line, "ERROR: ")
        }
    }

    if err := cmd.Wait(); err != nil {
        if lastError != "" {
            return fmt.Errorf("%s", lastError)
        }
        // Show full stderr for debugging
        errOutput := stderrBuf.String()
        if len(errOutput) > 200 {
            errOutput = errOutput[:200]
        }
        return fmt.Errorf("download failed: %s", strings.TrimSpace(errOutput))
    }

    // Convert to MP3 if requested
    if mp3Convert {
        progress <- TransferProgress{Phase: "converting to MP3"}
        if err := convertToMP3(jsOutputPath, destPath); err != nil {
            return err
        }
    }

    // Don't reconnect — device needs replug after exploit session.
    // The TUI will show "disconnected" until user rescans.
    return nil
}

func (s *NetMDService) reconnect() {
    // Wait for device to recover from exploit session
    time.Sleep(3 * time.Second)
    md, err := netmd.NewNetMD(0, s.debug)
    if err != nil {
        // Device may need replugging — that's OK
        return
    }
    s.md = md
    s.connected = true
    s.name = md.DeviceName()
}

func findDownloadScript() (string, error) {
    candidates := []string{
        "scripts/download.mjs",
    }

    // Check relative to executable (resolve symlinks like /proc/self/exe)
    if exe, err := os.Executable(); err == nil {
        if resolved, err := filepath.EvalSymlinks(exe); err == nil {
            exe = resolved
        }
        dir := filepath.Dir(exe)
        candidates = append(candidates,
            filepath.Join(dir, "scripts", "download.mjs"),
        )
    }

    // Check relative to working directory
    if wd, err := os.Getwd(); err == nil {
        candidates = append(candidates,
            filepath.Join(wd, "scripts", "download.mjs"),
        )
    }

    for _, path := range candidates {
        if _, err := os.Stat(path); err == nil {
            abs, _ := filepath.Abs(path)
            return abs, nil
        }
    }

    return "", fmt.Errorf("download script not found — ensure scripts/download.mjs exists and run 'npm install' in scripts/")
}

func (s *NetMDService) RenameTrack(index int, title string) error {
    if s.md == nil {
        return fmt.Errorf("not connected")
    }
    return s.md.SetTrackTitle(index, title, false)
}

func (s *NetMDService) RenameDisc(title string) error {
    if s.md == nil {
        return fmt.Errorf("not connected")
    }
    return s.md.SetDiscHeader(title)
}

func (s *NetMDService) DeleteTrack(index int) error {
    if s.md == nil {
        return fmt.Errorf("not connected")
    }
    return s.md.EraseTrack(index)
}

func (s *NetMDService) MoveTrack(from, to int) error {
    if s.md == nil {
        return fmt.Errorf("not connected")
    }
    return s.md.MoveTrack(from, to)
}

func (s *NetMDService) WipeDisc() error {
    if s.md == nil {
        return fmt.Errorf("not connected")
    }
    count, err := s.md.RequestTrackCount()
    if err != nil {
        return err
    }
    for i := count - 1; i >= 0; i-- {
        if err := s.md.EraseTrack(i); err != nil {
            return fmt.Errorf("failed to erase track %d: %w", i, err)
        }
    }
    return s.md.SetDiscHeader("")
}

// isATRAC3WAV checks if a WAV file already has ATRAC3 encoding (format tag 624).
func isATRAC3WAV(path string) bool {
    f, err := os.Open(path)
    if err != nil {
        return false
    }
    defer f.Close()
    header := make([]byte, 22)
    if _, err := f.Read(header); err != nil {
        return false
    }
    if string(header[0:4]) != "RIFF" || string(header[8:12]) != "WAVE" {
        return false
    }
    formatTag := int(header[20]) | int(header[21])<<8
    return formatTag == 624
}

// convertSectorsToMP3 extracts ATRAC3 frames from raw sector data,
// wraps them in an ATRAC3 WAV container, and converts to MP3 via ffmpeg.
func convertSectorsToMP3(sectorData []byte, mp3Path string) error {
    // Extract ATRAC3 frames from sectors (same layout as file storage)
    // Each sector: 20-byte header + 11 × (12-byte SG header + 192-byte frame + 8-byte padding)
    const (
        sectorSize   = 2352
        sectorHeader = 20
        sgHeader     = 12
        sgFrame      = 192
        sgPadding    = 8
        sgTotal      = sgHeader + sgFrame + sgPadding // 212
        sgPerSector  = 11
    )

    numSectors := len(sectorData) / sectorSize
    var frames []byte
    for s := 0; s < numSectors; s++ {
        for sg := 0; sg < sgPerSector; sg++ {
            frameStart := s*sectorSize + sectorHeader + sg*sgTotal + sgHeader
            if frameStart+sgFrame > len(sectorData) {
                break
            }
            frames = append(frames, sectorData[frameStart:frameStart+sgFrame]...)
        }
    }

    // Wrap in ATRAC3 WAV container
    wav := mdstore.BuildATRAC3WAV(frames)

    // Write temp WAV
    tmp, err := os.CreateTemp("", "md-tui-atrac-*.wav")
    if err != nil {
        return fmt.Errorf("failed to create temp file: %w", err)
    }
    tmpPath := tmp.Name()
    defer os.Remove(tmpPath)
    if _, err := tmp.Write(wav); err != nil {
        tmp.Close()
        return fmt.Errorf("failed to write temp WAV: %w", err)
    }
    tmp.Close()

    // Convert to MP3
    return convertToMP3(tmpPath, mp3Path)
}

func convertToMP3(src, dst string) error {
    if _, err := exec.LookPath("ffmpeg"); err != nil {
        return fmt.Errorf("ffmpeg not found — install ffmpeg to convert downloads to MP3")
    }
    cmd := exec.Command("ffmpeg", "-i", src, "-codec:a", "libmp3lame",
        "-q:a", "2", "-y", dst)
    if out, err := cmd.CombinedOutput(); err != nil {
        return fmt.Errorf("MP3 conversion failed: %w\n%s", err, out)
    }
    return nil
}

func wantsMP3(path string) bool {
    return strings.ToLower(filepath.Ext(path)) == ".mp3"
}

func needsConversion(path string) bool {
    ext := strings.ToLower(filepath.Ext(path))
    return ext != ".wav"
}

func convertToWAV(src string) (string, error) {
    if _, err := exec.LookPath("ffmpeg"); err != nil {
        return "", fmt.Errorf("ffmpeg not found — install ffmpeg to upload non-WAV files")
    }
    tmp, err := os.CreateTemp("", "md-tui-*.wav")
    if err != nil {
        return "", fmt.Errorf("failed to create temp file: %w", err)
    }
    tmp.Close()
    cmd := exec.Command("ffmpeg", "-i", src, "-ar", "44100",
        "-sample_fmt", "s16", "-ac", "2", "-y", tmp.Name())
    if out, err := cmd.CombinedOutput(); err != nil {
        os.Remove(tmp.Name())
        return "", fmt.Errorf("ffmpeg conversion failed: %w\n%s", err, out)
    }
    return tmp.Name(), nil
}

func convertToATRAC3(src string, format UploadFormat) (string, error) {
    if _, err := exec.LookPath("atracdenc"); err != nil {
        return "", fmt.Errorf("atracdenc not found — install atracdenc for LP2/LP4 uploads")
    }
    tmp, err := os.CreateTemp("", "md-tui-atrac-*.wav")
    if err != nil {
        return "", fmt.Errorf("failed to create temp file: %w", err)
    }
    tmp.Close()
    args := []string{"-e", "atrac3", "-i", src, "-o", tmp.Name()}
    if format == FormatLP4 {
        args = []string{"-e", "atrac3", "--bitrate", "64", "-i", src, "-o", tmp.Name()}
    }
    cmd := exec.Command("atracdenc", args...)
    if out, err := cmd.CombinedOutput(); err != nil {
        os.Remove(tmp.Name())
        return "", fmt.Errorf("atracdenc encoding failed: %w\n%s", err, out)
    }
    return tmp.Name(), nil
}
