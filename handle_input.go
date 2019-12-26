package main

import (
	"bufio"
	"errors"
	"io"
	"unicode/utf8"
)

const (
	ESC = 0x1b
	CSI = 0x5b
	DEL = 0x7f
)

var (
	ErrInvalidEscapeSequence = errors.New("invalid escape sequence")
)

func IgnoreEscapeSequence(r io.ByteReader) error {
	var b byte
	var err error
	// Read the second byte.
	b, err = r.ReadByte()
	if err != nil {
		return err
	}
	if !(b >= 0x40 && b <= 0x5f) {
		return ErrInvalidEscapeSequence
	}
	if b != CSI {
		// Not a CSI sequence. We're done.
		return nil
	}
	b, err = r.ReadByte()
	if err != nil {
		return err
	}
	// Read the parameter bytes.
	for {
		if !(b >= 0x30 && b <= 0x3f) {
			break
		}
		b, err = r.ReadByte()
		if err != nil {
			return err
		}
	}
	// Read the intermediate bytes.
	for {
		if !(b >= 0x20 && b <= 0x2f) {
			break
		}
		b, err = r.ReadByte()
		if err != nil {
			return err
		}
	}
	// Read the final byte.
	if !(b >= 0x40 && b <= 0x7e) {
		return ErrInvalidEscapeSequence
	}
	return nil
}

func HandleRune(notes *Notes, c rune) {
	switch c {
	case DEL:
		notes.Backspace()
	default:
		notes.Append(c)
	}
}

// TODO: refactor into RuneReader implementation that ignores escape sequences
func HandleInput(notes *Notes, in io.Reader) error {
	r := bufio.NewReader(in)
	buf := make([]byte, 0, 4)
	for {
		b, err := r.ReadByte()
		if err != nil {
			return err
		}
		if b == ESC {
			err := IgnoreEscapeSequence(r)
			if err != nil {
				return err
			}
			continue
		}
		buf = append(buf, b)
		if utf8.FullRune(buf) {
			c, _ := utf8.DecodeRune(buf)
			buf = buf[:0]
			HandleRune(notes, c)
		}
	}
}
