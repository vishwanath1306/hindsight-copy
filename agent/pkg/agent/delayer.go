package agent

import (
	"container/list"
	"time"

	"github.com/geraldleizhang/hindsight/agent/pkg/memory"
)

/*
The delayer drains entries from the incoming channel and,
after a configured delay, writes them to the outgoing
channel.
*/
type Delayer struct {
	Incoming chan []memory.Trigger
	Outgoing chan []memory.Trigger

	delay   time.Duration
	delayed *list.List
	timer   *time.Timer
}

type DelayedTrigger struct {
	trigger_at time.Time
	triggers   []memory.Trigger
}

func delayTriggers(delay time.Duration, incoming chan []memory.Trigger) *Delayer {
	var delayer Delayer
	delayer.Incoming = incoming
	delayer.Outgoing = make(chan []memory.Trigger, 10000)
	delayer.delay = delay
	delayer.delayed = list.New()
	delayer.timer = time.NewTimer(0)
	<-delayer.timer.C // Stops the timer

	go delayer.applyDelay()
	return &delayer
}

func (d *Delayer) addTriggers(triggers []memory.Trigger) {
	var t DelayedTrigger
	t.trigger_at = time.Now().Add(d.delay)
	t.triggers = triggers

	if d.delayed.Len() == 0 {
		d.timer.Reset(d.delay)
	}

	d.delayed.PushBack(&t)
}

/* Checks if we've hit any timeouts
and forwards the triggers if so */
func (d *Delayer) checkTimeouts() {
	now := time.Now()
	for d.delayed.Len() > 0 {
		next := d.delayed.Front()
		t := next.Value.(*DelayedTrigger)
		if t.trigger_at.Before(now) {
			d.delayed.Remove(next)
			d.Outgoing <- t.triggers
		} else {
			d.timer.Reset(t.trigger_at.Sub(now))
			return
		}
	}
}

func (d *Delayer) applyDelay() {
	for {
		if d.delayed.Len() == 0 {
			select {
			case triggers := <-d.Incoming:
				d.addTriggers(triggers)
			}
		} else {
			select {
			case triggers := <-d.Incoming:
				d.addTriggers(triggers)
			case <-d.timer.C:
				d.checkTimeouts()
			}
		}
	}
}
