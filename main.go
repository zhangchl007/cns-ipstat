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

func main() {
	client, err := api.NewClient(api.Config{
		Address: "http://127.0.0.1:9090",
	})
	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}

	v1api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, warnings, err := v1api.Query(ctx, "cx_ipam_pod_allocated_ips", time.Now(), v1.WithTimeout(5*time.Second))
	if err != nil {
		fmt.Printf("Error querying Prometheus: %v\n", err)
		os.Exit(1)
	}
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}

	fmt.Printf("Result:\n%v\n", result)
	switch {
	case result.Type() == model.ValScalar:
		fmt.Printf("Scalar: %v\n", result)
	case result.Type() == model.ValVector:
		var sum = make(map[string]int)
		var subnet_cidr = make(map[string]string)
		for _, v := range result.(model.Vector) {
			sum[fmt.Sprintf("%s", v.Metric["subnet"])] += int(v.Value)
			subnet_cidr[fmt.Sprintf("%s", v.Metric["subnet"])] = fmt.Sprintf("%s", v.Metric["subnet_cidr"])
		}
		fmt.Printf("Vector: subnet_cidr: %v,%s\n", sum, subnet_cidr)
	case result.Type() == model.ValMatrix:
		for _, v := range result.(model.Matrix) {
			fmt.Printf("Matrix: %v\n", v.Values)
		}
	case result.Type() == model.ValString:
		fmt.Printf("String: %v\n", result)
	case result.Type() == model.ValNone:
		fmt.Printf("None: %v\n", result)
	default:
		fmt.Printf("Unknown: %v\n", result)

	}

}
