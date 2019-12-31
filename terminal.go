package main

import (
	"bytes"
	"fmt"
	"io"
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
	termios, err := unix.IoctlGetTermios(syscall.Stdin, unix.TIOCGETA)
	if err != nil {
		return err
	}
	// Save terminal attributes.
	original := &unix.Termios{}
	*original = *termios
	SetRaw(termios)    // Set terminal to raw mode.
	SetSignal(termios) // Allow keyboard signaling.
	err = unix.IoctlSetTermios(syscall.Stdin, unix.TIOCSETA, termios)
	if err != nil {
		return err
	}
	defer func() {
		r := recover()
		// Restore terminal attributes.
		err := unix.IoctlSetTermios(syscall.Stdin, unix.TIOCSETA, original)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
		}
		if r != nil {
			panic(r)
		}
	}()
	return f()
}

type BytesBuilder struct {
	buf bytes.Buffer
}

func (bb *BytesBuilder) WriteBytes(bytes ...byte) {
	_, err := bb.buf.Write(bytes)
	if err != nil {
		panic(err)
	}
}

func (bb *BytesBuilder) WriteInteger(n int32) {
	_, err := fmt.Fprintf(&bb.buf, "%d", n)
	if err != nil {
		panic(err)
	}
}

func (bb *BytesBuilder) Build(w io.Writer) error {
	_, err := w.Write(bb.buf.Bytes())
	if err == nil {
		bb.buf.Reset()
	}
	return err
}

type ANSI struct {
	io.Writer
}

// Cursor Down
func (a ANSI) CUD(n int32) error {
	var bb BytesBuilder
	bb.WriteBytes(ESC, CSI)
	bb.WriteInteger(n)
	bb.WriteBytes('B')
	return bb.Build(a)
}

// Cursor Forward
func (a ANSI) CUF(n int32) error {
	var bb BytesBuilder
	bb.WriteBytes(ESC, CSI)
	bb.WriteInteger(n)
	bb.WriteBytes('C')
	return bb.Build(a)
}

// Cursor Position
func (a ANSI) CUP(n, m int32) error {
	var bb BytesBuilder
	bb.WriteBytes(ESC, CSI)
	bb.WriteInteger(n)
	bb.WriteBytes(';')
	bb.WriteInteger(m)
	bb.WriteBytes('H')
	return bb.Build(a)
}

const (
	EL_END   = 0
	EL_BEGIN = 1
	EL_ALL   = 2
)

// Erase in Line
func (a ANSI) EL(n int32) error {
	var bb BytesBuilder
	bb.WriteBytes(ESC, CSI)
	bb.WriteInteger(n)
	bb.WriteBytes('K')
	return bb.Build(a)
}

// Device Status Report
func (a ANSI) DSR() error {
	var bb BytesBuilder
	bb.WriteBytes(ESC, CSI, '6', 'n')
	return bb.Build(a)
}

// Save Cursor Position
func (a ANSI) SCP() error {
	var bb BytesBuilder
	bb.WriteBytes(ESC, CSI, 's')
	return bb.Build(a)
}

// Restore Cursor Position
func (a ANSI) RCP() error {
	var bb BytesBuilder
	bb.WriteBytes(ESC, CSI, 'u')
	return bb.Build(a)
}
