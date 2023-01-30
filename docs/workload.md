# Workload Generator

Hindsight comes with a simple workload generator that can be used to generate trace data.  This is useful during development.

After building Hindsight, the workload generator will be found at [bin/trigger_benchmark_test](client/test/trigger_benchmark.c).

## Basic Usage

To run the workload generator:

```
bin/trigger_benchmark_test example
```

In the above, `example` is a required `process_name`.  Note that the workload generator ***does not use config files*** -- all configuration comes from command line parameters.

This will start printing output.  By default this will run a loop generating trace data as fast as possible.

# Fine-Tuning

### Configuring data

* By default the workload generates trace data as fast a possible, which is a lot.  You can slow things down by adding a value for sleep `-S` or `--sleep`.  e.g.  `-S 100000` generates about 20 traces/second compared to about 1 million/sec with no sleep
* The buffer size and pool size can be configured with `--buffer_count` and `--buffer_size` however the defaults are sensible.
* The volume of data in each trace can be configured with `--tracepoints` and `--payload_size` which configure the number of trace points per trace, and the amount of data per tracepoint, respectively.  However, the defaults are sensible (100 tracepoints, 400 bytes per tracepoint = 40kB per trace).
* By default the benchmark runs one thread; this can be increased with `-t` or `--threads`.  One thread is sufficient to generate a lot of data.

### Configuring triggers

* By default there are no triggers.  You can add a trigger with `-p` or `--trigger` and specifying the trigger probability.  For example giving `-p 0.1` adds a trigger with probability 0.1 of firing; it is automatically given a queue_id of 10.  Multiple triggers can be specified; they will be assigned queue ids incrementally starting from `10`.
* By default head-based sampling is disabled, which is sensible.  If you want to also enable headbased sampling it can be done with `--headsampling` or `-H`.  This is optional and disabled by default.  On or off is sensible.  If on, the value should be pretty low, e.g. 0.01 is a sensible head sampling probability.
* By default, every request is traced, which is sensible for retroactive tracing.  We support the performance-conscious setting by enabling retroactive tracing to only apply to a subset of requests, which can be done by setting `--retroactive` or `-R`.  For example a value of 0.5 will enable retroactive tracing only for half of all requests.

### Running the other components.

To run a 'full' setup, you should also run an agent, collector, and coordinator:

```
go run cmd/agent2/main.go --serv example
go run cmd/coordinator/main.go
go run cmd/collector/main.go
```

All of the configuration options for [agent](agent.md), [coordinator](coordinator.go) and [collector](collector.go) are valid.  In particular, it can be useful and interesting to configure rate-limits in the agent for any triggers configured with `-p` or `-trigger`.

### Configuring a multi-benchmark setting

You can run more than one benchmark process and establish breadcrumbs between them.   You can run the benchmarks on the same machine, but it will be important to ensure they use different service names and different addresses.

For example, we could run two different benchmarks:

```
bin/trigger_benchmark_test service_a
bin/trigger_benchmark_test service_b
```

And two corresponding agents, with different ports

```
go run cmd/agent2/main.go --serv service_a -port 5053
go run cmd/agent2/main.go --serv service_b -port 5054
```

We can configure the benchmark to establish breadcrumbs artificially.  For example, suppose we want to establish a breadcrumb from b -> a.  We can do so as follows, by using the address of service_a

```
bin/trigger_benchmark_test service_b --breadcrumb=127.0.0.1:5053
```

We can now combine this with a trigger on service_b, which will exercise Hindsight's distributed retrieval from both service_a and service_b:

```
bin/trigger_benchmark_test service_b --breadcrumb=127.0.0.1:5053 --trigger=0.1
```

**Side note: the benchmarks don't actually make RPCs to each other, but they have sequential trace IDs starting from 0, so for a given trace ID, data will exist on all benchmarks.  This benchmark test is not a substitute for proper integration tests but is useful for testing the distributed components of Hindsight.**


### Detailed Usage

```
$ bin/trigger_benchmark_test --usage
Usage: trigger_benchmark_test [-?V] [-a HOST:PORT] [-b HOST:PORT] [-c NUM]
            [-d NUM] [-H NUM] [-n NUM] [-o FILE] [-p NUM] [-R NUM] [-s NUM]
            [-S NUM] [-t NUM] [-w NUM] [--addr=HOST:PORT]
            [--breadcrumb=HOST:PORT] [--buffer_count=NUM] [--duration=NUM]
            [--headsampling=NUM] [--tracepoints=NUM] [--output=FILE]
            [--trigger=NUM] [--retroactive=NUM] [--buffer_size=NUM]
            [--sleep=NUM] [--threads=NUM] [--payload_size=NUM] [--help]
            [--usage] [--version] PROCNAME
```

```
$ bin/trigger_benchmark_test --help
Usage: trigger_benchmark_test [OPTION...] PROCNAME
Simple hindsight benchmarking program.  PROCNAME is required for
hindsight_init

  -a, --addr=HOST:PORT       The address of the agent to use as local
                             breadcrumb.
  -b, --breadcrumb=HOST:PORT The address of other agents to add as breadcrumbs.

  -c, --buffer_count=NUM     Number of buffers in pool, int, default 10000
                             (320MB pool)
  -d, --duration=NUM         Duration in seconds before exiting, default 0. A
                             value of 0 runs forever
  -H, --headsampling=NUM     The default head-based sampling probability
                             between 0 and 1, default 0.0 (disabled), float
  -n, --tracepoints=NUM      Number of tracepoints per trace, int, default 100
                             tracepoints (40kB per trace)
  -o, --output=FILE          Output stats to FILE.  Not currently implemented.
  -p, --trigger=NUM          Adds a trigger with probability p, float.  This
                             can be provided multiple times.  For example '-p
                             0.5 -p 0.1' will create two triggers -- trigger 10
                             that fires 50%% of the time and trigger 11 that
                             fires 10%% of the time..
  -R, --retroactive=NUM      Set the percentage of requests that will generate
                             data, by default this is 1.0 (all requests
                             generate data)
  -s, --buffer_size=NUM      Buffer size in bytes, int, default 32kB
  -S, --sleep=NUM            Sleep time to add, in microseconds, after each
                             trace.  Default 0.
  -t, --threads=NUM          Number of benchmark threads to run, int, default 1

  -w, --payload_size=NUM     Payload size written by each tracepoint, int,
                             default 400 bytes per tracepoint
  -?, --help                 Give this help list
      --usage                Give a short usage message
  -V, --version              Print program version

Mandatory or optional arguments to long options are also mandatory or optional
for any corresponding short options.

Report bugs to <bug-gnu-utils@gnu.org>.
```

### Additional comments

* `-a` `--addr` has no significance currently.  It is used to correctly set the breadcrumb of the local agent, but the benchmark doesn't use this feature of Hindsight at the moment.



