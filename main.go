package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	err := WithTerminalAttributes(func() error {
		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
		fail := make(chan error)
		notes := &Notes{}
		go func() { fail <- HandleInput(notes, os.Stdin) }()
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
