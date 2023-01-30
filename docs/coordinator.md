# Running Hindsight's Coordinator

Hindsight's coordinator is responsible for disseminating breadcrumbs and triggers between agents.

## Basic coordinator

Run the coordinator:

```
go run cmd/coordinator/main.go
```

Expected output:

```
2022/04/07 20:31:10 Running coordinator
  port=5252 (command line)
2022/04/07 20:31:10 CoordinatorServer main goroutine running
2022/04/07 20:31:10 Listening for agent connections on port 5252
```

## Coordinator with logging

This logs statistics about breadcrumb traversal time to a file

Run the coordinator:

```
go run cmd/coordinator/main.go -out example.out
```

Expected output:

```
2022/04/07 20:32:20 Running coordinator
  port=5252 (command line)
2022/04/07 20:32:20 Logging breadcrumb stats to example.out
2022/04/07 20:32:20 Logger goroutine running
2022/04/07 20:32:20 CoordinatorServer main goroutine running
2022/04/07 20:32:20 Listening for agent connections on port 5252
```

# Configuring the Coordinator

By default Hindsight's coordinator will listen on port `5252`.  You can change the port of the coordinator with the `-port` flag.  

See the full coordinator options with the `--help` flag:

```Usage of /tmp/go-build1913113332/b001/exe/main:
  -out string
        Output filename for writing breadcrumb dissemination statistics.  If not specified, will not be written to file
  -port lc_port
        Coordinator port.  If not specified, uses lc_port from the legacy config lc.conf file. (default "5252")
```

# Configuring Agents to Point to the Coordinator

Hindsight agents report breadcrumbs and triggers to the coordinator, and thus they need the address of the coordinator.  If the coordinator isn't running or if it is misconfigured, then the agent will periodically retry connecting in the background and data will not be reported.  For example you will see the following output when running an agent:

```
Error in TriggersLoop: rpc error: code = Unavailable desc = connection error: desc = "transport: Error while dialing dial tcp 127.0.0.1:5252: connect: connection refused"  -- will retry every 2 seconds
Error in BreadcrumbsLoop: rpc error: code = Unavailable desc = connection error: desc = "transport: Error while dialing dial tcp 127.0.0.1:5252: connect: connection refused"  -- will retry every 2 seconds
```

This error will only show up lazily when we attempt to report the first triggers and breadcrumbs to the coordinator.

### Configuring Agents via the Command Line

The coordinator address can be given to agents with the `-lc` flag, e.g.

```
go run cmd/agent2/main.go -lc 127.0.0.1:5252
```

Alternatively, it can be configured in the conf file with the `lc_addr` and `lc_port` keys e.g.

`my_agent.conf`:
```
lc_addr 127.0.0.1
lc_port 5252
```

```
go run cmd/agent2/main.go --serv my_agent
```

# Breadcrumb traversal stats

The coordinator writes breadcrumb traversal statistics to the output file (if you specified it as a cmd line argument)

The file is a simple CSV with one row per trigger, e.g.

```
t,queue,total_agents,dissemination_time_ms
21,7,2,7
21,7,2,7
21,7,2,7
21,7,2,7
21,7,2,7
21,7,2,7
21,7,2,7
21,7,2,7
21,7,2,7
21,7,2,7
21,7,2,7
21,7,2,7
21,7,2,7
21,7,2,7
```

The columns are:

* `t` time in seconds
* `queue` the queue_id for the fired trigger
* `total_agents` the total number of agents traversed by breadcrumbs
* `dissemination_time_ms` the total time between the coordinator first learning of the trigger, and the final breadcrumb received
