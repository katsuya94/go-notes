package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

const DEFAULT_EDITOR = "vim"

var Logger *log.Logger

func Run() error {
	var (
		notePath string
		err      error
	)
	err = WithTerminalAttributes(func() error {
		var err error

		file, err := os.OpenFile(
			"/tmp/go-notes.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		Logger = log.New(file, "", log.Ldate|log.Ltime|log.Lshortfile)

		fail := make(chan error)

		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

		winch := make(chan os.Signal)
		signal.Notify(winch, syscall.SIGWINCH)

		selection := make(chan string)

		Logger.Print("Initializing managers")
		searchManager := NewSearchManager(
			SearchManagerOptions{Selection: selection})
		terminalDimensionsManager := NewTerminalDimensionsManager(
			TerminalDimensionsManagerOptions{Writer: os.Stdout})
		drawManager := NewDrawManager(DrawManagerOptions{Writer: os.Stdout})
		inputManager := NewInputManager(InputManagerOptions{Reader: os.Stdin})

		terminalDimensionsManager.Options.WinchSubscription =
			NewSignalSubscription(winch)
		drawManager.Options.Search = searchManager.Client()
		drawManager.Options.TerminalDimensions =
			terminalDimensionsManager.Client()
		inputManager.Options.Search = searchManager.Client()
		inputManager.Options.HandleCPR = terminalDimensionsManager.HandleCPR

		Logger.Print("Starting managers")
		go func() { fail <- searchManager.Start() }()
		go func() { fail <- terminalDimensionsManager.Start() }()
		go func() { fail <- drawManager.Start() }()
		go func() { fail <- inputManager.Start() }()

		defer drawManager.Cleanup()

		select {
		case err := <-fail:
			return err
		case <-quit:
			return nil
		case notePath = <-selection:
			return nil
		}
	})
	if err != nil {
		return err
	}
	if notePath == "" {
		return nil
	}
	editor := os.Getenv("VISUAL")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = DEFAULT_EDITOR
	}
	editorPath, err := exec.LookPath(editor)
	if err != nil {
		return err
	}
	err = syscall.Exec(editorPath, []string{editor, notePath}, os.Environ())
	if err != nil {
		return err
	}
	return nil // NEVER RUN
}

func main() {
	err := Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
