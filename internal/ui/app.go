package ui

import (
    "fmt"
    "time"

    "github.com/c0mmrade/md-tui/internal/device"
    "github.com/c0mmrade/md-tui/internal/ui/confirm"
    "github.com/c0mmrade/md-tui/internal/ui/deviceselect"
    "github.com/c0mmrade/md-tui/internal/ui/discview"
    "github.com/c0mmrade/md-tui/internal/ui/download"
    "github.com/c0mmrade/md-tui/internal/ui/statusbar"
    "github.com/c0mmrade/md-tui/internal/ui/theme"
    "github.com/c0mmrade/md-tui/internal/ui/trackedit"
    "github.com/c0mmrade/md-tui/internal/ui/upload"
    "github.com/charmbracelet/bubbles/key"
    "github.com/charmbracelet/bubbles/spinner"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

type ViewState int

const (
    ViewDeviceSelect ViewState = iota
    ViewDisc
    ViewUpload
    ViewDownload
    ViewTrackEdit
    ViewConfirm
)

type App struct {
    state   ViewState
    device  device.DeviceService
    disc    *device.Disc
    width   int
    height  int
    err     error
    program *tea.Program

    // Sub-models
    deviceSel    deviceselect.Model
    discView     discview.Model
    trackEdit    trackedit.Model
    confirmDlg   confirm.Model
    uploadView   upload.Model
    downloadView download.Model
    statusBar    statusbar.Model
    spinner      spinner.Model

    // View transition animation
    transitionPhase int // 0=normal, 1=dim old, 2=dim new
    pendingState    ViewState

    // Modal slide-in animation
    modalSlideFrame int // 0-3, where 3=fully positioned
}

func NewApp(svc device.DeviceService) *App {
    sp := spinner.New()
    sp.Spinner = spinner.Dot
    sp.Style = lipgloss.NewStyle().Foreground(theme.AccentColor)

    return &App{
        state:        ViewDeviceSelect,
        device:       svc,
        width:        80,
        height:       24,
        deviceSel:    deviceselect.New(80),
        discView:     discview.New(80, 24),
        trackEdit:    trackedit.New(),
        confirmDlg:   confirm.New(),
        uploadView:   upload.New(),
        downloadView: download.New(),
        statusBar:    statusbar.New(80),
        spinner:      sp,
    }
}

func (a *App) SetProgram(p *tea.Program) {
    a.program = p
}

func (a *App) Init() tea.Cmd {
    return tea.Batch(a.scanDevices(), a.spinner.Tick)
}

// downloadCompleteMsg is handled before anything else to avoid modal interception
type downloadCompleteMsg struct {
    err error
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Handle download completion FIRST — before any modal forwarding
    if msg, ok := msg.(downloadCompleteMsg); ok {
        a.downloadView.Close()
        if msg.err != nil {
            return a, a.setError(msg.err)
        }
        a.statusBar.Message = "Download complete"
        return a, nil
    }

    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        a.width = msg.Width
        a.height = msg.Height
        a.deviceSel.SetWidth(msg.Width)
        a.discView.SetSize(msg.Width, msg.Height-3)
        a.statusBar.Width = msg.Width
        if a.uploadView.IsActive() {
            a.uploadView.SetSize(msg.Width, msg.Height)
        }
        return a, nil

    case tea.KeyMsg:
        // Modal views consume all keys first
        if a.trackEdit.IsActive() {
            var cmd tea.Cmd
            a.trackEdit, cmd = a.trackEdit.Update(msg)
            return a, cmd
        }
        if a.confirmDlg.IsActive() {
            var cmd tea.Cmd
            a.confirmDlg, cmd = a.confirmDlg.Update(msg)
            return a, cmd
        }
        if a.uploadView.IsActive() {
            var cmd tea.Cmd
            a.uploadView, cmd = a.uploadView.Update(msg)
            return a, cmd
        }
        if a.downloadView.IsActive() {
            var cmd tea.Cmd
            a.downloadView, cmd = a.downloadView.Update(msg)
            return a, cmd
        }

        // Global quit
        if key.Matches(msg, theme.Keys.Quit) && !a.discView.InMoveMode() {
            return a, tea.Quit
        }

        // Theme switching (works from any non-modal view)
        if key.Matches(msg, theme.Keys.ThemeNext) {
            name := theme.CycleTheme(true)
            a.statusBar.Message = "Theme: " + name
            a.spinner.Style = lipgloss.NewStyle().Foreground(theme.AccentColor)
            return a, nil
        }
        if key.Matches(msg, theme.Keys.ThemePrev) {
            name := theme.CycleTheme(false)
            a.statusBar.Message = "Theme: " + name
            a.spinner.Style = lipgloss.NewStyle().Foreground(theme.AccentColor)
            return a, nil
        }

        switch a.state {
        case ViewDeviceSelect:
            if key.Matches(msg, theme.Keys.Rescan) {
                a.err = nil
                a.deviceSel.SetScanning()
                return a, tea.Batch(a.scanDevices(), a.spinner.Tick)
            }
            var cmd tea.Cmd
            a.deviceSel, cmd = a.deviceSel.Update(msg)
            if cmd != nil {
                return a, cmd
            }
            return a, nil
        case ViewDisc:
            return a.handleDiscKeys(msg)
        }

    case DeviceScanResultMsg:
        if msg.Err != nil {
            a.deviceSel.SetDevices(nil)
            return a, a.setError(msg.Err)
        }
        a.deviceSel.SetDevices(msg.Devices)
        if len(msg.Devices) == 1 {
            // Auto-connect single device
            return a, a.connectDevice(msg.Devices[0].Index)
        }
        return a, nil

    case deviceselect.SelectMsg:
        return a, a.connectDevice(msg.Index)

    case DeviceConnectedMsg:
        a.statusBar.Connected = true
        a.statusBar.DeviceName = msg.Name
        return a, tea.Batch(a.transitionTo(ViewDisc), a.refreshDisc())

    case DiscLoadedMsg:
        a.disc = msg.Disc
        a.discView.SetDisc(msg.Disc)
        a.err = nil
        return a, nil

    case DiscLoadErrorMsg:
        return a, a.setError(msg.Err)

    // Track edit results
    case trackedit.DoneMsg:
        switch msg.Mode {
        case trackedit.ModeRenameTrack:
            return a, a.renameTrack(msg.Index, msg.NewTitle)
        case trackedit.ModeRenameDisc:
            return a, a.renameDisc(msg.NewTitle)
        }
        return a, nil

    case trackedit.CancelMsg:
        return a, nil

    // Confirm results
    case confirm.ResultMsg:
        if !msg.Confirmed {
            return a, nil
        }
        switch msg.Action {
        case "delete":
            idx := a.discView.SelectedTrackIndex()
            return a, a.deleteTrack(idx)
        case "wipe":
            return a, a.wipeDisc()
        }
        return a, nil

    // Upload flow
    case upload.StartUploadMsg:
        path, title, format := a.uploadView.GetUploadParams()
        dotCmd := a.uploadView.SetUploading()
        return a, tea.Batch(dotCmd, a.startUpload(path, title, format))

    case upload.ProgressMsg:
        var cmd tea.Cmd
        a.uploadView, cmd = a.uploadView.Update(msg)
        return a, cmd

    case UploadFailedMsg:
        a.uploadView.Close()
        return a, tea.Batch(a.setError(msg.Err), a.refreshDisc())

    case upload.DoneMsg:
        // Check if there are more files in the batch queue
        if a.uploadView.AdvanceQueue() {
            dotCmd := a.uploadView.SetUploading()
            path, title, format := a.uploadView.GetUploadParams()
            return a, tea.Batch(dotCmd, a.startUpload(path, title, format))
        }
        // Batch complete — set disc title to folder name, then refresh
        if a.uploadView.IsBatchMode() {
            dirName := a.uploadView.BatchDirName()
            a.uploadView.Close()
            return a, a.renameDiscThenRefresh(dirName)
        }
        a.uploadView.Close()
        return a, a.refreshDisc()

    case upload.CancelMsg:
        return a, nil

    // Download flow
    case download.StartDownloadMsg:
        idx, destPath := a.downloadView.GetDownloadParams()
        dotCmd := a.downloadView.SetDownloading()
        // Run download entirely in background — tea.Cmd would block event loop
        go func() {
            progress := make(chan device.TransferProgress, 100)
            go func() {
                for p := range progress {
                    if a.program != nil {
                        a.program.Send(download.ProgressMsg{Progress: p})
                    }
                }
            }()
            err := a.device.Download(idx, destPath, progress)
            if a.program != nil {
                a.program.Send(downloadCompleteMsg{err: err})
            }
        }()
        return a, dotCmd

    case download.ProgressMsg:
        var cmd tea.Cmd
        a.downloadView, cmd = a.downloadView.Update(msg)
        return a, cmd

    case download.CancelMsg:
        return a, nil

    // Mutation results — refresh disc
    case TrackRenamedMsg, TrackDeletedMsg, TrackMovedMsg, DiscRenamedMsg, DiscWipedMsg:
        return a, a.refreshDisc()

    case discview.MoveTrackMsg:
        return a, a.moveTrack(msg.From, msg.To)

    case ErrorMsg:
        return a, a.setError(msg.Err)

    case ClearErrorMsg:
        a.err = nil
        return a, nil

    case spinner.TickMsg:
        if a.state == ViewDeviceSelect {
            var cmd tea.Cmd
            a.spinner, cmd = a.spinner.Update(msg)
            return a, cmd
        }

    case transitionTickMsg:
        switch a.transitionPhase {
        case 1:
            a.state = a.pendingState
            a.transitionPhase = 2
            return a, tea.Tick(50*time.Millisecond, func(time.Time) tea.Msg {
                return transitionTickMsg{}
            })
        case 2:
            a.transitionPhase = 0
        }
        return a, nil

    case modalSlideTickMsg:
        if a.modalSlideFrame < 3 {
            a.modalSlideFrame++
            return a, tea.Tick(30*time.Millisecond, func(time.Time) tea.Msg {
                return modalSlideTickMsg{}
            })
        }
        return a, nil
    }

    // Forward non-handled messages to active modal views
    if a.uploadView.IsActive() {
        var cmd tea.Cmd
        a.uploadView, cmd = a.uploadView.Update(msg)
        return a, cmd
    }
    if a.downloadView.IsActive() {
        var cmd tea.Cmd
        a.downloadView, cmd = a.downloadView.Update(msg)
        return a, cmd
    }

    // Delegate to active sub-model
    if a.state == ViewDisc {
        var cmd tea.Cmd
        a.discView, cmd = a.discView.Update(msg)
        return a, cmd
    }

    return a, nil
}

func (a *App) handleDiscKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch {
    case key.Matches(msg, theme.Keys.Upload):
        if a.disc != nil && !a.disc.WriteProtected {
            cmd := a.uploadView.Open(a.width, a.height)
            return a, cmd
        }
        return a, nil
    case key.Matches(msg, theme.Keys.Download):
        if a.discView.HasTracks() && a.disc != nil {
            idx := a.discView.SelectedTrackIndex()
            title := ""
            if idx >= 0 && idx < len(a.disc.Tracks) {
                title = a.disc.Tracks[idx].Title
            }
            a.downloadView.Open(idx, title, a.width)
        }
        return a, a.startModalSlide()
    case key.Matches(msg, theme.Keys.Rename):
        if a.discView.HasTracks() && a.disc != nil && !a.disc.WriteProtected {
            idx := a.discView.SelectedTrackIndex()
            title := ""
            if idx >= 0 && idx < len(a.disc.Tracks) {
                title = a.disc.Tracks[idx].Title
            }
            a.trackEdit.Open(trackedit.ModeRenameTrack, idx, title, a.width)
        }
        return a, a.startModalSlide()
    case key.Matches(msg, theme.Keys.DiscName):
        if a.disc != nil && !a.disc.WriteProtected {
            a.trackEdit.Open(trackedit.ModeRenameDisc, 0, a.disc.Title, a.width)
        }
        return a, a.startModalSlide()
    case key.Matches(msg, theme.Keys.Delete):
        if a.discView.HasTracks() && a.disc != nil && !a.disc.WriteProtected {
            idx := a.discView.SelectedTrackIndex()
            title := "(untitled)"
            if idx >= 0 && idx < len(a.disc.Tracks) && a.disc.Tracks[idx].Title != "" {
                title = a.disc.Tracks[idx].Title
            }
            a.confirmDlg.Open(
                fmt.Sprintf("Delete track %d: %q?", idx+1, title),
                "delete",
                a.width,
            )
        }
        return a, a.startModalSlide()
    case key.Matches(msg, theme.Keys.Wipe):
        if a.disc != nil && !a.disc.WriteProtected {
            a.confirmDlg.Open("Wipe entire disc? This cannot be undone.", "wipe", a.width)
        }
        return a, a.startModalSlide()
    }

    var cmd tea.Cmd
    a.discView, cmd = a.discView.Update(msg)
    return a, cmd
}

func (a *App) View() string {
    title := theme.TitleStyle.Render(" md-tui ")
    if a.statusBar.Connected {
        title += " " + lipgloss.NewStyle().Foreground(theme.DimTextColor).Render(a.device.DeviceName())
    }
    title += "\n"

    var content string
    switch a.state {
    case ViewDeviceSelect:
        content = a.viewDeviceSelect()
    case ViewDisc:
        content = a.discView.View()
    default:
        content = "  Loading..."
    }

    errBanner := ""
    if a.err != nil {
        errBanner = theme.ErrorBannerStyle.Render(" Error: "+a.err.Error()+" ") + "\n"
    }

    statusLine := a.statusBar.View()

    view := title + errBanner + content + "\n" + statusLine

    // Apply dim effect during view transitions
    if a.transitionPhase > 0 {
        view = lipgloss.NewStyle().Faint(true).Render(view)
    }

    // Overlay modals
    if a.trackEdit.IsActive() {
        view = a.overlayModal(a.trackEdit.View())
    }
    if a.confirmDlg.IsActive() {
        view = a.overlayModal(a.confirmDlg.View())
    }
    if a.uploadView.IsActive() {
        view = a.uploadView.View()
    }
    if a.downloadView.IsActive() {
        view = a.overlayModal(a.downloadView.View())
    }

    return view
}

func (a *App) overlayModal(modal string) string {
    vPos := lipgloss.Center
    // Slide-in: offset modal downward during first few frames
    if a.modalSlideFrame < 3 {
        // Use lipgloss.Place with vertical position as a fraction
        // Shift down by (3 - frame) * 2 lines from center
        offset := (3 - a.modalSlideFrame) * 2
        return lipgloss.Place(
            a.width, a.height,
            lipgloss.Center, lipgloss.Position(0.5+float64(offset)/float64(a.height)),
            modal,
        )
    }
    return lipgloss.Place(
        a.width, a.height,
        lipgloss.Center, vPos,
        modal,
    )
}

func (a *App) viewDeviceSelect() string {
    devView := a.deviceSel.View()
    if devView != "" {
        return "\n" + devView + "\n\n" +
            theme.KeyDescStyle.Render("  Press ") +
            theme.KeyStyle.Render("r") +
            theme.KeyDescStyle.Render(" to rescan, ") +
            theme.KeyStyle.Render("q") +
            theme.KeyDescStyle.Render(" to quit")
    }
    // Scanning state
    s := "\n  " + a.spinner.View() + " Scanning for NetMD devices...\n\n"
    s += theme.KeyDescStyle.Render("  Press ") +
        theme.KeyStyle.Render("r") +
        theme.KeyDescStyle.Render(" to rescan, ") +
        theme.KeyStyle.Render("q") +
        theme.KeyDescStyle.Render(" to quit")
    return s
}

func (a *App) transitionTo(newState ViewState) tea.Cmd {
    a.transitionPhase = 1
    a.pendingState = newState
    return tea.Tick(50*time.Millisecond, func(time.Time) tea.Msg {
        return transitionTickMsg{}
    })
}

func (a *App) startModalSlide() tea.Cmd {
    a.modalSlideFrame = 0
    return tea.Tick(30*time.Millisecond, func(time.Time) tea.Msg {
        return modalSlideTickMsg{}
    })
}

func (a *App) setError(err error) tea.Cmd {
    a.err = err
    return tea.Tick(5*time.Second, func(time.Time) tea.Msg {
        return ClearErrorMsg{}
    })
}

// Commands

func (a *App) scanDevices() tea.Cmd {
    return func() tea.Msg {
        devices, err := a.device.Scan()
        return DeviceScanResultMsg{Devices: devices, Err: err}
    }
}

func (a *App) connectDevice(index int) tea.Cmd {
    return func() tea.Msg {
        if err := a.device.Connect(index); err != nil {
            return ErrorMsg{Err: err}
        }
        return DeviceConnectedMsg{Name: a.device.DeviceName()}
    }
}

func (a *App) refreshDisc() tea.Cmd {
    return func() tea.Msg {
        disc, err := a.device.ListContent()
        if err != nil {
            return DiscLoadErrorMsg{Err: err}
        }
        return DiscLoadedMsg{Disc: disc}
    }
}

func (a *App) moveTrack(from, to int) tea.Cmd {
    return func() tea.Msg {
        if err := a.device.MoveTrack(from, to); err != nil {
            return ErrorMsg{Err: err}
        }
        return TrackMovedMsg{}
    }
}

func (a *App) renameTrack(index int, title string) tea.Cmd {
    return func() tea.Msg {
        if err := a.device.RenameTrack(index, title); err != nil {
            return ErrorMsg{Err: err}
        }
        return TrackRenamedMsg{}
    }
}

func (a *App) renameDisc(title string) tea.Cmd {
    return func() tea.Msg {
        if err := a.device.RenameDisc(title); err != nil {
            return ErrorMsg{Err: err}
        }
        return DiscRenamedMsg{}
    }
}

func (a *App) renameDiscThenRefresh(title string) tea.Cmd {
    return func() tea.Msg {
        // Rename first, then refresh — sequential to avoid USB conflicts
        a.device.RenameDisc(title)
        disc, err := a.device.ListContent()
        if err != nil {
            return DiscLoadErrorMsg{Err: err}
        }
        return DiscLoadedMsg{Disc: disc}
    }
}

func (a *App) deleteTrack(index int) tea.Cmd {
    return func() tea.Msg {
        if err := a.device.DeleteTrack(index); err != nil {
            return ErrorMsg{Err: err}
        }
        return TrackDeletedMsg{}
    }
}

func (a *App) wipeDisc() tea.Cmd {
    return func() tea.Msg {
        if err := a.device.WipeDisc(); err != nil {
            return ErrorMsg{Err: err}
        }
        return DiscWipedMsg{}
    }
}

func (a *App) startUpload(path, title string, format device.UploadFormat) tea.Cmd {
    return func() tea.Msg {
        progress := make(chan device.TransferProgress)
        done := make(chan error, 1)

        go func() {
            done <- a.device.Upload(path, title, format, progress)
        }()

        go func() {
            for p := range progress {
                if a.program != nil {
                    a.program.Send(upload.ProgressMsg{Progress: p})
                }
            }
        }()

        err := <-done
        if err != nil {
            return UploadFailedMsg{Err: err}
        }
        return upload.DoneMsg{}
    }
}

