#include <stdio.h>
#include <stdlib.h>
#include <pthread.h>
#include <time.h>
#include "assert.h"

#include "buffer.h"
#include "tracestate.h"
#include <pthread.h>

void test_buffer_simple() {
	Buffer b = buffer_create();
	assert(b.id == -1);
	assert(b.remaining == 0);
	assert(b.ptr == 0);

	printf("test_buffer_simple passed\n");
}

void test_buffer_write() {
	int buf_id = 5;
	size_t buf_size = 21;
	char buf[buf_size];

	Buffer b = buffer_create();
	b.id = buf_id;
	b.ptr = buf;
	b.remaining = buf_size;

	char* dst;
	size_t dst_size;
	buffer_write(&b, 5, &dst, &dst_size);

	assert(dst == buf);
	assert(dst_size == 5);
	assert(b.id == buf_id);
	assert(b.remaining == 16);
	assert(b.ptr == (buf + 5));
	assert(!buffer_is_full(&b));

	buffer_write(&b, 5, &dst, &dst_size);

	assert(dst == (buf + 5));
	assert(dst_size == 5);
	assert(b.id == buf_id);
	assert(b.remaining == 11);
	assert(b.ptr == (buf + 10));
	assert(!buffer_is_full(&b));

	buffer_write(&b, 5, &dst, &dst_size);

	assert(dst == (buf + 10));
	assert(dst_size == 5);
	assert(b.id == buf_id);
	assert(b.remaining == 6);
	assert(b.ptr == (buf + 15));
	assert(!buffer_is_full(&b));

	buffer_write(&b, 5, &dst, &dst_size);

	assert(dst == (buf + 15));
	assert(dst_size == 5);
	assert(b.id == buf_id);
	assert(b.remaining == 1);
	assert(b.ptr == (buf + 20));
	assert(!buffer_is_full(&b));

	buffer_write(&b, 5, &dst, &dst_size);

	assert(dst == (buf + 20));
	assert(dst_size == 1);
	assert(b.id == buf_id);
	assert(b.remaining == 0);
	assert(b.ptr == (buf + 21));
	assert(buffer_is_full(&b));

	buffer_write(&b, 5, &dst, &dst_size);

	assert(dst == (buf + 21));
	assert(dst_size == 0);
	assert(b.id == buf_id);
	assert(b.remaining == 0);
	assert(b.ptr == (buf + 21));
	assert(buffer_is_full(&b));

	printf("test_buffer_write passed\n");	
}

void put_available(Queue q, int bufid) {
	AvailableBuffer av = {bufid};
	queue_put_blocking(&q, (char*) &av);
}

void test_bufmanager() {
	size_t capacity = 10;
	BufManager mgr = bufmanager_init("test_bufmanager", capacity, 100);

	for (int i = 0; i < capacity; i++) {
		Buffer buf = buffer_create();
		bufmanager_acquire(&mgr, &buf);
		assert(buf.id == i);
	}

	put_available(mgr.available, 7);

	Buffer buf = buffer_create();
	bufmanager_acquire(&mgr, &buf);
	assert(buf.id == 7);
	buffer_clear(&buf);


	put_available(mgr.available, 11);
	bufmanager_acquire(&mgr, &buf);
	assert(buf.id == 11);
	buffer_clear(&buf);

	for (unsigned i = 0; i < 100; i++) {
		put_available(mgr.available, i);
		bufmanager_acquire(&mgr, &buf);
		assert(buf.id == i);
		buffer_clear(&buf);
	}

	for (unsigned i = 0; i < 100; i+=2) {
		put_available(mgr.available, i);
		put_available(mgr.available, i+1);
		bufmanager_acquire(&mgr, &buf);
		assert(buf.id == i);
		buffer_clear(&buf);
		bufmanager_acquire(&mgr, &buf);
		assert(buf.id == i+1);
		buffer_clear(&buf);
	}

	for (unsigned i = 0; i < 100; i+=2) {
		put_available(mgr.available, i+1);
		put_available(mgr.available, i);
		bufmanager_acquire(&mgr, &buf);
		assert(buf.id == i+1);
		buffer_clear(&buf);
		bufmanager_acquire(&mgr, &buf);
		assert(buf.id == i);
		buffer_clear(&buf);
	}

	printf("test_bufmanager passed\n");	
}

char* make_data(size_t size) {
	int8_t* data = malloc(size);
	for (unsigned i = 0; i < size; i++) {
		data[i] = i;
	}
	return (char*) data;
}

void test_tracestate() {
	int buffer_count = 10;
	size_t buffer_size = 100;
	BufManager mgr = bufmanager_init("test_tracestate", buffer_count, buffer_size);

	TraceState trace = tracestate_create();

	tracestate_begin(&trace, &mgr, 3000);
	printf("Traceheader size %ld\n", sizeof(TraceHeader));
	assert(sizeof(TraceHeader) == 32);
	assert(trace.buffer.id == 0);
	assert(trace.buffer.remaining == (buffer_size - sizeof(TraceHeader)));

	char* data = make_data(50);

	tracestate_write(&trace, &mgr, data, 50);
	assert(trace.buffer.id == 0);
	assert(trace.buffer.remaining == (buffer_size - sizeof(TraceHeader) - 50));
	for (int i = 0; i < 50; i++) {
		assert(mgr.pool[sizeof(TraceHeader)+i] == i);
	}

	char* data2 = make_data(60);
	tracestate_write(&trace, &mgr, data2, 60);
	assert(trace.buffer.id == 1);
	assert(trace.buffer.remaining == (2 * buffer_size - 2 * sizeof(TraceHeader) - 50 - 60));
	for (int i = 0; i < 18; i++) {
		int base = sizeof(TraceHeader) + 50;
		assert(mgr.pool[base+i] == i);
	}
	assert(mgr.pool[100] != 26);
	for (int i = 18; i < 60; i++) {
		int base = 2 * sizeof(TraceHeader) + 50;
		assert(mgr.pool[base+i] == i);
	}
	
	printf("test_tracestate passed\n");	
}

void test_tracestate_nullbuffer() {
	int buffer_count = 10;
	size_t buffer_size = 100;
	BufManager mgr = bufmanager_init("test_tracestate_nullbuffer", buffer_count, buffer_size);
	TraceState trace = tracestate_create();

	for (int i = 0; i < buffer_count; i++) {
		Buffer buf = buffer_create();
		bufmanager_acquire(&mgr, &buf);
		assert(buf.id == i);
	}

	for (unsigned i = 0; i < 2; i++) {
		put_available(mgr.available, i);
	}

	tracestate_begin(&trace, &mgr, 3000);
	assert(sizeof(TraceHeader) == 32);
	assert(trace.buffer.id == 0);

	char* data = make_data(50);

	tracestate_write(&trace, &mgr, data, 50);
	assert(trace.buffer.id == 0);

	char* data2 = make_data(60);
	tracestate_write(&trace, &mgr, data2, 60);
	assert(trace.buffer.id == 1);
	

	char* data3 = make_data(100);
	tracestate_write(&trace, &mgr, data3, 82);
	assert(trace.buffer.id == -2);

	tracestate_write(&trace, &mgr, data3, 82);
	assert(trace.buffer.id == -2);

	put_available(mgr.available, 7);

	tracestate_write(&trace, &mgr, data3, 82);
	assert(trace.buffer.id == 7);

	tracestate_write(&trace, &mgr, data3, 82);
	assert(trace.buffer.id == -2);

	CompleteBuffer cb;
	assert(queue_get_nonblocking(&mgr.complete, &cb));
	assert(cb.trace_id == 3000);
	assert(cb.buffer_id == 0);
	assert(queue_get_nonblocking(&mgr.complete, &cb));
	assert(cb.trace_id == 3000);
	assert(cb.buffer_id == 1);
	assert(queue_get_nonblocking(&mgr.complete, &cb));
	assert(cb.trace_id == 3000);
	assert(cb.buffer_id == 7);
	assert(!queue_get_nonblocking(&mgr.complete, &cb));

	
	printf("test_tracestate_nullbuffer passed\n");	
}

int main(int argc, char const *argv[])
{
	printf("Testing buffer!\n");
	test_buffer_simple();
	test_buffer_write();
	test_bufmanager();
	test_tracestate();
	test_tracestate_nullbuffer();
	return 0;
}