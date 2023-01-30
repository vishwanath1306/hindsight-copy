package coordinator

import (
	"container/list"
	"time"
)

type TriggerID struct {
	queue_id      int
	base_trace_id uint64
}

type Trigger struct {
	id        TriggerID
	trace_ids []uint64
}

type FinishedTrigger struct {
	queue_id           int
	total_agents       int
	dissemination_time time.Duration
}

type Coordinator struct {
	now         time.Time
	traces      map[uint64]*tracestate      // All known traces
	triggers    map[TriggerID]*triggerstate // All known triggers
	trigger_lru *list.List                  // For expiring triggers
	trace_lru   *list.List                  // For expiring traces
}

func (c *Coordinator) Init() {
	c.now = time.Now()
	c.traces = make(map[uint64]*tracestate)
	c.triggers = make(map[TriggerID]*triggerstate)
	c.trigger_lru = list.New()
	c.trace_lru = list.New()
}

/* Trace representation internal to the Coordinator */
type tracestate struct {
	id              uint64
	known_at        map[string]struct{}         // Agents where this trace is known
	triggers        map[TriggerID]*triggerstate // Triggers of this trace
	created         time.Time
	last_modified   time.Time
	last_breadcrumb time.Time
	lru_entry       *list.Element
}

/* Trigger representation internal to the coordinator */
type triggerstate struct {
	id              TriggerID
	known_at        map[string]struct{}    // Agents where this trigger is known
	traces          map[uint64]*tracestate // Traces of this trigger
	created         time.Time
	last_modified   time.Time
	last_breadcrumb time.Time
	lru_entry       *list.Element
}

func (ts *triggerstate) Trigger() Trigger {
	var t Trigger
	t.id = ts.id
	for trace_id, _ := range ts.traces {
		t.trace_ids = append(t.trace_ids, trace_id)
	}
	return t
}

func (c *Coordinator) getTrigger(id TriggerID) *triggerstate {
	if trigger, ok := c.triggers[id]; ok {
		return trigger
	}
	var trigger triggerstate
	trigger.id = id
	trigger.known_at = make(map[string]struct{})
	trigger.traces = make(map[uint64]*tracestate)
	trigger.created = c.now
	trigger.last_modified = c.now
	trigger.lru_entry = c.trigger_lru.PushFront(&trigger)
	c.triggers[id] = &trigger
	return &trigger
}

func (c *Coordinator) getTrace(id uint64) *tracestate {
	if trace, ok := c.traces[id]; ok {
		return trace
	}
	var trace tracestate
	trace.id = id
	trace.known_at = make(map[string]struct{})
	trace.triggers = make(map[TriggerID]*triggerstate)
	trace.last_modified = c.now
	trace.created = c.now
	trace.lru_entry = c.trace_lru.PushFront(&trace)
	c.traces[id] = &trace
	return &trace
}

func (c *Coordinator) checkTriggerExpiration(cutoff time.Time) (finished []FinishedTrigger) {
	for c.trigger_lru.Len() > 0 {
		trigger := c.trigger_lru.Back().Value.(*triggerstate)
		if trigger.last_modified.After(cutoff) {
			break
		}

		for _, tracestate := range trigger.traces {
			delete(tracestate.triggers, trigger.id)
		}
		delete(c.triggers, trigger.id)
		c.trigger_lru.Remove(trigger.lru_entry)

		var ft FinishedTrigger
		ft.queue_id = trigger.id.queue_id
		ft.total_agents = len(trigger.known_at)
		ft.dissemination_time = trigger.last_modified.Sub(trigger.created)
		finished = append(finished, ft)
	}
	return
}

func (c *Coordinator) checkTraceExpiration(cutoff time.Time) {
	for c.trace_lru.Len() > 0 {
		trace := c.trace_lru.Back().Value.(*tracestate)
		if trace.last_modified.After(cutoff) {
			break
		}

		for _, triggerstate := range trace.triggers {
			delete(triggerstate.traces, trace.id)
		}
		delete(c.traces, trace.id)
		c.trace_lru.Remove(trace.lru_entry)
	}
}

/*
An agent has sent us a trigger, which specifies a few traces.

If those traces are already known to the coordinator, then this
method will return zero or more breadcrumbs of agents that need
to learn of the trigger.

This method adds the trigger to the coordinator if it does not
already exist, and will expire after a timeout.
*/
func (c *Coordinator) AddTrigger(src string, t Trigger) []string {
	trigger := c.getTrigger(t.id)
	trigger.known_at[src] = struct{}{}
	c.trigger_lru.MoveToFront(trigger.lru_entry)
	trigger.last_modified = c.now

	/*
		We now need to disseminate the trigger as follows:
		* Send it to any breadcrumbs where it isn't known
		* If the set of trace_ids of this trigger changed, redistribute it to all addrs except src
	*/
	needs_rebroadcasting := false
	for _, trace_id := range t.trace_ids {
		if _, ok := trigger.traces[trace_id]; ok {
			continue // trace already attached to this trigger, do nothing
		}

		needs_rebroadcasting = true

		// Link the trigger and trace
		trace := c.getTrace(trace_id)
		trigger.traces[trace_id] = trace
		trace.triggers[trigger.id] = trigger

		// Update where the trigger and trace are known
		for addr, _ := range trace.known_at {
			trigger.known_at[addr] = struct{}{}
		}
		trace.known_at[src] = struct{}{}

		// Touch LRU
		c.trace_lru.MoveToFront(trace.lru_entry)
		trace.last_modified = c.now
	}

	/*
		Rebroadcast to all addresses where the trigger is known, except src,
		which already has the up-to-date trigger
	*/
	var send_to []string
	if needs_rebroadcasting {
		for addr, _ := range trigger.known_at {
			if addr != src {
				send_to = append(send_to, addr)
			}
		}
	}
	return send_to
}

/*
An agent has sent us some breadcrumbs of a trace.

Store the breadcrumb in the coordinator and return triggers that must
be now disseminated and the addresses to which they must be sent.
*/
func (c *Coordinator) AddBreadcrumb(src string, trace_id uint64, addrs []string) map[string][]Trigger {
	trace := c.getTrace(trace_id)

	/*
		For each trigger this trace belongs to, compare the breadcrumbs:
		- the trigger is not known at the crumb yet
		- the trigger is known at the crumb already
	*/
	addrs = append(addrs, src) // might also need to disseminate triggers to src

	to_disseminate := make(map[string][]Trigger)
	for _, addr := range addrs {
		if _, ok := trace.known_at[addr]; ok {
			continue // trace already known at this addr and by extension so too are triggers
		}

		/*
			The trace isn't known at this address yet; triggers should be disseminated
			to that address if they haven't been already
		*/
		var triggers_to_disseminate []Trigger
		for _, trigger := range trace.triggers {
			if _, ok := trigger.known_at[addr]; !ok {
				// trigger isn't known at this address yet; must disseminate
				triggers_to_disseminate = append(triggers_to_disseminate, trigger.Trigger())
				trigger.known_at[addr] = struct{}{}
				trigger.last_breadcrumb = c.now
			}
		}

		if len(triggers_to_disseminate) > 0 {
			to_disseminate[addr] = triggers_to_disseminate
		}

		trace.known_at[addr] = struct{}{}
		trace.last_breadcrumb = c.now
	}

	// Update trigger and trace LRUs
	for _, trigger := range trace.triggers {
		c.trigger_lru.MoveToFront(trigger.lru_entry)
		trigger.last_modified = c.now
	}
	c.trace_lru.MoveToFront(trace.lru_entry)
	trace.last_modified = c.now

	return to_disseminate
}
