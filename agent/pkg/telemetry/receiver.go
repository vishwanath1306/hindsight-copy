package telemetry

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

/* Interface for recipient of telemetry.  Hindsight processes
have flags for specifying whether to report telemetry to a local file,
to stdout, or over the network.  Each option is a different
Receiver implementation */
type Receiver interface {
	Init(headers []string) error
	Close() error
	Report(rows []map[string]string) error
}

/* Drops all telemetry */
type NullReceiver struct {
}

func (r *NullReceiver) Init(headers []string) error {
	return nil
}

func (r *NullReceiver) Close() error {
	return nil
}

func (r *NullReceiver) Report(rows []map[string]string) error {
	return nil
}

type MultiReceiver struct {
	receivers []Receiver
}

func NewMultiReceiver(receivers []Receiver) *MultiReceiver {
	var r MultiReceiver
	r.receivers = receivers
	return &r
}

func (r *MultiReceiver) Init(headers []string) error {
	var errors []error
	for _, receiver := range r.receivers {
		err := receiver.Init(headers)
		if err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("%d errors initializing MultiReceiver, %v", len(errors), errors)
	}
	return nil
}

func (r *MultiReceiver) Close() error {
	var errors []error
	for _, receiver := range r.receivers {
		err := receiver.Close()
		if err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("%d errors closing MultiReceiver, %v", len(errors), errors)
	}
	return nil
}

func (r *MultiReceiver) Report(rows []map[string]string) error {
	var errors []error
	for _, receiver := range r.receivers {
		err := receiver.Report(rows)
		if err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("%d errors reporting MultiReceiver, %v", len(errors), errors)
	}
	return nil
}

type StdoutReceiver struct {
	headers   []string
	separator string
}

func NewStdoutReceiver(separator string) *StdoutReceiver {
	var r StdoutReceiver
	r.separator = separator
	return &r
}

func (r *StdoutReceiver) Init(headers []string) error {
	r.headers = headers
	fmt.Println("TelemetryHeaders:", strings.Join(headers, r.separator))
	return nil
}

func (r *StdoutReceiver) Close() error {
	return nil
}

func (r *StdoutReceiver) Report(rows []map[string]string) error {
	for _, row := range rows {
		var values []string
		for _, column := range r.headers {
			if value, ok := row[column]; ok {
				values = append(values, value)
			} else {
				values = append(values, "")
			}
		}
		fmt.Println("Telemetry:", strings.Join(values, r.separator))
	}
	return nil
}

type CsvReceiver struct {
	filename string
	file     *os.File
	writer   *csv.Writer

	headers []string
}

func NewCsvReceiver(filename string) (r *CsvReceiver, err error) {
	r = new(CsvReceiver)
	r.file, err = os.Create(filename)
	if err != nil {
		return
	}
	r.writer = csv.NewWriter(r.file)
	return
}

func (r *CsvReceiver) Init(headers []string) error {
	r.headers = headers
	return r.writer.Write(headers)
}

func (r *CsvReceiver) Close() error {
	return r.file.Close()
}

func (r *CsvReceiver) Report(rows []map[string]string) error {
	var records [][]string
	for _, row := range rows {
		var record []string
		for _, column := range r.headers {
			if v, ok := row[column]; ok {
				record = append(record, v)
			} else {
				record = append(record, "")
			}
		}
		records = append(records, record)
	}
	return r.writer.WriteAll(records)
}
