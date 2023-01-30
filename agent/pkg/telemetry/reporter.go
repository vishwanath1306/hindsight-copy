package telemetry

import (
	"context"
	"fmt"
	"time"
)

/* Ties generators to receivers; reports telemetry with a configurable interval */
type Reporter struct {
	interval  time.Duration
	generator Generator
	receiver  Receiver
}

func (reporter *Reporter) Init(interval time.Duration, generator Generator, receiver Receiver) {
	reporter.interval = interval
	reporter.generator = generator
	reporter.receiver = receiver
}

func (reporter *Reporter) Run(ctx context.Context) (err error) {
	// User must call Init before calling Run
	if reporter.generator == nil || reporter.receiver == nil {
		return fmt.Errorf("Attempted to run an uninitialized reporter")
	}
	// First write the headers
	err = reporter.receiver.Init(reporter.generator.Headers())
	if err != nil {
		return
	}

	// Write data forever
	ticker := time.NewTicker(reporter.interval)
	last_report := time.Now()
	for {
		select {
		case <-ctx.Done():
			err = reporter.receiver.Close()
			return
		case <-ticker.C:
			now := time.Now()
			interval := now.Sub(last_report)
			err = reporter.receiver.Report(reporter.generator.NextData(now, interval))
			if err != nil {
				return
			}
			last_report = now
		}
	}
}
