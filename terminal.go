package main

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

// SetRaw sets terminal attributes equivalent to cfmakeraw as decribed in
// termios(3).
func SetRaw(termios *unix.Termios) {
	termios.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK |
		unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
	termios.Oflag &^= unix.OPOST
	termios.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG |
		unix.IEXTEN
	termios.Cflag &^= unix.CSIZE | unix.PARENB
	termios.Cflag |= unix.CS8
	termios.Cc[unix.VMIN] = 1
	termios.Cc[unix.VTIME] = 0
}

// SetSignal sets the ISIG flag as described in termios(3) "When any of the
// characters INTR, QUIT, SUSP, or DSUSP are received, generate the
// corresponding signal."
func SetSignal(termios *unix.Termios) {
	termios.Lflag |= unix.ISIG
}

func WithTerminalAttributes(f func() error) error {
	var err error
	termios, err := unix.IoctlGetTermios(syscall.Stdin, ioctlGetTermios)
	if err != nil {
		return err
	}
	// Save terminal attributes.
	original := &unix.Termios{}
	*original = *termios
	SetRaw(termios)    // Set terminal to raw mode.
	SetSignal(termios) // Allow keyboard signaling.
	err = unix.IoctlSetTermios(syscall.Stdin, ioctlSetTermios, termios)
	if err != nil {
		return err
	}
	defer func() {
		r := recover()
		// Restore terminal attributes.
		err := unix.IoctlSetTermios(syscall.Stdin, ioctlSetTermios, original)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
		}
		if r != nil {
			panic(r)
		}
	}()
	return f()
}
