package collector

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/geraldleizhang/hindsight/agent/pkg/memory"
)

type Collector struct {
	tracefile string
	port      string
	incoming  chan *ReceivedBuffer
}

func (c *Collector) Init(port string, tracefile string) {
	c.tracefile = tracefile
	c.port = port
	c.incoming = make(chan *ReceivedBuffer, 1000)
}

func (c *Collector) Run(ctx context.Context) {
	fmt.Println("Collector listening on TCP port", c.port)
	listener, err := net.Listen("tcp", ":"+c.port)
	if err != nil {
		fmt.Println("Error listening on port", c.port, err)
		return
	}

	if c.tracefile != "" {
		log.Println("Writing trace data to", c.tracefile)
		go c.fileWriter(c.tracefile)
	} else {
		log.Println("Not writing trace data to disk")
		go c.printer()
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting new connection", err)
			return
		}
		go c.handleConnection(conn)
	}
}

func doRead(conn net.Conn, dst []byte) error {
	for len(dst) > 0 {
		read, err := conn.Read(dst)
		if err != nil {
			return err
		}

		dst = dst[read:]
	}
	return nil
}

func readLengthPrefixed(conn net.Conn) (buf []byte, err error) {
	szbuf := make([]byte, 4)
	err = doRead(conn, szbuf)
	if err != nil {
		return
	}
	sz := binary.LittleEndian.Uint32(szbuf)
	buf = make([]byte, sz)
	err = doRead(conn, buf)
	return
}

func (c *Collector) handleConnection(conn net.Conn) {
	buf, err := readLengthPrefixed(conn)
	if err != nil {
		fmt.Println("Error receiving handshake from new agent connection", err)
		return
	}
	agent_addr := string(buf)
	fmt.Println("New connection from", agent_addr)

	for {
		buf, err := readLengthPrefixed(conn)
		if err != nil {
			fmt.Println("Error in handleConnection receiving next buffer", err)
			return
		}
		if len(buf) >= 32 { // 32 is size of buffer header
			header := memory.ExtractBufferHeader(buf)
			// header := memory.ExtractBufferHeader(buf)
			// fmt.Println("Read buf of size", len(buf), header)
			// fmt.Println("Buffer ID was", header.Buffer_id, "Prev Buffer ID was", header.Prev_buffer_id)
			var r ReceivedBuffer
			r.trace_id = header.Trace_id
			r.source_agent = agent_addr
			r.buffer = buf
			c.incoming <- &r
		}
	}
}

func (c *Collector) fileWriter(filename string) {
	f, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating /local/buffers.out: ", err)
		return
	}
	ticker := time.NewTicker(1 * time.Second)
	count := 0
	last_report := time.Now()
	for {
		select {
		case <-ticker.C:
			{
				now := time.Now()
				interval := now.Sub(last_report)
				last_report = now
				tput := (float64(count) / interval.Seconds()) / (1024 * 1024)
				log.Printf("%.2f MB/s\n", tput)
				count = 0
			}
		case r := <-c.incoming:
			{
				count += len(r.buffer)
				err = r.WriteToFile(f)
				if err != nil {
					fmt.Println("Error writing buffer to file: ", err)
				}
			}
		}
	}
}

func (c *Collector) printer() {
	ticker := time.NewTicker(1 * time.Second)
	count := 0
	last_report := time.Now()
	for {
		select {
		case <-ticker.C:
			{
				now := time.Now()
				interval := now.Sub(last_report)
				last_report = now
				tput := (float64(count) / interval.Seconds()) / (1024 * 1024)
				log.Printf("%.2f MB/s\n", tput)
				count = 0
			}
		case r := <-c.incoming:
			{
				count += len(r.buffer)
			}
		}
	}
}
