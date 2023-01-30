# Configuration

For environment setup, see [environment](environment.md)

The default Hindsight configuration can be found in `conf/default.conf`:

```
cap 10000
buf_length 32768
addr 127.0.0.1
port 5050
lc_addr 127.0.0.1
lc_port 5252
r_addr 127.0.0.1
r_port 5253
retroactive_sampling_percentage 1.0
head_sampling_probability 0.0
```

There are some other configuration values that are used for experiments and are here for convenience until they can be refactored:

```
payload 1000
```

**Reminder:** there are four categories of process that use Hindsight: [clients](clients.md), [agents](agents.md), the [coordinator](coordinator.md) and the [collector](collector.md).  Some configuration values are used by multiple prcesses.

Configuring the per-node buffer pool (relevant to clients and agents)

* `cap`: The default number of buffers in Hindsight's buffer pool.  Used by clients exclusively.
* `buf_length`: The default buffer size, in bytes, of buffers in Hindsight's buffer pools.  Used by clients exclusively.
  * *The size, in bytes, of Hindsight's buffer pool is `cap * buf_length`*

Configuring client-side sampling probabilities

* `head_sampling_probability`: An optional head-based sampling probability, by default set to 0. Accepts values 0 to 1.  If set, then traces will be eagerly triggered with a random probability.
* `retroactive_sampling_percentage`: By default Hindsight does retroactive tracing for 100% of requests.  This config value reduces this percentage, which thereby reduces overheads.  For example if set to 50% then 50% of requests will not generate any data at all.

Configuring addresses 

* `addr`: The local hostname or IP address of the host loading this config file.  Hindsight clients will use this as their local breadcrumb
* `port`: The port to be used by the agent.  Hindsight clients will use this as their local breadcrumb.  The Hindsight agent will listen on this port.
* `lc_addr`: The hostname or IP address of the [coordinator](coordinator.md).
* `lc_port`: The port of the coordinator of the [coordinator](coordinator.md)
* `r_addr`: The hostname or IP address of the [collector](collector.md).
* `r_port`: The port of the [collector](collector.md).

Experiment-specific

* `payload`: No idea

## Process names

Hindsight clients and agents identify each other with process names:
* Within a client, when Hindsight is initialized, `hindsight_init(char* process_name)` receives the process name; this must be provided by the caller
* With the co-located agent, the process name is passed as a command line argument: `go run cmd/agent2/main.go --serv process_name`
* Both the client and the agent will load the config file from `/etc/hindsight_conf/{process_name}.conf`, e.g. `/etc/hindsight_conf/my_process.conf`
  * If no config file exists, it will load `default.conf`; this is fine for a single-node setup
* The shared-memory files will be prefixed with the `process_name`; if the agent is given a different process_name than the client, then they won't find each other.

## Writing a configuration

For a node named `datanode`, copy default.conf into a new file `datanode.conf`.  Update the `addr` and `lc_addr` with appropriate values, e.g.
```
cd client
cp conf/default.conf conf/datanode.conf
```
Then edit `conf/datanode.conf`:
```
cap 10000
buf_length 32768
addr 127.0.0.1
port 5050
lc_addr 127.0.0.1
lc_port 5252
r_addr 127.0.0.1
r_port 5253
retroactive_sampling_percentage 1.0
head_sampling_probability 0.0
```
Then
```
sudo make install
```
This will copy all config files to `/etc/hindsight_conf`.  e.g.
```
less /etc/hindsight_conf/datanode.conf
```
