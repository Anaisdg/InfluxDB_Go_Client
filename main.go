package main

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go"
)

func main() {
	// create new client with default option for server url authenticate by token
	client := influxdb2.NewClient("http://localhost:9999", "my-token")
	// user blocking write client for writes to desired bucket
	writeApi := client.WriteApiBlocking("my-org", "my-bucket")
	// create point using full params constructor
	p := influxdb2.NewPoint("stat",
		map[string]string{"unit": "temperature"},
		map[string]interface{}{"avg": 24.5, "max": 45},
		time.Now())
	// write point immediately
	writeApi.WritePoint(context.Background(), p)
	// create point using fluent style
	p = influxdb2.NewPointWithMeasurement("stat").
		AddTag("unit", "temperature").
		AddField("avg", 23.2).
		AddField("max", 45).
		SetTime(time.Now())
	writeApi.WritePoint(context.Background(), p)

	// Or write directly line protocol
	line := fmt.Sprintf("stat,unit=temperature avg=%f,max=%f", 23.5, 45.0)
	writeApi.WriteRecord(context.Background(), line)

	// get query client
	queryApi := client.QueryApi("my-org")
	// get parser flux query result
	result, err := queryApi.Query(context.Background(), `from(bucket:"my-bucket")|> range(start: -1h) |> filter(fn: (r) => r._measurement == "stat")`)
	if err == nil {
		// Use Next() to iterate over query result lines
		for result.Next() {
			// Observe when there is new grouping key producing new table
			if result.TableChanged() {
				fmt.Printf("table: %s\n", result.TableMetadata().String())
			}
			// read result
			fmt.Printf("row: %s\n", result.Record().String())
		}
		if result.Err() != nil {
			fmt.Printf("Query error: %s\n", result.Err().Error())
		}
	}
	// Ensures background processes finishes
	client.Close()
}
