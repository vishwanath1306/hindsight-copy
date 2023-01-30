package agent

import (
	"container/list"
	"log"
	"time"
)

/* We represent a trace with a simple state machine that
can transition between untriggered (default), triggered (no data),
and reporting (triggered and reporting data) */
type Trace struct {
	id    uint64
	state tracestate
}

/* Implements state machine transitions of a trace.  States are:
    * untriggeredTrace
		* triggeredTrace
		* reportingTrace
		* nil (invalid)
*/
type tracestate interface {
	addBuffers(dm *DataManager, trace *Trace, buffers []int) tracestate
	addBreadcrumbs(dm *DataManager, trace *Trace, breadcrumbs []string) (tracestate, []string)
	addTrigger(dm *DataManager, trace *Trace, f *FiredTrigger) (tracestate, []string)
	removeTrigger(dm *DataManager, trace *Trace, f *FiredTrigger) (tracestate, []int)
	takeBuffers(dm *DataManager, trace *Trace) (tracestate, []int)
}

func (t *Trace) AddBuffers(dm *DataManager, buffers []int) {
	t.state = t.state.addBuffers(dm, t, buffers)
}

/* Adds breadcrumbs to the trace.
If the trace is untriggered then the breadcrumbs are buffered.
If the trace is triggered, this method returns breadcrumbs that
must be immediately reported */
func (t *Trace) AddBreadcrumbs(dm *DataManager, breadcrumbs []string) []string {
	var to_report []string
	t.state, to_report = t.state.addBreadcrumbs(dm, t, breadcrumbs)
	return to_report
}

/* Triggers the trace.  Returns any breadcrumbs that must be
immediately reported */
func (t *Trace) AddTrigger(dm *DataManager, f *FiredTrigger) []string {
	var to_report []string
	t.state, to_report = t.state.addTrigger(dm, t, f)
	return to_report
}

/* Untriggers the trace.  If there are no other triggers, then
removes and returns the trace's buffers */
func (t *Trace) RemoveTrigger(dm *DataManager, f *FiredTrigger) []int {
	var buffers []int
	t.state, buffers = t.state.removeTrigger(dm, t, f)
	return buffers
}

/* Removes and returns the trace's buffers for reporting */
func (t *Trace) TakeBuffers(dm *DataManager) []int {
	var buffers []int
	t.state, buffers = t.state.takeBuffers(dm, t)
	return buffers
}

/* An untriggeredTrace simply accumulates buffer and breadcrumb data until
it is eventually either triggered or evicted */
type untriggeredTrace struct {
	/* Trace data */
	buffers     []int
	breadcrumbs []string

	/* For eviction from the data manager */
	last_modified  time.Time
	dm_lru_element *list.Element
}

/* Called by the data manager the first time a trace is seen */
func initUntriggeredTrace(dm *DataManager, trace_id uint64) *Trace {
	var t Trace
	t.id = trace_id

	dm.traces[trace_id] = &t

	var ut untriggeredTrace
	ut.buffers = nil
	ut.breadcrumbs = nil
	ut.last_modified = dm.now
	ut.dm_lru_element = dm.untriggered.lru.PushFront(&t)
	t.state = ut

	dm.trace_count += 1
	dm.untriggered.trace_count += 1

	return &t
}

/* If a trace is untriggered, we simply accumulate buffers and update
the LRU in the datamanager */
func (t untriggeredTrace) addBuffers(dm *DataManager, trace *Trace, buffers []int) tracestate {
	t.buffers = append(t.buffers, buffers...)
	t.last_modified = dm.now
	dm.untriggered.lru.MoveToFront(t.dm_lru_element)
	dm.buffer_count += len(buffers)
	dm.untriggered.buffer_count += len(buffers)
	return t
}

/* If a trace is untriggered, we simply accumulate breadcrumbs and update
the LRU in the datamanager.  No breadcrumbs need to be reported immediately */
func (t untriggeredTrace) addBreadcrumbs(dm *DataManager, trace *Trace, breadcrumbs []string) (tracestate, []string) {
	t.breadcrumbs = append(t.breadcrumbs, breadcrumbs...)
	t.last_modified = dm.now
	dm.untriggered.lru.MoveToFront(t.dm_lru_element)
	return t, nil
}

/* Adding a trigger means transitioning from an untriggered to a triggered trace.
We remove the trace from the untriggered LRU, transition to the new state, and return
any breadcrumbs accumulated by the untriggeredTrace. */
func (ut untriggeredTrace) addTrigger(dm *DataManager, trace *Trace, fired *FiredTrigger) (tracestate, []string) {
	// Remove from the datamanager's untriggered LRU
	dm.untriggered.lru.Remove(ut.dm_lru_element)

	// Update counters
	dm.untriggered.trace_count -= 1
	dm.untriggered.buffer_count -= len(ut.buffers)
	dm.triggered.trace_count += 1
	dm.triggered.buffer_count += len(ut.buffers)
	fired.queue.trace_count += 1

	// Transition to the new state.  In both cases we return the trace's
	// breadcrumbs for immediate dissemination.
	if len(ut.buffers) > 0 {
		// If there are buffers then we transition to reporting
		fired.buffersAdded(len(ut.buffers))
		var t reportingTrace
		t.buffers = ut.buffers
		t.triggers = map[TriggerID]*FiredTrigger{fired.id: fired}
		return t, ut.breadcrumbs
	} else {
		// If there are no buffers then we transition to triggered
		var t triggeredTrace
		t.triggers = map[TriggerID]*FiredTrigger{fired.id: fired}
		return t, ut.breadcrumbs
	}
}

/* In an untriggered state, takeBuffers is used when evicting a trace,  */
func (ut untriggeredTrace) takeBuffers(dm *DataManager, trace *Trace) (tracestate, []int) {
	buffers := ut.buffers
	ut.buffers = nil
	dm.trace_count -= 1
	dm.buffer_count -= len(buffers)
	dm.untriggered.trace_count -= 1
	dm.untriggered.buffer_count -= len(buffers)
	dm.untriggered.lru.Remove(ut.dm_lru_element)
	dm.untriggered.event_horizion = ut.last_modified
	delete(dm.traces, trace.id)
	return ut, buffers
}

/* Invalid transition for untriggeredTrace */
func (ut untriggeredTrace) removeTrigger(dm *DataManager, trace *Trace, fired *FiredTrigger) (tracestate, []int) {
	log.Fatal("Cannot remove a trigger from an untriggered trace")
	return nil, nil
}

/* A triggeredTrace is one that is included in a FiredTrigger but
does not have any data to report.  When new data comes in, it will transition
to reportingTrace.  If new breadcrumbs come in, they will be immediately reported.
Eventually the trigger manager might untrigger the trace, in which case it will
transition to untriggeredTrace */
type triggeredTrace struct {
	/* The triggers that include this trace */
	triggers map[TriggerID]*FiredTrigger
}

/* If a trace is triggered, adding buffers means we must transition to reporting */
func (t triggeredTrace) addBuffers(dm *DataManager, trace *Trace, buffers []int) tracestate {
	// Update datamanager statistics
	bufcount := len(buffers)
	dm.buffer_count += bufcount
	dm.triggered.buffer_count += bufcount

	// Inform all triggers that they need to report data, which potentially
	// transitions them into reporting state
	for _, f := range t.triggers {
		f.buffersAdded(bufcount)
	}

	// Create the new state
	var rt reportingTrace
	rt.triggers = t.triggers
	rt.buffers = buffers
	return rt
}

/* If a trace is triggered, we immediately report breadcrumbs */
func (t triggeredTrace) addBreadcrumbs(dm *DataManager, trace *Trace, breadcrumbs []string) (tracestate, []string) {
	return t, breadcrumbs
}

/* If a trace is already triggered, adding a trigger does little extra */
func (t triggeredTrace) addTrigger(dm *DataManager, trace *Trace, fired *FiredTrigger) (tracestate, []string) {
	if _, ok := t.triggers[fired.id]; !ok {
		t.triggers[fired.id] = fired
		fired.queue.trace_count += 1
	}
	return t, nil
}

func (t triggeredTrace) takeBuffers(dm *DataManager, trace *Trace) (tracestate, []int) {
	return t, nil
}

/* In a triggered state, removeTrigger is used when expiring a trigger from timeout or eviction.
   If there are no other triggers, then the trace expires after this call. */
func (t triggeredTrace) removeTrigger(dm *DataManager, trace *Trace, f *FiredTrigger) (tracestate, []int) {
	if _, ok := t.triggers[f.id]; !ok {
		log.Fatal("Attempted to removeTrigger for f that does not exist")
	}

	delete(t.triggers, f.id)
	f.queue.trace_count -= 1

	if len(t.triggers) == 0 {
		// No triggers remaining; trace is evicted
		dm.trace_count -= 1
		dm.triggered.trace_count -= 1
		delete(dm.traces, trace.id)

		return nil, nil
	} else {
		// Some triggers still remain, trace is not evicted
		return t, nil
	}
}

/* A reportingTrace is one that is included in a FiredTrigger and
has data pending to report.  It will accumulate data until eventually being reported.
If new breadcrumbs come in, they will be immediately reported.  Once data is
reported, it will transition to triggeredTrace. */
type reportingTrace struct {
	/* Buffers to be reported */
	buffers []int

	/* The triggers that include this trace */
	triggers map[TriggerID]*FiredTrigger
}

/* If a trace is reporting, adding buffers simply adds to the data pending
to be reported */
func (rt reportingTrace) addBuffers(dm *DataManager, trace *Trace, buffers []int) tracestate {
	// Update datamanager statistics
	bufcount := len(buffers)
	dm.buffer_count += bufcount
	dm.triggered.buffer_count += bufcount

	// Save the buffers alongside the existing ones that are awaiting reporting
	rt.buffers = append(rt.buffers, buffers...)

	// Inform triggers of new buffers
	for _, f := range rt.triggers {
		f.buffersAdded(len(buffers))
	}
	return rt
}

/* If a trace is reporting, we immediately report breadcrumbs */
func (rt reportingTrace) addBreadcrumbs(dm *DataManager, trace *Trace, breadcrumbs []string) (tracestate, []string) {
	return rt, breadcrumbs
}

/* If a trace is already reporting for other triggers, then this trigger needs
to be reporting too */
func (rt reportingTrace) addTrigger(dm *DataManager, trace *Trace, fired *FiredTrigger) (tracestate, []string) {
	if _, ok := rt.triggers[fired.id]; !ok {
		rt.triggers[fired.id] = fired
		fired.queue.trace_count += 1
		fired.buffersAdded(len(rt.buffers))
	}
	return rt, nil
}

/* takeBuffers is used when actually reporting the data, and transitions us back
to triggered state */
func (rt reportingTrace) takeBuffers(dm *DataManager, trace *Trace) (tracestate, []int) {
	bufcount := len(rt.buffers)
	dm.buffer_count -= bufcount
	dm.triggered.buffer_count -= bufcount

	// Inform triggers of reported buffers
	for _, f := range rt.triggers {
		f.buffersRemoved(bufcount)
	}

	var t triggeredTrace
	t.triggers = rt.triggers
	return t, rt.buffers
}

/* In a triggered state, removeTrigger is used when evicting a trigger.
   If there are no other triggers, then the trace expires after this call. */
func (rt reportingTrace) removeTrigger(dm *DataManager, trace *Trace, f *FiredTrigger) (tracestate, []int) {
	if _, ok := rt.triggers[f.id]; !ok {
		log.Fatal("Attempted to removeTrigger in reportingTrace for f that does not exist")
	}

	delete(rt.triggers, f.id)
	f.queue.trace_count -= 1

	f.buffersRemoved(len(rt.buffers))

	if len(rt.triggers) == 0 {
		// No triggers remaining; evict the trace's buffers
		dm.trace_count -= 1
		dm.buffer_count -= len(rt.buffers)
		dm.triggered.trace_count -= 1
		dm.triggered.buffer_count -= len(rt.buffers)
		delete(dm.traces, trace.id)

		return nil, rt.buffers
	} else {
		// Some triggers remain, its buffers remain valid
		return rt, nil
	}
}
