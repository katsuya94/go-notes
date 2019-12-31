package main

import (
	"fmt"
	"io"
)

type DrawManagerOptions struct {
	Writer             io.Writer
	State              *StateClient
	TerminalDimensions *TerminalDimensionsClient
}

type DrawManager struct {
	Options DrawManagerOptions
}

func NewDrawManager(options DrawManagerOptions) *DrawManager {
	return &DrawManager{options}
}

func (dm *DrawManager) Start() error {
	if dm.Options.Writer == nil {
		return fmt.Errorf("no Writer")
	}
	if dm.Options.State == nil {
		return fmt.Errorf("no State")
	}
	if dm.Options.TerminalDimensions == nil {
		return fmt.Errorf("no TerminalDimensions")
	}
	subscription := NewAnySubscription(
		dm.Options.State, dm.Options.TerminalDimensions)
	var err error
	Logger.Print("Starting DrawManager")
	for {
		subscription.Wait()
		Logger.Print("Received draw")
		err = dm.draw()
		if err != nil {
			return err
		}
	}
}

func (dm *DrawManager) draw() error {
	width, _ := dm.Options.TerminalDimensions.Dimensions()
	ansi := ANSI{dm.Options.Writer}
	var err error
	err = ansi.EL(EL_ALL)
	if err != nil {
		return err
	}
	// TODO: handle wide unicaode characters
	query := dm.Options.State.Query()
	_, err = fmt.Fprint(ansi, "\r")
	if err != nil {
		return err
	}
	if len(query) > width {
		format := fmt.Sprintf("%%.%ds…", width-1)
		_, err = fmt.Fprintf(ansi, format, dm.Options.State.Query())
	} else {
		_, err = fmt.Fprint(ansi, query)
	}
	if err != nil {
		return err
	}
	return nil
}
