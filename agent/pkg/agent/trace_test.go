package agent

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Uint64() uint64 {
	return uint64(rand.Uint32())<<32 + uint64(rand.Uint32())
}

func TestDataManagerFromScratch(t *testing.T) {
	assert := assert.New(t)

	dm := InitDataManager()

	/*
		First, add a trace, and check it gets inserted correctly
		with the correct buffers and correct counts
	*/
	dm.AddBuffers(75, []int{3, 12})

	assert.Equal(dm.trace_count, 1, "Trace count")
	assert.Equal(dm.buffer_count, 2, "Buffer count")
	assert.Equal(dm.untriggered.trace_count, 1, "Untriggered trace count")
	assert.Equal(dm.untriggered.buffer_count, 2, "Untriggered buffer count")
	assert.Equal(dm.triggered.trace_count, 0, "Triggered trace count")
	assert.Equal(dm.triggered.buffer_count, 0, "Triggered buffer count")

	switch v := dm.traces[75].state.(type) {
	case untriggeredTrace:
		assert.Equal(v.buffers, []int{3, 12}, "Buffers are correct")
	default:
		assert.Fail("Unexpected trace type")
	}

	/*
		Add some more buffers to the trace, check it gets updated
	*/
	dm.AddBuffers(75, []int{55, 2})

	assert.Equal(dm.trace_count, 1, "Trace count")
	assert.Equal(dm.buffer_count, 4, "Buffer count")
	assert.Equal(dm.untriggered.trace_count, 1, "Untriggered trace count")
	assert.Equal(dm.untriggered.buffer_count, 4, "Untriggered buffer count")

	switch v := dm.traces[75].state.(type) {
	case untriggeredTrace:
		assert.Equal(v.buffers, []int{3, 12, 55, 2}, "Buffers are correct")
		assert.Equal(len(v.breadcrumbs), 0, "No breadcrumbs yet")
	default:
		assert.Fail("Unexpected trace type")
	}

	/*
		Add some breadcrumbs, check they get added too, and counts are correct
	*/
	dm.AddBreadcrumbs(75, []string{"hello", "world"})

	assert.Equal(dm.trace_count, 1, "Trace count")
	assert.Equal(dm.buffer_count, 4, "Buffer count")
	assert.Equal(dm.untriggered.trace_count, 1, "Untriggered trace count")
	assert.Equal(dm.untriggered.buffer_count, 4, "Untriggered buffer count")
	assert.Equal(dm.triggered.trace_count, 0, "Triggered trace count")
	assert.Equal(dm.triggered.buffer_count, 0, "Triggered buffer count")

	switch v := dm.traces[75].state.(type) {
	case untriggeredTrace:
		assert.Equal(v.buffers, []int{3, 12, 55, 2}, "Buffers are correct")
		assert.Equal(v.breadcrumbs, []string{"hello", "world"}, "Breadcrumbs are correct")
	default:
		assert.Fail("Unexpected trace type")
	}

	/*
		Add some other traces, check they get added and counts are correct
	*/
	dm.AddBuffers(25, []int{100, 101, 102, 103, 104})
	dm.AddBuffers(50, []int{200, 201, 202, 203, 204, 205, 206, 207})
	dm.AddBreadcrumbs(100, []string{"breadcrumbs", "only"})

	assert.Equal(dm.trace_count, 4, "Trace count")
	assert.Equal(dm.buffer_count, 17, "Buffer count")
	assert.Equal(dm.untriggered.trace_count, 4, "Untriggered trace count")
	assert.Equal(dm.untriggered.buffer_count, 17, "Untriggered buffer count")
	assert.Equal(dm.triggered.trace_count, 0, "Triggered trace count")
	assert.Equal(dm.triggered.buffer_count, 0, "Triggered buffer count")

	assert.Equal(len(dm.triggered.queues), 0, "Shouldn't have triggers yet")

	/*
		Transition to triggered, check that we get back the breadcrumbs immediately
		and that we transition to reporting
	*/
	breadcrumbs := dm.Trigger(1, 75, []uint64{75})

	assert.Equal(dm.trace_count, 4, "Trace count")
	assert.Equal(dm.buffer_count, 17, "Buffer count")
	assert.Equal(dm.untriggered.trace_count, 3, "Untriggered trace count")
	assert.Equal(dm.untriggered.buffer_count, 13, "Untriggered buffer count")
	assert.Equal(dm.triggered.trace_count, 1, "Triggered trace count")
	assert.Equal(dm.triggered.buffer_count, 4, "Triggered buffer count")

	assert.Equal(breadcrumbs, []string{"hello", "world"}, "Breadcrumbs were returned upon triggering")

	assert.Equal(len(dm.triggered.queues), 1, "Trigger queue was created")
	assert.NotNil(dm.triggered.queues[1], "Trigger queue was created")

	q := dm.triggered.queues[1]
	assert.Equal(q.id, 1, "Trigger queue ID created correctly")
	assert.Equal(q.trace_count, 1, "Trigger queue trace count")
	assert.Equal(q.buffer_count, 4, "Trigger queue buffer count")

	assert.Equal(len(q.fired), 1, "FiredTrigger exists")
	assert.NotNil(q.fired[75], "FiredTrigger exists")

	trigger := q.fired[75]
	assert.Equal(trigger.id, TriggerID{1, uint64(75)}, "FiredTrigger ID")
	assert.Equal(trigger.buffer_count, 4, "FiredTrigger buffer count")
	assert.Equal(trigger.queue, q, "FiredTrigger queue")
	assert.Equal(len(trigger.traces), 1, "Trace was added to FiredTrigger")
	assert.NotNil(trigger.traces[75], "Trace was added to FiredTrigger")

	switch v := dm.traces[75].state.(type) {
	case reportingTrace:
		assert.Equal(v.buffers, []int{3, 12, 55, 2}, "Buffers are correct")
		assert.Equal(len(v.triggers), 1, "Trigger was registered to the trace correctly")
	default:
		assert.Fail("Unexpected trace type, expected to be reportingTrace")
	}

	switch trigger.state.(type) {
	case reportingTrigger:
	default:
		assert.Fail("Unexpected trigger type, expected to be reportingTrigger")
	}
}

/*
Creates a data manager used by most tests.

Trace IDs:
 * [0, 10) have 1, 2, 3, 4, ... buffers respectively, are untriggered
 * [100, 110) have no buffers and are triggered
 * [200, 210) have no buffers and are triggered by multiple
 * [300, 310) have no buffers and are all triggered by one trigger
 * [400, 410) have no buffers and are all triggered by two triggers
 * [500, 510) have buffers and are reporting
 * [600, 610) have buffers and are reporting by multiple
 * [700, 710) have buffers and are all reporting by one trigger
 * [800, 810) have buffers and are all reporting by two triggers
 * [900, 910) are all triggered by one trigger, with [900, 905) reporting and [905, 910) have no buffers
*/
func initDataManagerForTest() *DataManager {
	dm := InitDataManager()

	for i := 0; i < 10; i++ {
		var buffers []int
		for j := 0; j < i+1; j++ {
			buffers = append(buffers, 1000*(i+1)+j)
		}
		dm.AddBuffers(uint64(i), buffers)
	}

	return dm
}

/*
Preconditions:
* Expect the datamanager to be prepopulated with 10 untriggered traces
*/
func TestUntriggeredLRU(t *testing.T) {
	assert := assert.New(t)

	dm := initDataManagerForTest()

	/*
		Preconditions: expect the DM to be prepopulated with 10 untriggered traces
	*/
	assert.Equal(dm.untriggered.trace_count, 10, "Untriggered trace count")

	/*
		Drain all untriggered traces
	*/
	trace_count := dm.trace_count
	buf_count := dm.buffer_count
	untriggered_buf_count := dm.untriggered.buffer_count
	triggered_buf_count := dm.triggered.buffer_count
	triggered_trace_count := dm.triggered.trace_count

	for i := 9; i >= 0; i-- {
		evicted := dm.Evict()

		untriggered_buf_count -= len(evicted)
		buf_count -= len(evicted)
		trace_count -= 1

		assert.Equal(trace_count, dm.trace_count, "Global trace count decremented")
		assert.Equal(i, dm.untriggered.trace_count, "Untriggered trace count decremented")
		assert.Equal(triggered_trace_count, dm.triggered.trace_count, "Triggered trace count remains unchanged")

		assert.Equal(buf_count, dm.buffer_count, "Global buffer count decremented")
		assert.Equal(untriggered_buf_count, dm.untriggered.buffer_count, "Untriggered buffer count decremented")
		assert.Equal(triggered_buf_count, dm.triggered.buffer_count, "Triggered buffer count remains unchanged")
	}

	assert.Equal(0, dm.untriggered.buffer_count, "No untriggered buffers remaining")
	assert.Equal(0, dm.untriggered.trace_count, "No untriggered traces remaining")
	assert.Equal(dm.buffer_count, dm.triggered.buffer_count, "Global buffer count is equal to triggered buffer count")
	assert.Equal(dm.trace_count, dm.triggered.trace_count, "Global trace count is equal to triggered trace count")
	assert.Equal(triggered_buf_count, dm.triggered.buffer_count, "Triggered buffer count remains unchanged")
	assert.Equal(triggered_trace_count, dm.triggered.trace_count, "Triggered trace count remains unchanged")

	/*
		Evict should return nil when nothing to evict and should leave
		triggered traces unaffected
	*/

	evicted := dm.Evict()
	assert.Nil(evicted, "Nothing evicted when nothing exists")

	assert.Equal(0, dm.untriggered.buffer_count, "No untriggered buffers remaining")
	assert.Equal(0, dm.untriggered.trace_count, "No untriggered traces remaining")
	assert.Equal(dm.buffer_count, dm.triggered.buffer_count, "Global buffer count is equal to triggered buffer count")
	assert.Equal(dm.trace_count, dm.triggered.trace_count, "Global trace count is equal to triggered trace count")
	assert.Equal(triggered_buf_count, dm.triggered.buffer_count, "Triggered buffer count remains unchanged")
	assert.Equal(triggered_trace_count, dm.triggered.trace_count, "Triggered trace count remains unchanged")

	/*
		Add some untriggered traces
	*/
	for i := 0; i < 20; i++ {
		dm.AddBuffers(uint64(i), []int{2 * i, 2*i + 1})
		assert.Equal(2*(i+1), dm.untriggered.buffer_count, "Untriggered buffers were added")
		assert.Equal(i+1, dm.untriggered.trace_count, "Untriggered traces were added")
	}

	/*
		Evict should evict in LRU order
	*/
	for i := 0; i < 10; i++ {
		evicted := dm.Evict()
		assert.Equal(2, len(evicted), "Evicted 2 buffers")
		assert.Equal(2*i, evicted[0], "Evicted the right buffers")
		assert.Equal(2*i+1, evicted[1], "Evicted the right buffers")
	}

	/*
		Adding buffers should update LRU
	*/
	for i := 19; i >= 10; i-- {
		dm.AddBuffers(uint64(i), []int{2*i + 2, 2*i + 3})
	}

	/*
		Evict should evict in LRU order
	*/
	for i := 19; i >= 10; i-- {
		evicted := dm.Evict()
		assert.Equal(4, len(evicted), "Evicted 4 buffers")
		assert.Equal(2*i, evicted[0], "Evicted the right buffers")
		assert.Equal(2*i+1, evicted[1], "Evicted the right buffers")
		assert.Equal(2*i+2, evicted[2], "Evicted the right buffers")
		assert.Equal(2*i+3, evicted[3], "Evicted the right buffers")
	}

	/*
		Shouldn't be any untriggered buffers or traces remaining
	*/
	assert.Equal(0, dm.untriggered.buffer_count, "No untriggered buffers remaining")
	assert.Equal(0, dm.untriggered.trace_count, "No untriggered traces remaining")
	assert.Equal(dm.buffer_count, dm.triggered.buffer_count, "Global buffer count is equal to triggered buffer count")
	assert.Equal(dm.trace_count, dm.triggered.trace_count, "Global trace count is equal to triggered trace count")
	assert.Equal(triggered_buf_count, dm.triggered.buffer_count, "Triggered buffer count remains unchanged")
	assert.Equal(triggered_trace_count, dm.triggered.trace_count, "Triggered trace count remains unchanged")

}

func TestTriggeredArentEvicted(t *testing.T) {
	assert := assert.New(t)

	dm := initDataManagerForTest()

	/*
		Preconditions: expect the DM to be prepopulated with 10 untriggered traces
		with trace IDs [0, 10)
	*/
	assert.Equal(dm.untriggered.trace_count, 10, "Untriggered trace count")

	trace_count_before := dm.trace_count
	buf_count_before := dm.buffer_count

	/*
		Trigger all untriggered traces
	*/
	for i := 0; i < 10; i++ {
		dm.Trigger(0, uint64(i), []uint64{uint64(i)})
	}

	assert.Equal(0, dm.untriggered.trace_count, "Expect no untriggered traces")
	assert.Equal(0, dm.untriggered.buffer_count, "Expect no untriggered traces")
	assert.Equal(trace_count_before, dm.trace_count, "Total traces remains unchanged")
	assert.Equal(buf_count_before, dm.buffer_count, "Total buffers remains unchanged")
	assert.Equal(dm.trace_count, dm.triggered.trace_count, "All traces are triggered")
	assert.Equal(dm.buffer_count, dm.triggered.buffer_count, "All buffers are triggered")

	/*
		Evict should return nil when nothing to evict and should leave
		triggered traces unaffected
	*/

	evicted := dm.Evict()
	assert.Nil(evicted, "Nothing evicted when nothing exists")

	assert.Equal(0, dm.untriggered.trace_count, "Expect no untriggered traces")
	assert.Equal(0, dm.untriggered.buffer_count, "Expect no untriggered traces")
	assert.Equal(trace_count_before, dm.trace_count, "Total traces remains unchanged")
	assert.Equal(buf_count_before, dm.buffer_count, "Total buffers remains unchanged")
	assert.Equal(dm.trace_count, dm.triggered.trace_count, "All traces are triggered")
	assert.Equal(dm.buffer_count, dm.triggered.buffer_count, "All buffers are triggered")
}

func TestMultipleTriggers(t *testing.T) {
	assert := assert.New(t)

	dm := initDataManagerForTest()

	dm.AddBuffers(uint64(75), []int{1, 2, 3, 4, 5})

	/*
		Trigger on two queues, multiple times
	*/

	dm.Trigger(1, uint64(75), []uint64{uint64(75)})
	dm.Trigger(1, uint64(75), []uint64{uint64(75)})
	dm.Trigger(1, uint64(75), []uint64{uint64(75)})
	dm.Trigger(1, uint64(75), []uint64{uint64(75)})
	dm.Trigger(2, uint64(75), []uint64{uint64(75)})
	dm.Trigger(2, uint64(75), []uint64{uint64(75)})
	dm.Trigger(2, uint64(75), []uint64{uint64(75)})
	dm.Trigger(2, uint64(75), []uint64{uint64(75)})

	assert.Equal(2, len(dm.triggered.queues), "Two queues exist")
	assert.Equal(1, dm.triggered.queues[1].trace_count, "Trace is in first queue")
	assert.Equal(5, dm.triggered.queues[1].buffer_count, "Trace buffers are in first queue")
	assert.Equal(1, dm.triggered.queues[2].trace_count, "Trace is in second queue")
	assert.Equal(5, dm.triggered.queues[2].buffer_count, "Trace buffers are in second queue")
	assert.Equal(1, dm.triggered.trace_count, "Only one trace is triggered")
	assert.Equal(5, dm.triggered.buffer_count, "Only 5 buffers are triggered")

	assert.Equal(0, len(dm.triggered.queues[1].EvictNext()), "First trigger eviction doesn't drop buffers")
	assert.Equal(5, len(dm.triggered.queues[2].EvictNext()), "Second trigger eviction drops buffers")
}

func TestMultipleTriggers2(t *testing.T) {
	assert := assert.New(t)

	dm := initDataManagerForTest()

	dm.AddBuffers(uint64(75), []int{1, 2, 3, 4, 5})
	dm.AddBuffers(uint64(76), []int{7, 8, 9})

	/*
		Trigger on two queues, multiple times
	*/

	dm.Trigger(1, uint64(75), []uint64{uint64(75)})
	dm.Trigger(2, uint64(76), []uint64{uint64(75), uint64(76)})
	dm.Trigger(3, uint64(76), []uint64{uint64(76)})

	q := dm.triggered.queues

	assert.Equal(3, len(q), "Two queues exist")
	assert.Equal(1, q[1].trace_count, "Trace is in first queue")
	assert.Equal(5, q[1].buffer_count, "Trace buffers are in first queue")
	assert.Equal(2, q[2].trace_count, "Traces are in second queue")
	assert.Equal(8, q[2].buffer_count, "Trace buffers are in second queue")
	assert.Equal(1, q[3].trace_count, "Trace is in third queue")
	assert.Equal(3, q[3].buffer_count, "Trace buffers are in third queue")
	assert.Equal(2, dm.triggered.trace_count, "Only two traces are triggered")
	assert.Equal(8, dm.triggered.buffer_count, "Only 8 buffers are triggered")

	q[1].ReportNext()
	assert.Equal(1, q[1].trace_count, "Trace remains in first queue")
	assert.Equal(0, q[1].buffer_count, "No buffers to report for first queue")
	assert.Equal(0, q[1].reporting.Size(), "No triggers to report in first queue")
	assert.Equal(2, q[2].trace_count, "Traces are in second queue")
	assert.Equal(3, q[2].buffer_count, "Some trace buffers remain in second queue")
	assert.Equal(1, q[2].reporting.Size(), "1 triggers to report in first queue")
	assert.Equal(1, q[3].trace_count, "Trace is in third queue")
	assert.Equal(3, q[3].buffer_count, "Trace buffers are in third queue")
	assert.Equal(1, q[3].reporting.Size(), "1 trigger to report in third queue")
	assert.Equal(2, dm.triggered.trace_count, "Only two traces are triggered")
	assert.Equal(3, dm.triggered.buffer_count, "Only 3 buffers are triggered")

	q[3].ReportNext()
	assert.Equal(1, q[1].trace_count, "Trace remains in first queue")
	assert.Equal(0, q[1].buffer_count, "No buffers to report for first queue")
	assert.Equal(0, q[1].reporting.Size(), "No triggers to report in first queue")
	assert.Equal(2, q[2].trace_count, "Traces are in second queue")
	assert.Equal(0, q[2].buffer_count, "No buffers to report in second queue")
	assert.Equal(1, q[2].reporting.Size(), "1 trigger to report in second queue")
	assert.Equal(1, q[3].trace_count, "Trace is in third queue")
	assert.Equal(0, q[3].buffer_count, "No buffers to report in third queue")
	assert.Equal(0, q[3].reporting.Size(), "No triggers to report in third queue")
	assert.Equal(2, dm.triggered.trace_count, "Only two traces are triggered")
	assert.Equal(0, dm.triggered.buffer_count, "No buffers are triggered")

	q[2].ReportNext()
	assert.Equal(1, q[1].trace_count, "Trace remains in first queue")
	assert.Equal(0, q[1].buffer_count, "No buffers to report for first queue")
	assert.Equal(0, q[1].reporting.Size(), "No triggers to report in first queue")
	assert.Equal(2, q[2].trace_count, "Traces are in second queue")
	assert.Equal(0, q[2].buffer_count, "No buffers to report in second queue")
	assert.Equal(0, q[2].reporting.Size(), "1 trigger to report in second queue")
	assert.Equal(1, q[3].trace_count, "Trace is in third queue")
	assert.Equal(0, q[3].buffer_count, "No buffers to report in third queue")
	assert.Equal(0, q[3].reporting.Size(), "No triggers to report in third queue")
	assert.Equal(2, dm.triggered.trace_count, "Only two traces are triggered")
	assert.Equal(0, dm.triggered.buffer_count, "No buffers are triggered")

	dm.AddBuffers(uint64(75), []int{1, 2, 3, 4, 5})
	assert.Equal(1, q[1].trace_count, "Trace is in first queue")
	assert.Equal(5, q[1].buffer_count, "Buffers are in first queue")
	assert.Equal(1, q[1].reporting.Size(), "No triggers to report in first queue")
	assert.Equal(2, q[2].trace_count, "Traces are in second queue")
	assert.Equal(5, q[2].buffer_count, "Buffers to report in second queue")
	assert.Equal(1, q[2].reporting.Size(), "1 trigger to report in second queue")
	assert.Equal(1, q[3].trace_count, "Trace is in third queue")
	assert.Equal(0, q[3].buffer_count, "No buffers to report in third queue")
	assert.Equal(0, q[3].reporting.Size(), "No triggers to report in third queue")
	assert.Equal(2, dm.triggered.trace_count, "Only two traces are triggered")
	assert.Equal(5, dm.triggered.buffer_count, "5 buffers are triggered")

	dm.AddBuffers(uint64(76), []int{7, 8, 9})
	assert.Equal(1, q[1].trace_count, "Trace is in first queue")
	assert.Equal(5, q[1].buffer_count, "Trace buffers are in first queue")
	assert.Equal(2, q[2].trace_count, "Traces are in second queue")
	assert.Equal(8, q[2].buffer_count, "Trace buffers are in second queue")
	assert.Equal(1, q[3].trace_count, "Trace is in third queue")
	assert.Equal(3, q[3].buffer_count, "Trace buffers are in third queue")
	assert.Equal(2, dm.triggered.trace_count, "Only two traces are triggered")
	assert.Equal(8, dm.triggered.buffer_count, "Only 8 buffers are triggered")

	evicted := q[2].EvictNext()
	assert.Equal(0, len(evicted), "No buffers were evicted yet")
	assert.Equal(1, q[1].trace_count, "Trace remains in first queue")
	assert.Equal(5, q[1].buffer_count, "5 buffers to report for first queue")
	assert.Equal(1, q[1].reporting.Size(), "1 trigger to report in first queue")
	assert.Equal(0, q[2].trace_count, "Traces are not in second queue")
	assert.Equal(0, q[2].buffer_count, "No buffers to report in second queue")
	assert.Equal(0, q[2].reporting.Size(), "No triggers to report in second queue")
	assert.Equal(1, q[3].trace_count, "Trace is in third queue")
	assert.Equal(3, q[3].buffer_count, "3 buffers to report in third queue")
	assert.Equal(1, q[3].reporting.Size(), "1 trigger to report in third queue")
	assert.Equal(2, dm.triggered.trace_count, "Only two traces are triggered")
	assert.Equal(8, dm.triggered.buffer_count, "8 buffers are triggered")

	evicted = q[1].EvictNext()
	assert.Equal(5, len(evicted), "5 buffers were evicted")
	assert.Equal(0, q[1].trace_count, "Trace is not in first queue")
	assert.Equal(0, q[1].buffer_count, "No buffers to report for first queue")
	assert.Equal(0, q[1].reporting.Size(), "No trigger to report in first queue")
	assert.Equal(0, q[2].trace_count, "Traces are not in second queue")
	assert.Equal(0, q[2].buffer_count, "No buffers to report in second queue")
	assert.Equal(0, q[2].reporting.Size(), "No triggers to report in second queue")
	assert.Equal(1, q[3].trace_count, "Trace is in third queue")
	assert.Equal(3, q[3].buffer_count, "3 buffers to report in third queue")
	assert.Equal(1, q[3].reporting.Size(), "1 trigger to report in third queue")
	assert.Equal(1, dm.triggered.trace_count, "Only one trace is triggered")
	assert.Equal(3, dm.triggered.buffer_count, "3 buffers are triggered")

	evicted = q[3].EvictNext()
	assert.Equal(3, len(evicted), "3 buffers were evicted")
	assert.Equal(0, q[1].trace_count, "Trace is not in first queue")
	assert.Equal(0, q[1].buffer_count, "No buffers to report for first queue")
	assert.Equal(0, q[1].reporting.Size(), "No trigger to report in first queue")
	assert.Equal(0, q[2].trace_count, "Traces are not in second queue")
	assert.Equal(0, q[2].buffer_count, "No buffers to report in second queue")
	assert.Equal(0, q[2].reporting.Size(), "No triggers to report in second queue")
	assert.Equal(0, q[3].trace_count, "Trace is not in third queue")
	assert.Equal(0, q[3].buffer_count, "No buffers to report in third queue")
	assert.Equal(0, q[3].reporting.Size(), "No trigger to report in third queue")
	assert.Equal(0, dm.triggered.trace_count, "No traces are triggered")
	assert.Equal(0, dm.triggered.buffer_count, "No buffers are triggered")
}

func TestDataManagerEvictionToTargetCapacity(t *testing.T) {
	assert := assert.New(t)

	dm := InitDataManager()

	for i := 0; i < 1000; i++ {
		dm.AddBuffers(uint64(i), []int{i})
		dm.Trigger(1, uint64(i), []uint64{uint64(i)})
	}

	assert.Equal(1000, dm.triggered.trace_count, "1000 triggered traces")
	assert.Equal(1000, dm.triggered.buffer_count, "1000 triggered buffers")

	for i := 0; i < 10; i++ {
		bufs := dm.triggered.queues[1].ReportNext()
		assert.Equal([]int{i}, bufs, fmt.Sprintf("Report trace %d", i))
	}

	evicted := dm.triggered.queues[1].EvictToCapacity(900)
	assert.Equal(90, len(evicted), "Evicted 90 buffers")

	evicted = dm.triggered.queues[1].EvictToCapacity(801)
	assert.Equal(99, len(evicted), "Evicted 99 buffers")

	evicted = dm.triggered.queues[1].EvictToCapacity(800)
	assert.Equal(8, len(evicted), "Evicted 8 buffers")
	assert.Equal(793, dm.triggered.buffer_count, "792 buffers remain")

}

func TestDataManagerEvictionPriority(t *testing.T) {
	assert := assert.New(t)

	dm := InitDataManager()

	for i := 0; i < 1000; i++ {
		dm.AddBuffers(uint64(i), []int{i})
		dm.Trigger(1, uint64(i), []uint64{uint64(i)})
	}

	assert.Equal(1000, dm.triggered.trace_count, "1000 triggered traces")
	assert.Equal(1000, dm.triggered.buffer_count, "1000 triggered buffers")

	for i := 0; i < 400; i++ {
		bufs := dm.triggered.queues[1].ReportNext()
		assert.Equal([]int{i}, bufs, fmt.Sprintf("Report trace %d", i))

		evicted := dm.triggered.queues[1].EvictNext()
		assert.Greater(evicted[0], 500, "Evict a trace with ID > 500")
	}
}

func TestDataManagerReportingPriority(t *testing.T) {
	assert := assert.New(t)

	dm := InitDataManager()

	dm.AddBuffers(77, []int{7, 8, 9})
	dm.Trigger(1, uint64(77), []uint64{uint64(77)})

	dm.AddBuffers(79, []int{21, 22, 23})
	dm.Trigger(1, uint64(79), []uint64{uint64(79)})

	dm.AddBuffers(75, []int{1, 2, 3, 4, 5})
	dm.Trigger(1, uint64(75), []uint64{uint64(75)})

	dm.AddBuffers(76, []int{6})
	dm.Trigger(1, uint64(76), []uint64{uint64(76)})

	dm.AddBuffers(78, []int{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20})
	dm.Trigger(1, uint64(78), []uint64{uint64(78)})

	assert.Equal(1, len(dm.triggered.queues), "One queue exists")
	assert.Equal(5, len(dm.triggered.queues[1].fired), "5 Fired triggers")
	assert.Equal(5, dm.triggered.queues[1].trace_count, "Traces are all in queue")
	assert.Equal(23, dm.triggered.queues[1].buffer_count, "Buffers are all in queue")
	assert.Equal(5, dm.triggered.trace_count, "Traces are all triggered")
	assert.Equal(23, dm.triggered.buffer_count, "Buffers are all triggered")
	assert.Equal(0, dm.triggered.queues[1].idle.Len(), "No idle triggers")

	bufs := dm.triggered.queues[1].ReportNext()
	assert.Equal([]int{1, 2, 3, 4, 5}, bufs, "Expect trace ID 75 to be reported first")
	assert.Equal(5, dm.triggered.trace_count, "Traces are all triggered")
	assert.Equal(5, len(dm.triggered.queues[1].fired), "5 Fired triggers")
	assert.Equal(18, dm.triggered.queues[1].buffer_count, "Buffers are all in queue")
	assert.Equal(18, dm.triggered.buffer_count, "Buffers are all triggered")
	assert.Equal(1, dm.triggered.queues[1].idle.Len(), "Reported trigger is now idle")

	bufs = dm.triggered.queues[1].ReportNext()
	assert.Equal([]int{6}, bufs, "Expect trace ID 76 to be reported next")
	assert.Equal(5, dm.triggered.trace_count, "Traces are all triggered")
	assert.Equal(5, len(dm.triggered.queues[1].fired), "5 Fired triggers")
	assert.Equal(17, dm.triggered.queues[1].buffer_count, "Buffers are all in queue")
	assert.Equal(17, dm.triggered.buffer_count, "Buffers are all triggered")
	assert.Equal(2, dm.triggered.queues[1].idle.Len(), "Reported trigger is now idle")

	bufs = dm.triggered.queues[1].ReportNext()
	assert.Equal([]int{7, 8, 9}, bufs, "Expect trace ID 77 to be reported next")
	assert.Equal(5, dm.triggered.trace_count, "Traces are all triggered")
	assert.Equal(5, len(dm.triggered.queues[1].fired), "5 Fired triggers")
	assert.Equal(14, dm.triggered.queues[1].buffer_count, "Buffers are all in queue")
	assert.Equal(14, dm.triggered.buffer_count, "Buffers are all triggered")
	assert.Equal(3, dm.triggered.queues[1].idle.Len(), "Reported trigger is now idle")

	dm.AddBuffers(80, []int{24, 25, 26, 27, 28})
	dm.Trigger(1, uint64(80), []uint64{uint64(80)})
	assert.Equal(6, dm.triggered.trace_count, "Traces are all triggered")
	assert.Equal(6, len(dm.triggered.queues[1].fired), "5 Fired triggers")
	assert.Equal(19, dm.triggered.queues[1].buffer_count, "Buffers are all in queue")
	assert.Equal(19, dm.triggered.buffer_count, "Buffers are all triggered")
	assert.Equal(3, dm.triggered.queues[1].idle.Len(), "Reported trigger is now idle")

	bufs = dm.triggered.queues[1].ReportNext()
	assert.Equal([]int{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}, bufs, "Expect trace ID 78 to be reported next")
	assert.Equal(6, dm.triggered.trace_count, "Traces are all triggered")
	assert.Equal(6, len(dm.triggered.queues[1].fired), "5 Fired triggers")
	assert.Equal(8, dm.triggered.queues[1].buffer_count, "Buffers are all in queue")
	assert.Equal(8, dm.triggered.buffer_count, "Buffers are all triggered")
	assert.Equal(4, dm.triggered.queues[1].idle.Len(), "Reported trigger is now idle")

	dm.AddBuffers(75, []int{29, 30})
	assert.Equal(6, dm.triggered.trace_count, "Traces are all triggered")
	assert.Equal(6, len(dm.triggered.queues[1].fired), "5 Fired triggers")
	assert.Equal(10, dm.triggered.queues[1].buffer_count, "Buffers are all in queue")
	assert.Equal(10, dm.triggered.buffer_count, "Buffers are all triggered")
	assert.Equal(3, dm.triggered.queues[1].idle.Len(), "Reported trigger is no longer idle")

	bufs = dm.triggered.queues[1].ReportNext()
	assert.Equal([]int{29, 30}, bufs, "Expect trace ID 75 to be reported next")
	assert.Equal(6, dm.triggered.trace_count, "Traces are all triggered")
	assert.Equal(6, len(dm.triggered.queues[1].fired), "5 Fired triggers")
	assert.Equal(8, dm.triggered.queues[1].buffer_count, "Buffers are all in queue")
	assert.Equal(8, dm.triggered.buffer_count, "Buffers are all triggered")
	assert.Equal(4, dm.triggered.queues[1].idle.Len(), "Reported trigger is now idle")

	dm.AddBuffers(70, []int{31, 32, 33})
	dm.Trigger(1, uint64(70), []uint64{uint64(70)})
	bufs = dm.triggered.queues[1].ReportNext()
	assert.Equal([]int{31, 32, 33}, bufs, "Expect trace ID 70 to be reported next")
	assert.Equal(7, dm.triggered.trace_count, "Traces are all triggered")
	assert.Equal(7, len(dm.triggered.queues[1].fired), "5 Fired triggers")
	assert.Equal(5, dm.triggered.queues[1].idle.Len(), "Reported trigger is now idle")
}

func TestDataManagerEvictIdleTriggers(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(1, 1, "hello world")

	dm := InitDataManager()

	for i := 0; i < 10; i++ {
		dm.AddBuffers(uint64(i), []int{i})
		dm.Trigger(1, uint64(i), []uint64{uint64(i)})
	}

	assert.Equal(10, dm.triggered.trace_count, "Traces are all triggered")
	assert.Equal(10, dm.triggered.buffer_count, "Buffers are all triggered")
	assert.Equal(10, len(dm.triggered.queues[1].fired), "10 triggers")

	for i := 0; i < 5; i++ {
		bufs := dm.triggered.queues[1].ReportNext()
		assert.Equal([]int{i}, bufs, fmt.Sprintf("Trace %d was reported", i))
	}

	assert.Equal(5, dm.triggered.queues[1].idle.Len(), "5 idle triggers")
	assert.Equal(5, dm.triggered.buffer_count, "5 Buffers are triggered")
	assert.Equal(10, len(dm.triggered.queues[1].fired), "10 triggers")

	dm.triggered.queues[1].CheckIdleTriggers(time.Now().Add(time.Duration(-1) * time.Hour))
	assert.Equal(5, dm.triggered.queues[1].idle.Len(), "5 idle triggers still, eviction time hasn't been reached")
	assert.Equal(5, dm.triggered.buffer_count, "5 Buffers are triggered")
	assert.Equal(10, len(dm.triggered.queues[1].fired), "10 triggers")

	dm.triggered.queues[1].CheckIdleTriggers(time.Now().Add(time.Duration(1) * time.Hour))
	assert.Equal(0, dm.triggered.queues[1].idle.Len(), "No idle triggers remain")
	assert.Equal(5, dm.triggered.buffer_count, "5 Buffers are triggered")
	assert.Equal(5, len(dm.triggered.queues[1].fired), "5 triggers")

	for i := 5; i < 10; i++ {
		bufs := dm.triggered.queues[1].ReportNext()
		assert.Equal([]int{i}, bufs, fmt.Sprintf("Trace %d was reported", i))
	}

	assert.Equal(5, dm.triggered.queues[1].idle.Len(), "5 idle triggers")
	assert.Equal(0, dm.triggered.buffer_count, "0 Buffers are triggered")
	assert.Equal(5, len(dm.triggered.queues[1].fired), "5 triggers")

	dm.triggered.queues[1].CheckIdleTriggers(time.Now().Add(time.Duration(-1) * time.Hour))
	assert.Equal(5, dm.triggered.queues[1].idle.Len(), "5 idle triggers still, eviction time hasn't been reached")
	assert.Equal(0, dm.triggered.buffer_count, "0 Buffers are triggered")
	assert.Equal(5, len(dm.triggered.queues[1].fired), "5 triggers")

	dm.triggered.queues[1].CheckIdleTriggers(time.Now().Add(time.Duration(1) * time.Hour))
	assert.Equal(0, dm.triggered.queues[1].idle.Len(), "No idle triggers remain")
	assert.Equal(0, dm.triggered.buffer_count, "0 Buffers are triggered")
	assert.Equal(0, len(dm.triggered.queues[1].fired), "0 triggers")

}

func TestDataManagerEvictUntriggeredLRU(t *testing.T) {
	assert := assert.New(t)

	dm := InitDataManager()

	/*
		First, add a trace, and check it gets inserted correctly
		with the correct buffers and correct counts
	*/
	dm.AddBuffers(75, []int{1, 2, 3, 4, 5})
	dm.AddBuffers(76, []int{6})
	dm.AddBuffers(77, []int{7, 8, 9})
	dm.AddBuffers(78, []int{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20})
	dm.AddBuffers(79, []int{21, 22, 23})

	assert.Equal(dm.trace_count, 5, "Trace count")
	assert.Equal(dm.buffer_count, 23, "Buffer count")
	assert.Equal(dm.untriggered.trace_count, 5, "Untriggered trace count")
	assert.Equal(dm.untriggered.buffer_count, 23, "Untriggered buffer count")
	assert.Equal(dm.triggered.trace_count, 0, "Triggered trace count")
	assert.Equal(dm.triggered.buffer_count, 0, "Triggered buffer count")

	dm.EvictToCapacity(100)
	assert.Equal(dm.trace_count, 5, "Trace count")
	assert.Equal(dm.buffer_count, 23, "Buffer count")

	dm.EvictToCapacity(20)
	assert.Equal(dm.trace_count, 4, "Trace count")
	assert.Equal(dm.buffer_count, 18, "Buffer count")

	dm.EvictToCapacity(18)
	assert.Equal(dm.trace_count, 4, "Trace count")
	assert.Equal(dm.buffer_count, 18, "Buffer count")

	dm.EvictToCapacity(15)
	assert.Equal(dm.trace_count, 2, "Trace count")
	assert.Equal(dm.buffer_count, 14, "Buffer count")

	dm.EvictToCapacity(2)
	assert.Equal(dm.trace_count, 0, "Trace count")
	assert.Equal(dm.buffer_count, 0, "Buffer count")

	// Populate dm with 50005 buffers
	for i := 0; i <= 10000; i++ {
		dm.AddBuffers(uint64(i), []int{i, i + 1, i + 2, i + 3, i + 4})
	}
	assert.Equal(10001, dm.trace_count, "Trace count")
	assert.Equal(50005, dm.buffer_count, "Buffer count")

	/* Test eviction occurs in batches of size 0.01 * capacity.
	This should evict 500 buffers / 100 traces */
	evicted := dm.EvictToCapacity(50000)
	assert.Equal(500, len(evicted), "Evicting in batches")
	assert.Equal(9901, dm.trace_count, "Trace count")
	assert.Equal(49505, dm.buffer_count, "Buffer count")

	evicted = dm.EvictToCapacity(1000000)
	assert.Equal(0, len(evicted), "Evicted nothing")
	assert.Equal(9901, dm.trace_count, "Trace count")
	assert.Equal(49505, dm.buffer_count, "Buffer count")

	evicted = dm.EvictToCapacity(0)
	assert.Equal(49505, len(evicted), "Evicted everything")
	assert.Equal(0, dm.trace_count, "Trace count")
	assert.Equal(0, dm.buffer_count, "Buffer count")

}

func TestDataManagerEvictFromLargestQueue(t *testing.T) {
	assert := assert.New(t)

	dm := InitDataManager()

	for j := 0; j < 2; j++ {
		for i := 0; i < 1000; i++ {
			dm.AddBuffers(uint64(i), []int{2 * i, 2*i + 1})
			dm.Trigger(1, uint64(i), []uint64{uint64(i)})
		}

		for i := 1000; i < 2000; i++ {
			dm.AddBuffers(uint64(i), []int{2 * i})
			dm.Trigger(2, uint64(i), []uint64{uint64(i)})
		}

		assert.Equal(dm.trace_count, 2000, "2000 traces")
		assert.Equal(dm.buffer_count, 3000, "3000 buffers")
		assert.Equal(dm.triggered.queues[1].trace_count, 1000, "1000 traces in queue 1")
		assert.Equal(dm.triggered.queues[2].trace_count, 1000, "1000 traces in queue 2")
		assert.Equal(dm.triggered.queues[1].buffer_count, 2000, "2000 buffers in queue 1")
		assert.Equal(dm.triggered.queues[2].buffer_count, 1000, "1000 buffers in queue 2")

		evicted := dm.EvictedTriggeredToCapacity(3000)
		assert.Equal(0, len(evicted), "Nothing evicted yet")

		evicted = dm.EvictedTriggeredToCapacity(2500)
		assert.Equal(500, len(evicted), "500 buffers / 250 traces evicted, all from queue 1")
		assert.Equal(1750, dm.trace_count, "1750 traces")
		assert.Equal(2500, dm.buffer_count, "2500 buffers")
		assert.Equal(750, dm.triggered.queues[1].trace_count, "750 traces in queue 1")
		assert.Equal(1000, dm.triggered.queues[2].trace_count, "1000 traces in queue 2")
		assert.Equal(1500, dm.triggered.queues[1].buffer_count, "1500 buffers in queue 1")
		assert.Equal(1000, dm.triggered.queues[2].buffer_count, "1000 buffers in queue 2")

		evicted = dm.EvictedTriggeredToCapacity(2100)
		assert.Equal(400, len(evicted), "400 buffers / 200 traces evicted, all from queue 1")
		assert.Equal(1550, dm.trace_count, "1550 traces")
		assert.Equal(2100, dm.buffer_count, "2100 buffers")
		assert.Equal(550, dm.triggered.queues[1].trace_count, "550 traces in queue 1")
		assert.Equal(1000, dm.triggered.queues[2].trace_count, "1000 traces in queue 2")
		assert.Equal(1100, dm.triggered.queues[1].buffer_count, "1100 buffers in queue 1")
		assert.Equal(1000, dm.triggered.queues[2].buffer_count, "1000 buffers in queue 2")

		evicted = dm.EvictedTriggeredToCapacity(1900)
		assert.Equal(200, len(evicted), "200 buffers / 100 traces evicted, all from queue 1")
		assert.Equal(1450, dm.trace_count, "1450 traces")
		assert.Equal(1900, dm.buffer_count, "1900 buffers")
		assert.Equal(450, dm.triggered.queues[1].trace_count, "450 traces in queue 1")
		assert.Equal(1000, dm.triggered.queues[2].trace_count, "1000 traces in queue 2")
		assert.Equal(900, dm.triggered.queues[1].buffer_count, "900 buffers in queue 1")
		assert.Equal(1000, dm.triggered.queues[2].buffer_count, "1000 buffers in queue 2")

		evicted = dm.EvictedTriggeredToCapacity(1700)
		assert.Equal(200, len(evicted), "200 buffers / 200 traces evicted, all from queue 2")
		assert.Equal(1250, dm.trace_count, "1250 traces")
		assert.Equal(1700, dm.buffer_count, "1700 buffers")
		assert.Equal(450, dm.triggered.queues[1].trace_count, "450 traces in queue 1")
		assert.Equal(800, dm.triggered.queues[2].trace_count, "800 traces in queue 2")
		assert.Equal(900, dm.triggered.queues[1].buffer_count, "900 buffers in queue 1")
		assert.Equal(800, dm.triggered.queues[2].buffer_count, "800 buffers in queue 2")

		evicted = dm.EvictedTriggeredToCapacity(0)
		assert.Equal(900, len(evicted), "900 buffers / 450 traces evicted, all from queue 1")
		assert.Equal(800, dm.trace_count, "800 traces")
		assert.Equal(800, dm.buffer_count, "800 buffers")
		assert.Equal(0, dm.triggered.queues[1].trace_count, "0 traces in queue 1")
		assert.Equal(800, dm.triggered.queues[2].trace_count, "800 traces in queue 2")
		assert.Equal(0, dm.triggered.queues[1].buffer_count, "0 buffers in queue 1")
		assert.Equal(800, dm.triggered.queues[2].buffer_count, "800 buffers in queue 2")

		evicted = dm.EvictedTriggeredToCapacity(0)
		assert.Equal(800, len(evicted), "800 buffers / 800 traces evicted, all from queue 2")
		assert.Equal(0, dm.trace_count, "0 traces")
		assert.Equal(0, dm.buffer_count, "0 buffers")
		assert.Equal(0, dm.triggered.queues[1].trace_count, "0 traces in queue 1")
		assert.Equal(0, dm.triggered.queues[2].trace_count, "0 traces in queue 2")
		assert.Equal(0, dm.triggered.queues[1].buffer_count, "0 buffers in queue 1")
		assert.Equal(0, dm.triggered.queues[2].buffer_count, "0 buffers in queue 2")
	}

}
