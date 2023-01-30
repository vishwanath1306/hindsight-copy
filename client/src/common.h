#ifndef _HINDSIGHT_CLIENT_COMMON_H_
#define _HINDSIGHT_CLIENT_COMMON_H_

#include <stdint.h>
#include <stdio.h>

char* get_shm_fname(char* dst1, char* dst2);

void truncate_string(char* dst, const char* src, size_t max_size);

uint64_t multiply_by(uint64_t v, float f);

uint64_t nanos();

uint64_t rand_uint64();

#endif // _HINDSIGHT_CLIENT_COMMON_H_