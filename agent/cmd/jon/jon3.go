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
	"context"
	"fmt"
	"time"

	"github.com/geraldleizhang/hindsight/agent/pkg/memory"
)

func drain_forever(api *memory.GoAgentAPI, ctx context.Context) {
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

		select {
		case batch := <-api.Complete:
			count += 1
			for _, buffer_ids := range batch {
				sum += len(buffer_ids)
				api.Available <- buffer_ids
			}
		}
	}
	select {
	case <-ctx.Done():
		return
	}
}

func main() {
	fmt.Println("Hello world!")

	fname := "hs_integration_test"

	agentapi := memory.InitGoAgentAPI(fname)

	ctx, _ := context.WithCancel(context.Background())

	go agentapi.Run(ctx)
	drain_forever(agentapi, ctx)

}
