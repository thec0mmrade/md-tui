package device

import (
    "fmt"
    "time"
)

type MockService struct {
    connected  bool
    deviceName string
    disc       *Disc
}

func NewMockService() *MockService {
    return &MockService{
        deviceName: "Mock MZ-N510",
        disc: &Disc{
            Title: "My Mix Disc",
            Tracks: []Track{
                {Index: 0, Title: "Autumn Leaves", Duration: 3*time.Minute + 42*time.Second, Encoding: EncodingSP, Channels: 2},
                {Index: 1, Title: "Blue in Green", Duration: 5*time.Minute + 27*time.Second, Encoding: EncodingSP, Channels: 2},
                {Index: 2, Title: "So What", Duration: 9*time.Minute + 22*time.Second, Encoding: EncodingLP2, Channels: 2},
                {Index: 3, Title: "", Duration: 4*time.Minute + 15*time.Second, Encoding: EncodingLP2, Channels: 2},
                {Index: 4, Title: "Take Five", Duration: 5*time.Minute + 24*time.Second, Encoding: EncodingLP4, Channels: 2},
                {Index: 5, Title: "A Love Supreme", Duration: 7*time.Minute + 44*time.Second, Encoding: EncodingSP, Channels: 2},
                {Index: 6, Title: "Round Midnight", Duration: 3*time.Minute + 18*time.Second, Encoding: EncodingLP2, Channels: 1},
                {Index: 7, Title: "Naima", Duration: 4*time.Minute + 23*time.Second, Encoding: EncodingLP4, Channels: 2},
            },
            WriteProtected: false,
            TotalSeconds:   4800, // 80 minutes
            UsedSeconds:    2595,
            FreeSeconds:    2205,
        },
    }
}

func (m *MockService) Scan() ([]DeviceInfo, error) {
    return []DeviceInfo{
        {Index: 0, Name: m.deviceName},
    }, nil
}

func (m *MockService) Connect(index int) error {
    m.connected = true
    return nil
}

func (m *MockService) Close() {
    m.connected = false
}

func (m *MockService) DeviceName() string {
    return m.deviceName
}

func (m *MockService) IsConnected() bool {
    return m.connected
}

func (m *MockService) ListContent() (*Disc, error) {
    if !m.connected {
        return nil, fmt.Errorf("not connected")
    }
    return m.disc, nil
}

func (m *MockService) Upload(filePath, title string, format UploadFormat, progress chan<- TransferProgress) error {
    if !m.connected {
        return fmt.Errorf("not connected")
    }
    defer close(progress)

    total := int64(3_000_000)
    for sent := int64(0); sent < total; sent += 100_000 {
        phase := "sending"
        if sent < 300_000 {
            phase = "encoding"
        }
        progress <- TransferProgress{BytesSent: sent, TotalBytes: total, Phase: phase}
        time.Sleep(80 * time.Millisecond)
    }
    progress <- TransferProgress{BytesSent: total, TotalBytes: total, Phase: "finalizing"}

    enc := EncodingSP
    switch format {
    case FormatLP2:
        enc = EncodingLP2
    case FormatLP4:
        enc = EncodingLP4
    }

    idx := len(m.disc.Tracks)
    m.disc.Tracks = append(m.disc.Tracks, Track{
        Index:    idx,
        Title:    title,
        Duration: 3*time.Minute + 30*time.Second,
        Encoding: enc,
        Channels: 2,
    })
    m.disc.UsedSeconds += 210
    m.disc.FreeSeconds -= 210
    return nil
}

func (m *MockService) Download(trackIndex int, destPath string, progress chan<- TransferProgress) error {
    if !m.connected {
        return fmt.Errorf("not connected")
    }
    if trackIndex < 0 || trackIndex >= len(m.disc.Tracks) {
        return fmt.Errorf("track index out of range")
    }
    defer close(progress)

    total := int64(5_000_000)
    for sent := int64(0); sent < total; sent += 200_000 {
        progress <- TransferProgress{BytesSent: sent, TotalBytes: total, Phase: "reading"}
        time.Sleep(60 * time.Millisecond)
    }
    progress <- TransferProgress{BytesSent: total, TotalBytes: total, Phase: "done"}
    return nil
}

func (m *MockService) RenameTrack(index int, title string) error {
    if !m.connected {
        return fmt.Errorf("not connected")
    }
    if index < 0 || index >= len(m.disc.Tracks) {
        return fmt.Errorf("track index out of range")
    }
    m.disc.Tracks[index].Title = title
    return nil
}

func (m *MockService) RenameDisc(title string) error {
    if !m.connected {
        return fmt.Errorf("not connected")
    }
    m.disc.Title = title
    return nil
}

func (m *MockService) DeleteTrack(index int) error {
    if !m.connected {
        return fmt.Errorf("not connected")
    }
    if index < 0 || index >= len(m.disc.Tracks) {
        return fmt.Errorf("track index out of range")
    }
    m.disc.Tracks = append(m.disc.Tracks[:index], m.disc.Tracks[index+1:]...)
    // Reindex
    for i := range m.disc.Tracks {
        m.disc.Tracks[i].Index = i
    }
    return nil
}

func (m *MockService) MoveTrack(from, to int) error {
    if !m.connected {
        return fmt.Errorf("not connected")
    }
    if from < 0 || from >= len(m.disc.Tracks) || to < 0 || to >= len(m.disc.Tracks) {
        return fmt.Errorf("track index out of range")
    }
    track := m.disc.Tracks[from]
    m.disc.Tracks = append(m.disc.Tracks[:from], m.disc.Tracks[from+1:]...)
    rear := make([]Track, len(m.disc.Tracks[to:]))
    copy(rear, m.disc.Tracks[to:])
    m.disc.Tracks = append(m.disc.Tracks[:to], track)
    m.disc.Tracks = append(m.disc.Tracks, rear...)
    // Reindex
    for i := range m.disc.Tracks {
        m.disc.Tracks[i].Index = i
    }
    return nil
}

func (m *MockService) WipeDisc() error {
    if !m.connected {
        return fmt.Errorf("not connected")
    }
    m.disc.Tracks = nil
    m.disc.Title = ""
    m.disc.UsedSeconds = 0
    m.disc.FreeSeconds = m.disc.TotalSeconds
    return nil
}
