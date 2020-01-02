package main

import (
	"os"
	"sync"
)

type Subscription interface {
	Wait()
}

type Trigger struct {
	state  int
	locker sync.Locker
	cond   *sync.Cond
}

func NewTrigger() *Trigger {
	mutex := &sync.Mutex{}
	return &Trigger{0, mutex, sync.NewCond(mutex)}
}

func (t *Trigger) Subscribe() Subscription {
	t.locker.Lock()
	snapshot := t.state
	t.locker.Unlock()
	return &TriggerSubscription{t, snapshot}
}

func (t *Trigger) Notify() {
	t.locker.Lock()
	t.state++
	t.locker.Unlock()
	t.cond.Broadcast()
}

type TriggerSubscription struct {
	trigger  *Trigger
	snapshot int
}

func (ts *TriggerSubscription) Wait() {
	ts.trigger.locker.Lock()
	var state int
	for {
		state = ts.trigger.state
		if ts.snapshot != state {
			break
		}
		ts.trigger.cond.Wait()
	}
	ts.trigger.locker.Unlock()
	ts.snapshot = state
}

// Returns a subscription that waits until a signal is passed to the channel.
func NewSignalSubscription(c chan os.Signal) Subscription {
	trigger := NewTrigger()
	go func() {
		for {
			<-c
			trigger.Notify()
		}
	}()
	return trigger.Subscribe()

}

// Returns a subscription that waits for any of its subscriptions.
func NewAnySubscription(subscriptions ...Subscription) Subscription {
	trigger := NewTrigger()
	for _, subscription := range subscriptions {
		go func(subscription Subscription) {
			for {
				subscription.Wait()
				trigger.Notify()
			}
		}(subscription)
	}
	return trigger.Subscribe()
}
