package agent

import (
	"container/list"
	"time"

	"github.com/geraldleizhang/hindsight/agent/pkg/util"
)

type TriggerQueue struct {
	id           int
	trace_count  int
	buffer_count int
	fired        map[uint64]*FiredTrigger
	reporting    *util.TreeNode // queue for FiredTriggers with data to report
	idle         *list.List     // LRU for idle FiredTriggers
	dm           *DataManager
	metrics      TriggerMetrics
}

/*
Initializes a trigger queue instance.  Only called the first time we see a trigger
for a given queue.
*/
func (dm *DataManager) initTriggerQueue(id int) *TriggerQueue {
	var queue TriggerQueue
	queue.id = id
	queue.buffer_count = 0
	queue.fired = make(map[uint64]*FiredTrigger)
	queue.reporting = util.InitPartialPriorityTree()
	queue.idle = list.New()
	queue.dm = dm
	dm.triggered.queues[id] = &queue
	return &queue
}

/*
When a trigger fires locally or remotely, we call this method to create (or retrieve existing)
metadata about the fired trigger.  The common case is that for any given base_trace_id,
this method is likely only called once (because it's unlikely to fire a trigger for the same
trace ID multiple times).  Therefore the common case for this method is to create a new
FiredTrigger instance.
*/
func (queue *TriggerQueue) getOrCreateTrigger(base_trace_id uint64) *FiredTrigger {
	if trigger, ok := queue.fired[base_trace_id]; ok {
		return trigger
	} else {
		return queue.initIdleTrigger(base_trace_id)
	}
}

/* Evicts one fired trigger from the queue, and returns buffers to be freed */
func (queue *TriggerQueue) EvictNext() []int {
	id := queue.reporting.PopNearMax()
	if trigger, ok := queue.fired[id]; ok {
		return trigger.Evict()
	} else {
		return nil
	}
}

/* Evicts triggers from the specified queue until the total number of triggered buffers is
reduced below the specified target_capacity */
func (queue *TriggerQueue) EvictToCapacity(target_capacity int) []int {
	if queue.dm.triggered.buffer_count <= target_capacity || target_capacity < 0 {
		return nil
	}

	/*
		We evict in batches for efficiency rather than one at a time; here
		calculate the number to actually evict, rounding up
	*/
	num_to_evict := queue.dm.triggered.buffer_count - target_capacity
	min_to_evict := target_capacity / 100
	if num_to_evict < min_to_evict {
		num_to_evict = min_to_evict
	}

	/*
		Do the eviction; it's possible this can completely drain a queue without
		reaching the target_capacity, which is OK
	*/
	var evicted []int
	eviction_count := 0
	for len(evicted) < num_to_evict && queue.reporting.Size() > 0 {
		id := queue.reporting.PopNearMax()
		trigger := queue.fired[id]
		evicted = append(evicted, trigger.Evict()...)
		eviction_count++
	}

	queue.metrics.evicted += eviction_count
	queue.metrics.evicted_buffers += len(evicted)

	return evicted
}

/* Pops one fired trigger from the specified queue, and returns buffers to be reported and freed */
func (queue *TriggerQueue) ReportNext() []int {
	id := queue.reporting.PopMin()
	if trigger, ok := queue.fired[id]; ok {
		buffers := trigger.GetBuffersForReport()

		queue.metrics.reported_buffers += len(buffers)

		return buffers

	} else {
		return nil
	}
}

/* Remove triggers that have been idle since before the specified time.
Since they are idle, this should not return any buffers; instead returns the number
of idle triggers that were evicted (used only for testing) */
func (queue *TriggerQueue) CheckIdleTriggers(before time.Time) int {
	eviction_count := 0
	for queue.idle.Len() > 0 {
		oldest := queue.idle.Back().Value.(*FiredTrigger)
		if !oldest.CheckTimeout(before) {
			break
		}
		eviction_count += 1
	}
	return eviction_count
}

/* Add a new trigger to the queue or update an existing trigger if it already exist */
func (queue *TriggerQueue) Trigger(trigger_id uint64, trace_ids []uint64) map[uint64][]string {
	trigger := queue.getOrCreateTrigger(trigger_id)
	breadcrumbs := make(map[uint64][]string)
	for _, trace_id := range trace_ids {
		trace := queue.dm.getOrCreateTrace(trace_id)
		to_report := trigger.AddTrace(trace)
		if len(to_report) > 0 {
			breadcrumbs[trace_id] = to_report
		}
	}
	return breadcrumbs
}
