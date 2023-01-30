#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

#include "buffer.h"
#include "hindsight.h"
#include "agentapi.h"
#include "common.h"
#include <time.h>

#define PROCESS_NAME "hs_integration_test"

// TODO: configurable number of each thread.  Implement drainer in go. Compare

HindsightConfig config() {
	HindsightConfig conf = hindsight_default_config();
	conf.pool_capacity = 10000;
	conf.buffer_size = 4000;
	conf.breadcrumbs_capacity = conf.pool_capacity;
	conf.triggers_capacity = conf.pool_capacity;
	conf.address = malloc(32 * sizeof(char));
	return conf;
}

void reset_available_buffers(HindsightAgentAPI* api) {
	printf("Resetting buffers: draining any existing buffers...\n");

	size_t davailable = 0;
	while (true) {
		AvailableBuffers ab;

		hindsight_agentapi_get_available_nonblocking(api, &ab);

		if (ab.count == 0) {
			break;
		}

		davailable += ab.count;
	}

	size_t dcomplete = 0;
	while (true) {
		CompleteBuffers cb;

		hindsight_agentapi_get_complete_nonblocking(api, &cb);

		if (cb.count == 0) {
			break;
		}

		dcomplete += cb.count;
	}

	printf("Resetting buffers: drained %ld available and %ld complete.\n", davailable, dcomplete);
}

void make_all_buffers_available(HindsightAgentAPI* api) {
	printf("Initialize buffers: making %ld buffers available...\n", api->mgr.meta->capacity);

	int next_buffer_id = 0;
	size_t remaining = api->mgr.meta->capacity;
	while (remaining > 0) {
		AvailableBuffers av;
		av.count = 100;
		if (av.count > remaining) {
			av.count = remaining;
		}
		remaining -= av.count;

		for (int i = 0; i < av.count; i++) {
			av.bufs[i].buffer_id = next_buffer_id;
			next_buffer_id++;
		}

		hindsight_agentapi_put_available_blocking(api, &av);
	}

	printf("Initialize buffers: done\n");
	printf("Queue states:\n");
	printf("  Available ");
	queue_print(&api->mgr.available);
	printf("  Complete ");
	queue_print(&api->mgr.complete);	
}

HindsightAgentAPI* init_agentapi(const char* name) {
	HindsightAgentAPI* api = hindsight_agentapi_init(name);

	printf("Inited existing bufmanager %s\n", name);

	reset_available_buffers(api);

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

void breadcrumbs_agent(HindsightAgentAPI* api) {
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

		BreadcrumbBatch batch;
		hindsight_agentapi_get_breadcrumbs_nonblocking(api, &batch);

		count += batch.count;
		for (int i = 0; i < batch.count; i++) {
			printf("BC%d: %s\n", i, batch.breadcrumb_addrs[i]);
		}
	}
}

void breadcrumbs_client() {
	printf("Beginning client trace\n");
	hindsight_begin(700);


	printf("Beginning client loop\n");
	uint64_t last_print = nanos();
	uint64_t print_every = 1000000000UL;
	uint64_t count = 0;
	uint64_t v = 0;
	while (true) {
		uint64_t now = nanos();
		// printf("nanos %ld\n", now);
		if ((now - last_print) > print_every) {
			// Calculate throughputs
			uint64_t tput = (count * print_every) / (now - last_print);

			printf("Breadcrumbs tput %ld\n", tput);
			last_print = now;
			count = 0;
		}

		char addr[32];
		sprintf(addr, "Hello %ld", v++);

		hindsight_breadcrumb(addr);
		count ++;
	}	
}

void client() {
	hindsight_init_with_config(PROCESS_NAME, config());

	breadcrumbs_client();
}

void agent() {
	HindsightAgentAPI* api = init_agentapi(PROCESS_NAME);

	make_all_buffers_available(api);

	breadcrumbs_agent(api);
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