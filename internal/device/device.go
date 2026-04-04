package device

import "time"

type Encoding int

const (
    EncodingSP Encoding = iota
    EncodingLP2
    EncodingLP4
)

func (e Encoding) String() string {
    switch e {
    case EncodingSP:
        return "SP"
    case EncodingLP2:
        return "LP2"
    case EncodingLP4:
        return "LP4"
    default:
        return "???"
    }
}

type UploadFormat int

const (
    FormatSP UploadFormat = iota
    FormatLP2
    FormatLP4
)

func (f UploadFormat) String() string {
    switch f {
    case FormatSP:
        return "SP"
    case FormatLP2:
        return "LP2"
    case FormatLP4:
        return "LP4"
    default:
        return "???"
    }
}

type Track struct {
    Index    int
    Title    string
    Duration time.Duration
    Encoding Encoding
    Channels int // 1=mono, 2=stereo
}

type Disc struct {
    Title          string
    Tracks         []Track
    WriteProtected bool
    UsedSeconds    int
    TotalSeconds   int
    FreeSeconds    int
}

type DeviceInfo struct {
    Index int
    Name  string
}

type TransferProgress struct {
    BytesSent  int64
    TotalBytes int64
    Phase      string // "encoding", "sending", "finalizing"
}

type DeviceService interface {
    // Discovery & connection
    Scan() ([]DeviceInfo, error)
    Connect(index int) error
    Close()
    DeviceName() string
    IsConnected() bool

    // Read
    ListContent() (*Disc, error)

    // Write
    Upload(filePath, title string, format UploadFormat, progress chan<- TransferProgress) error
    Download(trackIndex int, destPath string, progress chan<- TransferProgress) error
    RenameTrack(index int, title string) error
    RenameDisc(title string) error
    DeleteTrack(index int) error
    MoveTrack(from, to int) error
    WipeDisc() error
}
