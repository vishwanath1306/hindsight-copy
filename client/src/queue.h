#ifndef _HINDSIGHT_QUEUE_H_
#define _HINDSIGHT_QUEUE_H_

#include <stdlib.h>
#include <stdint.h>
#include <stdbool.h>

// Metadata at the start of the shared memory region of the queue
typedef struct QueueMetadata {
    bool initialized; // Set to true once everything is set up
    size_t capacity; // Capacity in number of elements
    size_t element_metadata_size; // Size of element metadata
    size_t element_size; // Size of one element content
    size_t element_total_size; // metadata + content
    __attribute__((aligned(64))) size_t head; // Index (not ptr) of the head of the queue
    __attribute__((aligned(64))) size_t tail; // Index (not ptr) of the tail of the queue
} QueueMetadata;

// Metadata at the start of each queue element
typedef struct QueueElementMetadata {
    int status; // 0=empty, 1=writing, 2=full, 3=reading
} QueueElementMetadata;

typedef struct Queue {
    // shmem pointers:
    QueueMetadata* meta; // Metadata of the queue, **within** the shmem region
    char* baseptr; // Baseptr of the shmem region
    char* queue; // Baseptr of the queue region, comes after the metadata
} Queue;

// Return true if a shmem queue exists for the specified name
bool queue_exists(const char* fname);

// Creates a new shm queue at the specified filename
Queue queue_init(const char* fname, size_t element_size, size_t capacity);

// Uses an existing shm queue at the specified filename.
// Blocks until the file exists
Queue queue_init_existing(const char* fname);

void queue_print(Queue* q);

void queue_put_blocking(Queue* q, char* element);
void queue_put_blocking_multi(Queue* q, char* elements, size_t num_elements);
bool queue_put_nonblocking(Queue* q, char* element);
size_t queue_put_nonblocking_multi(Queue* q, char* elements, size_t num_elements);

void queue_get_blocking(Queue* q, char* dst_element);
bool queue_get_nonblocking(Queue* q, char* dst_element);
size_t queue_get_nonblocking_multi(Queue* q, char* elements, size_t max_elements);

#endif // _HINDSIGHT_QUEUE_H_