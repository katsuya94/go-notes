package main

import (
	"bytes"
	"fmt"
	"io"
)

type BytesBuilder struct {
	buf bytes.Buffer
}

func (bb *BytesBuilder) WriteBytes(bytes ...byte) {
	_, err := bb.buf.Write(bytes)
	if err != nil {
		panic(err)
	}
}

func (bb *BytesBuilder) WriteInteger(n int) {
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

// Carriage Return
func (a ANSI) CR() error {
	var bb BytesBuilder
	bb.WriteBytes('\r')
	return bb.Build(a)
}

// New Line
func (a ANSI) NL() error {
	var bb BytesBuilder
	bb.WriteBytes('\n')
	return bb.Build(a)
}

// Cursor Up
func (a ANSI) CUU(n int) error {
	var bb BytesBuilder
	bb.WriteBytes(ESC, CSI)
	bb.WriteInteger(n)
	bb.WriteBytes('A')
	return bb.Build(a)
}

// Cursor Down
func (a ANSI) CUD(n int) error {
	var bb BytesBuilder
	bb.WriteBytes(ESC, CSI)
	bb.WriteInteger(n)
	bb.WriteBytes('B')
	return bb.Build(a)
}

// Cursor Forward
func (a ANSI) CUF(n int) error {
	var bb BytesBuilder
	bb.WriteBytes(ESC, CSI)
	bb.WriteInteger(n)
	bb.WriteBytes('C')
	return bb.Build(a)
}

// Cursor Back
func (a ANSI) CUB(n int) error {
	var bb BytesBuilder
	bb.WriteBytes(ESC, CSI)
	bb.WriteInteger(n)
	bb.WriteBytes('D')
	return bb.Build(a)
}

// Erase in Display
func (a ANSI) ED(n int) error {
	var bb BytesBuilder
	bb.WriteBytes(ESC, CSI)
	bb.WriteInteger(n)
	bb.WriteBytes('J')
	return bb.Build(a)
}

// Erase in Line
func (a ANSI) EL(n int) error {
	var bb BytesBuilder
	bb.WriteBytes(ESC, CSI)
	bb.WriteInteger(n)
	bb.WriteBytes('K')
	return bb.Build(a)
}

// Select Graphic Rendition
func (a ANSI) SGR(n int) error {
	var bb BytesBuilder
	bb.WriteBytes(ESC, CSI)
	bb.WriteInteger(n)
	bb.WriteBytes('m')
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

// Show/hide cursor
func (a ANSI) DECTCEM(show bool) error {
	var c byte
	if show {
		c = 'h'
	} else {
		c = 'l'
	}
	var bb BytesBuilder
	bb.WriteBytes(ESC, CSI, '?', '2', '5', c)
	return bb.Build(a)
}
