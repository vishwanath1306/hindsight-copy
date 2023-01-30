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

	"github.com/geraldleizhang/hindsight/agent/pkg/memory"
)

func drain_forever(api *memory.AgentAPI) {
	last_print := int(time.Now().UnixNano())
	print_every := 1000000000
	count := 0
	sum := 0

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

		var buffers []memory.CompleteBuffer
		max_backoff := 100000
		backoff := int(10)
		for {
			buffers = api.GetComplete()
			if len(buffers) > 0 {
				break
			}

			time.Sleep(time.Duration(backoff) * time.Nanosecond)
			backoff *= 2
			if backoff > max_backoff {
				backoff = max_backoff
			}
		}

		count += 1
		sum += len(buffers)

		ids := make([]int, len(buffers))
		for i, buf := range buffers {
			ids[i] = buf.Buffer_id
		}

		api.PutAvailable(ids)
	}
}

func main() {
	fmt.Println("Hello world!")

	fname := "hs_integration_test"

	agentapi := memory.InitAgentAPI(fname)
	drain_forever(agentapi)
}
