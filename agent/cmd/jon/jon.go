package main

// IMPORTANT: for the below to work, must do:
//   export CGO_LDFLAGS_ALLOW=".*"

/*
#cgo CFLAGS: -I${SRCDIR}/../../../client/src -I${SRCDIR}/../../../client/include
#cgo LDFLAGS: ${SRCDIR}/../../../client/lib/libtracer.a -lm

#include "agentapi.h"

*/
import "C"

import (
	"fmt"
	"time"
)

func init_agentapi(fname string) *C.HindsightAgentAPI {
	agentapi := C.hindsight_agentapi_init(C.CString(fname))
	fmt.Println("Inited existing bufmanager", fname)
	return agentapi
}

func drain_forever(api *C.HindsightAgentAPI) {
	last_print := int(time.Now().UnixNano())
	print_every := 1000000000
	count := 0
	sum := 0

	var cb C.CompleteBuffers
	for {
		now := int(time.Now().UnixNano())
		if (now - last_print) > print_every {
			tput := (sum * print_every) / (now - last_print)
			batchsize := float32(sum) / float32(count)
			fmt.Println("Throughput:", tput, "Average batch:", batchsize)
			last_print = now
			count = 0
			sum = 0
		}

		max_backoff := 100000
		backoff := int(10)
		for {
			C.hindsight_agentapi_get_complete_nonblocking(api, &cb)
			if int(cb.count) > 0 {
				break
			}

			time.Sleep(time.Duration(backoff) * time.Nanosecond)
			backoff *= 2
			if backoff > max_backoff {
				backoff = max_backoff
			}
		}

		count += 1
		sum += int(cb.count)

		var ab C.AvailableBuffers
		ab.count = cb.count

		limit := int(cb.count)
		for i := 1; i < limit; i++ {
			ab.bufs[i].buffer_id = cb.bufs[i].buffer_id
		}

		C.hindsight_agentapi_put_available_blocking(api, &ab)
	}
}

func main() {
	fmt.Println("Hello world!")

	fname := "hs_integration_test"

	agentapi := init_agentapi(fname)
	drain_forever(agentapi)
}
