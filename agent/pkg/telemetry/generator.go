package telemetry

import "time"

/* Interface for generating telemetry.  Hindsight's
agent, coordinator, and collector all implement this interface
to provide telemetry data */
type Generator interface {
	Headers() []string
	NextData(now time.Time, interval time.Duration) []map[string]string
}
