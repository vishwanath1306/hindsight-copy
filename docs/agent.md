# Running Hindsight's Agent

For every client application using Hindsight, we require a corresponding Hindsight agent running on the same host.  Within the client application, Hindsight is initialized by specifying a `service_name`, e.g. `hindsight_init(service_name)`, which will load `service_name.conf`.  Correspondingly, you must run a Hindsight agent with the same `service_name` as follows:

```
go run cmd/agent2/main.go -serv service_name
```

By default this will load the same `service_name.conf` file, though there are various command line parameters available too.

Expected output: if the client application is not running yet, then you will see:

```
Running agent my_service
  hostname=127.0.0.1 (my_service.conf)
  port=5050 (my_service.conf)
  lc_addr=127.0.0.1:5252 (my_service.conf)
  r_addr=127.0.0.1:5253 (default fallback)
Init agent my_service
  Triggers fire immediately
  Unrestricted reporting bandwidth
/dev/shm/my_service__pool does not exist, waiting...
```

If the client application is already running you will see:

```
Loaded existing buffer pool, capacity=10000 buffer_size=32768 at /dev/shm/my_service__pool
Loaded existing queue capacity=10000 element_size=4 element_total_size=8 at /dev/shm/my_service__available_queue (0x7f72e8172000)
Loaded existing queue capacity=10000 element_size=16 element_total_size=20 at /dev/shm/my_service__complete_queue (0x7f72e8141000)
Loaded existing queue capacity=10000 element_size=24 element_total_size=28 at /dev/shm/my_service__triggers_queue (0x7f72e80fc000)
Loaded existing queue capacity=10000 element_size=48 element_total_size=52 at /dev/shm/my_service__breadcrumbs_queue (0x7f72e807d000)
Initialize buffers: done
Queue states:
  Available occupancy=9932 remaining=68 head=68 tail=10000
  Complete occupancy=68 remaining=9932 head=0 tail=68
Go Agent cache capacity 8000
Reporting goroutine running
DataLoop connecting to 127.0.0.1:5253
Agent goroutine running
Coordinator goroutine running - coordinator at:  127.0.0.1:5252
Unable to connect to reporting backend, retrying every 2 seconds 127.0.0.1:5253 dial tcp 127.0.0.1:5253: connect: connection refused
shm queue goroutine running
completeLoop
```

## Important edge case

If you are repeatedly restarting your clients, then it is possible that shared memory state is leftover that needs to be cleaned up.  This can be handled in two ways:

### Option 1: Always Run Client Application First

If the client application starts before the Hindsight agent, then you will not encounter any issues

### Option 2: Manually delete shm files

If you want to start the Hindsight agent before the client application, then you should delete shm files with the `service_name` prefix, e.g.:

```
rm /dev/shm/my_service_*
```

# Configuring the agent

Hindsight's agent has a number of configuration options.  The most important required flag is `--serv` for specifying the `service_name` of this agent.  To see the full list of options run:

```
go run cmd/agent2/main.go --help
```

```
Usage of /tmp/go-build2373230828/b001/exe/main:
  -delay int
        Used for experimental purposes.  If specified, this delays the reporting
        of triggers by the specified delay (in nanoseconds).  Default to 0 - no 
        delay.
  -host addr
        Hostname or IP of this agent.  If not specified, uses addr from the lega
        cy config file
  -l value
        A per-trigger reporting rate limit in the form queue_id,rate where queue
        _id is an integer and rate is a float representing a reporting limit in 
        MB/s.  This flag can be set multiple times to provide rate limits for di
        fferent triggers.
  -lc lc_addr
        Address of the log collector in form hostname:port.  If not specified, u
        ses lc_addr:`lc_port` from the legacy config file.
  -output string
        Filename for outputting agent telemetry.  If specified, will write a csv of agent telemetry data.  Disabled by default.
  -port port
        Port to run the agent on.  If not specified, uses port from the legacy c
        onfig file.
  -r r_addr
        Address of the reporting backend in form hostname:port.  If not specifie
        d, uses r_addr:`r_port` from the legacy config file.
  -rate float
        Rate limit for reporting traces in MB/s.  Set to 0 to disable.  Default 
        0.
  -serv string
        Service name
  -triggerrate float
        Rate limit for a spammy trigger in triggers/s.  Set to 0 to disable.  De
        fault 10000. (default 10000)
  -verbose
        If set to true, prints telemetry to the command line.  False by default.
```

Where noted, some port and address configurations can be specified via the `service_name.conf` file, and overridden by command line arguments.  For information about the configuration file, see [configuration.md](configuration.md)

# Defaults

By default the Hindsight agent will not apply any rate limitings or delays to reporting.  However it is sensible to configure this.

### Global reporting rate

The `-rate` flag specifies in MB/s a rate limit for reporting data to the Hindsight backend over the network.  By default there is no limit, but in practice it may be useful to impose e.g. a 10 MB/s reporting limit with `-rate 10`.

When an agent reaches this reporting limit, it will selectively drop data according to Hindsight's prioritization schemes.

### Triggering rate

In the 'golden case' triggers fire rarely and all is well, but if an application is misconfigured or there is an unanticipated edge case then the client might inadvertently fire too many triggers.  By default Hindsight imposes a limit of 10,000 triggers per second for each distinct trigger queue.  This is a very high value and it might be desirable to reduce this further.  Too many triggers imposes high network and coordination overhead.  To set e.g. 100 triggers/second you can specify `-triggerrate 100`.  When set, local triggers might be preemptively dropped if above this rate.  Triggerrate does not affect remote triggers due to Hindsight's prioritization schemes.

### Per-trigger reporting rates

Some triggers might be spammy while others might only have a few traces.  You can configure per-trigger rate limits with the `-l` flag.  For this you need to know the `queue_id` of the trigger used by the client application.  Rate limits are specified in MB/s.  For example, to rate-limit reporting from queue 1 to 5 MB/s, you can provide `-l 1,5`.    If a rate limit isn't specified for a queue then it is unlimited and will only be affected by a global reporting rate limit if specified.

# Example:

```
go run cmd/agent2/main.go -serv my_service -l 1,5 -l 2,1 -rate 10 -triggerrate 1000
```

In the above example for a service called `my_service` we have imposed a global reporting rate limit of 10 MB/s, and additionally restrict queue 1 to 5 MB/s and queue 2 to 1 MB/s.  We also prevent more than 1,000 local triggers/second.

# Telemetry

For details on the `-output` and `-verbose` flags, see [telemetry.md](telemetry.md)