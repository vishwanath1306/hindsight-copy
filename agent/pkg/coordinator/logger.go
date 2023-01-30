package coordinator

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

type CsvLogger struct {
	filename string
	file     *os.File
	writer   *csv.Writer

	begin            time.Time
	wg               *sync.WaitGroup
	Finished         chan []FinishedTrigger
	dropped_finished int
}

func NewCsvLogger(filename string) (r *CsvLogger, err error) {
	r = new(CsvLogger)
	r.file, err = os.Create(filename)
	if err != nil {
		return
	}
	r.writer = csv.NewWriter(r.file)
	headers := []string{"t", "queue", "total_agents", "dissemination_time_ms"}
	r.writer.Write(headers)
	r.begin = time.Now()
	r.wg = new(sync.WaitGroup)
	r.Finished = make(chan []FinishedTrigger, 1000)
	return
}

func (r *CsvLogger) AwaitCompletion() {
	r.wg.Wait()
}

func (r *CsvLogger) write(finished []FinishedTrigger) {
	now := fmt.Sprintf("%.0f", time.Now().Sub(r.begin).Seconds())
	var rows [][]string
	for _, t := range finished {
		queue := fmt.Sprintf("%d", t.queue_id)
		total_agents := fmt.Sprintf("%d", t.total_agents)
		dissemination_time := fmt.Sprintf("%d", t.dissemination_time.Milliseconds())

		row := []string{now, queue, total_agents, dissemination_time}
		rows = append(rows, row)
	}
	r.writer.WriteAll(rows)
}

func (r *CsvLogger) Run() context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	r.wg.Add(1)
	go func() {
		log.Println("Logger goroutine running")
		for {
			select {
			case <-ctx.Done():
				{
					log.Println("Logger draining remaining stats to file")
					for {
						select {
						case f := <-r.Finished:
							r.write(f)
						default:
							log.Println("Logger complete")
							r.writer.Flush()
							err := r.file.Close()
							if err != nil {
								fmt.Println("Logger error closing file", err)
							}
							r.wg.Done()
							return
						}
					}
				}
			case f := <-r.Finished:
				r.write(f)
			}
		}
	}()
	return cancel
}

// func (r *CsvLogger) Report() error {
// 	var records [][]string
// 	for _, row := range rows {
// 		var record []string
// 		for _, column := range r.headers {
// 			if v, ok := row[column]; ok {
// 				record = append(record, v)
// 			} else {
// 				record = append(record, "")
// 			}
// 		}
// 		records = append(records, record)
// 	}
// 	return r.writer.WriteAll(records)
// }
