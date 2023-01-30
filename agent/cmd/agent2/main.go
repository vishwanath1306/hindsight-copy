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

	"github.com/geraldleizhang/hindsight/agent/pkg/agent"
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

func resolveConfigValue(key string, value string, legacyconfigvalue string, defaultvalue string, service_name string) string {
	if value == "" {
		value = legacyconfigvalue

		if value == "" || value == ":" {
			value = defaultvalue
			fmt.Printf("  %s=%s (default fallback)\n", key, value)
		} else {
			fmt.Printf("  %s=%s (%s.conf)\n", key, value, service_name)
		}
	} else {
		fmt.Printf("  %s=%s (command line)\n", key, value)
	}
	return value
}

// TODO different main methods for different cmds..........
func main() {

	serv := flag.String("serv", "", "Service name")
	hostname := flag.String("host", "", "Hostname or IP of this agent.  If not specified, uses `addr` from the legacy config file")
	port := flag.String("port", "", "Port to run the agent on.  If not specified, uses `port` from the legacy config file.")
	lc_addr := flag.String("lc", "", "Address of the log collector in form hostname:port.  If not specified, uses `lc_addr`:`lc_port` from the legacy config file.")
	r_addr := flag.String("r", "", "Address of the reporting backend in form hostname:port.  If not specified, uses `r_addr`:`r_port` from the legacy config file.")
	// isReport := flag.Bool("report", true, "If report to LC (or local mode)")
	delayf := flag.Int("delay", 0, "Used for experimental purposes.  If specified, this delays the reporting of triggers by the specified delay (in milliseconds).  Default to 0 - no delay.")
	reportingratelimit := flag.Float64("rate", 0, "Rate limit for reporting traces in MB/s.  Set to 0 to disable.  Default 0.")
	triggerratelimit := flag.Float64("triggerrate", 10000, "Rate limit for a spammy trigger in triggers/s.  Set to 0 to disable.  Default 10000.")
	outputfile := flag.String("output", "", "Filename for outputting agent telemetry.  If specified, will write a csv of agent telemetry data.  Disabled by default.")
	verbose := flag.Bool("verbose", false, "If set to true, prints telemetry to the command line.  False by default.")

	per_trigger_limits := make(triggerRateLimitFlags)
	flag.Var(&per_trigger_limits, "l", "A per-trigger reporting rate limit in the form queue_id,rate where queue_id is an integer and rate is a float representing a reporting limit in MB/s.  This flag can be set multiple times to provide rate limits for different triggers.")

	flag.Parse()

	delay := uint64((*delayf))

	isConfig := util.Conf_init(*serv)
	if !isConfig {
		fmt.Println("Failed to load config file for", *serv)
		return
	}

	log.Println("Running agent", *serv)
	*hostname = resolveConfigValue("hostname", *hostname, util.Server_addr, "127.0.0.1", *serv)
	*port = resolveConfigValue("port", *port, util.Server_port, "5050", *serv)
	*lc_addr = resolveConfigValue("lc_addr", *lc_addr, util.Coordinator_addr+":"+util.Coordinator_port, "127.0.0.1:5252", *serv)
	*r_addr = resolveConfigValue("r_addr", *r_addr, util.Reporting_addr+":"+util.Reporting_port, "127.0.0.1:5253", *serv)

	ctx, cancel := context.WithCancel(context.Background())

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
			os.Exit(0)
		}
	}()

	agent := agent.InitAgent2(*serv, *hostname, *port, *lc_addr, *r_addr, delay, *reportingratelimit, *triggerratelimit, per_trigger_limits, *outputfile, *verbose)
	agent.Run(ctx, cancel)
	log.Println("Agent exiting")
	os.Exit(0)
}
