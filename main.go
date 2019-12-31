package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var Logger *log.Logger

func main() {
	err := WithTerminalAttributes(func() error {
		file, err := os.OpenFile(
			"/tmp/go-notes.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		Logger = log.New(file, "", log.Ldate|log.Ltime|log.Lshortfile)

		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

		winch := make(chan os.Signal)
		signal.Notify(winch, syscall.SIGWINCH)

		fail := make(chan error)

		Logger.Print("Initializing managers")
		stateManager := NewStateManager()
		terminalDimensionsManager := NewTerminalDimensionsManager(
			TerminalDimensionsManagerOptions{Writer: os.Stdout})
		drawManager := NewDrawManager(DrawManagerOptions{Writer: os.Stdout})
		inputManager := NewInputManager(InputManagerOptions{Reader: os.Stdin})

		terminalDimensionsManager.Options.WinchSubscription =
			NewSignalSubscription(winch)
		drawManager.Options.State = stateManager.Client()
		drawManager.Options.TerminalDimensions =
			terminalDimensionsManager.Client()
		inputManager.Options.State = stateManager.Client()
		inputManager.Options.HandleCPR = terminalDimensionsManager.HandleCPR

		Logger.Print("Starting managers")
		go func() { fail <- terminalDimensionsManager.Start() }()
		go func() { fail <- drawManager.Start() }()
		go func() { fail <- inputManager.Start() }()

		select {
		case err := <-fail:
			return err
		case <-quit:
			return nil
		}
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
