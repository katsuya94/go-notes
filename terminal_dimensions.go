package main

import (
	"bufio"
	"fmt"
	"io"
	"sync"
)

const MAX_TERMINAL_DIMENSION = 999

type TerminalDimensionsManagerOptions struct {
	Writer            io.Writer
	WinchSubscription Subscription
}

type TerminalDimensionsManager struct {
	Options       TerminalDimensionsManagerOptions
	w             *bufio.Writer
	width, height int
	trigger       *Trigger
	mutex         *sync.RWMutex
}

func NewTerminalDimensionsManager(
	options TerminalDimensionsManagerOptions) *TerminalDimensionsManager {
	return &TerminalDimensionsManager{
		options,
		nil,
		0, 0,
		NewTrigger(),
		&sync.RWMutex{}}
}

func (tdm *TerminalDimensionsManager) Client() *TerminalDimensionsClient {
	return &TerminalDimensionsClient{tdm}
}

func (tdm *TerminalDimensionsManager) notify() {
	Logger.Print("TerminalDimensionsManager Notify")
	tdm.trigger.Notify()
}

func (tdm *TerminalDimensionsManager) HandleCPR(n, m int) {
	tdm.mutex.Lock()
	tdm.width = m
	tdm.height = n
	Logger.Print("New terminal width, height := ", tdm.width, tdm.height)
	tdm.mutex.Unlock()
	tdm.notify()
}

func (tdm *TerminalDimensionsManager) requestCPR() error {
	Logger.Print("Requesting CPR")
	ansi := ANSI{tdm.w}
	var err error
	err = ansi.SCP()
	if err != nil {
		return err
	}
	err = ansi.CUF(MAX_TERMINAL_DIMENSION)
	if err != nil {
		return err
	}
	err = ansi.CUD(MAX_TERMINAL_DIMENSION)
	if err != nil {
		return err
	}
	err = ansi.DSR()
	if err != nil {
		return err
	}
	err = ansi.RCP()
	if err != nil {
		return err
	}
	return tdm.w.Flush()
}

func (tdm *TerminalDimensionsManager) Start() error {
	if tdm.Options.Writer == nil {
		return fmt.Errorf("no Writer")
	}
	if tdm.Options.WinchSubscription == nil {
		return fmt.Errorf("no WinchSubscription")
	}
	tdm.w = bufio.NewWriter(tdm.Options.Writer)
	var err error
	Logger.Print("Starting TerminalDimensionsManager")
	for {
		err = tdm.requestCPR()
		if err != nil {
			return err
		}
		tdm.Options.WinchSubscription.Wait()
	}
}

type TerminalDimensionsClient struct {
	tdm *TerminalDimensionsManager
}

func (tdc *TerminalDimensionsClient) Dimensions() (int, int) {
	tdc.tdm.mutex.RLock()
	width, height := tdc.tdm.width, tdc.tdm.height
	tdc.tdm.mutex.RUnlock()
	return width, height
}

func (tdc *TerminalDimensionsClient) Subscribe() Subscription {
	return tdc.tdm.trigger.Subscribe()
}
