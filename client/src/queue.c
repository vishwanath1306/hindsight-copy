#include <stdbool.h>
#include <stdio.h>
#include <string.h>
#include <unistd.h>
#include <sys/mman.h>
#include <fcntl.h>
#include <assert.h>
#include <sched.h>

#include "queue.h"
#include "memory.h"



Queue queue_init(const char* fname, size_t element_size, size_t capacity) {
    Queue q;

    size_t element_metadata_size = sizeof(QueueElementMetadata);
    size_t element_total_size = element_size + element_metadata_size;

    size_t shmem_size = sizeof(QueueMetadata) + capacity * element_total_size;

    int fd = open(fname, O_RDWR | O_CREAT, 0666);
    assert(fd >= 0);

    int i = ftruncate(fd, shmem_size);
    assert(i == 0);

    void* shm = mmap(NULL, shmem_size, PROT_READ | PROT_WRITE, MAP_SHARED, fd, 0);
    assert(shm != MAP_FAILED);  
    close(fd);

    memset(shm, 0, shmem_size);

    q.meta = (QueueMetadata*) shm;
    q.baseptr = (char*) shm;
    q.queue = q.baseptr + sizeof(QueueMetadata);

    q.meta->head = 0;
    q.meta->tail = 0;
    q.meta->capacity = capacity;
    q.meta->element_metadata_size = element_metadata_size;
    q.meta->element_size = element_size;
    q.meta->element_total_size = element_total_size;
    q.meta->initialized = true;

    printf("Created queue ");
    printf("capacity=%ld ", q.meta->capacity);
    printf("element_size=%ld ", q.meta->element_size);
    printf("element_total_size=%ld ", q.meta->element_total_size);
    printf("at %s (%p)\n", fname, q.baseptr);

    return q;
}

Queue queue_init_existing(const char* fname) {
    Queue q;

    // Wait until the file exists
    while (access(fname, F_OK) != 0) {
        printf("%s does not exist, waiting...\n", fname);
        usleep(1000000);
    }
    
    // Open the file, get its length
    int fd = open(fname, O_RDWR, 0666);
    assert(fd >= 0);

    struct stat st;
    fstat(fd, &st);
    size_t shmem_size = st.st_size;

    void* shm = mmap(NULL, shmem_size, PROT_READ | PROT_WRITE, MAP_SHARED, fd, 0);
    assert(shm != MAP_FAILED);  
    close(fd);

    q.meta = (QueueMetadata*) shm;
    q.baseptr = (char*) shm;
    q.queue = q.baseptr + sizeof(QueueMetadata);

    while (!q.meta->initialized) {
        printf("Waiting for initialization of %s...\n", fname);
        usleep(1000000);
    }

    printf("Loaded existing queue ");
    printf("capacity=%ld ", q.meta->capacity);
    printf("element_size=%ld ", q.meta->element_size);
    printf("element_total_size=%ld ", q.meta->element_total_size);
    printf("at %s (%p)\n", fname, q.baseptr);

    return q;
}
void queue_print(Queue* q) {
    size_t head = q->meta->head;
    size_t tail = q->meta->tail;
    size_t occupancy = tail-head;
    size_t remaining = q->meta->capacity - occupancy;
    printf("occupancy=%ld remaining=%ld head=%ld tail=%ld\n", occupancy, remaining, head, tail);
}

char* queue_ptr(Queue* q, size_t index) {
    index = index % q->meta->capacity;
    return q->queue + (q->meta->element_total_size * index);
}

size_t queue_get_nonblocking_multi(Queue* q, char* elements, size_t max_elements) {
    assert(max_elements != 0);
    while (true) {
        // First, read the current head and tail values of the queue
        __sync_synchronize();
        size_t head = q->meta->head;
        size_t tail = q->meta->tail;

        // If the queue is currently empty, we can return
        int64_t delta = tail-head;
        if (delta <= 0) {
            // printf("Head=%ld Tail=%ld Delta=%ld\n", head, tail, delta);
            return 0;
        }

        // We'll dequeue the max of delta or max_elements
        if (delta > max_elements) {
            delta = max_elements;
        }

        // Try updating the head pointer; somebody else might have taken it
        if (!__sync_bool_compare_and_swap(&q->meta->head, head, head+delta)) {
            continue;
        }

        // Relevant sizes
        size_t element_size = q->meta->element_size;

        // We got our elements.
        for (int i = 0; i < delta; i++) {
            // Grab the element
            char* e_ptr = queue_ptr(q, head+i);
            QueueElementMetadata* e_md = (QueueElementMetadata*) e_ptr;

            // It's possible a writer is still writing this element
            // Even though this should be a non-blocking call, we will block here :(
            int max_backoff = 100000; // 100ms
            int backoff = 10;
            while (!__sync_bool_compare_and_swap(&e_md->status, 2, 3)) {
                usleep(backoff);
                backoff *= 2;
                if (backoff > max_backoff) {
                    backoff = max_backoff;
                }
            }

            // Read the element
            char* e_content = e_ptr + sizeof(QueueElementMetadata);
            char* dst_ptr = elements + (i * element_size);
            memcpy(dst_ptr, e_content, element_size);

            // Update status, fail if somebody else touched it
            assert(__sync_bool_compare_and_swap(&e_md->status, 3, 0));
        }

        return delta;
    }

}

bool queue_get_nonblocking(Queue* q, char* element) {
    return queue_get_nonblocking_multi(q, element, 1) == 1;
}

void queue_get_blocking(Queue* q, char* element) {
    // Only allowed to put if (tail-head) < capacity
    int max_backoff = 100000; // 100ms
    int backoff = 10;

    // Call non-blocking impl and backoff
    while (!queue_get_nonblocking(q, element)) {
        usleep(backoff);
        backoff *= 2;
        if (backoff > max_backoff) {
            backoff = max_backoff;
        }
    }
}

size_t queue_put_nonblocking_multi(Queue* q, char* elements, size_t num_elements) {
    while (true) {
        // First, read the current head and tail values of the queue
        __sync_synchronize();
        size_t tail = q->meta->tail;
        size_t head = q->meta->head;

        // If the queue is currently full, we can simply return
        int64_t delta = tail-head;
        if (delta >= q->meta->capacity) {
            return 0;
        }

        // Figure out how many we can put
        size_t num_to_write = q->meta->capacity - delta;
        if (num_to_write > num_elements) {
            num_to_write = num_elements;
        }

        // Try updating the tail pointer; somebody else might have taken it
        if (!__sync_bool_compare_and_swap(&q->meta->tail, tail, tail+num_to_write)) {
            continue;
        }

        size_t element_size = q->meta->element_size;

        // We can do all of the writes.
        for (int i = 0; i < num_to_write; i++) {            
            char* e_ptr = queue_ptr(q, tail+i);
            QueueElementMetadata* e_md = (QueueElementMetadata*) e_ptr;

            // It's possible a reader is still reading this element
            // Even though this should be a non-blocking call, we will block here :(
            int max_backoff = 100000; // 100ms
            int backoff = 10;
            while (!__sync_bool_compare_and_swap(&e_md->status, 0, 1)) {
                usleep(backoff);
                backoff *= 2;
                if (backoff > max_backoff) {
                    backoff = max_backoff;
                }
            }

            // Write the element
            char* src_ptr = elements + (i * element_size);
            char* e_content = e_ptr + sizeof(QueueElementMetadata);
            memcpy(e_content, src_ptr, element_size);

            // Update status, fail if somebody else touched it
            assert(__sync_bool_compare_and_swap(&e_md->status, 1, 2));
        }

        return num_to_write;
    }

}

bool queue_put_nonblocking(Queue* q, char* element) {
    return queue_put_nonblocking_multi(q, element, 1) == 1;
}

void queue_put_blocking(Queue* q, char* element) {
    queue_put_blocking_multi(q, element, 1);
}

void queue_put_blocking_multi(Queue* q, char* elements, size_t num_elements) {
    assert(num_elements > 0);

    int max_backoff = 100000; // 100ms
    int backoff = 10;

    size_t element_size = q->meta->element_size;

    while (true) {
        size_t num_written = queue_put_nonblocking_multi(q, elements, num_elements);

        elements = elements + (num_written * element_size);
        num_elements -= num_written;

        if (num_elements == 0) {
            return;
        }

        if (num_written > 0) {
            backoff = 10;
        }
        usleep(backoff);
        backoff *= 2;
        if (backoff > max_backoff) {
            backoff = max_backoff;
        }
    }
}


