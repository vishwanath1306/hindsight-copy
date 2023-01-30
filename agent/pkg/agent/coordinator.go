package agent

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	"github.com/geraldleizhang/hindsight/agent/pkg/datapb"
	"github.com/geraldleizhang/hindsight/agent/pkg/memory"
	"google.golang.org/grpc"
)

type Coordinator struct {
	datapb.UnimplementedAgentServer

	enabled bool

	local_addr  string // The address that the coordinator uses to contact us
	local_port  string
	remote_addr string // Address of the coordinator

	localtriggers  chan []memory.Trigger    // Local triggers to be reported to coordinator
	breadcrumbs    chan map[uint64][]string // Breadcrumbs to be reported to coordinator
	remotetriggers chan []memory.Trigger    // Remote triggers received from coordinator
}

func InitCoordinator(enabled bool, local_hostname string, local_port string, remote_addr string) *Coordinator {
	var r Coordinator
	r.Init(enabled, local_hostname, local_port, remote_addr)
	return &r
}

func (r *Coordinator) Init(enabled bool, local_hostname string, local_port string, remote_addr string) {
	r.enabled = enabled // used for testing/dev

	r.local_port = local_port
	r.local_addr = local_hostname + ":" + local_port
	r.remote_addr = remote_addr

	r.localtriggers = make(chan []memory.Trigger, 500)
	r.breadcrumbs = make(chan map[uint64][]string, 500)
	r.remotetriggers = make(chan []memory.Trigger, 500)
}

/* Send a batch of local triggers to the coordinator */
func (r *Coordinator) sendTriggers(rpcclient datapb.CoordinatorClient, triggers []memory.Trigger) error {
	var request datapb.TriggerRequest
	request.Src = r.local_addr

	for _, trigger := range triggers {
		var t datapb.Trigger
		t.QueueId = int32(trigger.Queue_id)
		t.BaseTraceId = trigger.Base_trace_id
		t.TraceIds = []uint64{trigger.Trace_id}
		request.Triggers = append(request.Triggers, &t)
	}

	if r.enabled {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_, err := rpcclient.LocalTrigger(ctx, &request)

		return err
	}

	return nil
}

/* Received a batch of remote triggers from the coordinator */
func (r *Coordinator) RemoteTrigger(ctx context.Context, in *datapb.TriggerRequest) (*datapb.TriggerReply, error) {

	var triggers []memory.Trigger
	for _, trigger := range in.Triggers {
		for _, traceid := range trigger.GetTraceIds() {
			// TODO: update memory.Trigger with list of trace ids
			var mt memory.Trigger
			mt.Queue_id = int(trigger.QueueId)
			mt.Base_trace_id = trigger.BaseTraceId
			mt.Trace_id = traceid
			triggers = append(triggers, mt)
		}
	}

	if len(triggers) > 0 {
		select {
		case r.remotetriggers <- triggers:
			break
		default:
			// Agent is bottlenecked, drop remote triggers
			// TODO: counters here
		}
	}

	return &datapb.TriggerReply{}, nil
}

/* After we are connected to the coordinator, this loops over the outgoing breadcrumbs, reporting them in batches */
func (r *Coordinator) BreadcrumbsLoop(ctx context.Context) {
	firsttime := true
	for {
		select {
		case <-ctx.Done():
			return
		default:
			break
		}

		conn, err := grpc.Dial(r.remote_addr, grpc.WithInsecure(), grpc.WithTimeout(10*time.Second))
		if err != nil {
			if firsttime {
				log.Printf("Unable to send breadcrumbs to %s; will retry every 2 seconds (reason: %s)\n", r.remote_addr, err.Error())
				firsttime = false
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
				continue
			}
			continue
		}
		defer conn.Close()

		rpcclient := datapb.NewCoordinatorClient(conn)

		err = r.ReportBreadcrumbs(ctx, rpcclient)
		if err != nil {
			if firsttime {
				log.Printf("Unable to send breadcrumbs to %s; will retry every 2 seconds (reason: %s)\n", r.remote_addr, err.Error())
				firsttime = false
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
				continue
			}
			continue
		}

		firsttime = true
	}
}

func (r *Coordinator) ReportBreadcrumbs(ctx context.Context, rpcclient datapb.CoordinatorClient) error {
	addr_to_id := make(map[string]int32)
	seed := int32(0)

	for {
		// Accumulate a batch of up to 100 breadcrumbs
		var accumulated []map[uint64][]string

		// Block waiting for some breadcrumbs
		for len(accumulated) == 0 {
			select {
			case <-ctx.Done():
				return nil
			case breadcrumbs := <-r.breadcrumbs:
				if len(breadcrumbs) > 0 {
					accumulated = append(accumulated, breadcrumbs)
				}
			}
		}

		// Now try to batch as many additional breadcrumbs as possible (without blocking)
	Accumulation:
		for len(accumulated) < 100 {
			select {
			case <-ctx.Done():
				return nil
			case breadcrumbs := <-r.breadcrumbs:
				if len(breadcrumbs) > 0 {
					accumulated = append(accumulated, breadcrumbs)
				}
			default:
				break Accumulation
			}
		}

		// Construct RPC request object, mapping from string addrs to ints
		var request datapb.BreadcrumbsRequest
		request.Src = r.local_addr
		for _, breadcrumbs := range accumulated {
			for trace_id, addrs := range breadcrumbs {
				var bcs datapb.Breadcrumbs
				bcs.TraceId = trace_id
				request.Breadcrumbs = append(request.Breadcrumbs, &bcs)

				for _, addr := range addrs {
					if addr_id, ok := addr_to_id[addr]; ok {
						bcs.Addrs = append(bcs.Addrs, addr_id)

					} else {
						var bca datapb.BreadcrumbAddress
						bca.Addr = addr
						bca.Id = seed
						request.Addresses = append(request.Addresses, &bca)

						addr_to_id[addr] = seed
						bcs.Addrs = append(bcs.Addrs, seed)
						seed++
					}
				}
			}
		}

		if r.enabled {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			_, err := rpcclient.Breadcrumbs(ctx, &request)
			if err != nil {
				return err
			}
		}
	}
}

func (r *Coordinator) TriggersLoop(ctx context.Context) {
	firsttime := true
	for {
		select {
		case <-ctx.Done():
			return
		default:
			break
		}

		conn, err := grpc.Dial(r.remote_addr, grpc.WithInsecure(), grpc.WithTimeout(10*time.Second))
		if err != nil {
			if firsttime {
				log.Printf("Unable to send local triggers to %s; will retry every 2 seconds (reason: %s)\n", r.remote_addr, err.Error())
				firsttime = false
			}
			time.Sleep(time.Duration(2) * time.Second)
			continue
		}
		defer conn.Close()

		rpcclient := datapb.NewCoordinatorClient(conn)

		err = r.ReportTriggers(ctx, rpcclient)
		if err != nil {
			if firsttime {
				log.Printf("Unable to send local triggers to %s; will retry every 2 seconds (reason: %s)\n", r.remote_addr, err.Error())
				firsttime = false
			}
			time.Sleep(time.Duration(2) * time.Second)
			continue
		}

		firsttime = true
	}
}

func (r *Coordinator) ReportTriggers(ctx context.Context, rpcclient datapb.CoordinatorClient) error {
	for {
		// Accumulate a batch of up to 100 triggers
		var accumulated []memory.Trigger

		// Block waiting for some triggers
		for len(accumulated) == 0 {
			select {
			case <-ctx.Done():
				return nil
			case triggers := <-r.localtriggers:
				accumulated = append(accumulated, triggers...)
			}
		}

		// Now try to batch as many additional triggers as possible (without blocking)
	Accumulation:
		for len(accumulated) < 100 {
			select {
			case <-ctx.Done():
				return nil
			case triggers := <-r.localtriggers:
				accumulated = append(accumulated, triggers...)
			default:
				break Accumulation
			}
		}

		// Send them
		err := r.sendTriggers(rpcclient, accumulated)

		if err != nil {
			return err
		}
	}
}

func (r *Coordinator) Run(ctx context.Context, cancel context.CancelFunc) {
	log.Println("Receiving remote triggers on:", r.local_port)

	lis, err := net.Listen("tcp", ":"+r.local_port)
	if err != nil {
		log.Printf("Coordinator unable to listen for remote triggers: %v\n", err)
		cancel()
		return
	}
	s := grpc.NewServer()
	datapb.RegisterAgentServer(s, r)

	go func() {
		select {
		case <-ctx.Done():
			s.GracefulStop()
		}
	}()

	// Run the GRPC server
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Printf("Coordinator server stopped unexpectedly: %v\n", err)
			cancel()
			return
		}
		log.Println("Stopped receiving remote triggers from coordinator")
	}()

	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		log.Println("Breadcrumbs will be reported to", r.remote_addr)
		r.BreadcrumbsLoop(ctx)
		log.Println("Stopped sending breadcrumbs to coordinator")
		wg.Done()
	}()
	go func() {
		log.Println("Triggers will be reported to", r.remote_addr)
		r.TriggersLoop(ctx)
		log.Println("Stopped sending local triggers to coordinator")
		wg.Done()
	}()
	wg.Wait()
	s.Stop()
}
