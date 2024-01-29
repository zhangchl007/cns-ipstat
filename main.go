package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

type Query struct {
	QueryString string
	Client      api.Client
}

type QueryResult struct {
	Query  string
	Result model.Value
	Err    error
}

func Prometheus_Query(query Query) QueryResult {
	v1api := v1.NewAPI(query.Client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, warnings, err := v1api.Query(ctx, query.QueryString, time.Now(), v1.WithTimeout(5*time.Second))
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}
	return QueryResult{Query: query.QueryString, Result: result, Err: err}
}

func QueryWorker(queries <-chan Query, results chan<- QueryResult) {
	for query := range queries {
		results <- Prometheus_Query(query)
	}
}

func main() {
	client, err := api.NewClient(api.Config{
		Address: "http://127.0.0.1:9090",
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}

	// List of queries
	queries := []Query{
		{Client: client, QueryString: "cx_ipam_pod_allocated_ips"},
		{Client: client, QueryString: "sum(kube_node_info)"},
		// Add more queries here
	}

	// Create channels
	queriesChan := make(chan Query, len(queries))
	resultsChan := make(chan QueryResult, len(queries))

	// Start workers
	for w := 0; w < 10; w++ {
		go QueryWorker(queriesChan, resultsChan)
	}

	// Send queries to workers
	for _, query := range queries {
		queriesChan <- query
	}
	close(queriesChan)

	// Collect results
	for range queries {
		result := <-resultsChan
		if result.Err != nil {
			fmt.Printf("Error querying %s: %v\n", result.Query, result.Err)
			continue
		}
		// Process result here

		result1 := result.Result
		switch {
		case result1.Type() == model.ValScalar:
			fmt.Printf("Scalar: %v\n", result)
		case result1.Type() == model.ValVector:
			var podallocatedip_sum = make(map[string]int)
			var subnet_cidr = make(map[string]string)
			for _, v := range result1.(model.Vector) {
				if v.Metric["subnet"] != "" {
					podallocatedip_sum[fmt.Sprintf("%s", v.Metric["subnet"])] += int(v.Value)
					subnet_cidr[fmt.Sprintf("%s", v.Metric["subnet"])] = fmt.Sprintf("%s", v.Metric["subnet_cidr"])
				} else {
					node_number := fmt.Sprintf("%s", v.Value)
					fmt.Printf("Node number: %v\n", node_number)
				}
			}
			if len(podallocatedip_sum) > 0 {
				fmt.Printf("Vector: subnet_cidr: %v,%s\n", podallocatedip_sum, subnet_cidr)
			}
		case result1.Type() == model.ValMatrix:
			for _, v := range result1.(model.Matrix) {
				fmt.Printf("Matrix: %v\n", v.Values)
			}
		case result1.Type() == model.ValString:
			fmt.Printf("String: %v\n", result1)
		case result1.Type() == model.ValNone:
			fmt.Printf("None: %v\n", result1)
		default:
			fmt.Printf("Unknown: %v\n", result1)

		}
	}
}
