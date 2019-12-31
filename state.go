package main

type StateManager struct {
	query   []rune
	trigger *Trigger
}

func NewStateManager() *StateManager {
	return &StateManager{nil, NewTrigger()}
}

func (sm *StateManager) Client() *StateClient {
	return &StateClient{sm, sm.trigger.Subscribe()}
}

type StateClient struct {
	sm           *StateManager
	subscription Subscription
}

func (sc *StateClient) Query() string {
	return string(sc.sm.query)
}

func (sc *StateClient) Append(c rune) {
	sc.sm.query = append(sc.sm.query, c)
	sc.sm.trigger.Notify()
}

func (sc *StateClient) Backspace() {
	if len(sc.sm.query) == 0 {
		return
	}
	sc.sm.query = sc.sm.query[:len(sc.sm.query)-1]
	sc.sm.trigger.Notify()
}

func (sc *StateClient) Wait() {
	sc.subscription.Wait()
}
