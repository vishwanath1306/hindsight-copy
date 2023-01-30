package coordinator

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/geraldleizhang/hindsight/agent/pkg/datapb"
	"google.golang.org/grpc"
)

type IncomingBreadcrumbs struct {
	req *datapb.BreadcrumbsRequest
	ret chan error
}

type IncomingTriggers struct {
	req *datapb.TriggerRequest
	ret chan error
}

type CoordinatorServer struct {
	datapb.UnimplementedCoordinatorServer

	ctx     context.Context   // For shutdown
	timeout time.Duration     // Time before expiring triggers / traces
	c       Coordinator       // Manages coordination data
	agents  map[string]*Agent // connections to agents

	listen_port string // Port to listen for connections from agents

	incoming_triggers    chan *IncomingTriggers
	incoming_breadcrumbs chan *IncomingBreadcrumbs

	dropped_incoming_triggers    uint64
	dropped_incoming_breadcrumbs uint64
	trigger_warn_mutex           sync.Mutex
	breadcrumb_warn_mutex        sync.Mutex
	last_trigger_warning         time.Time
	last_breadcrumb_warning      time.Time

	logger *CsvLogger
}

type Agent struct {
	addr              string
	id_to_addr        map[int32]string
	outgoing_triggers chan []Trigger
	dropped_triggers  int
	last_warn         time.Time
}

func (s *CoordinatorServer) Init(port string, logfile string) (err error) {
	s.c.Init()
	s.timeout = -60 * time.Second
	s.agents = make(map[string]*Agent)
	s.listen_port = port
	s.incoming_triggers = make(chan *IncomingTriggers, 10000)
	s.incoming_breadcrumbs = make(chan *IncomingBreadcrumbs, 10000)
	if logfile != "" {
		s.logger, err = NewCsvLogger(logfile)
	} else {
		s.logger = nil
	}
	s.last_trigger_warning = time.Now()
	s.last_breadcrumb_warning = time.Now()
	return
}

func (a *Agent) Init(addr string) {
	a.addr = addr
	a.id_to_addr = make(map[int32]string)
	a.outgoing_triggers = make(chan []Trigger, 10000)
	a.dropped_triggers = 0
	a.last_warn = time.Now()
}

func (s *CoordinatorServer) Run(ctx context.Context) {
	s.ctx = ctx
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		s.runServer(ctx)
		log.Println("Stopped coordinator server")
		wg.Done()
	}()
	go func() {
		s.runCoordinator(ctx)
		log.Println("Stopped main coordinator goroutine")
		wg.Done()
	}()
	if s.logger != nil {
		cancel := s.logger.Run()
		wg.Wait()
		cancel() // Done like this to ensure everything gets drained properly
		s.logger.AwaitCompletion()
	} else {
		wg.Wait()
	}
}

/* Run the RPC server that receives triggers and breadcrumbs */
func (cs *CoordinatorServer) runServer(ctx context.Context) {
	lis, err := net.Listen("tcp", ":"+cs.listen_port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Println("Listening for agent connections on port", cs.listen_port)
	grpcserver := grpc.NewServer()
	datapb.RegisterCoordinatorServer(grpcserver, cs)

	go func() {
		select {
		case <-ctx.Done():
			log.Println("Shutting down gRPC server")
			grpcserver.GracefulStop()
		}
	}()

	if err := grpcserver.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func (cs *CoordinatorServer) GetAgent(addr string) *Agent {
	if agent, ok := cs.agents[addr]; ok {
		return agent
	} else {
		var agent Agent
		agent.Init(addr)
		cs.agents[addr] = &agent
		agent.Run(cs.ctx)
		return &agent
	}
}

func (cs *CoordinatorServer) checkExpirations() {
	cs.c.now = time.Now()
	cs.c.checkTraceExpiration(cs.c.now.Add(cs.timeout))
	finished := cs.c.checkTriggerExpiration(cs.c.now.Add(cs.timeout))
	if cs.logger != nil && len(finished) > 0 {
		select {
		case cs.logger.Finished <- finished:
			return
		default:
			cs.logger.dropped_finished += len(finished)
		}
	}
}

func (cs *CoordinatorServer) processTriggersRequest(incoming *IncomingTriggers) {
	cs.c.now = time.Now()
	req := incoming.req

	triggers_to_forward := make(map[string][]Trigger)
	for _, t := range req.Triggers {
		// Store the received trigger
		var trigger Trigger
		trigger.id.queue_id = int(t.QueueId)
		trigger.id.base_trace_id = t.BaseTraceId
		trigger.trace_ids = t.TraceIds
		forwarding_addrs := cs.c.AddTrigger(req.Src, trigger)

		// Forward the trigger to any addresses specified
		for _, addr := range forwarding_addrs {
			triggers_to_forward[addr] = append(triggers_to_forward[addr], trigger)
		}
	}

	// Do the forwarding
	for addr, triggers := range triggers_to_forward {
		cs.GetAgent(addr).SendTriggers(triggers)
	}

	cs.checkExpirations()
	select {
	case incoming.ret <- nil:
	default:
	}
}

func (cs *CoordinatorServer) processBreadcrumbRequest(incoming *IncomingBreadcrumbs) {
	cs.c.now = time.Now()
	req := incoming.req

	origin := cs.GetAgent(req.Src)

	// Breadcrumbs are received as IDs; unravel into addr strings
	for _, a := range req.Addresses {
		origin.id_to_addr[a.Id] = a.Addr
	}

	breadcrumbs := make(map[uint64][]string)
	for _, b := range req.Breadcrumbs {
		for _, addr_id := range b.Addrs {
			if addr, ok := origin.id_to_addr[addr_id]; ok {
				breadcrumbs[b.TraceId] = append(breadcrumbs[b.TraceId], addr)
			} else {
				e := fmt.Errorf("Received addr_id %d from %s that hasn't been mapped to an address", addr_id, req.Src)
				select {
				case incoming.ret <- e:
				default:
				}
				return
			}
		}
	}

	// Now process them
	triggers_to_forward := make(map[string][]Trigger)
	for trace_id, addrs := range breadcrumbs {
		// Store the received breadcrumbs
		to_forward := cs.c.AddBreadcrumb(req.Src, trace_id, addrs)

		// Forward any necessary triggers
		for addr, triggers := range to_forward {
			if len(triggers) > 0 {
				triggers_to_forward[addr] = append(triggers_to_forward[addr], triggers...)
			}
		}
	}

	// Do the forwarding
	for addr, triggers := range triggers_to_forward {
		cs.GetAgent(addr).SendTriggers(triggers)
	}

	cs.checkExpirations()
	select {
	case incoming.ret <- nil:
	default:
	}
}

/* The "main" thread that receives incoming stuff and sends outgoing stuff */
func (cs *CoordinatorServer) runCoordinator(ctx context.Context) {
	log.Println("CoordinatorServer main goroutine running")
	for {
		select {
		case <-ctx.Done():
			if cs.logger != nil {
				log.Println("CoordinatorServer flushing logs")
				/* Expire everything, so that it flushes to log */
				finished := cs.c.checkTriggerExpiration(time.Now().Add(1 * time.Second))
				select {
				case cs.logger.Finished <- finished:
					break
				default:
					cs.logger.dropped_finished += len(finished)
				}
			}
			return
		case req := <-cs.incoming_triggers:
			/* Received some triggers from an agent over RPC*/
			cs.processTriggersRequest(req)
		case req := <-cs.incoming_breadcrumbs:
			/* Received some breadcrumbs from an agent over RPC */
			cs.processBreadcrumbRequest(req)
		}
	}
}

/* An agent has sent us a trigger */
func (s *CoordinatorServer) LocalTrigger(ctx context.Context, req *datapb.TriggerRequest) (rsp *datapb.TriggerReply, err error) {
	var incoming IncomingTriggers
	incoming.req = req
	incoming.ret = make(chan error, 1)

	rsp = &datapb.TriggerReply{}

	select {
	case s.incoming_triggers <- &incoming:
		select {
		case <-ctx.Done():
			return
		case err = <-incoming.ret:
			if err != nil {
				log.Println("Breadcrumbs error:", err.Error())
			}
		}
	default:
		atomic.AddUint64(&s.dropped_incoming_triggers, uint64(len(req.Triggers)))
		if s.trigger_warn_mutex.TryLock() {
			defer s.trigger_warn_mutex.Unlock()

			if time.Now().After(s.last_trigger_warning.Add(1 * time.Second)) {
				dropped := s.dropped_incoming_triggers
				atomic.AddUint64(&s.dropped_incoming_triggers, -dropped)

				s.last_trigger_warning = time.Now()
				log.Printf("Warning: coordinator is bottlenecked; %d incoming triggers dropped\n", dropped)
			}
		}
	}
	return
}

/* An agent has sent us breadcrumbs */
func (s *CoordinatorServer) Breadcrumbs(ctx context.Context, req *datapb.BreadcrumbsRequest) (rsp *datapb.BreadcrumbsReply, err error) {
	var incoming IncomingBreadcrumbs
	incoming.req = req
	incoming.ret = make(chan error, 1)

	rsp = &datapb.BreadcrumbsReply{}

	select {
	case s.incoming_breadcrumbs <- &incoming:
		select {
		case <-ctx.Done():
			return
		case err = <-incoming.ret:
			if err != nil {
				log.Println("Breadcrumbs error:", err)
			}
		}
	default:
		atomic.AddUint64(&s.dropped_incoming_breadcrumbs, uint64(len(req.Breadcrumbs)))
		if s.breadcrumb_warn_mutex.TryLock() {
			defer s.breadcrumb_warn_mutex.Unlock()

			if time.Now().After(s.last_breadcrumb_warning.Add(1 * time.Second)) {
				dropped := s.dropped_incoming_breadcrumbs
				atomic.AddUint64(&s.dropped_incoming_breadcrumbs, -dropped)

				s.last_breadcrumb_warning = time.Now()
				log.Printf("Warning: coordinator is bottlenecked; %d incoming breadcrumbs dropped\n", dropped)
			}
		}
	}
	return
}

func (a *Agent) Run(ctx context.Context) {
	go func() {
		a.AgentLoop(ctx)
	}()
}

/* Connects to an agent in a loop, then sends triggers once connected */
func (a *Agent) AgentLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			break
		}

		log.Println("Connecting to agent", a.addr)
		conn, err := grpc.Dial(a.addr, grpc.WithInsecure(), grpc.WithTimeout(2*time.Second))
		if err != nil {
			log.Println("Unable to connect to", a.addr, "retrying in 2 seconds:", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
				continue
			}
			continue
		}
		defer conn.Close()

		rpcclient := datapb.NewAgentClient(conn)

		err = a.ReportTriggers(ctx, rpcclient)
		if err != nil {
			log.Println("Connection error", a.addr, "retrying in 2 seconds:", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
				continue
			}
			continue
		}
	}
}

func (a *Agent) ReportTriggers(ctx context.Context, rpcclient datapb.AgentClient) error {
	for {
		// Accumulate a batch of up to 100 triggers
		var accumulated []Trigger

		// Block waiting for some triggers
		for len(accumulated) == 0 {
			select {
			case <-ctx.Done():
				return nil
			case triggers := <-a.outgoing_triggers:
				if len(triggers) > 0 {
					accumulated = append(accumulated, triggers...)
				}
			}
		}

		// Now try to batch as many additional triggers as possible (without blocking)
	Accumulation:
		for len(accumulated) < 100 {
			select {
			case <-ctx.Done():
				return nil
			case triggers := <-a.outgoing_triggers:
				if len(triggers) > 0 {
					accumulated = append(accumulated, triggers...)
				}
			default:
				break Accumulation
			}
		}

		// Send them
		err := a.doSend(rpcclient, accumulated)

		if err != nil {
			return err
		}
	}
}

/* Send a batch of remove triggers to an agent */
func (a *Agent) doSend(rpcclient datapb.AgentClient, triggers []Trigger) error {
	var request datapb.TriggerRequest
	for _, trigger := range triggers {
		var t datapb.Trigger
		t.QueueId = int32(trigger.id.queue_id)
		t.BaseTraceId = trigger.id.base_trace_id
		t.TraceIds = trigger.trace_ids
		request.Triggers = append(request.Triggers, &t)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := rpcclient.RemoteTrigger(ctx, &request)

	return err
}

func (a *Agent) SendTriggers(triggers []Trigger) {
	select {
	case a.outgoing_triggers <- triggers:
		break
	default:
		a.dropped_triggers += len(triggers)
		now := time.Now()
		if now.After(a.last_warn.Add(1 * time.Second)) {
			log.Printf("Warning: agent %s is bottlenecked; dropping %d triggers\n", a.addr, a.dropped_triggers)
			a.dropped_triggers = 0
			a.last_warn = now
		}
	}
}
