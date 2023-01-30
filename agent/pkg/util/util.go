package util

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

var DEBUG int

// addresses and message queues, for server and log collector
var Server_addr string
var Server_port string
var Coordinator_addr string
var Coordinator_port string
var Reporting_addr string
var Reporting_port string

func Conf_init(service_name string) bool {
	conf_file, err := os.Open("/etc/hindsight_conf/" + service_name + ".conf")
	if err != nil {
		conf_file, err = os.Open("/etc/hindsight_conf/default.conf")
		if err != nil {
			fmt.Println("Please check conf file")
			return false
		}
	}
	defer conf_file.Close()

	scanner := bufio.NewScanner(conf_file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		///// cap is deprecated; now read direct from shm
		// if strings.Contains(scanner.Text(), "cap") {
		// 	Cap, _ = strconv.Atoi(strings.Split(scanner.Text(), " ")[1])
		// }
		///// buf_length is deprecated; now read direct from shm
		// if strings.Contains(scanner.Text(), "buf_length") {
		// 	Buf_length, _ = strconv.Atoi(strings.Split(scanner.Text(), " ")[1])
		// }
		if strings.Contains(scanner.Text(), "addr") && !strings.Contains(scanner.Text(), "lc_") && !strings.Contains(scanner.Text(), "r_") {
			Server_addr = strings.Split(scanner.Text(), " ")[1]
		}
		if strings.Contains(scanner.Text(), "port") && !strings.Contains(scanner.Text(), "lc_") && !strings.Contains(scanner.Text(), "r_") {
			Server_port = strings.Split(scanner.Text(), " ")[1]
		}
		if strings.Contains(scanner.Text(), "lc_addr") {
			Coordinator_addr = strings.Split(scanner.Text(), " ")[1]
		}
		if strings.Contains(scanner.Text(), "lc_port") {
			Coordinator_port = strings.Split(scanner.Text(), " ")[1]
		}
		if strings.Contains(scanner.Text(), "r_addr") {
			Reporting_addr = strings.Split(scanner.Text(), " ")[1]
		}
		if strings.Contains(scanner.Text(), "r_port") {
			Reporting_port = strings.Split(scanner.Text(), " ")[1]
		}
	}

	return true
}

type MessageQueue struct {
	Req   map[int64]int
	Mutex sync.RWMutex
}

type ReportQueue struct {
	Req   map[int64]int64
	Mutex sync.RWMutex
}

type RetrievalQueue struct {
	Req   map[int64]map[string]int
	Mutex sync.RWMutex
}

// tools

func Int32ToBytes(data int32) []byte {
	bytebuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytebuf, uint32(data))
	return bytebuf
}

func BytesToInt32(bys []byte) int32 {
	return int32(binary.LittleEndian.Uint32(bys))
}

func Int64ToBytes(data int64) []byte {
	bytebuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytebuf, uint64(data))
	return bytebuf
}

func BytesToInt64(bys []byte) int64 {
	return int64(binary.LittleEndian.Uint64(bys))
}

func IsFileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func GetTime() int64 {
	now := time.Now()      // current local time
	nsec := now.UnixNano() // number of nanoseconds since January 1, 1970 UTC
	return nsec
}
