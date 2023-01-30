package telemetry

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

type generatorForTest struct {
}

func (generator *generatorForTest) Headers() []string {
	return []string{"time", "interval", "id", "c", "a", "f", "b", "d"}
}

func (generator *generatorForTest) NextData(now time.Time, interval time.Duration) []map[string]string {
	var rows []map[string]string
	for i := 0; i < 3; i++ {
		row := make(map[string]string)
		row["id"] = fmt.Sprintf("%d", i)
		row["time"] = fmt.Sprintf("%v", now)
		row["interval"] = fmt.Sprintf("%v", interval)
		if i == 0 {
			row["a"] = "va"
		}
		if i == 2 {
			row["b"] = "vb"
		}
		row["c"] = "vc"
		row["d"] = "vd"
		rows = append(rows, row)
	}
	return rows
}

func TestStdoutReceiver(t *testing.T) {
	// assert := assert.New(t)

	receiver := NewStdoutReceiver(" ")
	generator := new(generatorForTest)
	reporter := new(Reporter)

	reporter.Init(time.Duration(1)*time.Second, generator, receiver)

	ctx, _ := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)

	reporter.Run(ctx)

	fmt.Println("Done")
}

func TestCsvReceiver(t *testing.T) {
	assert := assert.New(t)

	receiver, err := NewCsvReceiver("test.csv")
	assert.NoError(err, "Unable to create test.csv")

	generator := new(generatorForTest)
	reporter := new(Reporter)

	reporter.Init(time.Duration(1)*time.Second, generator, receiver)

	ctx, _ := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)

	reporter.Run(ctx)

	fmt.Println("Done")
}

func TestMultiReceiver(t *testing.T) {
	assert := assert.New(t)

	csvreceiver, err := NewCsvReceiver("multitest.csv")
	assert.NoError(err, "Unable to create multitest.csv")
	stdoutreceiver := NewStdoutReceiver(" ")
	receiver := NewMultiReceiver([]Receiver{csvreceiver, stdoutreceiver})

	generator := new(generatorForTest)
	reporter := new(Reporter)

	reporter.Init(time.Duration(1)*time.Second, generator, receiver)

	ctx, _ := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)

	reporter.Run(ctx)

	fmt.Println("Done")
}
