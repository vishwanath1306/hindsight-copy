#ifndef _HINDSIGHT_CLIENT_BREADCRUMB_H_
#define _HINDSIGHT_CLIENT_BREADCRUMB_H_

#include <stddef.h>
#include <stdbool.h>

#include "queue.h"

#define ADDR_MAX_SIZE 32

typedef struct Breadcrumbs {
    const char* name; // Name of this service

    Queue queue; // Used to send breadcrumbs
} Breadcrumbs;

// For now, a breadcrumb is a fixed-length char array
typedef struct Breadcrumb {
    uint64_t trace_id; // The trace of this breadcrumb
    short type; // Regular (0) or Forward (1)
    char address[ADDR_MAX_SIZE]; // Literal host:port string
} Breadcrumb;

// name is used for mapping to the appropriate shmem file
// capacity is used to decide queue size
Breadcrumbs breadcrumbs_init(const char* name, size_t capacity);

Breadcrumbs breadcrumbs_init_existing(const char* name);

// Add a regular (backwards) breadcrumb
void breadcrumbs_add(Breadcrumbs* b, uint64_t trace_id, const char* address);

// Add a forward breadcrumb
void breadcrumbs_add_forward(Breadcrumbs* b, uint64_t trace_id, const char* address);


#endif // _HINDSIGHT_CLIENT_BREADCRUMB_H_