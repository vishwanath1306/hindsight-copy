package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/geraldleizhang/hindsight/agent/pkg/collector"
	"github.com/geraldleizhang/hindsight/agent/pkg/util"
)

func resolveConfigValue(key string, value string, legacyconfigvalue string, defaultvalue string, service_name string) string {
	if value == "" {
		value = legacyconfigvalue

		if value == "" {
			value = defaultvalue
			fmt.Printf("  %s=%s (default)\n", key, value)
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

	tracefile := flag.String("out", "", "Filename to write trace data to.  If not specified, trace data won't be written to disk.  If you're at MPI, don't write to your home directory!")
	port := flag.String("port", "", "Collector port.  If not specified, uses `r_port` from the legacy config lc.conf file, or 5253 as a backup")

	flag.Parse()

	isConfig := util.Conf_init("lc")
	if !isConfig {
		fmt.Println("Failed to load config file for lc")
		return
	}

	fmt.Println("Running coordinator")
	*port = resolveConfigValue("port", *port, util.Reporting_port, "5253", "lc")

	ctx, _ := context.WithCancel(context.Background())

	var c collector.Collector
	c.Init(*port, *tracefile)
	c.Run(ctx)
}
