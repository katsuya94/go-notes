package main

import (
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
	width, height int
	trigger       *Trigger
	mutex         *sync.RWMutex
}

func NewTerminalDimensionsManager(
	options TerminalDimensionsManagerOptions) *TerminalDimensionsManager {
	return &TerminalDimensionsManager{
		options,
		0, 0,
		NewTrigger(),
		&sync.RWMutex{}}
}

func (tdm *TerminalDimensionsManager) Client() *TerminalDimensionsClient {
	return &TerminalDimensionsClient{tdm, tdm.trigger.Subscribe()}
}

func (tdm *TerminalDimensionsManager) HandleCPR(n, m int) {
	tdm.mutex.Lock()
	tdm.width = m
	tdm.height = n
	Logger.Print("New terminal width, height := ", tdm.width, tdm.height)
	tdm.mutex.Unlock()
	tdm.trigger.Notify()
}

func (tdm *TerminalDimensionsManager) requestCUP() {
	ansi := ANSI{tdm.Options.Writer}
	OrDie(ansi.SCP())
	OrDie(ansi.CUF(MAX_TERMINAL_DIMENSION))
	OrDie(ansi.CUD(MAX_TERMINAL_DIMENSION))
	OrDie(ansi.DSR())
	OrDie(ansi.RCP())
}

func (tdm *TerminalDimensionsManager) Start() error {
	if tdm.Options.Writer == nil {
		return fmt.Errorf("no Writer")
	}
	if tdm.Options.WinchSubscription == nil {
		return fmt.Errorf("no WinchSubscription")
	}
	Logger.Print("Starting TerminalDimensionsManager")
	for {
		tdm.requestCUP()
		tdm.Options.WinchSubscription.Wait()
		Logger.Print("Received WINCH")
	}
}

type TerminalDimensionsClient struct {
	tdm          *TerminalDimensionsManager
	subscription Subscription
}

func (tdc *TerminalDimensionsClient) Dimensions() (int, int) {
	tdc.tdm.mutex.RLock()
	width, height := tdc.tdm.width, tdc.tdm.height
	tdc.tdm.mutex.RUnlock()
	return width, height
}

func (tdc *TerminalDimensionsClient) Wait() {
	tdc.subscription.Wait()
}
