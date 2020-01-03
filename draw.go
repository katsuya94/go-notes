package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type DrawManagerOptions struct {
	Writer             io.Writer
	Search             *SearchClient
	TerminalDimensions *TerminalDimensionsClient
}

type DrawManager struct {
	Options  DrawManagerOptions
	w        *bufio.Writer
	maxLines int
}

func NewDrawManager(options DrawManagerOptions) *DrawManager {
	return &DrawManager{options, nil, 1}
}

// TODO: this needs to be more robust
func (dm *DrawManager) Cleanup() error {
	var err error
	ansi := ANSI{os.Stdout}
	err = ansi.ED(0)
	if err != nil {
		return err
	}
	err = ansi.CR()
	if err != nil {
		return err
	}
	return nil
}

func (dm *DrawManager) Start() error {
	if dm.Options.Writer == nil {
		return fmt.Errorf("no Writer")
	}
	if dm.Options.Search == nil {
		return fmt.Errorf("no Search")
	}
	if dm.Options.TerminalDimensions == nil {
		return fmt.Errorf("no TerminalDimensions")
	}
	dm.w = bufio.NewWriter(dm.Options.Writer)
	subscription := NewAnySubscription(
		dm.Options.Search.Subscribe(),
		dm.Options.TerminalDimensions.Subscribe())
	var err error
	Logger.Print("Starting DrawManager")
	for {
		err = dm.draw()
		if err != nil {
			return err
		}
		subscription.Wait()
	}
}

// TODO: handle wide unicaode characters
func printLine(ansi ANSI, line string, width int, selected bool) error {
	var err error

	// Inverse, if selected
	if selected {
		err = ansi.SGR(7)
		if err != nil {
			return err
		}
	}

	// Erase line
	if selected {
		err = ansi.CR()
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(ansi, strings.Repeat(" ", width))
		if err != nil {
			return err
		}
	} else {
		err = ansi.EL(2)
		if err != nil {
			return err
		}
	}

	// Go to beginning of line
	err = ansi.CR()
	if err != nil {
		return err
	}

	// Print, potentially with truncation
	if len(line) > width {
		format := fmt.Sprintf("%%.%ds…", width-1)
		_, err = fmt.Fprintf(ansi, format, line)
		if err != nil {
			return err
		}
	} else {
		_, err = fmt.Fprint(ansi, line)
		if err != nil {
			return err
		}
	}

	// Disable inverse, if selected
	if selected {
		err = ansi.SGR(0)
		if err != nil {
			return err
		}
	}
	return nil
}

func (dm *DrawManager) draw() error {
	Logger.Print("Drawing")

	selection, results := dm.Options.Search.Results()
	width, _ := dm.Options.TerminalDimensions.Dimensions()
	ansi := ANSI{dm.w}
	var err error

	// Hide cursor
	err = ansi.DECTCEM(false)
	if err != nil {
		return err
	}

	// Write query
	query := dm.Options.Search.Query()
	err = printLine(ansi, query, width, selection == -1)
	if err != nil {
		return err
	}

	// Write results
	for i, result := range results {
		err = ansi.NL()
		if err != nil {
			return err
		}
		err = printLine(ansi, result, width, selection == i)
		if err != nil {
			return err
		}
	}

	// Clear rest of screen
	lines := len(results) + 1
	if lines > dm.maxLines {
		dm.maxLines = lines
	}
	if lines < dm.maxLines {
		err = ansi.CR()
		if err != nil {
			return err
		}
		err = ansi.NL()
		if err != nil {
			return err
		}
		lines++
		err = ansi.ED(0)
		if err != nil {
			return err
		}
	}

	// Set cursor at end of query
	if lines > 1 {
		err = ansi.CR()
		if err != nil {
			return err
		}
		err = ansi.CUU(lines - 1)
		if err != nil {
			return err
		}
		err = ansi.CUF(len(query))
		if err != nil {
			return err
		}
	}

	// Show cursor
	err = ansi.DECTCEM(true)
	if err != nil {
		return err
	}

	return dm.w.Flush()
}
