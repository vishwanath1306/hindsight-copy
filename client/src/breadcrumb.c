#include <stdio.h>
#include <fcntl.h>
#include <sys/mman.h>
#include <assert.h>
#include <string.h>

#include "breadcrumb.h"
#include "common.h"

#define BREADCRUMBS_SHM_FILENAME(name) get_shm_fname(name, "breadcrumbs_queue")

Breadcrumbs breadcrumbs_init(const char* name, size_t capacity) {
    Breadcrumbs b;
    b.name = name;
    b.queue = queue_init(BREADCRUMBS_SHM_FILENAME(name), sizeof(Breadcrumb), capacity);
    return b;
}

Breadcrumbs breadcrumbs_init_existing(const char* name) {
    Breadcrumbs b;
    b.name = name;
    b.queue = queue_init_existing(BREADCRUMBS_SHM_FILENAME(name));
    return b;
}

void breadcrumbs_add(Breadcrumbs* b, uint64_t trace_id, const char* address) {
    Breadcrumb crumb;
    crumb.trace_id = trace_id;
    crumb.type = 0;
    truncate_string(crumb.address, address, ADDR_MAX_SIZE);
    queue_put_nonblocking(&b->queue, (char*) &crumb);
}

// Add a forward breadcrumb
void breadcrumbs_add_forward(Breadcrumbs* b, uint64_t trace_id, const char* address) {
    Breadcrumb crumb;
    crumb.trace_id = trace_id;
    crumb.type = 1;
    truncate_string(crumb.address, address, ADDR_MAX_SIZE);
    queue_put_nonblocking(&b->queue, (char*) &crumb);
}
