package main

import (
	"context"
	"fmt"
	"os"
	"sync"
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

func Prometheus_Query(query Query, results chan<- QueryResult, wg *sync.WaitGroup) {
	defer wg.Done()
	v1api := v1.NewAPI(query.Client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, warnings, err := v1api.Query(ctx, query.QueryString, time.Now(), v1.WithTimeout(5*time.Second))
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	results <- QueryResult{Query: query.QueryString, Result: result, Err: err}
}

func processResult(result QueryResult) (int, map[string]string, map[string]int) {
	result1 := result.Result
	switch {
	case result1.Type() == model.ValScalar:
		fmt.Printf("Scalar: %v\n", result)
	case result1.Type() == model.ValVector:
		return processVectorResult(result1)
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
	return 0, nil, nil
}

func processVectorResult(result model.Value) (int, map[string]string, map[string]int) {
	var podallocatedip_sum = make(map[string]int)
	var subnet_cidr = make(map[string]string)
	var nodeNumber int
	for _, v := range result.(model.Vector) {
		if v.Metric["subnet"] != "" {
			podallocatedip_sum[fmt.Sprintf("%s", v.Metric["subnet"])] += int(v.Value)
			subnet_cidr[fmt.Sprintf("%s", v.Metric["subnet"])] = fmt.Sprintf("%s", v.Metric["subnet_cidr"])
		}

		if v.Metric.String() == "{}" {
			nodeNumber += int(v.Value)
		}

	}
	return nodeNumber, subnet_cidr, podallocatedip_sum
}

func main() {
	var podallocatedip_sum = make(map[string]int)
	var subnet_cidr = make(map[string]string)
	var nodeNumber, podip_maxmum int
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

	// Collect results
	results := make(chan QueryResult, len(queries))
	var wg sync.WaitGroup
	for _, query := range queries {
		wg.Add(1)
		go Prometheus_Query(query, results, &wg)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		nodeNumber1, subnet_cidr1, podallocatedip_sum1 := processResult(result)
		if nodeNumber1 > 0 {
			nodeNumber = nodeNumber1
		} else {
			podallocatedip_sum = podallocatedip_sum1
			subnet_cidr = subnet_cidr1
		}
	}

	fmt.Printf("the number of nodes: %v\n", nodeNumber)
	fmt.Printf("the pod ips allocated in cidr: %v, subnet_cidr: %s\n", podallocatedip_sum, subnet_cidr)

	for _, v := range podallocatedip_sum {
		podip_maxmum += int(v)
	}

	podip_maxmum = podip_maxmum + nodeNumber + nodeNumber*8
	fmt.Printf("the possible maxmum number of pod ips:%v\n", podip_maxmum)

}
