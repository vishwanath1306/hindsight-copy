# Telemetry

Hindsight telemetry is currently supported by Hindsight agents and is WIP for the other components.

The hindsight agent takes the following two command line arguments:

```
  -output string
        Filename for outputting agent telemetry.  If specified, will write a csv of agent telemetry data.  Disabled by default.
  -verbose
        If set to true, prints telemetry to the command line.  False by default.
```

If the `-output` flag is specified, the agent will write CSV telemetry data to the specified file.  Each row of the CSV file has some summary statistics about the agent.  The CSV file has headers for each column.  As Hindsight updates over time, the column index might change -- if you're using the telemetry data for further processing, you should rely on named headers rather than positional indices.

Example output telemetry file:

```
t,interval_ms,queue_id,data_mb,reported_mb,evicted_mb,triggers,local_triggers,remote_triggers,dropped_triggers,evicted_triggers,tput_data_mb,tput_reported_mb,tput_evicted_mb,tput_triggers,tput_local_triggers,tput_remote_triggers,tput_dropped_triggers,tput_evicted_triggers,cache_occupancy,eviction_percent,internal_bottleneck,event_horizon_ms,report_horizon_ms
1644919999673768532,1000,total,1148.94,0.94,0.00,22516,22516,0,2665,15648,1148.69,0.94,0.00,22511,22511,0,2664,15645,106.7,99.9,33.5,634,
1644919999673768532,1000,10,12.06,0.06,0.00,234,234,0,0,0,12.06,0.06,0.00,234,234,0,0,0,9.6,0.0,,,
1644919999673768532,1000,11,115.94,0.38,0.00,2255,2255,0,0,906,115.91,0.37,0.00,2255,2255,0,0,906,47.1,99.3,,,
```

If the `-verbose` flag is specified then telemetry is also printed to the command line, prefixed by the word `Telemetry: `.  