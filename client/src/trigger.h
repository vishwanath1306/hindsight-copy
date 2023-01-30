#ifndef _HINDSIGHT_CLIENT_TRIGGER_H_
#define _HINDSIGHT_CLIENT_TRIGGER_H_

#include <stddef.h>
#include <stdbool.h>

#include "queue.h"

typedef struct Triggers {
    const char* name; // Name of this service

    Queue queue; // Used to send triggers
} Triggers;

// For now, a trigger is just an ID and trace_id
typedef struct Trigger {
    int trigger_id; // The ID of the trigger that fired
    uint64_t base_trace_id; // The trace that fired it
    uint64_t trace_id; // The trace ID to report for this trigger (e.g. lateral trace ID)
} Trigger;

// name is used for mapping to the appropriate shmem file
// capacity is used to decide queue size
Triggers triggers_init(const char* name, size_t capacity);

// Initializes an existing file, waiting for it to exist
Triggers triggers_init_existing(const char* name);

// For now, we are just sen
void triggers_fire(Triggers* t, int trigger_id, uint64_t base_trace_id, uint64_t trace_id);

#endif // _HINDSIGHT_CLIENT_TRIGGER_H_