#include <stdio.h>
#include <stdlib.h>
#include <time.h>
#include "hindsight.h"
#include "assert.h"
#include "common.h"

void do_test_sampling_threshold(float f, uint64_t expected) {
	uint64_t actual = multiply_by(UINT64_MAX, f);
	printf("%.7f is %lu (expected %lu)\n", f, actual, expected);
	assert(expected == actual);
}

void test_sampling_threshold() {
	do_test_sampling_threshold(0.0, 0);	
	do_test_sampling_threshold(1.0, UINT64_MAX);
	do_test_sampling_threshold(0.5, UINT64_MAX/2);
	do_test_sampling_threshold(0.25, UINT64_MAX/4);
	do_test_sampling_threshold(0.1, UINT64_MAX/10);
	do_test_sampling_threshold(0.01, UINT64_MAX/100);
	do_test_sampling_threshold(0.001, UINT64_MAX/1000);
	do_test_sampling_threshold(0.0001, UINT64_MAX/10000);
	do_test_sampling_threshold(0.00001, UINT64_MAX/100000);
	do_test_sampling_threshold(0.000001, UINT64_MAX/1000000);
	do_test_sampling_threshold(0.0000001, UINT64_MAX/10000000);

	printf("test_sampling_threshold passed\n");

}

void test_default_sampling() {
	HindsightConfig config = hindsight_load_config("myservice");
	hindsight_print_config(&config);

	// Test the Hindsight default conf values
	assert(config.retroactive_sampling_percentage == 1.0);
	assert(config._retroactive_sampling_threshold == UINT64_MAX);

	assert(config.head_sampling_probability == 0.0);
	assert(config._head_sampling_threshold == 0);

	printf("test_default_sampling passed\n");
}

void test_sampling() {
	HindsightConfig config = hindsight_load_config_file("./test/sampling_test.conf");
	hindsight_print_config(&config);

	// Test the Hindsight default conf values
	assert(config.retroactive_sampling_percentage == 0.25f);
	assert(config._retroactive_sampling_threshold == UINT64_MAX/4);

	assert(config.head_sampling_probability == 0.01f);
	assert(config._head_sampling_threshold == UINT64_MAX/100);

	printf("test_default_sampling passed\n");
}

void test_sampling_enforced() {
	HindsightConfig config = hindsight_load_config_file("./test/sampling_test.conf");
	hindsight_print_config(&config);

	hindsight_init_with_config("sampling_test", config);

	hindsight_begin(10);

	assert(hindsight_is_active() == true);
	assert(hindsight_is_recording() == true);
	assert(hindsight_get_is_head_sampled() == true);

	hindsight_end();

	assert(hindsight_is_active() == false);
	assert(hindsight_is_recording() == false);
	assert(hindsight_get_is_head_sampled() == false);


	hindsight_begin(UINT64_MAX);

	assert(hindsight_is_active() == true);
	assert(hindsight_is_recording() == false);
	assert(hindsight_get_is_head_sampled() == false);

	hindsight_end();

	assert(hindsight_is_active() == false);
	assert(hindsight_is_recording() == false);
	assert(hindsight_get_is_head_sampled() == false);


	hindsight_begin(UINT64_MAX/10);

	assert(hindsight_is_active() == true);
	assert(hindsight_is_recording() == true);
	assert(hindsight_get_is_head_sampled() == false);

	hindsight_end();

	assert(hindsight_is_active() == false);
	assert(hindsight_is_recording() == false);
	assert(hindsight_get_is_head_sampled() == false);

	printf("test_sampling_enforced passed\n");
}

int main(int argc, char const *argv[])
{
	printf("Testing sampling!\n");
	test_sampling_threshold();
	test_default_sampling();
	test_sampling();
	test_sampling_enforced();
	return 0;
}