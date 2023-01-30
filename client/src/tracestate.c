#include "tracestate.h"

#include <assert.h>
#include <string.h>
#include <time.h>

TraceState tracestate_create() {
    TraceState trace = {false};
    return trace;
}

// Write the header to the current buffer
void tracestate_write_header(TraceState* trace) {
    char* base = trace->buffer.base;
    char* dst;
    size_t dst_size;
    buffer_write(&trace->buffer, sizeof(TraceHeader), &dst, &dst_size);

    // write_header should only be called on a fresh buffer.
    assert(dst_size == sizeof(TraceHeader));
    assert(base == dst);

    // Write the header
    *((TraceHeader*) dst) = trace->header;
    trace->current = (TraceHeader*) dst;
}

// time_t tracestate_get_time() {
//     struct timespec ts;
//     timespec_get(&ts, TIME_UTC);
//     return ts.tv_sec * 1000000000 + ts.tv_nsec; 
// }

uint64_t tracestate_get_time() {
    unsigned int lo, hi;

    // RDTSC copies contents of 64-bit TSC into EDX:EAX
    asm volatile("rdtsc" : "=a" (lo), "=d" (hi));
    return (unsigned long long)hi << 32 | lo;
}

void tracestate_begin(TraceState* trace, BufManager* mgr, uint64_t trace_id) {
    tracestate_begin_with_sampling(trace, mgr, trace_id, 0, UINT64_MAX);
}

void tracestate_begin_with_sampling(TraceState* trace, BufManager* mgr, uint64_t trace_id, uint64_t head_sampling_threshold, uint64_t retroactive_sampling_threshold) {
    if (trace->active) {
        // If the trace ID is the same, ignore this call
        if (trace_id == trace->header.trace_id) return;

        // If traceID is different, need to return the old buffer
        if (trace->recording) {
            // trace->current->completed = tracestate_get_time();
            trace->current->size = trace->buffer.ptr - trace->buffer.base;
            bufmanager_return(mgr, trace->header.trace_id, &trace->buffer);
        }
    }
    buffer_clear(&trace->buffer);

    trace->active = true;

    // Apply head sampling threshold
    trace->head_sampled = (trace_id <= head_sampling_threshold);

    // Apply retroactive sampling threshold
    trace->recording = trace->head_sampled || (trace_id <= retroactive_sampling_threshold);

    // Set the new header
    trace->header.trace_id = trace_id;
    trace->header.buffer_number = 0;
    trace->header.null_buffer_count = 0;
    trace->header.size = 0;
    trace->header.acquired = tracestate_get_time();

    // Acquire a fresh buffer and write the header
    if (trace->recording) {
        bufmanager_acquire(mgr, &trace->buffer);
        if (trace->buffer.id == -2) {
            // TODO: probably shouldn't be implemented like this
            trace->header.null_buffer_count++;
        }
        trace->header.buffer_id = trace->buffer.id;
        trace->header.prev_buffer_id = trace->header.buffer_id; // First buffer points to itself
        tracestate_write_header(trace);
    }
}

void tracestate_end(TraceState* trace, BufManager* mgr) {
    if (!trace->active) return;

    if (trace->recording) {
        // Finish buffer data
        // trace->current->completed = tracestate_get_time();
        trace->current->size = trace->buffer.ptr - trace->buffer.base;

        // Return the current buffer
        bufmanager_return(mgr, trace->header.trace_id, &trace->buffer);
    }

    trace->active = false;
    trace->head_sampled = false;
    trace->recording = false;

    // Clear the header
    trace->header.buffer_number = 0;
    trace->header.trace_id = 0;
    trace->header.null_buffer_count = 0;
    trace->header.size = 0;
    trace->current = &trace->header;
}

void tracestate_write_data(TraceState* trace, 
                           BufManager* mgr,
                           size_t write_size, 
                           char** dst, 
                           size_t* dst_size) {
    if (!trace->recording) return;

    // Common case: there's room in the current buffer. Write and return.
    buffer_write(&trace->buffer, write_size, dst, dst_size);
    if (*dst_size != 0) return;

    int prev_buffer_id = trace->header.buffer_id;

    // Buffer is full, return old buffer
    // trace->current->completed = tracestate_get_time();
    trace->current->size = trace->buffer.ptr - trace->buffer.base;
    bufmanager_return(mgr, trace->header.trace_id, &trace->buffer);

    // Acquire new buffer and write header
    bufmanager_acquire(mgr, &trace->buffer);
    trace->header.buffer_number++;
    trace->header.acquired = tracestate_get_time();
    trace->header.buffer_id = trace->buffer.id;
    trace->header.prev_buffer_id = prev_buffer_id;
    if (trace->buffer.id == -2) {
        // TODO: probably shouldn't be implemented like this
        trace->header.null_buffer_count++;
    }
    tracestate_write_header(trace);

    // Retry the write
    buffer_write(&trace->buffer, write_size, dst, dst_size);

    // Implies we're getting buffers with 0 capacity
    assert(*dst_size > 0);
}

// Writes data to the trace; called by tracepoint
void tracestate_write(TraceState* trace, 
                      BufManager* mgr,
                      char* buf,
                      size_t buf_size) {
    if (trace->recording) {
        char* dst;
        size_t dst_size;

        while (buf_size != 0) {
            // Try to write everything
            tracestate_write_data(trace, mgr, buf_size, &dst, &dst_size);

            // Write what we're allowed
            memcpy((void*) dst, (void*) buf, dst_size);

            buf += dst_size;
            buf_size -= dst_size;
        }
    }
}

// Writes data to the trace; called by tracepoint
bool tracestate_try_write(TraceState* trace,
                          char* buf,
                          size_t buf_size) {
    return trace->recording && buffer_try_write_all(&trace->buffer, buf, buf_size);
}
