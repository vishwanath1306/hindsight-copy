
#define _GNU_SOURCE
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <argp.h>
#include <pthread.h>

#include "buffer.h"
#include "hindsight.h"
#include "agentapi.h"
#include "common.h"
#include <time.h>
#include <sys/sysinfo.h>

#include <sched.h>

const char *argp_program_version = "argp-ex3 1.0";
const char *argp_program_bug_address = "<bug-gnu-utils@gnu.org>";
static char doc[] = "Simple hindsight benchmarking program.  PROCNAME is required for hindsight_init";
static char args_doc[] = "PROCNAME";

static struct argp_option options[] = {
  {"threads",  't', "NUM",  0,  "Number of benchmark threads to run" },
  {"buffer_size",  's', "NUM",  0,  "Buffer size in bytes" },
  {"buffer_count",  'c', "NUM",  0,  "Number of buffers in pool" },
  {"payload_size",  'w', "NUM",  0,  "Payload size written by each tracepoint" },
  {"tracepoints", 'n', "NUM", 0, "Number of tracepoints per trace"},
  {"trigger", 'p', "NUM", 0, "Trigger probability, default 0, float"},
  {"headsampling", 'H', "NUM", 0, "Head-based sampling probability between 0 and 1, default 0, float"},
  {"retroactive", 'R', "NUM", 0, "Retroactive sampling percentage between 0 and 1, default 1, float"},
  {"duration", 'd', "NUM", 0, "Duration in seconds before exiting. 0 to run forever"},
  {"output",   'o', "FILE", 0, "Output stats to FILE" },
  { 0 }
};

static inline uint64_t ticksbegin(void) {
    uint32_t lo, hi;
    asm volatile("lfence;rdtsc" : "=a" (lo), "=d" (hi));
    return (uint64_t) hi << 32 | lo;
}

static inline uint64_t ticksend(void) {
    uint32_t lo, hi;
    asm volatile("rdtscp;lfence" : "=a" (lo), "=d" (hi));
    return (uint64_t) hi << 32 | lo;
}

struct arguments {
  int num_threads;
  size_t buffer_size;
  size_t buffer_count;
  size_t payload_size;
  int tracepoints_per_request;
  bool trigger_enabled;
  float trigger_probability;
  float head_sampling_probability;
  float retroactive_sampling_percentage;
  uint64_t duration;
  char* output_file;
  char* process_name;
};

static error_t parse_opt (int key, char *arg, struct argp_state *state) {
  struct arguments *arguments = state->input;

  switch (key)
    {
    case 't':
      arguments->num_threads = atoi(arg);
      break;
    case 's':
      arguments->buffer_size = atoll(arg);
      break;
    case 'c':
      arguments->buffer_count = atoll(arg);
      break;
    case 'w':
      arguments->payload_size = atoll(arg);
      break;
    case 'n':
      arguments->tracepoints_per_request = atoi(arg);
      break;
    case 'p':
      arguments->trigger_enabled = true;
      arguments->trigger_probability = atof(arg);
      break;
    case 'H':
      arguments->head_sampling_probability = atof(arg);
      break;
    case 'R':
      arguments->retroactive_sampling_percentage = atof(arg);
      break;
    case 'd':
      arguments->duration = atoll(arg);
      break;
    case 'o':
      arguments->output_file = arg;
      break;

    case ARGP_KEY_ARG:
      if (state->arg_num >= 1)
        /* Too many arguments. */
        argp_usage (state);

      arguments->process_name = arg;

      break;

    case ARGP_KEY_END:
      if (state->arg_num < 1)
        /* Not enough arguments. */
        argp_usage (state);
      break;

    default:
      return ARGP_ERR_UNKNOWN;
    }
  return 0;
}

static struct argp argp = { options, parse_opt, args_doc, doc };

void init_hindsight_client(struct arguments *arguments) {
    HindsightConfig conf = hindsight_default_config();
    conf.pool_capacity = arguments->buffer_count;
    conf.buffer_size = arguments->buffer_size;
    conf.breadcrumbs_capacity = conf.pool_capacity;
    conf.triggers_capacity = conf.pool_capacity;
    conf.address = malloc(32 * sizeof(char));
    conf.retroactive_sampling_percentage = arguments->retroactive_sampling_percentage;
    conf.head_sampling_probability = arguments->head_sampling_probability;
    conf._retroactive_sampling_threshold = multiply_by(UINT64_MAX, conf.retroactive_sampling_percentage);
    conf._head_sampling_threshold = multiply_by(UINT64_MAX, conf.head_sampling_probability);

    hindsight_init_with_config(arguments->process_name, conf);
}

typedef struct exp_stats {
    uint64_t count;
    uint64_t traces;
    uint64_t invalid_traces;
    uint64_t begins;
    uint64_t tracepoints;
    uint64_t ends;
} exp_stats;

void set_cores(int* cores, size_t cores_size) {
  if (cores_size == 0) {
    printf("Trying to bind to empty core set\n");
  }
  cpu_set_t cpuset;
  CPU_ZERO(&cpuset);
  for (int i = 0; i < cores_size; i++) {
    CPU_SET(cores[i], &cpuset);
  }
  int rc = pthread_setaffinity_np(pthread_self(), sizeof(cpu_set_t), &cpuset);
  if (rc != 0) {
    printf("Error calling pthread_setaffinity_np: %d\n", rc);
  }
}

void client_thread_main(volatile int *alive, 
        int client_id, struct arguments *arguments, exp_stats* stats) {

    // Bind to core
    int cores[1];
    cores[0] = client_id % get_nprocs();
    set_cores(cores, 1);

    printf("Client %d started on core %d\n", client_id, cores[0]);

    size_t payload_src_size = arguments->payload_size;
    char payload[payload_src_size];
    int* payload_ints = (int*) payload;
    for (int i = 0; i < payload_src_size/4; i++) {
      payload_ints[i] = rand();
    }

    int tracepoints_per_request = arguments->tracepoints_per_request;
    uint64_t ts[6];

    int traces = 0;
    int invalid_traces = 0;
    int batchsize = 100;
    uint64_t count = 0;
    uint64_t sum_begins = 0;
    uint64_t sum_tracepoints = 0;
    uint64_t sum_ends = 0;

    uint64_t begin = nanos();
    uint64_t tbegin = ticksbegin();
    while (*alive) {
        uint64_t trace_id = rand_uint64();
        ts[0] = ticksbegin();
        hindsight_begin(trace_id);
        ts[1] = ticksend();
        ts[2] = ticksbegin();
        for (int i = 0; i < tracepoints_per_request; i++) {
            hindsight_tracepoint(payload, payload_src_size);
        }
        ts[3] = ticksend();
        ts[4] = ticksbegin();
        // usleep(50);
        hindsight_end();
        ts[5] = ticksend();
        uint64_t end = nanos();

        uint64_t duration = (end - begin);
        uint64_t ts_duration = ts[5] - tbegin;

        traces++;
        bool is_valid = (hindsight_null_buffer_count() == 0);
        if (!is_valid)
            invalid_traces++;
        count += tracepoints_per_request;
        sum_begins += (duration * (ts[1]-ts[0])) / ts_duration;
        sum_tracepoints += (duration * (ts[3]-ts[2])) / ts_duration;
        sum_ends += (duration * (ts[5]-ts[4])) / ts_duration;

        if (traces == batchsize) {
            __sync_fetch_and_add(&stats->count, count);
            __sync_fetch_and_add(&stats->traces, traces);
            __sync_fetch_and_add(&stats->invalid_traces, invalid_traces);
            __sync_fetch_and_add(&stats->begins, sum_begins);
            __sync_fetch_and_add(&stats->tracepoints, sum_tracepoints);
            __sync_fetch_and_add(&stats->ends, sum_ends);

            traces = 0;
            invalid_traces = 0;
            count = 0;
            sum_begins = 0;
            sum_tracepoints = 0;
            sum_ends = 0;
        }
    }
    printf("Client ended\n");
}

void print_thread_main(volatile int *alive, struct arguments *args, exp_stats* stats) {
    printf("Print thread beginning\n");
    uint64_t begin = nanos();
    uint64_t last_print = begin;
    uint64_t print_every = 1000000000UL;
    BufferStats prev_stats = {0,0,0,0};
    uint64_t prev_count = 0;
    uint64_t prev_traces_count = 0;
    uint64_t prev_invalid_traces = 0;
    uint64_t prev_begins = 0;
    uint64_t prev_tracepoints = 0;
    uint64_t prev_ends = 0;

    printf("headers:\tt\tduration\ttraces\tinvalidtraces\ttracepoints\ttracepoints_tput\tbytes\tpool_acquired\tpool_released\tnull_acquired\tnull_released\tbegin\ttracepoint\tend\n");

    while (*alive) {
        uint64_t now = nanos();
        if ((now - last_print) > print_every) {
            BufferStats new_stats = hindsight.mgr.stats;
            BufferStats delta = {
                new_stats.pool_acquired - prev_stats.pool_acquired,
                new_stats.null_acquired - prev_stats.null_acquired,
                new_stats.pool_released - prev_stats.pool_released,
                new_stats.null_released - prev_stats.null_released
            };

            uint64_t new_count = stats->count;
            uint64_t count = new_count - prev_count;

            uint64_t new_traces_count = stats->traces;
            uint64_t traces = new_traces_count - prev_traces_count;

            uint64_t new_invalid_traces = stats->invalid_traces;
            uint64_t invalid_traces = new_invalid_traces - prev_invalid_traces;

            uint64_t new_begins = stats->begins;
            uint64_t begins = new_begins - prev_begins;

            uint64_t new_tracepoints = stats->tracepoints;
            uint64_t tracepoints = new_tracepoints - prev_tracepoints;

            uint64_t new_ends = stats->ends;
            uint64_t ends = new_ends - prev_ends;

            // Calculate throughputs
            uint64_t tput = (count * print_every) / (now - last_print);
            delta.pool_acquired = (delta.pool_acquired * print_every) / (now - last_print);
            delta.null_acquired = (delta.null_acquired * print_every) / (now - last_print);
            delta.pool_released = (delta.pool_released * print_every) / (now - last_print);
            delta.null_released = (delta.null_released * print_every) / (now - last_print);

            printf("data:\t%ld\t%ld\t%ld\t%ld\t%ld\t%ld\t%ld\t%ld\t%ld\t%ld\t%ld\t%.2f\t%.2f\t%.2f\n",
                now - begin,
                now - last_print,
                traces,
                invalid_traces,
                count,
                tput,
                count * args->payload_size,
                delta.pool_acquired,
                delta.pool_released,
                delta.null_acquired,
                delta.null_released,
                traces == 0 ? 0 : begins / (float) traces,
                tracepoints == 0 ? 0 : tracepoints / (float) count,
                ends == 0 ? 0 : ends / (float) traces
                );
            last_print = now;
            prev_count = new_count;
            prev_stats = new_stats;
            prev_invalid_traces = new_invalid_traces;
            prev_traces_count = new_traces_count;
            prev_begins = new_begins;
            prev_tracepoints = new_tracepoints;
            prev_ends = new_ends;
        }

        usleep(100000);
    }
}

typedef struct client_args {
    volatile int *alive;
    int client_id;
    struct arguments* arguments;
    exp_stats* stats;
} client_args;

void* run_client_thread(void *vargp) {
    client_args* args = (client_args*) vargp;
    client_thread_main(args->alive, args->client_id, args->arguments, args->stats);
    return 0;
}

typedef struct print_args {
    volatile int *alive;
    struct arguments* arguments;
    exp_stats* stats;    
} print_args;

void* run_print_thread(void *vargp) {
    print_args* args = (print_args*) vargp;
    print_thread_main(args->alive, args->arguments, args->stats);
    return 0;
}

void run_clients(struct arguments *arguments) {
    volatile int alive = 1;
    exp_stats stats = {0,0,0,0,0,0};
    printf("Running clients\n");
    pthread_t threads[arguments->num_threads];
    client_args args[arguments->num_threads];
    for (int i = 0; i < arguments->num_threads; i++) {
        args[i].alive = &alive;
        args[i].client_id = i;
        args[i].arguments = arguments;
        args[i].stats = &stats;
        pthread_create(&threads[i], NULL, run_client_thread, (void*) &args[i]);
    }

    pthread_t printthread;
    print_args printargs;
    printargs.alive = &alive;
    printargs.arguments = arguments;
    printargs.stats = &stats;
    pthread_create(&printthread, NULL, run_print_thread, (void*)&printargs);


    uint64_t end = nanos() + arguments->duration * 1000000000LL;
    if (arguments->duration == 0) {
        end = -1;
    }

    while (true) {
        uint64_t now = nanos();
        if (now > end) {
            break;
        }
        usleep(1000000);
    }
    alive = 0;
    for (int i = 0; i < arguments->num_threads; i++) {
        pthread_join(threads[i], NULL);
    }
    pthread_join(printthread, NULL);
    printf("Clients complete.\n");
}


int main (int argc, char **argv) {
  struct arguments arguments;

  /* Default values. */
  arguments.num_threads = 1;
  arguments.buffer_size = 4096;
  arguments.buffer_count = 25000; // Default 100MB pool
  arguments.payload_size = 400;
  arguments.tracepoints_per_request = 100;
  arguments.trigger_enabled = false;
  arguments.trigger_probability = 0;
  arguments.head_sampling_probability = 0.0;
  arguments.retroactive_sampling_percentage = 1.0;
  arguments.duration = 0;


  arguments.process_name = 0;
  arguments.output_file = "-";

  /* Parse our arguments; every option seen by parse_opt will
     be reflected in arguments. */
  argp_parse (&argp, argc, argv, 0, 0, &arguments);

  printf("name=%s\nbuffer_size=%ld\nbuffer_count=%ld\nnum_threads=%d\npayload_size=%ld\ntp_per=%d\ntrigger=%.3f\nduration=%lu\n-------\n",
            arguments.process_name,
            arguments.buffer_size, arguments.buffer_count,
            arguments.num_threads, arguments.payload_size,
            arguments.tracepoints_per_request,
            arguments.trigger_probability,
            arguments.duration);

  init_hindsight_client(&arguments);
  printf("------\n");

  run_clients(&arguments);

  exit (0);
}