package agent

import (
	"context"
	"encoding/binary"
	"log"
	"net"
	"time"

	"github.com/geraldleizhang/hindsight/agent/pkg/memory"

	"github.com/juju/ratelimit"
)

type Reporting struct {
	api     *memory.GoAgentAPI // API to the shared memory for returning data buffers
	enabled bool

	rate_limit  float64
	buffer_size int
	bucket      *ratelimit.Bucket

	agent_addr  string     // Address of this agent
	remote_addr string     // Address of the trace data backend (not the coordinator)
	data        chan []int // Buffers to be reported to collector
}

func InitReporting(api *memory.GoAgentAPI, rate_limit_mb float64, enabled bool, remote_addr string,
	local_hostname string, local_port string) *Reporting {
	var r Reporting
	r.Init(api, rate_limit_mb, enabled, remote_addr, local_hostname, local_port)
	return &r
}

func (r *Reporting) Init(api *memory.GoAgentAPI, rate_limit_mb float64, enabled bool, remote_addr string,
	local_hostname string, local_port string) {
	r.api = api
	r.data = make(chan []int, 4)               // 4 somewhat arbitrary
	r.enabled = enabled                        // used for testing/dev
	r.rate_limit = rate_limit_mb * 1024 * 1024 // rate limit in bytes/s
	r.buffer_size = r.api.BufferSize()

	if r.rate_limit != 0 {
		r.bucket = ratelimit.NewBucketWithRate(r.rate_limit, int64(r.rate_limit))
	}

	r.agent_addr = local_hostname + ":" + local_port
	r.remote_addr = remote_addr
}

func doWrite(conn net.Conn, src []byte) error {
	for len(src) > 0 {
		written, err := conn.Write(src)
		if err != nil {
			return err
		}

		src = src[written:]
	}
	return nil
}

func writeLengthPrefixed(conn net.Conn, buf []byte) (err error) {
	sizebuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(sizebuf, uint32(len(buf)))
	err = doWrite(conn, sizebuf)
	if err != nil {
		return
	}
	err = doWrite(conn, buf)
	return
}

/* Reports trace data to the collector TODO grpc? */
func (r *Reporting) reportData(conn net.Conn, buffers []int) (err error) {
	// Apply rate-limiting
	if r.bucket != nil {
		r.bucket.Wait(int64(len(buffers) * r.buffer_size))
	}

	if r.enabled {
		for _, buffer_id := range buffers {
			// Get the buffer data from the buffer pool
			header, data := r.api.ExtractBuffer(buffer_id)
			data = data[0:header.Size]

			// Send it
			err = writeLengthPrefixed(conn, data)
			if err != nil {
				break // Stop writing and allow error to propagate; always return all buffers to pool
			}
		}
	}

	// Return the buffers
	if len(buffers) > 0 {
		r.api.Available <- buffers
	}

	return
}

/* We need to inform the reporting backend of this agent's identity */
func (r *Reporting) writeConnectionHandshake(conn net.Conn) error {
	agent_addr_bytes := []byte(r.agent_addr)
	return writeLengthPrefixed(conn, agent_addr_bytes)
}

func (r *Reporting) Run(ctx context.Context) {
	log.Println("Reporting triggered trace data to", r.remote_addr)
	firsttime := true
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopped reporting triggered trace data")
			return
		default:
			break
		}
		conn, err := net.Dial("tcp", r.remote_addr)
		if err != nil {
			if firsttime {
				log.Println("Unable to connect to reporting backend, retrying every 2 seconds", r.remote_addr, err)
				firsttime = false
			}
			select {
			case <-ctx.Done():
				log.Println("Stopped reporting triggered trace data")
				return
			case <-time.After(2 * time.Second):
				continue
			}
			continue
		}
		defer conn.Close()

		err = r.ReportData(ctx, conn)
		if err != nil {
			if firsttime {
				log.Println("Error in DataLoop:", err, " -- will retry every 2 seconds")
				firsttime = false
			}
			select {
			case <-ctx.Done():
				log.Println("Stopped reporting triggered trace data")
				return
			case <-time.After(2 * time.Second):
				continue
			}
		}

		firsttime = true
	}
}

func (r *Reporting) ReportData(ctx context.Context, conn net.Conn) (err error) {
	err = r.writeConnectionHandshake(conn)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case buffers := <-r.data:
			err = r.reportData(conn, buffers)
			if err != nil {
				return err
			}
		}
	}
}
