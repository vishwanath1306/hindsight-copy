package agent

import (
	"github.com/juju/ratelimit"
)

/*
This file defines the TriggerManager, which wraps the trigger queues of
the DataManager and applies rate limiting of triggers and of reporting.
*/
type TriggerManager struct {
	dm     *DataManager
	queues map[int]*ManagedQueue
	vc     int // Virtual clock used for fair sharing reporting across queues

	buffer_size     int     // Size of buffers in the cache
	trigger_limit   float64 // Default limit an individual queue can trigger per second
	reporting_limit float64 // Default limit an individual queue can report per second
	batch_size      int     // The number of buffers per report
}

/*
Wraps a TriggerQueue in the DataManager
*/
type ManagedQueue struct {
	tm                *TriggerManager
	queue             *TriggerQueue     // The actual DataManager queue
	trigger_limiter   *ratelimit.Bucket // Rate limiter for local triggers
	reporting_limiter *ratelimit.Bucket // Rate limiter for reporting
	vt                int               // Virtual time used for fair sharing of reporting
}

func (tm *TriggerManager) Init(dm *DataManager, buffer_size int, trigger_limit float64) {
	tm.dm = dm
	tm.queues = make(map[int]*ManagedQueue)
	tm.vc = 0
	tm.buffer_size = buffer_size
	tm.trigger_limit = trigger_limit             // TODO: not hardcoded, configured per trigger, or adaptive based on eviction rates
	tm.reporting_limit = 10 * 1024 * 1024 * 1024 // TODO: not hardcoded
	tm.batch_size = 1 + (128*1024)/buffer_size

	if tm.trigger_limit == 0 {
		tm.trigger_limit = 10 * 1024 * 1024 * 1024
	}
}

/*
Set per-trigger rate limits, configured via command line / config parameters

per_trigger_rate_limits is specified in MB/s
*/
func (tm *TriggerManager) ConfigureRateLimits(per_trigger_rate_limits map[int]float64) {
	for queue_id, limit := range per_trigger_rate_limits {
		queue := tm.getQueue(queue_id)
		limit_bytes := limit * 1024 * 1024
		queue.reporting_limiter = ratelimit.NewBucketWithRate(limit_bytes, int64(limit_bytes))
	}
}

func (tm *TriggerManager) getQueue(queue_id int) *ManagedQueue {
	if queue, ok := tm.queues[queue_id]; ok {
		return queue
	}

	var mq ManagedQueue
	mq.tm = tm
	mq.queue = tm.dm.GetQueue(queue_id)
	mq.trigger_limiter = ratelimit.NewBucketWithRate(tm.trigger_limit, int64(tm.trigger_limit))
	mq.reporting_limiter = ratelimit.NewBucketWithRate(tm.reporting_limit, int64(tm.reporting_limit))
	mq.vt = tm.vc

	tm.queues[queue_id] = &mq
	return &mq
}

func (mq *ManagedQueue) TriggerLocal(trigger_id uint64, trace_ids []uint64) (bool, map[uint64][]string) {
	// Rate limit local triggers
	if mq.trigger_limiter.Available() < 0 {
		mq.queue.metrics.dropped += 1
		return false, nil
	}
	mq.trigger_limiter.Take(1)
	mq.queue.metrics.local += 1

	// Send to DataManager, return any breadcrumbs that must be reported
	return true, mq.queue.Trigger(trigger_id, trace_ids)
}

func (mq *ManagedQueue) TriggerRemote(trigger_id uint64, trace_ids []uint64) map[uint64][]string {
	// TODO: consume tokens for remote triggers?
	// mq.trigger_limiter.Take(1)
	mq.queue.metrics.remote += 1

	// Send to DataManager, return any breadcrumbs that must be reported
	return mq.queue.Trigger(trigger_id, trace_ids)
}

/*
Get the next batch of buffers to be reported, up to the specified batch size
*/
func (tm *TriggerManager) GetNextBatchToReport() []int {
	var buffers []int
	for len(buffers) < tm.batch_size && tm.dm.triggered.buffer_count > 0 {
		buffers = append(buffers, tm.getNextBuffersToReport()...)
	}
	return buffers
}

/*
Get the next buffers to be reported
*/
func (tm *TriggerManager) getNextBuffersToReport() []int {
	// Find the next queue to report from based on fair sharing
	var mq *ManagedQueue
	for _, candidate := range tm.queues {
		if candidate.queue.buffer_count == 0 {
			candidate.vt = tm.vc // Catch up virtual clock
		} else {
			/* Apply rate limiting */
			if candidate.reporting_limiter.Available() < 0 {
				candidate.vt = tm.vc
				continue
			}

			/* Apply fair sharing -- pick the queue with lowest virtual time */
			if mq == nil || candidate.vt < mq.vt {
				mq = candidate
			}
		}
	}

	if mq == nil {
		return nil
	}

	buffers := mq.queue.ReportNext()
	if len(buffers) > 0 {
		mq.vt += len(buffers)
		tm.vc = mq.vt // Not fully correct but enough for now
		mq.reporting_limiter.Take(int64(len(buffers) * tm.buffer_size))
	}
	return buffers
}
