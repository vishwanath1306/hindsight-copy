package agent

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/geraldleizhang/hindsight/agent/pkg/memory"
	"github.com/geraldleizhang/hindsight/agent/pkg/telemetry"
)

type Agent struct {
	dm          DataManager       // the trace data
	api         memory.GoAgentAPI // API to the shared memory
	coordinator Coordinator       // Interface to coordinator
	reporting   Reporting         // Interface to trace data backend
	tm          TriggerManager    // Rate limits and fair shares the triggers

	// Constants for deciding when to evict
	cache_capacity     int           // Above this threshold, we should evict
	triggered_capacity int           // Above this threshold we evict from triggered
	trigger_timeout    time.Duration // How long a trigger remains idle before being deleted

	/* Wraps api.Triggers, possibly adding a delay for experiments */
	localtriggers <-chan []memory.Trigger // triggers from shm

	metrics  AgentMetrics
	reporter *telemetry.Reporter

	vc int // Virtual clock used for fair sharing
}

func InitAgent2(fname string, local_hostname string, local_port string, coordinator_addr string,
	reporting_addr string, trigger_delay uint64, reporting_rate_limit float64,
	trigger_rate_limit float64, per_trigger_rate_limits map[int]float64,
	telemetry_filename string, verbose bool) *Agent {
	fmt.Println("Init agent", fname)

	if trigger_delay > 0 {
		fmt.Printf("  Triggers are delayed by %d milliseconds before firing\n", trigger_delay)
	} else {
		fmt.Println("  Triggers fire immediately")
	}

	if reporting_rate_limit > 0 {
		fmt.Printf("  Reporting rate-limited to %.2f MB/s\n", reporting_rate_limit)
	} else {
		fmt.Println("  Unrestricted reporting bandwidth")
	}
	for trigger_id, rate := range per_trigger_rate_limits {
		fmt.Printf("    -Trigger %d rate limit %.2f MB/s\n", trigger_id, rate)
	}

	var agent Agent
	agent.dm.Init()
	agent.api.Init(fname)
	agent.reporting.Init(&agent.api, reporting_rate_limit, true, reporting_addr, local_hostname, local_port)
	agent.coordinator.Init(true, local_hostname, local_port, coordinator_addr)
	agent.tm.Init(&agent.dm, agent.api.BufferSize(), trigger_rate_limit)
	agent.tm.ConfigureRateLimits(per_trigger_rate_limits)

	agent.cache_capacity = (4 * agent.api.Capacity()) / 5 // TODO: not hardcoded
	agent.triggered_capacity = agent.cache_capacity / 2   // TODO: not hardcoded
	agent.trigger_timeout = time.Duration(-5) * time.Minute

	/* Trigger delay isn't a feature of Hindsight, but we use it for some of the
	Hindsight experiments to inject artificial delay in triggers firing. */
	if trigger_delay == 0 {
		agent.localtriggers = agent.api.Triggers
	} else {
		fmt.Printf("Delaying triggers by %d milliseconds", trigger_delay)
		delayer := delayTriggers(time.Duration(trigger_delay)*time.Millisecond, agent.api.Triggers)
		agent.localtriggers = delayer.Outgoing
	}

	fmt.Println("Go Agent cache capacity", agent.cache_capacity)

	/* Initialize the telemetry reporting */
	report_interval := time.Duration(1) * time.Second // Seems overkill to add this as an argument at the moment
	debug := true                                     // Currently, debug telemetry is lightweight and pretty useful.
	agent.initTelemetry(report_interval, telemetry_filename, verbose, debug)

	return &agent
}

/* Invoked during agent initialization; just creates and links up the
the appropriate telemetry loggers according to agent init arguments */
func (agent *Agent) initTelemetry(report_interval time.Duration, telemetry_filename string, verbose bool, debug bool) error {
	/* Create the receivers */
	var receivers []telemetry.Receiver
	if telemetry_filename != "" {
		fmt.Println("Outputting telemetry to", telemetry_filename)
		r, err := telemetry.NewCsvReceiver(telemetry_filename)
		if err != nil {
			return err
		}
		receivers = append(receivers, r)
	}
	if verbose {
		fmt.Println("Outputting telemetry to stdout")
		receivers = append(receivers, telemetry.NewStdoutReceiver(" "))
	}

	if len(receivers) == 0 {
		receivers = append(receivers, &telemetry.NullReceiver{})
	}

	/* If there are 2 or more receivers, wrap them in a MultiReceiver */
	var receiver telemetry.Receiver
	if len(receivers) == 1 {
		receiver = receivers[0]
	} else {
		receiver = telemetry.NewMultiReceiver(receivers)
	}

	/* Create the AgentTelemetryGenerator defined in metrics.go */
	var generator AgentTelemetryGenerator
	generator.Init(agent, debug)

	/* Link up with the reporter */
	agent.reporter = new(telemetry.Reporter)
	agent.reporter.Init(report_interval, &generator, receiver)
	return nil
}

/*
	Checks the capacity of the DataManager and evicts buffers if necessary as follows:
	* Evict untriggered traces if above global capacity threshold
	* Evict triggers if above triggered buffer capacity threshold
	* Time out idle triggers
	Returns any evicted buffers to the available queue
*/
func (agent *Agent) maybeEvict() {
	// Skip until the cache is full
	if agent.dm.buffer_count < agent.cache_capacity {
		return
	}

	// Clean up timed-out triggers
	agent.dm.CheckIdleTriggers(agent.dm.now.Add(agent.trigger_timeout))

	// Evict spammy triggers
	evicted := agent.dm.EvictedTriggeredToCapacity(agent.triggered_capacity)
	if len(evicted) > 0 {
		agent.api.Available <- evicted
	}

	// Evict untriggered trace data
	evicted = agent.dm.EvictToCapacity(agent.cache_capacity)
	if len(evicted) > 0 {
		agent.api.Available <- evicted
	}
	agent.metrics.event_horizon = agent.dm.now.Sub(agent.dm.untriggered.event_horizion)
}

/*
  Process a batch of buffers retrieved from the complete queue.
	This mainly just sends the buffers to the datamanager
*/
func (agent *Agent) processCompletedBuffers(batch memory.CompleteBatch) {
	var freed_buffers []int
	for trace_id, buffers := range batch {
		/* Update agent metrics */
		agent.metrics.complete_buffers += len(buffers)

		/* Ignore trace ID 0 */
		if trace_id == 0 {
			freed_buffers = append(freed_buffers, buffers...)
			continue
		}

		/* Add to the DataManager */
		agent.dm.AddBuffers(trace_id, buffers)
	}

	/* Trigger eviction if necessary */
	agent.maybeEvict()

	/* Send freed buffers */
	if len(freed_buffers) > 0 {
		agent.api.Available <- freed_buffers
	}

	agent.metrics.complete_batches++
}

/*
  Process a batch of breadcrumbs retrieved from the shm bc queue.
	This mainly just sends the breadcrumbs to the datamanager
*/
func (agent *Agent) processBreadcrumbs(batch memory.BreadcrumbBatch) {
	to_report := make(map[uint64][]string)
	num_to_report := 0
	for trace_id, breadcrumbs := range batch {
		/* Ignore trace ID 0 */
		if trace_id == 0 {
			continue
		}

		/* Add to the DataManager */
		breadcrumbs := agent.dm.AddBreadcrumbs(trace_id, breadcrumbs)
		if len(breadcrumbs) > 0 {
			to_report[trace_id] = breadcrumbs
			num_to_report += len(breadcrumbs)
		}
	}

	/* Forward breadcrumbs as needed */
	if len(to_report) > 0 {
		select {
		case agent.coordinator.breadcrumbs <- to_report:
			break
		default:
			agent.metrics.dropped_breadcrumbs += num_to_report
		}
	}
}

func (agent *Agent) processTriggers(batch []memory.Trigger) {
	triggers_to_forward := make([]memory.Trigger, 0, len(batch))
	breadcrumbs_to_forward := make(map[uint64][]string)
	num_breadcrumbs_to_forward := 0
	for _, t := range batch {
		/* Add to the DataManager */
		// TODO: update C struct to send lateral trace ids all in one or have two ids
		queue := agent.tm.getQueue(t.Queue_id)
		triggered, breadcrumbs := queue.TriggerLocal(t.Base_trace_id, []uint64{t.Trace_id})

		/* Forward trigger to coordinator */
		if triggered {
			triggers_to_forward = append(triggers_to_forward, t)
		}

		/* Accumulate breadcrumbs to forward */
		for trace_id, addrs := range breadcrumbs {
			if len(addrs) > 0 {
				breadcrumbs_to_forward[trace_id] = append(breadcrumbs_to_forward[trace_id], addrs...)
				num_breadcrumbs_to_forward += len(addrs)
			}
		}
	}

	/* Forward triggers and breadcrumbs */
	if len(triggers_to_forward) > 0 {
		select {
		case agent.coordinator.localtriggers <- triggers_to_forward:
			break
		default:
			// Connection to coordinator is bottlenecked; drop the triggers
			agent.metrics.dropped_triggers += len(triggers_to_forward)
		}
	}
	if len(breadcrumbs_to_forward) > 0 {
		select {
		case agent.coordinator.breadcrumbs <- breadcrumbs_to_forward:
			break
		default:
			// Connection to coordinator is bottlenecked; drop the breadcrumbs
			agent.metrics.dropped_breadcrumbs += num_breadcrumbs_to_forward
		}
	}
}

func (agent *Agent) processRemoteTriggers(batch []memory.Trigger) {
	breadcrumbs_to_forward := make(map[uint64][]string)
	num_breadcrumbs_to_forward := 0
	for _, t := range batch {
		queue := agent.tm.getQueue(t.Queue_id)
		// TODO: update C struct to send lateral trace ids all in one or have two ids
		breadcrumbs := queue.TriggerRemote(t.Base_trace_id, []uint64{t.Trace_id})

		/* Accumulate breadcrumbs to forward */
		for trace_id, addrs := range breadcrumbs {
			if len(addrs) > 0 {
				breadcrumbs_to_forward[trace_id] = append(breadcrumbs_to_forward[trace_id], addrs...)
				num_breadcrumbs_to_forward += len(addrs)
			}
		}
	}

	if len(breadcrumbs_to_forward) > 0 {
		select {
		case agent.coordinator.breadcrumbs <- breadcrumbs_to_forward:
			break
		default:
			// Connection to coordinator is bottlenecked; drop the breadcrumbs
			agent.metrics.dropped_breadcrumbs += num_breadcrumbs_to_forward
		}
	}
}

func (agent *Agent) RunProcessingLoop(ctx context.Context) {
	log.Println("Begun receiving trace data from application")
	var data_to_report []int
	timer := time.NewTimer(0 * time.Second)
	for {
		agent.dm.now = time.Now()

		if len(data_to_report) == 0 {
			/* We have no data to report currently, so we periodically
			check if there's anything to report */
			select {
			case <-ctx.Done():
				log.Println("Stopped receiving trace data from application")
				return
			case <-timer.C:
				data_to_report = agent.tm.GetNextBatchToReport()
				timer.Reset(100 * time.Millisecond)
			case triggers := <-agent.coordinator.remotetriggers:
				/* Received some triggers from the coordinator */
				agent.processRemoteTriggers(triggers)
			case triggers := <-agent.localtriggers:
				/* Received some triggers from the shm triggers queue */
				agent.processTriggers(triggers)
			case buffers := <-agent.api.Complete:
				/* Received some buffers from the shm complete queue */
				agent.processCompletedBuffers(buffers)
			case breadcrumbs := <-agent.api.Breadcrumbs:
				/* Received some breadcrumbs from the shm breadcrumbs queue */
				agent.processBreadcrumbs(breadcrumbs)
			}
		} else {
			/* We do have data to report, so we attempt to report it,
			and after doing so check if there's anything more to report */
			select {
			case <-ctx.Done():
				log.Println("Stopped receiving trace data from application")
				return
			case agent.reporting.data <- data_to_report:
				data_to_report = agent.tm.GetNextBatchToReport()
				timer.Reset(100 * time.Millisecond)
			case triggers := <-agent.coordinator.remotetriggers:
				/* Received some triggers from the coordinator */
				agent.processRemoteTriggers(triggers)
			case triggers := <-agent.localtriggers:
				/* Received some triggers from the shm triggers queue */
				agent.processTriggers(triggers)
			case buffers := <-agent.api.Complete:
				/* Received some buffers from the shm complete queue */
				agent.processCompletedBuffers(buffers)
			case breadcrumbs := <-agent.api.Breadcrumbs:
				/* Received some breadcrumbs from the shm breadcrumbs queue */
				agent.processBreadcrumbs(breadcrumbs)
			}
		}
	}
}

func (agent *Agent) Run(ctx context.Context, cancel context.CancelFunc) {
	wg := new(sync.WaitGroup)
	wg.Add(5)
	go func() {
		agent.RunProcessingLoop(ctx)
		wg.Done()
	}()
	go func() {
		agent.coordinator.Run(ctx, cancel)
		wg.Done()
	}()
	go func() {
		agent.reporting.Run(ctx)
		wg.Done()
	}()
	go func() {
		agent.api.Run(ctx)
		wg.Done()
	}()
	go func() {
		if agent.reporter != nil {
			err := agent.reporter.Run(ctx)
			if err != nil {
				fmt.Println("Error in telemetry reporter:", err)
			}
		}
		wg.Done()
	}()
	wg.Wait()
}
