#include "common.h"
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

#include <stdio.h>
#include <fcntl.h>
#include <sys/mman.h>
#include <assert.h>
#include <unistd.h>
#include <sys/stat.h>
#include <time.h>
#include <math.h>


char* get_shm_fname(char* dst1, char* dst2) {
    char* name = malloc(sizeof(char)*128);
    memset(name, 0, sizeof(char)*128);
    strcpy(name, "/dev/shm/");
    strcat(name, dst1);
    strcat(name, "__");
    strcat(name, dst2);
    return name;
}

void truncate_string(char* dst, const char* src, size_t max_size) {
    size_t src_len = strlen(src);
    if (src_len > max_size-1) {
        src_len = max_size-1;
    }
    memcpy(dst, src, src_len);
    dst[src_len] = '\0';
}

uint64_t nanos() {
    struct timespec t;
    clock_gettime(CLOCK_MONOTONIC_RAW, &t);
    uint64_t nanos = t.tv_sec * 1000000000UL + t.tv_nsec;
    return nanos;
}

uint64_t multiply_by(uint64_t v, float f) {
    if (f == 0) return 0;
    return v / (uint64_t) round(1.0/f);
}

static uint64_t g_seed;

// Compute a pseudorandom integer.
// Output value in range [0, 32767]
uint64_t rand_uint64(void) {
    g_seed = (214013*g_seed+2531011);
    return g_seed;
}

uint64_t slow_rand_uint64(void) {
  uint64_t r = 0;
  for (int i=0; i<64; i += 15 /*30*/) {
    r = r*((uint64_t)RAND_MAX + 1) + rand();
  }
  return r;
}