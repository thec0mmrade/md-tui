package main

import (
    "flag"
    "fmt"
    "os"

    "github.com/c0mmrade/md-tui/internal/device"
    "github.com/c0mmrade/md-tui/internal/ui"
    tea "github.com/charmbracelet/bubbletea"
)

func main() {
    mock := flag.Bool("mock", false, "use mock device for development")
    debug := flag.Bool("debug", false, "enable verbose USB logging")
    flag.Parse()

    var svc device.DeviceService
    if *mock {
        svc = device.NewMockService()
    } else {
        svc = device.NewNetMDService(*debug)
    }

    app := ui.NewApp(svc)
    p := tea.NewProgram(app, tea.WithAltScreen())
    app.SetProgram(p)

    if _, err := p.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "error: %v\n", err)
        os.Exit(1)
    }
}
