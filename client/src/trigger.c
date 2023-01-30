#include <stdio.h>
#include <fcntl.h>
#include <sys/mman.h>
#include <assert.h>
#include <string.h>

#include "trigger.h"
#include "common.h"

#define TRIGGERS_SHM_FILENAME(name) get_shm_fname(name, "triggers_queue")

Triggers triggers_init(const char* name, size_t capacity) {
    Triggers t;
    t.name = name;
    t.queue = queue_init(TRIGGERS_SHM_FILENAME(name), sizeof(Trigger), capacity);
    return t;
}

Triggers triggers_init_existing(const char* name) {
    Triggers t;
    t.name = name;
    t.queue = queue_init_existing(TRIGGERS_SHM_FILENAME(name));
    return t;
}

void triggers_fire(Triggers* t, int trigger_id, uint64_t base_trace_id, uint64_t trace_id) {
    Trigger trigger = {trigger_id, base_trace_id, trace_id};
    queue_put_nonblocking(&t->queue, (char*) &trigger);
}