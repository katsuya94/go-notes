package main

import (
	"fmt"
	"io/ioutil"
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
		config   *Config
		notePath string
		err      error
	)

	config, err = LoadConfig()
	if err != nil {
		return err
	}

	if config.LogFile == "" {
		Logger = log.New(ioutil.Discard, "", 0)
	} else {
		file, err := os.OpenFile(
			config.LogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		Logger = log.New(file, "", log.Ldate|log.Ltime|log.Lshortfile)
	}

	err = WithTerminalAttributes(func() error {
		fail := make(chan error)
		die := make(chan interface{})

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

		searchManager.Options.NotesDirectory = config.NotesDirectory
		terminalDimensionsManager.Options.WinchSubscription =
			NewSignalSubscription(winch)
		drawManager.Options.Search = searchManager.Client()
		drawManager.Options.TerminalDimensions =
			terminalDimensionsManager.Client()
		inputManager.Options.Search = searchManager.Client()
		inputManager.Options.TerminalDimensions =
			terminalDimensionsManager.Client()

		Logger.Print("Starting managers")

		start := func(f func() error) {
			go func() {
				defer func() {
					if r := recover(); r != nil {
						die <- r
					}
				}()
				fail <- f()
			}()
		}

		start(searchManager.Start)
		start(terminalDimensionsManager.Start)
		start(drawManager.Start)
		start(inputManager.Start)

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
