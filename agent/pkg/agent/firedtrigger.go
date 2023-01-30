package agent

import (
	"container/list"
	"log"
	"time"
)

type TriggerID struct {
	queue_id      int
	base_trace_id uint64
}

/*
A fired trigger represents a specific instance of a trigger going off.
We represent a fired trigger with a simple state machine that can transition
between triggered (no data to report) and reporting (some traces have data to report)

The trace IDs specified by the fired trigger will now be reported.

The trigger will remain fired until it is either reported or untriggered;
this process is driven by the TriggerManager.
*/
type FiredTrigger struct {
	/* Each fired trigger has an ID that typically corresponds to a 'base' trace ID
	responsible for firing the trigger.  The ID is used as the trigger's reporting priority */
	id TriggerID

	/* The queue that this FiredTrigger belongs to */
	queue *TriggerQueue

	/* Each fired trigger specifies one or more traces that should be reported.
	The traces are all reported with the same priority as the FiredTrigger's ID.
	Note that any given trace can appear in multiple different FiredTriggers.
	These traces are the 'lateral traces' discussed in the Hindsight paper. */
	traces map[uint64]*Trace

	buffer_count int

	/* We use a simple state machine for fired triggers */
	state firedtriggerstate
}

/* Implements state machine transitions of a trace.  States are:
     * reportingTrigger
		 * idleTrigger
		 * nil (invalid)
*/
type firedtriggerstate interface {
	buffersAdded(f *FiredTrigger) firedtriggerstate
	getBuffersForReport(f *FiredTrigger) (firedtriggerstate, []int)
	evictTrigger(f *FiredTrigger) (firedtriggerstate, []int)
	checkTimeout(f *FiredTrigger, before time.Time) (firedtriggerstate, bool)
}

/*
Adds a trace to this trigger; the trace's data will now be reported.  The trace's
data will share fate with the 'base' traceID of the FiredTrigger.
If the trace has data pending, then this will transition to reportingTrigger.
It returns any breadcrumbs that need to be immediately reported.
*/
func (f *FiredTrigger) AddTrace(trace *Trace) []string {
	f.traces[trace.id] = trace
	return trace.AddTrigger(f.queue.dm, f)
}

/*
Takes any buffers from this trigger that are ready to be reported.
This potentially transitions the firedtrigger into idle state
*/
func (f *FiredTrigger) GetBuffersForReport() []int {
	var buffers []int
	f.state, buffers = f.state.getBuffersForReport(f)
	return buffers
}

/*
Called by the TriggerManager to evict a low priority fired trigger.
Returns any evicted buffers
*/
func (f *FiredTrigger) Evict() []int {
	var buffers []int
	f.state, buffers = f.state.evictTrigger(f)
	return buffers
}

/*
Evicts an idle trigger if it was last modified before the specified time.
Returns true if timed out.
*/
func (f *FiredTrigger) CheckTimeout(before time.Time) bool {
	var timedout bool
	f.state, timedout = f.state.checkTimeout(f, before)
	return timedout
}

/*
Informs the trigger that one of its traces has received new data
that should be reported.  This transitions the firedtrigger
into reporting state if it is not already reporting.
*/
func (f *FiredTrigger) buffersAdded(count int) {
	f.buffer_count += count
	f.queue.buffer_count += count
	f.queue.metrics.buffers += count
	f.state = f.state.buffersAdded(f)
}

/*
Informs the trigger that one of its traces has reported some
of its data.  This might be because TakeBuffers was called on this
trigger, or because TakeBuffers was called on a different trigger
that shares a trace ID.  We do not transition into idle state here,
even if there are no buffers remaining.
*/
func (f *FiredTrigger) buffersRemoved(count int) {
	f.buffer_count -= count
	f.queue.buffer_count -= count
	// No state transition -- handled by TakeBuffers
}

/*
A trigger is idle when all trace data has been reported.  It remains idle
for a period of time until either new data arrives, or it times out and is
removed.
*/
type idleTrigger struct {
	last_modified  time.Time
	tq_lru_element *list.Element
}

/*
Most of the time, when a trigger fires, it does not already exist, and we create
an idle FiredTrigger
*/
func (queue *TriggerQueue) initIdleTrigger(base_trace_id uint64) *FiredTrigger {
	var f FiredTrigger
	f.id.base_trace_id = base_trace_id
	f.id.queue_id = queue.id
	f.queue = queue
	f.traces = make(map[uint64]*Trace)
	f.buffer_count = 0

	var it idleTrigger
	it.last_modified = queue.dm.now
	it.tq_lru_element = queue.idle.PushFront(&f)
	f.state = it

	queue.fired[base_trace_id] = &f
	queue.metrics.count++

	return &f
}

/* Transition to reporting */
func (it idleTrigger) buffersAdded(f *FiredTrigger) firedtriggerstate {
	f.queue.idle.Remove(it.tq_lru_element)
	f.queue.reporting.Insert(f.id.base_trace_id)

	var rt reportingTrigger
	return rt
}

func (it idleTrigger) getBuffersForReport(f *FiredTrigger) (firedtriggerstate, []int) {
	log.Fatal("Attempted to takeBuffers for idleTrigger")
	return nil, nil
}

func (it idleTrigger) evictTrigger(f *FiredTrigger) (firedtriggerstate, []int) {
	log.Fatal("idleTrigger cannot be evicted")
	return nil, nil
}

func (it idleTrigger) checkTimeout(f *FiredTrigger, before time.Time) (firedtriggerstate, bool) {
	if it.last_modified.After(before) {
		return it, false // Hasn't timed out yet
	}

	// Unhook from all traces
	for _, t := range f.traces {
		t.RemoveTrigger(f.queue.dm, f) // Shouldn't return any buffers
	}

	// Delete the fired trigger
	f.queue.idle.Remove(it.tq_lru_element)
	delete(f.queue.fired, f.id.base_trace_id)

	return nil, true
}

/*
A trigger transitions to reporting when one or more of its traces
has data to be reported.  A reportingTrigger may remain in this state
despite data being reported already, because traces can belong to more
than one trigger
*/
type reportingTrigger struct {
}

/* This trigger is already in a reporting state, so more buffers doesn't
change our state */
func (rt reportingTrigger) buffersAdded(f *FiredTrigger) firedtriggerstate {
	return rt
}

/* Get all buffers pending for report, then transition to idle.
TODO: no reason why we have to do ALL traces at a time, could do a subset */
func (rt reportingTrigger) getBuffersForReport(f *FiredTrigger) (firedtriggerstate, []int) {
	var buffers []int
	for _, t := range f.traces {
		buffers = append(buffers, t.TakeBuffers(f.queue.dm)...)
	}

	if f.buffer_count != 0 {
		log.Fatal("Buffers remain after takeBuffers")
		return nil, nil
	}

	var it idleTrigger
	it.last_modified = f.queue.dm.now
	it.tq_lru_element = f.queue.idle.PushFront(f)
	return it, buffers
}

/* Get any buffers that should be evicted.  No transition after this, trigger
becomes invalid */
func (rt reportingTrigger) evictTrigger(f *FiredTrigger) (firedtriggerstate, []int) {
	var buffers []int
	for _, t := range f.traces {
		buffers = append(buffers, t.RemoveTrigger(f.queue.dm, f)...)
	}

	// No more buffers should be associated with this trigger
	if f.buffer_count != 0 {
		log.Fatal("Buffers remain after eviction")
	}

	delete(f.queue.fired, f.id.base_trace_id)

	return nil, buffers
}

func (rt reportingTrigger) checkTimeout(f *FiredTrigger, before time.Time) (firedtriggerstate, bool) {
	return rt, false // Not allowed to time out reportingTriggers
}
