package ui

import "github.com/c0mmrade/md-tui/internal/device"

// Device connection messages
type DeviceConnectedMsg struct {
    Name string
}

type DeviceDisconnectedMsg struct{}

type DeviceScanResultMsg struct {
    Devices []device.DeviceInfo
    Err     error
}

// Disc content messages
type DiscLoadedMsg struct {
    Disc *device.Disc
}

type DiscLoadErrorMsg struct {
    Err error
}

// Track operation messages
type TrackRenamedMsg struct{}
type TrackDeletedMsg struct{}
type TrackMovedMsg struct{}
type DiscRenamedMsg struct{}
type DiscWipedMsg struct{}

// Upload/download messages
type UploadProgressMsg struct {
    Progress device.TransferProgress
}

type UploadDoneMsg struct {
    Err error
}

type DownloadProgressMsg struct {
    Progress device.TransferProgress
}

type DownloadDoneMsg struct {
    Err error
}

// Navigation
type NavigateMsg struct {
    View ViewState
}

// Transfer failures (close the modal + show error)
type UploadFailedMsg struct {
    Err error
}

type DownloadFailedMsg struct {
    Err error
}

// Error display
type ErrorMsg struct {
    Err error
}

type ClearErrorMsg struct{}
