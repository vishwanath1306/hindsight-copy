package agent

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"
)

type TriggerMetrics struct {
	count            int // Number of triggers
	local            int // Number of local triggers
	remote           int // Number of remote triggers
	dropped          int // Number of triggers dropped by rate limiting in ManagedQueue
	evicted          int // Number of triggers evicted by cache pressure
	buffers          int // Total number of triggered buffers
	reported_buffers int // Reporting throughput of this trigger
	evicted_buffers  int // Buffers that should have been reported but were evicted
}

type AgentMetrics struct {
	complete_batches    int
	complete_buffers    int
	event_horizon       time.Duration
	dropped_triggers    int
	dropped_breadcrumbs int
}

type Stats struct {
	complete_batches     int
	complete_buffers     int
	buffer_throughput    float64
	buffer_throughput_mb float64
	mean_batchsize       float64
	event_horizon        time.Duration
	dropped_triggers     int
	dropped_breadcrumbs  int

	queue_totals QueueStats
	queue_ids    []int
	queues       []QueueStats

	diagnostics *Diagnostics
}

type QueueStats struct {
	trigger_count                 int
	eviction_count                int
	local_trigger_count           int
	remote_trigger_count          int
	dropped_count                 int
	buffers                       int
	reported_buffers              int
	evicted_buffers               int
	queue_throughput              float64
	reported_buffer_throughput    float64
	reported_buffer_throughput_mb float64
	evicted_buffer_throughput     float64
	evicted_buffer_throughput_mb  float64
	eviction_percent              float64

	diagnostics *QueueDiagnostics
}

func (stats *QueueStats) add(other *QueueStats) {
	stats.trigger_count += other.trigger_count
	stats.eviction_count += other.eviction_count
	stats.local_trigger_count += other.local_trigger_count
	stats.remote_trigger_count += other.remote_trigger_count
	stats.dropped_count += other.dropped_count
	stats.buffers += other.buffers
	stats.reported_buffers += other.reported_buffers
	stats.queue_throughput += other.queue_throughput
	stats.reported_buffer_throughput += other.reported_buffer_throughput
	stats.reported_buffer_throughput_mb += other.reported_buffer_throughput_mb
	stats.evicted_buffer_throughput += other.evicted_buffer_throughput
	stats.evicted_buffer_throughput_mb += other.evicted_buffer_throughput_mb

	if stats.reported_buffer_throughput+stats.evicted_buffer_throughput > 0 {
		stats.eviction_percent = 100 * stats.evicted_buffer_throughput / (stats.reported_buffer_throughput + stats.evicted_buffer_throughput)
	} else {
		stats.eviction_percent = 0
	}

	if stats.diagnostics != nil {
		stats.diagnostics.add(other.diagnostics)
	}
}

func (s *QueueStats) Str() string {
	var b strings.Builder
	fmt.Fprintf(&b, " %.1f trigs/s ", s.queue_throughput)
	fmt.Fprintf(&b, " %.1f MB/s ", s.reported_buffer_throughput_mb)
	fmt.Fprintf(&b, "(%.0f bufs/s, %d total) ", s.reported_buffer_throughput, s.reported_buffers)
	fmt.Fprintf(&b, "%.0f%% loss (%.1f MB/s)", s.eviction_percent, s.evicted_buffer_throughput_mb)
	if s.diagnostics != nil {
		fmt.Fprintf(&b, "  ||   %v", s.diagnostics.Str())
	}
	return b.String()
}

func (s *Stats) Str() string {
	var b strings.Builder
	fmt.Fprintf(&b, "EH: %d ms ", s.event_horizon/time.Millisecond)
	fmt.Fprintf(&b, "%.3f MB/s ", s.buffer_throughput_mb)
	fmt.Fprintf(&b, "(%.0f bufs/s, %d bufs total), ", s.buffer_throughput, s.complete_buffers)
	fmt.Fprintf(&b, "Avg batch %.1f, ", s.mean_batchsize)
	fmt.Fprintf(&b, "Drops %d,%d ", s.dropped_triggers, s.dropped_breadcrumbs)
	if s.diagnostics != nil {
		fmt.Fprintf(&b, "  ||  %v", s.diagnostics.Str())
	}
	fmt.Fprintf(&b, "\n")
	fmt.Fprintf(&b, "  -- Triggers %v\n", s.queue_totals.Str())

	for i, trigger_id := range s.queue_ids {
		ts := s.queues[i]
		fmt.Fprintf(&b, "            %d - ", trigger_id)
		fmt.Fprintf(&b, "%v\n", ts.Str())
	}

	return b.String()
}

/* Calculates agent stats and resets for next iteration */
func (agent *Agent) calculateAgentStats(duration_nanos float64, debug bool) Stats {
	/* Get and reset the agent's metrics */
	metrics := agent.metrics
	agent.metrics = AgentMetrics{}

	/* Calculate stats */
	var stats Stats
	stats.complete_batches = metrics.complete_batches
	stats.complete_buffers = metrics.complete_buffers
	stats.buffer_throughput = float64(uint64(metrics.complete_buffers)*1000000000) / duration_nanos
	stats.buffer_throughput_mb = (stats.buffer_throughput * float64(agent.tm.buffer_size)) / (1024 * 1024)
	if metrics.complete_batches > 0 {
		stats.mean_batchsize = float64(metrics.complete_buffers) / float64(metrics.complete_batches)
	}
	stats.event_horizon = metrics.event_horizon
	stats.dropped_triggers = metrics.dropped_triggers
	stats.dropped_breadcrumbs = metrics.dropped_breadcrumbs

	if debug {
		diagnostics := agent.calculateDiagnostics()
		stats.diagnostics = &diagnostics
		stats.queue_totals.diagnostics = &QueueDiagnostics{}
	}

	/* Get and sort queue ids */
	for queue_id := range agent.tm.queues {
		stats.queue_ids = append(stats.queue_ids, queue_id)
	}
	sort.Ints(stats.queue_ids)

	/* Calculate stats for each trigger, plus totals */
	for _, queue_id := range stats.queue_ids {
		queue := agent.tm.queues[queue_id]
		queue_stats := agent.calculateQueueStats(duration_nanos, queue)

		if debug {
			queue_diagnostics := agent.calculateQueueDiagnostics(queue)
			queue_stats.diagnostics = &queue_diagnostics
		}

		stats.queues = append(stats.queues, queue_stats)
		stats.queue_totals.add(&queue_stats)
	}

	return stats
}

func (agent *Agent) calculateQueueStats(duration_nanos float64, queue *ManagedQueue) QueueStats {
	/* Get and reset the trigger's metrics */
	metrics := queue.queue.metrics
	queue.queue.metrics = TriggerMetrics{}

	/* Calculate stats */
	var stats QueueStats
	stats.trigger_count = metrics.count
	stats.eviction_count = metrics.evicted
	stats.local_trigger_count = metrics.local
	stats.remote_trigger_count = metrics.remote
	stats.dropped_count = metrics.dropped
	stats.buffers = metrics.buffers
	stats.reported_buffers = metrics.reported_buffers
	stats.queue_throughput = float64(metrics.count*1000000000) / duration_nanos
	stats.reported_buffer_throughput = float64(metrics.reported_buffers*1000000000) / duration_nanos
	stats.reported_buffer_throughput_mb = (stats.reported_buffer_throughput * float64(agent.tm.buffer_size)) / (1024 * 1024)
	stats.evicted_buffer_throughput = float64(metrics.evicted_buffers*1000000000) / duration_nanos
	stats.evicted_buffer_throughput_mb = (stats.evicted_buffer_throughput * float64(agent.tm.buffer_size)) / (1024 * 1024)

	if stats.reported_buffer_throughput+stats.evicted_buffer_throughput > 0 {
		stats.eviction_percent = 100 * stats.evicted_buffer_throughput / (stats.reported_buffer_throughput + stats.evicted_buffer_throughput)
	} else {
		stats.eviction_percent = 0
	}

	return stats
}

type Diagnostics struct {
	cache_size        int
	cache_percent     float64
	triggered_size    int
	triggered_percent float64
	lru_size          int
}

type QueueDiagnostics struct {
	buffers         int
	buffers_percent float64
	pending_reports int
	lru_size        int
	lru_percent     float64
}

func (d *QueueDiagnostics) Str() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%d buffers (%.0f%%) ", d.buffers, d.buffers_percent)
	fmt.Fprintf(&b, "%d pending, ", d.pending_reports)
	fmt.Fprintf(&b, "lru %.0f%% (%d)", d.lru_percent, d.lru_size)
	return b.String()
}

func (d *Diagnostics) Str() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Cache %.0f%% (%d) ", d.cache_percent, d.cache_size)
	fmt.Fprintf(&b, "Triggered %.0f%% (%d) ", d.triggered_percent, d.triggered_size)
	fmt.Fprintf(&b, "LRU %d ", d.lru_size)
	return b.String()
}

func (d *QueueDiagnostics) add(other *QueueDiagnostics) {
	d.buffers += other.buffers
	d.buffers_percent += other.buffers_percent
	d.pending_reports += other.pending_reports
	d.lru_size += other.lru_size
	d.lru_percent += other.lru_percent
}

func (agent *Agent) calculateQueueDiagnostics(queue *ManagedQueue) QueueDiagnostics {
	var d QueueDiagnostics
	d.buffers = queue.queue.buffer_count
	d.buffers_percent = 100 * float64(d.buffers) / float64(agent.triggered_capacity)
	d.pending_reports = queue.queue.reporting.Size()
	d.lru_size = queue.queue.idle.Len()
	d.lru_percent = 0.001 // Deprecated
	return d
}

func (agent *Agent) calculateDiagnostics() Diagnostics {
	var d Diagnostics
	d.cache_size = agent.dm.buffer_count
	d.cache_percent = 100 * float64(d.cache_size) / float64(agent.cache_capacity)
	d.triggered_size = agent.dm.triggered.buffer_count
	d.triggered_percent = 100 * float64(d.triggered_size) / float64(agent.triggered_capacity)
	d.lru_size = agent.dm.untriggered.lru.Len()
	return d
}

type AgentTelemetryGenerator struct {
	agent         *Agent
	debug         bool
	print_summary bool
}

func (g *AgentTelemetryGenerator) Init(agent *Agent, debug bool) {
	g.agent = agent
	g.debug = debug
	g.print_summary = true
}

/* TelemetryGenerator interface */
func (g *AgentTelemetryGenerator) Headers() []string {
	return []string{
		// Preamble
		"t",
		"interval_ms",
		"queue_id",

		// Totals
		"data_mb",          // Total data in MB generated by the local application
		"reported_mb",      // Total data in MB successfully reported to trace collection backends
		"evicted_mb",       // Total data in MB that should have been reported but was dropped due to cache pressure
		"triggers",         // local_triggers + remote_triggers + dropped_triggers
		"local_triggers",   // Triggers fired locally that weren't dropped
		"remote_triggers",  // Triggers received from the coordinator
		"dropped_triggers", // Local triggers dropped by rate-limiting
		"evicted_triggers", // Triggers that are evicted by cache pressure

		// Throughputs of the above totals
		"tput_data_mb",
		"tput_reported_mb",
		"tput_evicted_mb",
		"tput_triggers",
		"tput_local_triggers",
		"tput_remote_triggers",
		"tput_dropped_triggers",
		"tput_evicted_triggers",

		// Other percentages and instantaneous measurements
		"cache_occupancy",     // How much of the cache capacity is used by data from this queue
		"eviction_percent",    // How much triggered trace data is being lost
		"internal_bottleneck", // For diagnostics - should be 0 most of the time - measures data dropped internally due to bottlenecks
		"event_horizon_ms",    // For untriggered data, the time in cache before being evicted
		"report_horizon_ms",   // For triggered traces, mean time until data is reported
	}
}

func generateDataRow(agent *Agent, now time.Time, interval time.Duration, queue *QueueStats, queueid string) map[string]string {
	row := make(map[string]string)

	data_mb := float64(queue.buffers*agent.tm.buffer_size) / float64(1024*1024)
	reported_mb := float64(queue.reported_buffers*agent.tm.buffer_size) / float64(1024*1024)
	evicted_mb := float64(queue.evicted_buffers*agent.tm.buffer_size) / float64(1024*1024)

	// Preamble
	row["t"] = strconv.FormatInt(now.UTC().UnixNano(), 10)
	row["interval_ms"] = strconv.FormatInt(interval.Milliseconds(), 10)
	row["queue_id"] = queueid

	// Totals
	row["data_mb"] = strconv.FormatFloat(data_mb, 'f', 2, 64)
	row["reported_mb"] = strconv.FormatFloat(reported_mb, 'f', 2, 64)
	row["evicted_mb"] = strconv.FormatFloat(evicted_mb, 'f', 2, 64)
	row["triggers"] = strconv.Itoa(queue.trigger_count)
	row["local_triggers"] = strconv.Itoa(queue.local_trigger_count)
	row["remote_triggers"] = strconv.Itoa(queue.remote_trigger_count)
	row["dropped_triggers"] = strconv.Itoa(queue.dropped_count)
	row["evicted_triggers"] = strconv.Itoa(queue.eviction_count)

	// Throughputs
	interval_s := float64(interval) / float64(time.Second)
	row["tput_data_mb"] = strconv.FormatFloat(data_mb/interval_s, 'f', 2, 64)
	row["tput_reported_mb"] = strconv.FormatFloat(reported_mb/interval_s, 'f', 2, 64)
	row["tput_evicted_mb"] = strconv.FormatFloat(evicted_mb/interval_s, 'f', 2, 64)
	row["tput_triggers"] = strconv.FormatFloat(float64(queue.trigger_count)/interval_s, 'f', 0, 64)
	row["tput_local_triggers"] = strconv.FormatFloat(float64(queue.local_trigger_count)/interval_s, 'f', 0, 64)
	row["tput_remote_triggers"] = strconv.FormatFloat(float64(queue.remote_trigger_count)/interval_s, 'f', 0, 64)
	row["tput_dropped_triggers"] = strconv.FormatFloat(float64(queue.dropped_count)/interval_s, 'f', 0, 64)
	row["tput_evicted_triggers"] = strconv.FormatFloat(float64(queue.eviction_count)/interval_s, 'f', 0, 64)

	// Other percentages and instantaneous measurements
	if queue.diagnostics != nil {
		row["cache_occupancy"] = strconv.FormatFloat(queue.diagnostics.buffers_percent, 'f', 1, 64)
	}
	row["eviction_percent"] = strconv.FormatFloat(queue.eviction_percent, 'f', 1, 64)

	return row
}

func generateTotalsRow(agent *Agent, now time.Time, interval time.Duration, stats *Stats) map[string]string {
	// Most totals data is generated in the same way as per-queue data
	row := generateDataRow(agent, now, interval, &stats.queue_totals, "total")

	// Totals use different calculation for data_mb
	data_mb := float64(stats.complete_buffers*agent.tm.buffer_size) / float64(1024*1024)
	interval_s := float64(interval) / float64(time.Second)
	row["data_mb"] = strconv.FormatFloat(data_mb, 'f', 2, 64)
	row["tput_data_mb"] = strconv.FormatFloat(data_mb/interval_s, 'f', 2, 64)

	// internal_bottleneck and event_horizon are only reported for the totals, not per-queue
	internal_bottleneck := 100 * float64(stats.dropped_triggers) / float64(stats.queue_totals.trigger_count)
	row["internal_bottleneck"] = strconv.FormatFloat(internal_bottleneck, 'f', 1, 64)
	row["event_horizon_ms"] = strconv.FormatFloat(float64(stats.event_horizon)/float64(time.Millisecond), 'f', 0, 64)

	return row
}

/* TelemetryGenerator interface */
func (g *AgentTelemetryGenerator) NextData(now time.Time, interval time.Duration) (rows []map[string]string) {
	stats := g.agent.calculateAgentStats(float64(interval.Nanoseconds()), g.debug)

	// Add the totals
	rows = append(rows, generateTotalsRow(g.agent, now, interval, &stats))

	// Add a row for each queue
	for i, queue_id := range stats.queue_ids {
		queue_stats := stats.queues[i]
		row := generateDataRow(g.agent, now, interval, &queue_stats, strconv.Itoa(queue_id))
		rows = append(rows, row)
	}

	if g.print_summary {
		log.Print(stats.Str())
	}

	return rows
}
