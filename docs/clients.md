# Clients

Client applications must be instrumented with Hindsight in order to generate data.

Hindsight offers two ways to instrument client applications.  Applications can use Hindsight's OpenTelemetry integration and instrument against OpenTelemetry APIS.  (TODO: instructions).

Alternatively applications can instrument directly against Hindsight's APIs defined in [hindsight.h](../client/include/hindsight.h)

# Initializing

A process must initialize Hindsight upon startup.  This can be done by either loading a config file, or programmatically constructing the config.

```
void hindsight_init(const char* service_name);
void hindsight_init_with_config(const char* service_name, HindsightConfig config);
```

By default you should use `hindsight_init` and specify a `service_name`.  The corresponding `service_name.conf` file will be used for the client's configuration.  See [configuration.md](configuration.md) for details on this configuration file.  `service_name` is an important consideration, as you will need to also give the same `service_name` when starting Hindsight's agent process on the same machine.

# Starting and Stopping Traces

When a request begins executing in a thread, call `hindsight_begin`:
```
void hindsight_begin(uint64_t trace_id);
```

When a request finishes executing in a thread, call `hindsight_end`:
```
void hindsight_end();
```

# Recording Trace Data

To report data:
```
void hindsight_tracepoint(char* buf, size_t buf_size);
```

Hindsight's direct APIs are low-level and thus the caller can determine their own serialization format for tracepoint data.

Alternatively to write directly into Hindsight buffers, the following method returns a pointer to a buffer
that the caller can write to.
```
void hindsight_tracepoint_write(size_t write_size, char** dst, size_t* dst_size);
```

# Propagating Trace Contexts

When making a call to a different machine, a client application should get the Hindsight context and include it in the remote communication:
```
char* hindsight_serialize();
```

Likewise on the receiver side, the context should be reinstated:
```
void hindsight_deserialize(char* baggage);
```

# Manually adding breadcrumbs

Using `hindsight_serialize` and `hindsight_deserialize` will automatically establish a breadcrumb from caller to callee on the callee side.  In addition, the caller can establish a *forward breadcrumb* on the caller side as follows:
```
void hindsight_forward_breadcrumb(const char* addr);
```