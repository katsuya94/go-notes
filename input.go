package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"unicode/utf8"
)

const (
	ESC = 0x1b
	CSI = 0x5b
	DEL = 0x7f
)

var ErrInvalidEscapeSequence = errors.New("invalid escape sequence")

type InputManagerOptions struct {
	Reader    io.Reader
	State     *StateClient
	HandleCPR func(width, height int)
}

type InputManager struct {
	Options InputManagerOptions
	reader  io.ByteReader
}

func NewInputManager(options InputManagerOptions) *InputManager {
	return &InputManager{options, nil}
}

func (im *InputManager) handleEscapeSequence() error {
	var b byte
	var err error
	// Read the second byte.
	b, err = im.reader.ReadByte()
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
	b, err = im.reader.ReadByte()
	if err != nil {
		return err
	}
	// Read the parameter bytes.
	parameterBytes := []byte{}
	for {
		if !(b >= 0x30 && b <= 0x3f) {
			break
		}
		parameterBytes = append(parameterBytes, b)
		b, err = im.reader.ReadByte()
		if err != nil {
			return err
		}
	}
	Logger.Print("Escape sequence parameter bytes ", string(parameterBytes))
	// Read the intermediate bytes.
	for {
		if !(b >= 0x20 && b <= 0x2f) {
			break
		}
		b, err = im.reader.ReadByte()
		if err != nil {
			return err
		}
	}
	// Read the final byte.
	if !(b >= 0x40 && b <= 0x7e) {
		return ErrInvalidEscapeSequence
	}
	switch b {
	case 'R':
		var n, m int
		_, err := fmt.Sscanf(string(parameterBytes), "%d;%d", &n, &m)
		if err != nil {
			return err
		}
		im.Options.HandleCPR(n, m)
	}
	return nil
}

func (im *InputManager) handleRune(c rune) {
	switch c {
	case DEL:
		im.Options.State.Backspace()
	default:
		im.Options.State.Append(c)
	}
}

// TODO: refactor into RuneReader implementation that ignores escape sequences
func (im *InputManager) Start() error {
	if im.Options.Reader == nil {
		return fmt.Errorf("no Reader")
	}
	if im.Options.State == nil {
		return fmt.Errorf("no State")
	}
	if im.Options.HandleCPR == nil {
		return fmt.Errorf("no HandleCPR")
	}
	im.reader = bufio.NewReader(im.Options.Reader)
	buf := make([]byte, 0, 4)
	Logger.Print("Starting InputManager")
	for {
		b, err := im.reader.ReadByte()
		if err != nil {
			return err
		}
		if b == ESC {
			Logger.Print("Start handling escape sequence")
			err := im.handleEscapeSequence()
			if err != nil {
				return err
			}
			continue
		}
		buf = append(buf, b)
		if utf8.FullRune(buf) {
			Logger.Print("Received input")
			c, _ := utf8.DecodeRune(buf)
			buf = buf[:0]
			im.handleRune(c)
		}
	}
}
