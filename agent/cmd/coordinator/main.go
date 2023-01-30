package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/geraldleizhang/hindsight/agent/pkg/coordinator"
	"github.com/geraldleizhang/hindsight/agent/pkg/util"
)

type triggerRateLimitFlags map[int]float64

func (rates *triggerRateLimitFlags) String() string {
	var b strings.Builder
	for trigger_id, rate := range *rates {
		fmt.Fprintf(&b, "%d=%.1f ", trigger_id, rate)
	}
	return b.String()
}

func (i *triggerRateLimitFlags) Set(value string) error {
	splits := strings.Split(value, ",")
	if len(splits) != 2 {
		return fmt.Errorf("Invalid rate %v -- must be of the form int,float", value)
	}
	trigger_id, err := strconv.ParseInt(splits[0], 10, 64)
	if err != nil {
		return err
	}
	rate, err := strconv.ParseFloat(splits[1], 64)
	if err != nil {
		return err
	}

	(*i)[int(trigger_id)] = rate
	return nil
}

func resolveConfigValue(key string, value string, legacyconfigvalue string, service_name string) string {
	if value == "" {
		value = legacyconfigvalue
		fmt.Printf("  %s=%s (%s.conf)\n", key, value, service_name)
	} else {
		fmt.Printf("  %s=%s (command line)\n", key, value)
	}
	return value
}

// TODO different main methods for different cmds..........
func main() {

	port := flag.String("port", "5252", "Coordinator port.  If not specified, uses `lc_port` from the legacy config lc.conf file.")
	outfile := flag.String("out", "", "Output filename for writing breadcrumb dissemination statistics.  If not specified, will not be written to file")

	flag.Parse()

	isConfig := util.Conf_init("lc")
	if !isConfig {
		log.Println("Failed to load config file for lc")
		return
	}

	log.Println("Running coordinator")
	*port = resolveConfigValue("port", *port, util.Server_port, "lc")

	ctx, cancel := context.WithCancel(context.Background())

	// // Not sure if needed
	// util.Conf_init("lc")

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		log.Println("Initiating graceful shutdown...")

		go func() {
			<-ch
			log.Println("Exiting without graceful shutdown")
			os.Exit(0)
		}()

		cancel()
		select {
		case <-time.After(5 * time.Second):
			log.Println("Shutdown timeout expired, exiting")
			os.Exit(0)
		}
	}()

	if *outfile != "" {
		log.Println("Logging breadcrumb stats to", *outfile)
	}

	var c coordinator.CoordinatorServer
	err := c.Init(*port, *outfile)

	if err != nil {
		fmt.Println("Error initializing coordinator:", err)
	} else {
		c.Run(ctx)
	}
}
