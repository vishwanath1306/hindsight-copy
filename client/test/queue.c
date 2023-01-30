#include <stdio.h>
#include <stdlib.h>
#include <pthread.h>
#include <time.h>
#include "assert.h"

#include "buffer.h"
#include "tracestate.h"
#include "queue.h"
#include <pthread.h>


void test_queue_simple() {
	Queue q = queue_init("/dev/shm/test_queue_simple", sizeof(int), 10);

	int rv;
	assert(!queue_get_nonblocking(&q, (char*) &rv));
	assert(!queue_get_nonblocking(&q, (char*) &rv));
	assert(!queue_get_nonblocking(&q, (char*) &rv));

	int x = 99;
	assert(queue_put_nonblocking(&q, (char*) &x));
	assert(queue_get_nonblocking(&q, (char*) &rv));
	assert(rv == x);

	printf("test_queue_simple passedd\n");
}

void test_queue_nonblocking() {
	Queue q = queue_init("/dev/shm/test_queue_nonblocking", sizeof(int), 10);

	for (int i = 0; i < 10; i++) {
		int y = i + 7;
		assert(queue_put_nonblocking(&q, (char*) &y));
		assert(y == i+7);
	}
	int nonex = -300;
	assert(!queue_put_nonblocking(&q, (char*) &nonex));
	assert(!queue_put_nonblocking(&q, (char*) &nonex));
	assert(!queue_put_nonblocking(&q, (char*) &nonex));
	assert(nonex == -300);
	for (int i = 0; i < 10; i++) {
		int y = -1;
		assert(queue_get_nonblocking(&q, (char*) &y));
		assert(y == i + 7);
	}
	int noney = -75;
	assert(!queue_get_nonblocking(&q, (char*) &noney));
	assert(noney == -75);
	for (int i = 0; i < 10; i++) {
		int y = i + 7;
		assert(queue_put_nonblocking(&q, (char*) &y));
		assert(y == i+7);
	}
	assert(!queue_put_nonblocking(&q, (char*) &nonex));
	assert(!queue_put_nonblocking(&q, (char*) &nonex));
	assert(!queue_put_nonblocking(&q, (char*) &nonex));
	assert(nonex == -300);
	for (int i = 0; i < 10; i++) {
		int y = -1;
		assert(queue_get_nonblocking(&q, (char*) &y));
		assert(y == i + 7);
	}

	for (int i = 0; i < 3; i++) {
		int y = i + 7;
		assert(queue_put_nonblocking(&q, (char*) &y));
		assert(y == i+7);
	}
	for (int i = 0; i < 3; i++) {
		int y = -1;
		assert(queue_get_nonblocking(&q, (char*) &y));
		assert(y == i + 7);
	}

	for (int i = 0; i < 3; i++) {
		int y = i + 100000;
		assert(queue_put_nonblocking(&q, (char*) &y));
		assert(y == i + 100000);
	}
	for (int i = 0; i < 3; i++) {
		int y = -1;
		assert(queue_get_nonblocking(&q, (char*) &y));
		assert(y == i + 100000);
	}

	for (int i = 0; i < 3; i++) {
		int y = i - 1000;
		assert(queue_put_nonblocking(&q, (char*) &y));
		assert(y == i -1000);
	}
	for (int i = 0; i < 3; i++) {
		int y = -1;
		assert(queue_get_nonblocking(&q, (char*) &y));
		assert(y == i- 1000);
	}

	int next_putv = 3;
	int next_getv = 3;

	for (int i = 0; i < 3; i++) {
		assert(queue_put_nonblocking(&q, (char*) &next_putv));
		next_putv++;
	}

	for (int i = 0; i < 10; i++) {
		for (int j = 0; j < 5; j++) {
			assert(queue_put_nonblocking(&q, (char*) &next_putv));
			next_putv++;			
		}

		for (int j = 0; j < 5; j++) {
			int y = -1;
			assert(queue_get_nonblocking(&q, (char*) &y));
			assert(y == next_getv);
			next_getv++;
		}
	}

	for (int i = 0; i < 3; i++) {
		int rv = -1;
		assert(queue_get_nonblocking(&q, (char*) &rv));
		assert(rv == next_getv);
		next_getv++;
	}
	assert(!queue_get_nonblocking(&q, (char*) &noney));
	assert(noney == -75);

	printf("test_queue_nonblocking passed\n");
}

void test_queue_nonblocking_multi() {
	Queue q = queue_init("/dev/shm/test_queue_nonblocking_multi", sizeof(int), 10);

	{
		size_t max_elements = 10;
		int nones[max_elements];
		assert(queue_get_nonblocking_multi(&q, nones, max_elements) == 0);
	}

	for (int i = 0; i < 10; i++) {
		int y = i + 7;
		assert(queue_put_nonblocking(&q, (char*) &y));
		assert(y == i+7);
	}

	{
		size_t max_elements = 10;
		int ys[max_elements];
		size_t dequeued = queue_get_nonblocking_multi(&q, ys, max_elements);
		assert(dequeued == max_elements);
	}

	{
		size_t max_elements = 10;
		int nones[max_elements];
		assert(queue_get_nonblocking_multi(&q, nones, max_elements) == 0);
	}

	for (int i = 0; i < 5; i++) {
		int y = i + 7;
		assert(queue_put_nonblocking(&q, (char*) &y));
		assert(y == i+7);
	}

	{
		size_t max_elements = 10;
		int ys[max_elements];
		size_t dequeued = queue_get_nonblocking_multi(&q, ys, max_elements);
		assert(dequeued == 5);
	}

	for (int i = 0; i < 10; i++) {
		int y = i + 7;
		assert(queue_put_nonblocking(&q, (char*) &y));
		assert(y == i+7);
	}

	{
		size_t max_elements = 5;
		int ys[max_elements];
		size_t dequeued = queue_get_nonblocking_multi(&q, ys, max_elements);
		assert(dequeued == max_elements);
	}

	{
		size_t max_elements = 3;
		int ys[max_elements];
		size_t dequeued = queue_get_nonblocking_multi(&q, ys, max_elements);
		assert(dequeued == max_elements);
	}

	{
		size_t max_elements = 1;
		int ys[max_elements];
		size_t dequeued = queue_get_nonblocking_multi(&q, ys, max_elements);
		assert(dequeued == max_elements);
	}

	{
		size_t max_elements = 10;
		int ys[max_elements];
		size_t dequeued = queue_get_nonblocking_multi(&q, ys, max_elements);
		assert(dequeued == 1);
	}
	
	printf("test_queue_nonblocking_multi passed\n");
}

void test_queue_put_nonblocking_multi() {
	Queue q = queue_init("/dev/shm/test_queue_nonblocking_multi", sizeof(int), 10);

	{
		size_t max_elements = 10;
		int nones[max_elements];
		assert(queue_get_nonblocking_multi(&q, nones, max_elements) == 0);
	}

	{
		size_t num_writes = 10;
		int ys[num_writes];
		for (int i = 0; i < num_writes; i++) {
			ys[i] = i + 7;
		}
		size_t num_written = queue_put_nonblocking_multi(&q, ys, num_writes);
		assert(num_written == num_writes);
	}

	{
		size_t max_elements = 10;
		int ys[max_elements];
		size_t dequeued = queue_get_nonblocking_multi(&q, ys, max_elements);
		assert(dequeued == max_elements);

		for (int i = 0; i < max_elements; i++) {
			assert(ys[i] == (i+7));
		}
	}

	{
		size_t max_elements = 10;
		int nones[max_elements];
		assert(queue_get_nonblocking_multi(&q, nones, max_elements) == 0);
	}

	{
		size_t num_writes = 5;
		int ys[num_writes];
		for (int i = 0; i < num_writes; i++) {
			ys[i] = i + 50;
		}
		size_t num_written = queue_put_nonblocking_multi(&q, ys, num_writes);
		assert(num_written == num_writes);
	}

	{
		size_t max_elements = 10;
		int ys[max_elements];
		size_t dequeued = queue_get_nonblocking_multi(&q, ys, max_elements);
		assert(dequeued == 5);

		for (int i = 0; i < 5; i++) {
			assert(ys[i] == (i+50));
		}
	}

	{
		size_t num_writes = 5;
		int ys[num_writes];
		for (int i = 0; i < num_writes; i++) {
			ys[i] = i + 50;
		}
		size_t num_written = queue_put_nonblocking_multi(&q, ys, num_writes);
		assert(num_written == num_writes);
	}

	{
		size_t num_writes = 10;
		int ys[num_writes];
		for (int i = 0; i < num_writes; i++) {
			ys[i] = i + 50;
		}
		size_t num_written = queue_put_nonblocking_multi(&q, ys, num_writes);
		assert(num_written == 5);
	}

	{
		size_t max_elements = 20;
		int ys[max_elements];
		size_t dequeued = queue_get_nonblocking_multi(&q, ys, max_elements);
		assert(dequeued == 10);

		for (int i = 0; i < 5; i++) {
			assert(ys[i] == (i+50));
		}
		for (int i = 5; i < 10; i++) {
			assert(ys[i] == (i+45));
		}
	}

	printf("test_queue_put_nonblocking_multi passed\n");
}

void test_queue_blocking() {
	Queue q = queue_init("/dev/shm/test_queue_blocking", sizeof(int), 10);

	for (int i = 0; i < 10; i++) {
		int y = i + 7;
		queue_put_blocking(&q, (char*) &y);
		assert(y == i+7);
	}
	int nonex = -300;
	assert(!queue_put_nonblocking(&q, (char*) &nonex));
	assert(nonex == -300);

	for (int i = 0; i < 10; i++) {
		int y = i + 7;
		queue_get_blocking(&q, (char*) &y);
		assert(y == i+7);
	}
	assert(!queue_get_nonblocking(&q, (char*) &nonex));

	printf("test_queue_blocking passed\n");
}

void test_tiny_queue() {
	Queue q = queue_init("/dev/shm/test_tiny_queue", sizeof(int), 1);

	for (int i = 0; i < 1; i++) {
		int y = i + 7;
		assert(queue_put_nonblocking(&q, (char*) &y));
		assert(y == i+7);
	}
	int nonex = -300;
	assert(!queue_put_nonblocking(&q, (char*) &nonex));
	assert(!queue_put_nonblocking(&q, (char*) &nonex));
	assert(!queue_put_nonblocking(&q, (char*) &nonex));
	assert(nonex == -300);
	for (int i = 0; i < 1; i++) {
		int y = -1;
		assert(queue_get_nonblocking(&q, (char*) &y));
		assert(y == i + 7);
	}
	int noney = -75;
	assert(!queue_get_nonblocking(&q, (char*) &noney));
	assert(noney == -75);
	for (int i = 0; i < 1; i++) {
		int y = i + 7;
		assert(queue_put_nonblocking(&q, (char*) &y));
		assert(y == i+7);
	}
	assert(!queue_put_nonblocking(&q, (char*) &nonex));
	assert(!queue_put_nonblocking(&q, (char*) &nonex));
	assert(!queue_put_nonblocking(&q, (char*) &nonex));
	assert(nonex == -300);
	assert(nonex == -300);
	for (int i = 0; i < 1; i++) {
		int y = -1;
		assert(queue_get_nonblocking(&q, (char*) &y));
		assert(y == i + 7);
	}
	assert(!queue_get_nonblocking(&q, (char*) &noney));
	assert(noney == -75);


	printf("test_tiny_queue passed\n");	
}

typedef struct StructForQueueTest {
	int a;
	int b;
	int c;
	int64_t d;
	int ab;
	int bb;
	int cb;
	int64_t db;
	int64_t dba;
	int64_t dbs;
	int64_t dbd;
	int64_t dbf;
} StructForQueueTest;

void test_queue_struct() {
	Queue q = queue_init("/dev/shm/test_queue_struct", sizeof(StructForQueueTest), 10);

	for (int i = 0; i < 10; i++) {
		StructForQueueTest v = {i*10, i*20, i*30, i*40};
		assert(queue_put_nonblocking(&q, (char*) &v));
	}
	StructForQueueTest none = {-1, -2, -3, -4};
	assert(!queue_put_nonblocking(&q, (char*) &none));

	for (int i = 0; i < 10; i++) {
		StructForQueueTest v = {0,0,0,0};
		assert(queue_get_nonblocking(&q, (char*) &v));
		assert(v.a == i*10);
		assert(v.b == i*20);
		assert(v.c == i*30);
		assert(v.d == i*40);
	}
	assert(!queue_get_nonblocking(&q, (char*) &none));


	printf("test_queue_struct passed\n");		
}

typedef struct TestArgs {
	int thread_num;
	int enqueue_count;
	Queue* queue;
	pthread_barrier_t* barrier;
} TestArgs;

void queue_blocking_multithread_thread(void* arg) {
	TestArgs* args = (TestArgs*) arg;

	pthread_barrier_wait(args->barrier);
	printf("  test_queue_blocking_multithread: thread %d start\n", args->thread_num);

	for(int i = 0; i < args->enqueue_count; i++) {
		int v = args->thread_num * args->enqueue_count + i;
		queue_put_blocking(args->queue, (char*) &v);
		usleep(1000);
	}
	printf("  test_queue_blocking_multithread: thread %d exit\n", args->thread_num);
}

void test_queue_blocking_multithread() {
	Queue q = queue_init("/dev/shm/test_queue_blocking_multithread", sizeof(int), 10);

	int num_threads = 10;
	int enqueue_count = 1000;

	pthread_barrier_t barrier;
	pthread_barrier_init(&barrier, NULL, num_threads);

	pthread_t threads[num_threads];
	for (int i=0; i<num_threads; i++) {
		TestArgs* args = (TestArgs*) malloc(sizeof(TestArgs));
		args->thread_num = i;
		args->queue = &q;
		args->barrier = &barrier;
		args->enqueue_count = enqueue_count;
		pthread_create(&threads[i], NULL, &queue_blocking_multithread_thread, args);
	}

	size_t total = num_threads * enqueue_count;
	bool seen[total];
	for (int i = 0; i < total; i++) {
		seen[total] = false;
	}

	for (int i = 0; i < enqueue_count; i++) {
		for (int j = 0; j < num_threads; j++) {
			int v = -1;
			queue_get_blocking(&q, &v);
			seen[v] = true;
		}
	}

	for (int i = 0; i < total; i++) {
		assert(seen[i]);
	}

	for (int i = 0; i < num_threads; i++) {
		pthread_join(threads[i], NULL);		
	}

	printf("test_queue_blocking_multithread passed\n");
}

int main(int argc, char const *argv[])
{
	printf("Testing queue implementation\n");
	test_queue_simple();
	test_queue_nonblocking();
	test_queue_blocking();
	test_tiny_queue();
	test_queue_struct();
	test_queue_blocking_multithread();
	test_queue_nonblocking_multi();
	test_queue_put_nonblocking_multi();
	return 0;
}