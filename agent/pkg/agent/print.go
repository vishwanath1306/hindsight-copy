package agent

import (
	"fmt"
	"strings"
)

func (t Trace) String() string {
	return fmt.Sprintf("trace %d %s", t.id, t.state)
}

func (ut untriggeredTrace) String() string {
	return fmt.Sprintf("untriggered bufs=%v bcs=%v", ut.buffers, ut.breadcrumbs)
}

func (t triggeredTrace) String() string {
	return fmt.Sprintf("triggered")
}

func (r reportingTrace) String() string {
	return fmt.Sprintf("reporting bufs=%v", r.buffers)
}

func (dm *DataManager) String() string {
	return fmt.Sprintf("DM %d trace(s)\n%v\n%v", len(dm.traces), &dm.triggered, &dm.untriggered)
}

func (ut *UntriggeredData) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(" untriggered %d total:", ut.buffer_count))
	for e := ut.lru.Front(); e != nil; e = e.Next() {
		trace := e.Value.(*Trace)
		sb.WriteString(fmt.Sprintf("\n   %v", trace))
	}
	return sb.String()
}

func (t *TriggeredData) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(" triggered %d buffers %d queues", t.buffer_count, len(t.queues)))
	for _, queue := range t.queues {
		sb.WriteString(fmt.Sprintf("\n  %v", queue))
	}
	return sb.String()
}

func (q *TriggerQueue) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("queue %d with %d fired %d reporting, %d buffers total:", q.id, len(q.fired), q.reporting.Size(), q.buffer_count))
	for _, trigger := range q.fired {
		sb.WriteString(fmt.Sprintf("\n   %v", trigger))
	}
	return sb.String()
}

func (t *FiredTrigger) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("fired %d", t.id))
	for _, trace := range t.traces {
		sb.WriteString(fmt.Sprintf("\n    %v", trace))
	}
	return sb.String()
}
