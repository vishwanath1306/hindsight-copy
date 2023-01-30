package agent

import (
	"container/list"
	"time"
)

/*
The DataManager stores all trace and trigger data.  The DataManager
is not responsible for implementing eviction policies, timing traces out,
and so on - that is handled externally by the Agent
*/
type DataManager struct {
	now          time.Time
	traces       map[uint64]*Trace
	untriggered  UntriggeredData
	triggered    TriggeredData
	trace_count  int
	buffer_count int
}

type UntriggeredData struct {
	lru            *list.List // LRU of untriggered traces
	trace_count    int
	buffer_count   int
	event_horizion time.Time
}

type TriggeredData struct {
	trace_count  int
	buffer_count int
	queues       map[int]*TriggerQueue
}

/* Called upon agent startup */
func InitDataManager() *DataManager {
	var dm DataManager
	dm.Init()
	return &dm
}

func (dm *DataManager) Init() {
	dm.now = time.Now()
	dm.traces = make(map[uint64]*Trace)

	dm.triggered.buffer_count = 0
	dm.triggered.queues = make(map[int]*TriggerQueue)

	dm.untriggered.lru = list.New()
	dm.untriggered.buffer_count = 0
}

/*
We only create the trigger metadata the first time a trigger fires for a
queue ID.  A trigger is never destroyed, for now.

TODO: a buggy trigger could fire for random queueIds resulting in too many
queues being created.  A future fix would be to limit the number of
allowed empty queues and tear them down on an LRU basis.
*/
func (dm *DataManager) GetQueue(queue_id int) *TriggerQueue {
	if queue, ok := dm.triggered.queues[queue_id]; ok {
		return queue
	} else {
		return dm.initTriggerQueue(queue_id)
	}
}

/* Buffers received from the shm queues */
func (dm *DataManager) AddBuffers(trace_id uint64, buffers []int) {
	trace := dm.getOrCreateTrace(trace_id)
	trace.AddBuffers(dm, buffers)
}

/* Breadcrumbs received from the shm queues */
func (dm *DataManager) AddBreadcrumbs(trace_id uint64, breadcrumbs []string) []string {
	trace := dm.getOrCreateTrace(trace_id)
	return trace.AddBreadcrumbs(dm, breadcrumbs)
}

/* A trigger has fired. */
func (dm *DataManager) Trigger(queue_id int, trigger_id uint64, trace_ids []uint64) map[uint64][]string {
	queue := dm.GetQueue(queue_id)
	return queue.Trigger(trigger_id, trace_ids)
}

/*
Evicts one untriggered trace according to the least-recently-used policy.
Returns any buffers of this trace, that must then be freed by the caller.
*/
func (dm *DataManager) Evict() []int {
	if dm.untriggered.trace_count == 0 {
		return nil
	}

	trace := dm.untriggered.lru.Back().Value.(*Trace)
	return trace.TakeBuffers(dm)
}

/*
Repeatedly evicts untriggered traces until the buffer_count of the DataManager
is below the specified target_capacity.  Returns all evicted buffers that must
then be freed by the caller.
*/
func (dm *DataManager) EvictToCapacity(target_capacity int) []int {
	if dm.buffer_count <= target_capacity || target_capacity < 0 {
		return nil
	}

	/*
		We evict in batches for efficiency rather than one at a time; here
		calculate the number to actually evict, rounding up
	*/
	num_to_evict := dm.buffer_count - target_capacity
	min_to_evict := target_capacity / 100
	if num_to_evict < min_to_evict {
		num_to_evict = min_to_evict
	}

	/*
		Do the eviction
	*/
	var evicted []int
	for len(evicted) < num_to_evict && dm.untriggered.lru.Len() > 0 {
		trace := dm.untriggered.lru.Back().Value.(*Trace)
		evicted = append(evicted, trace.TakeBuffers(dm)...)
	}
	return evicted
}

/*
Repeatedly evicts triggers until the buffer_count of dm.triggered is below
the specified target_capacity.  Returns all evicted buffers that must
then be freed by the caller.
*/
func (dm *DataManager) EvictedTriggeredToCapacity(target_capacity int) []int {
	if dm.triggered.buffer_count <= target_capacity || target_capacity < 0 {
		return nil
	}

	/*
		Evict from the current largest queue.  Only evict from the largest queue
		each time; don't try to be clever
	*/
	var queue *TriggerQueue
	for _, candidate := range dm.triggered.queues {
		if queue == nil || candidate.buffer_count > queue.buffer_count {
			queue = candidate
		}
	}
	return queue.EvictToCapacity(target_capacity)
}

/*
Time-out any triggers that have been idle since before the specified time.
Idle triggers don't have any buffers, therefore this simply untriggers them.
If new data arrives in future for the same trace IDs, it will behave like
we've never seen those trace IDs before.
*/
func (dm *DataManager) CheckIdleTriggers(before time.Time) {
	for _, queue := range dm.triggered.queues {
		queue.CheckIdleTriggers(before)
	}
}

func (dm *DataManager) getOrCreateTrace(trace_id uint64) *Trace {
	if trace, ok := dm.traces[trace_id]; ok {
		return trace
	} else {
		return initUntriggeredTrace(dm, trace_id)
	}
}
