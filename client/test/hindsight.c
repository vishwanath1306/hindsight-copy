#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

#include "buffer.h"
#include "hindsight.h"
#include "agentapi.h"
#include "common.h"
#include <time.h>

#define PROCESS_NAME "hs_integration_test"
#define BUFFERSIZE 128000
#define TRACEPOINTSPERTRACE 3000
#define WRITESIZE 1000
#define CHECKEVERY 3000

// TODO: configurable number of each thread.  Implement drainer in go. Compare

HindsightConfig config() {
	HindsightConfig conf = hindsight_default_config();
	conf.pool_capacity = 10000;
	conf.buffer_size = BUFFERSIZE;
	conf.breadcrumbs_capacity = conf.pool_capacity;
	conf.triggers_capacity = conf.pool_capacity;
	conf.address = malloc(32 * sizeof(char));
	return conf;
}

HindsightAgentAPI* init_agentapi(const char* name) {
	HindsightAgentAPI* api = hindsight_agentapi_init(name);

	printf("Inited existing bufmanager %s\n", name);

	// reset_available_buffers(api);

	return api;
}

CompleteBuffers await_complete(HindsightAgentAPI* api) {
	CompleteBuffers complete;

	int max_backoff = 100000; // 100ms
	int backoff = 10;

	while (true) {
		hindsight_agentapi_get_complete_nonblocking(api, &complete);
		if (complete.count > 0) {
			return complete;
		}
		usleep(backoff);
		backoff *= 2;
		if (backoff > max_backoff) {
			backoff = max_backoff;
		}
	}
}

void drain_forever_agent(HindsightAgentAPI* api) {
	uint64_t last_print = nanos();
	uint64_t print_every = 1000000000UL;
	uint64_t count = 0;
	while (true) {
		uint64_t now = nanos();
		if ((now - last_print) > print_every) {
			uint64_t tput = (count * print_every) / (now - last_print);
			printf("Throughput: %ld\n", tput);
			last_print = now;
			count = 0;
		}

		// printf("Awaiting complete\n");
		CompleteBuffers complete = await_complete(api);
		count += complete.count;

		AvailableBuffers available;
		available.count = complete.count;
		for (int i = 0; i < complete.count; i++) {
			available.bufs[i].buffer_id = complete.bufs[i].buffer_id;
		}

		hindsight_agentapi_put_available_blocking(api, &available);
	}
}

void drain_forever_client() {
	printf("Beginning client trace\n");

	size_t tracepoints_per_trace = TRACEPOINTSPERTRACE;
	size_t write_size = WRITESIZE;

	printf("Beginning client loop\n");
	uint64_t last_print = nanos();
	uint64_t print_every = 1000000000UL;
	uint64_t count = 0;
	BufferStats stats = {0,0,0,0};
	uint64_t trace_id = 700;
	int check_every = CHECKEVERY;

	size_t total_buf_size = write_size * check_every;
	char* buf = (char*) malloc(total_buf_size);
	for (int i = 0; i < total_buf_size/4; i++) {
		((int*) buf)[i] = rand();
	}

	while (true) {
		hindsight_begin(++trace_id);
		for (int i = 0; i < tracepoints_per_trace; i+=check_every) {
			uint64_t now = nanos();
			// printf("nanos %ld\n", now);
			if ((now - last_print) > print_every) {
				BufferStats current = hindsight.mgr.stats;
				BufferStats delta = {
					current.pool_acquired - stats.pool_acquired,
					current.null_acquired - stats.null_acquired,
					current.pool_released - stats.pool_released,
					current.null_released - stats.null_released
				};

				// Calculate throughputs
				uint64_t tput = (count * print_every) / (now - last_print);
				delta.pool_acquired = (delta.pool_acquired * print_every) / (now - last_print);
				delta.null_acquired = (delta.null_acquired * print_every) / (now - last_print);
				delta.pool_released = (delta.pool_released * print_every) / (now - last_print);
				delta.null_released = (delta.null_released * print_every) / (now - last_print);

				printf("Tracepoints %ld - Pool: %ld %ld - NULL %ld %ld\n", tput, 
					delta.pool_acquired, delta.pool_released, 
					delta.null_acquired, delta.null_released);
				last_print = now;
				count = 0;
				stats = current;
			}

			for (size_t j = 0; j < total_buf_size; j += write_size) {
				hindsight_tracepoint(buf + j, write_size);
			}
			count += check_every;
		}
		if ((trace_id % 100000) == 0) {
			hindsight_trigger(0);
		}
		hindsight_breadcrumb("Hello World!");
		hindsight_end();
	}	
}

void client() {
	hindsight_init_with_config(PROCESS_NAME, config());

	drain_forever_client();
}

void agent() {
	HindsightAgentAPI* api = init_agentapi(PROCESS_NAME);

	drain_forever_agent(api);
}


int main(int argc, char const *argv[])
{
	// TODO: capacity as argument
	// TODO: buffer size as argument
	// TODO: flag for agent/client

	if (argc <= 1) {
		printf("Running as agent\n");
		agent();
	} else {
		printf("Running as client\n");
		client();
	}



	return 0;
}