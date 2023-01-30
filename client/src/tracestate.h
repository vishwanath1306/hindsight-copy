#ifndef _HINDSIGHT_TRACESTATE_H_
#define _HINDSIGHT_TRACESTATE_H_

#include <stdbool.h>

#include "buffer.h"

// TraceHeader represents the header data that Hindsight inserts at the start of every buffer
// It could include stuff like span IDs, but I'm not sure that's necessary in the header
typedef struct TraceHeader {
    uint64_t trace_id;
    uint64_t acquired;
    // uint64_t completed;
    int buffer_id;
    int prev_buffer_id;
    uint32_t size;
    short buffer_number;
    short null_buffer_count;
} TraceHeader;

// TraceState represents an active, ongoing trace in the current process
typedef struct TraceState {
    bool active;
    bool head_sampled;   // Is the trace sampled by head-based sampling?
    bool recording; // Are we actually recording data? true if head-sampled or retro-sampled
    TraceHeader header; // The trace header data. Gets written to every buffer.
    TraceHeader* current; // Header within the current active buffer.
    Buffer buffer; // The current active buffer.
} TraceState;

// TraceState can also be initialized to {false}
TraceState tracestate_create();

// Starts a new trace state for the specified trace ID
// This call will always enable retroactive sampling, and will never apply head-sampling
void tracestate_begin(TraceState* trace, BufManager* mgr, uint64_t trace_id);

// Version of tracestate_begin that will potentially not sample the trace if retroactive_sampling_percentage is set
// This call will only start a trace if trace_id <= retroactive_sampling_threshold
void tracestate_begin_with_sampling(TraceState* trace, BufManager* mgr, uint64_t trace_id, uint64_t head_sampling_threshold, uint64_t retroactive_sampling_threshold);

// Ends the current trace state, flushes the buffer
void tracestate_end(TraceState* trace, BufManager* mgr);

// Acquires a buffer to write to, that will be in the trace
void tracestate_write_data(TraceState* trace, 
                           BufManager* mgr,
                           size_t write_size, 
                           char** dst, 
                           size_t* dst_size);

// Writes a buffer to the trace; called by tracepoint
void tracestate_write(TraceState* trace,
                      BufManager* mgr,
                      char* buf,
                      size_t buf_size);

// Attempts to one-shot write buffer
bool tracestate_try_write(TraceState* trace,
                          char* buf,
                          size_t buf_size);




#endif // _HINDSIGHT_TRACESTATE_H_
