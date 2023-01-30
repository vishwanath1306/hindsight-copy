#ifndef _HINDSIGHT_CLIENT_BUFFER_H_
#define _HINDSIGHT_CLIENT_BUFFER_H_

#include <stddef.h>
#include <stdbool.h>

#include "queue.h"

// Metadata stored at the head of the bufmanager shmem region
typedef struct PoolMetadata {
    bool initialized;
    size_t capacity;
    size_t buffer_size;
} PoolMetadata;

// Some local stats just for convenience
typedef struct BufferStats {
    size_t pool_acquired;
    size_t null_acquired;
    size_t pool_released;
    size_t null_released;
} BufferStats;

// Manages shared memory buffers
typedef struct BufManager {
    const char* name; // Name of this service

    BufferStats stats; // Client-side stats

    char* baseptr; // Pointer to start of shared-memory region
    PoolMetadata* meta; // Metadata to this pool; lives at start of shmem region
    char* pool; // Pointer to shared-memory region used for buffers

    Queue available; // Used for receiving fresh buffers
    Queue complete; // Used for sending completed buffers.  
                    // TODO: queue impl will need to be updated to send both (traceid, bufid)

    char* null_buffer; // Used if unable to acquire a buffer from available queue

    uint32_t null_buffer_index;
} BufManager;

// Points to a buffer allocated in shared memory
// Includes some metadata not stored in shared memory
typedef struct Buffer {
    int id; // Equivalent to the index of this buffer in the buffer pool
    size_t remaining; // Space remaining in the underlying buffer
    char* ptr; // Pointer to next available byte in buffer
    char* base; // Base pointer
} Buffer;

// This struct is read from the available queue
typedef struct AvailableBuffer {
    int buffer_id; // The ID of the available buffer
} AvailableBuffer;

// This struct is written to the complete queue
typedef struct CompleteBuffer {
    uint64_t trace_id; // The trace ID that used this buffer
    int buffer_id;
} CompleteBuffer;

// Initializes a bufmanager, creating shmem regions and queues
BufManager bufmanager_init(const char* name,
                           size_t capacity,
                           size_t buffer_size);

// Initializes a bufmanager with existing shm regions and queues
BufManager bufmanager_init_existing(const char* name);

// Makes all buffers available; called on initialization
void bufmanager_make_all_buffers_available(BufManager* mgr);

// Acquires a buffer from the queue, setting it in dst.
// Doesn't block -- will set the null buffer if nothing can be acquired
void bufmanager_acquire(BufManager* mgr, Buffer* dst);

// Returns the current buffer and clears it
void bufmanager_return(BufManager* mgr, uint64_t trace_id, Buffer* dst);

// Initializes a buffer with ID -1, nullptr, and 0 remaining
Buffer buffer_create();

// Sets a buffer to ID -1, nullptr, and 0 remaining
void buffer_clear(Buffer* b);

// True if remaining is 0, false otherwise
bool buffer_is_full(Buffer* b);

// Returns remaining space in buffer
bool buffer_remaining(Buffer* b);

// True if buffer ID is >= 0, which is set whenever buffer_clear is called
bool buffer_is_valid(Buffer* b);

// Requests to write `size`-much data to the buffer.  The caller
// will receive a pointer in `dst` and will be responsible for actually
// writing the data to `dst`.  The caller can write up to `dst_size`
// data.
//
// If the caller requests more room than available (ie `size` > `remaining`)
// then `dst_size` will only be the remaining capacity, and the caller
// must acquire a new buffer to write the remaining data.
void buffer_write(Buffer* b, size_t size, char** dst, size_t* dst_size);
bool buffer_try_write_all(Buffer* b, char* buf, size_t buf_size);


#endif // _HINDSIGHT_CLIENT_BUFFER_H_