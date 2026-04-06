package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/c0mmrade/md-tui/internal/device"
	"github.com/c0mmrade/md-tui/internal/mdstore"
	"github.com/c0mmrade/md-tui/internal/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	mock := flag.Bool("mock", false, "use mock device for development")
	debug := flag.Bool("debug", false, "enable verbose USB logging")
	store := flag.String("store", "", "file storage command: encode, decode, calibrate, analyze")
	flag.Parse()

	// Handle --store subcommands before launching TUI
	if *store != "" {
		args := flag.Args()
		if err := runStoreCommand(*store, args); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

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

func runStoreCommand(cmd string, args []string) error {
	switch cmd {
	case "encode":
		if len(args) < 2 {
			return fmt.Errorf("usage: --store encode <input-file> <output.wav>")
		}
		return mdstore.EncodeFile(args[0], args[1])

	case "decode":
		if len(args) < 2 {
			return fmt.Errorf("usage: --store decode <raw-file> <output-dir>")
		}
		_, err := mdstore.DecodeFile(args[0], args[1])
		return err

	case "calibrate":
		if len(args) < 1 {
			return fmt.Errorf("usage: --store calibrate <output.wav> [num-frames]")
		}
		numFrames := 64
		if len(args) >= 2 {
			n, err := strconv.Atoi(args[1])
			if err == nil && n > 0 {
				numFrames = n
			}
		}
		return mdstore.GenerateCalibrationWAV(args[0], numFrames)

	case "analyze":
		if len(args) < 1 {
			return fmt.Errorf("usage: --store analyze <raw-file>")
		}
		report, err := mdstore.AnalyzeCalibration(args[0])
		if err != nil {
			return err
		}
		fmt.Print(report)
		return nil

	default:
		return fmt.Errorf("unknown store command %q (use: encode, decode, calibrate, analyze)", cmd)
	}
}
