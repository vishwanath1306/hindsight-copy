#include <stdio.h>
#include <fcntl.h>
#include <sys/mman.h>
#include <assert.h>
#include <string.h>
#include "common.h"

#include "hindsight.h"

Hindsight hindsight;
BufManager* mgr;
__thread TraceState hindsight_tls = {false};

void hindsight_print_config(HindsightConfig* conf) {
    printf("Hindsight Config:\n");
    printf("  Buffer pool cap=%ld buf_length=%ld\n", conf->pool_capacity, conf->buffer_size);
    printf("  Service addr=%s\n", conf->address);
    printf("  Queue sizes breadcrumbs_cap=%ld triggers_cap=%ld\n", conf->breadcrumbs_capacity, conf->triggers_capacity);
    printf("  Head-based sampling p=%.5f (traceid threshold %lu)\n", conf->head_sampling_probability, conf->_head_sampling_threshold);
    printf("  Retroactive sampling p=%.5f (traceid threshold %lu)\n", conf->retroactive_sampling_percentage, conf->_retroactive_sampling_threshold);
}

HindsightConfig hindsight_default_config() {
    HindsightConfig conf;
    conf.pool_capacity = -1; // size_t doesn't have negatives but we won't use comparisons
    conf.buffer_size = -1;
    conf.breadcrumbs_capacity = -1;
    conf.triggers_capacity = -1;
    conf.retroactive_sampling_percentage = 1.0;
    conf._retroactive_sampling_threshold = UINT64_MAX;
    conf.head_sampling_probability = 0.0;
    conf._head_sampling_threshold = 0;
    return conf;
}

HindsightConfig hindsight_load_config(const char* service_name) {
    // Load Hindsight conf for this service from default location
    char config_fname[128];
    strcpy(config_fname, "/etc/hindsight_conf/");
    strcat(config_fname, service_name);
    strcat(config_fname, ".conf");
    return hindsight_load_config_file(config_fname);
}

HindsightConfig hindsight_load_config_file(const char* fname) {
    // Initialize config with defaults
    HindsightConfig conf = hindsight_default_config();

    // Addr in the conf file is specified as separate address and port strings
    char* conf_addr = (char*) malloc(32 * sizeof(char));
    char* conf_port = (char*) malloc(32 * sizeof(char));
    memset(conf_addr, 0, 32*sizeof(char));
    memset(conf_port, 0, 32*sizeof(char));
    
    // Open the specified file, with defaults as backup
    FILE* config_file;
    config_file = fopen(fname,"r");
    if (config_file == NULL) {
        config_file = fopen(HINDSIGHT_DEFAULT_CONFIG,"r");
    }

    // Read the config
    char* line = NULL;
    size_t read;
    size_t len = 0;
    while((read = getline(&line, &len, config_file)) != -1) {
        char* temp = strchr(line, '\n');
        int index = (int)(temp - line);

        char* new_line = malloc(sizeof(char)*64);
        memset(new_line, 0, 64*sizeof(char));
        if (index == strlen(line)-1) {
            strncpy(new_line, line, index);
        } else {
            strncpy(new_line, line, strlen(line));
        }

        char* var = malloc(sizeof(char)*32);
        memset(var, 0, 32*sizeof(char));
        char* value = malloc(sizeof(char)*32);
        memset(value, 0, 32*sizeof(char));
        sscanf(new_line, "%s %s", var, value);

        if (!strcmp(var, "cap")) {
            conf.pool_capacity = atoi(value);
        }

        if (!strcmp(var, "buf_length")) {
            conf.buffer_size = atoi(value);
        }

        if (!strcmp(var, "addr")) {
            conf_addr = value;
        }

        if (!strcmp(var, "port")) {
            conf_port = value;
        }       

        if (!strcmp(var, "breadcrumbs_cap")) {
            conf.breadcrumbs_capacity = atoi(value);
        }       

        if (!strcmp(var, "triggers_cap")) {
            conf.triggers_capacity = atoi(value);
        }

        if (!strcmp(var, "retroactive_sampling_percentage")) {
            conf.retroactive_sampling_percentage = atof(value);
        }

        if (!strcmp(var, "head_sampling_probability")) {
            conf.head_sampling_probability = atof(value);
        }
    }
    fclose(config_file);

    if (line) free(line);

    
    // Addr in the conf struct is a single string of address:port
    conf.address = (char*) malloc(32 * sizeof(char));
    memset(conf.address, 0, 32*sizeof(char));
    strcpy(conf.address, conf_addr);
    strcat(conf.address, ":");
    strncat(conf.address, conf_port, 4);

    if (conf.pool_capacity == -1) conf.pool_capacity = 1000;
    if (conf.buffer_size == -1) conf.buffer_size = 1000;
    if (conf.breadcrumbs_capacity == -1) conf.breadcrumbs_capacity = conf.pool_capacity;
    if (conf.triggers_capacity == -1) conf.triggers_capacity = conf.pool_capacity;

    conf._retroactive_sampling_threshold = multiply_by(UINT64_MAX, conf.retroactive_sampling_percentage);
    conf._head_sampling_threshold = multiply_by(UINT64_MAX, conf.head_sampling_probability);

    return conf;
}

void hindsight_init(const char* service_name) {
    hindsight_init_with_config(service_name, hindsight_load_config(service_name));
}

void hindsight_init_with_config(const char* service_name, HindsightConfig config) {
    hindsight.config = config;
    hindsight_print_config(&hindsight.config);

    // Create pools and queues
    hindsight.mgr = bufmanager_init(
        service_name,
        hindsight.config.pool_capacity,
        hindsight.config.buffer_size);

    hindsight.breadcrumbs = breadcrumbs_init(
        service_name, 
        hindsight.config.breadcrumbs_capacity);

    hindsight.triggers = triggers_init(
        service_name,
        hindsight.config.triggers_capacity);

    mgr = &hindsight.mgr;
}

void hindsight_begin(uint64_t trace_id) {
    tracestate_begin_with_sampling(&hindsight_tls, mgr, trace_id, hindsight.config._head_sampling_threshold, hindsight.config._retroactive_sampling_threshold);
    if (hindsight_get_is_head_sampled()) {
        hindsight_trigger(TRIGGER_ID_HEAD_BASED_SAMPLING);
    }
}

void hindsight_begin_sampled(uint64_t trace_id) {
    tracestate_begin_with_sampling(&hindsight_tls, mgr, trace_id, UINT64_MAX, hindsight.config._retroactive_sampling_threshold);
    hindsight_trigger(TRIGGER_ID_HEAD_BASED_SAMPLING);
}

void hindsight_end() {
    tracestate_end(&hindsight_tls, mgr);
}

TraceState hindsight_detach() {
    TraceState current = hindsight_tls;
    hindsight_tls = tracestate_create();
    return current;
}

void hindsight_attach(TraceState* state) {
    hindsight_end();
    hindsight_tls = *state;
}

void hindsight_tracepoint(char* buf, size_t buf_size) {
    if (tracestate_try_write(&hindsight_tls, buf, buf_size)) return;
    tracestate_write(&hindsight_tls, mgr, buf, buf_size);
}

void hindsight_tracepoint_write(size_t write_size, char** dst, size_t* dst_size) {
    tracestate_write_data(&hindsight_tls, mgr, 
        write_size, dst, dst_size);
}

void hindsight_breadcrumb(const char* addr) {
    breadcrumbs_add(&hindsight.breadcrumbs, hindsight_tls.header.trace_id, addr);
}

void hindsight_forward_breadcrumb(const char* addr) {
    breadcrumbs_add_forward(&hindsight.breadcrumbs, hindsight_tls.header.trace_id, addr);
}

void hindsight_trigger(int trigger_id) {
    uint64_t trace_id = hindsight_tls.header.trace_id;
    triggers_fire(&hindsight.triggers, trigger_id, trace_id, trace_id);
}

void hindsight_trigger_manual(uint64_t trace_id, int trigger_id) {
    triggers_fire(&hindsight.triggers, trigger_id, trace_id, trace_id);   
}

void hindsight_trigger_lateral(int trigger_id, uint64_t base_trace_id, uint64_t lateral_trace_id) {
    triggers_fire(&hindsight.triggers, trigger_id, base_trace_id, lateral_trace_id);  
}

uint64_t hindsight_get_traceid() {
    return hindsight_tls.header.trace_id;
}

char* hindsight_get_local_address() {
    return hindsight.config.address;
}

bool hindsight_get_is_head_sampled() {
    return hindsight_tls.head_sampled;
}

char* hindsight_serialize() {
    return hindsight_get_local_address();
}

void hindsight_deserialize(char* baggage) {
    hindsight_breadcrumb(baggage);
}

float hindsight_retroactive_sampling_percentage() {
    return hindsight.config.retroactive_sampling_percentage;
}

float hindsight_head_sampling_probability() {
    return hindsight.config.head_sampling_probability;
}

int hindsight_null_buffer_count() {
    return hindsight_tls.header.null_buffer_count;
}

bool hindsight_is_active() {
    return hindsight_tls.active;
}

bool hindsight_is_recording() {
    return hindsight_tls.recording;
}