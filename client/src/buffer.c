#include <stdio.h>

#include <fcntl.h>
#include <sys/mman.h>
#include <assert.h>
#include <unistd.h>
#include <sys/stat.h>

#include "buffer.h"
#include "common.h"

#define POOL_SHM_FILENAME(name) get_shm_fname(name, "pool")
#define AVAILABLE_SHM_FILENAME(name) get_shm_fname(name, "available_queue")
#define COMPLETE_SHM_FILENAME(name) get_shm_fname(name, "complete_queue")
#define NULL_BUFFER_COUNT 1000

char* bufmanager_pool_init(const char* fname, size_t fsize) {
    void* shm;
    
    int fd = open(fname, O_RDWR | O_CREAT, 0666);
    assert(fd >= 0);

    int i = ftruncate(fd, fsize);
    assert(i == 0);

    shm = mmap(NULL, fsize, PROT_READ | PROT_WRITE, MAP_SHARED, fd, 0);
    assert(shm != MAP_FAILED);
    close(fd);

    memset(shm, 0, fsize);

    return (char*) shm;
}

char* bufmanager_pool_init_existing(const char* fname) {
    void* shm;

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
    size_t fsize = st.st_size;

    // Map it
    shm = mmap(NULL, fsize, PROT_READ | PROT_WRITE, MAP_SHARED, fd, 0);
    assert(shm != MAP_FAILED);
    close(fd);

    return (char*) shm;
}

void bufmanager_make_all_buffers_available(BufManager* mgr) {
    for (size_t i = 0; i < mgr->meta->capacity; i++) {
        AvailableBuffer av;
        av.buffer_id = i;
        bool success = queue_put_nonblocking(&mgr->available, (char*) &av);

        assert(success);
    }

    // Double check queue is now full
    AvailableBuffer av;
    bool success = queue_put_nonblocking(&mgr->available, (char*) &av);
    assert(!success);
}

BufManager bufmanager_init(const char* name,
                           size_t capacity,
                           size_t buffer_size) {
    BufManager m;
    m.name = name;

    // Initialize the stats
    m.stats.pool_acquired = 0;
    m.stats.null_acquired = 0;
    m.stats.pool_released = 0;
    m.stats.null_released = 0;
    
    size_t metadata_size = sizeof(PoolMetadata);
    if (metadata_size % 1024 != 0) {
        /* Align metadata to 1024 boundary to avoid fragmentation of buffers */
        metadata_size = (1 + metadata_size / 1024) * 1024;
    }

    const char* fname = POOL_SHM_FILENAME(name);
    size_t pool_size = metadata_size + capacity * buffer_size;
    m.baseptr = bufmanager_pool_init(fname, pool_size);
    m.meta = (PoolMetadata*) m.baseptr;
    m.meta->capacity = capacity;
    m.meta->buffer_size = buffer_size;
    m.pool = m.baseptr + metadata_size;

    printf("Created buffer pool, ");
    printf("capacity=%ld ", m.meta->capacity);
    printf("buffer_size=%ld ", m.meta->buffer_size);
    printf("at %s\n", fname);

    m.available = queue_init(AVAILABLE_SHM_FILENAME(name), sizeof(AvailableBuffer), capacity);
    m.complete = queue_init(COMPLETE_SHM_FILENAME(name), sizeof(CompleteBuffer), capacity);

    m.null_buffer = (char*) malloc(m.meta->buffer_size * NULL_BUFFER_COUNT);
    m.null_buffer_index = 0;

    bufmanager_make_all_buffers_available(&m);

    m.meta->initialized = true;
    return m;
}

BufManager bufmanager_init_existing(const char* name) {
    BufManager m;
    m.name = name;

    // Initialize the stats
    m.stats.pool_acquired = 0;
    m.stats.null_acquired = 0;
    m.stats.pool_released = 0;
    m.stats.null_released = 0;

    size_t metadata_size = sizeof(PoolMetadata);
    if (metadata_size % 1024 != 0) {
        /* Align metadata to 1024 boundary to avoid fragmentation of buffers */
        metadata_size = (1 + metadata_size / 1024) * 1024;
    }

    const char* fname = POOL_SHM_FILENAME(name);
    m.baseptr = bufmanager_pool_init_existing(fname);
    m.meta = (PoolMetadata*) m.baseptr;
    m.pool = m.baseptr + metadata_size;

    while (!m.meta->initialized) {
        printf("Waiting for pool initialization...\n");
        usleep(1000000);
    }

    printf("Loaded existing buffer pool, ");
    printf("capacity=%ld ", m.meta->capacity);
    printf("buffer_size=%ld ", m.meta->buffer_size);
    printf("at %s\n", fname);

    m.available = queue_init_existing(AVAILABLE_SHM_FILENAME(name));
    m.complete = queue_init_existing(COMPLETE_SHM_FILENAME(name));

    assert(m.available.meta->element_size == sizeof(AvailableBuffer));
    assert(m.complete.meta->element_size == sizeof(CompleteBuffer));

    m.null_buffer = (char*) malloc(m.meta->buffer_size * NULL_BUFFER_COUNT);
    m.null_buffer_index = 0;
    return m;   
}

void bufmanager_acquire(BufManager* mgr, Buffer* dst) {
    // Shouldn't be acquiring into a buffer that hasn't been released
    assert(!buffer_is_valid(dst));

    AvailableBuffer av = {-1};
    if (queue_get_nonblocking(&mgr->available, (char*) &av)) {
        dst->id = av.buffer_id;
        dst->remaining = mgr->meta->buffer_size;
        dst->ptr = mgr->pool + (av.buffer_id * mgr->meta->buffer_size);
        dst->base = dst->ptr;

        // TODO: allow to #define away
        __sync_fetch_and_add(&mgr->stats.pool_acquired, 1);
    } else {
        uint32_t null_i = __sync_fetch_and_add(&mgr->null_buffer_index, 1);
        size_t null_offset = (null_i % NULL_BUFFER_COUNT) * mgr->meta->buffer_size;

        dst->id = -2;
        dst->remaining = mgr->meta->buffer_size;
        dst->ptr = mgr->null_buffer + null_offset;
        dst->base = dst->ptr;

        // TODO: allow to #define away
        __sync_fetch_and_add(&mgr->stats.null_acquired, 1);
    }
}

void bufmanager_return(BufManager* mgr, uint64_t trace_id, Buffer* dst) {
    // No asserts; allowed to return an invalid buffer
    if (dst->id >= 0) {
        CompleteBuffer b = {trace_id, dst->id};
        queue_put_blocking(&mgr->complete, (char*) &b);

        // TODO: allow to #define away
        __sync_fetch_and_add(&mgr->stats.pool_released, 1);
    } else if (dst->id == -2) {
        // TODO: allow to #define away
        __sync_fetch_and_add(&mgr->stats.null_released, 1);
    }
    buffer_clear(dst);
}


Buffer buffer_create() {
    Buffer b;
    buffer_clear(&b);
    return b;
}

void buffer_clear(Buffer* b) {
    b->id = -1;
    b->ptr = 0;
    b->base = 0;
    b->remaining = 0;
}

bool buffer_is_full(Buffer* b) {
    return b->remaining == 0;
}

bool buffer_remaining(Buffer* b) {
    return b->remaining;
}

bool buffer_is_valid(Buffer* b) {
    return b->id >= 0;
}

void buffer_write(Buffer* b, size_t size, char** dst, size_t* dst_size) {
    if (b->remaining < size) size = b->remaining;

    *dst_size = size;
    *dst = b->ptr;

    b->remaining -= size;
    b->ptr += size;
}

// bool buffer_try_write_all(Buffer *b, char* buf, size_t buf_size) {
//     if (b->remaining < buf_size) return false;
//     printf("Buffer is: %s\n", buf);
//     printf("Size of buffer: %zu", buf_size);
//     memcpy((void*) b->ptr, (void*) buf, buf_size);
//     b->remaining -= buf_size;
//     b->ptr += buf_size;
//     return true;
// }

bool buffer_try_write_all(Buffer *b, char* buf, size_t buf_size) {
    if (b->remaining < buf_size+sizeof(size_t)) return false;
    printf("Buffer is: %s\n", buf);
    printf("Size of buffer: %d\n", buf_size);
    printf("Size of size(t): %d\n", sizeof(size_t));
    *(size_t *) b->ptr = buf_size; // might have to use htonl() and ntohl() if non-x86 boxes involved
    b->ptr += sizeof(size_t);
    memcpy((void*) b->ptr, (void*) buf, buf_size);
    b->remaining -= buf_size + sizeof(size_t);
    b->ptr += buf_size;
    return true;
}