package main

import (
	"context"
	"fmt"
	"os"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/common/config"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

const subscriptionID = "ffa067fd-36c1-4774-a161-7ebdac9a934f"

func main() {

	// Get aad access token by rest api

	authendpoint := "https://login.microsoftonline.com/16b3c013-d300-468d-ac64-7eda0820b6d3/oauth2/token"
	body := url.Values(map[string][]string{
		"resource":      {"https://prometheus.monitor.azure.com"},
		"client_id":     {"660ccd58-9a5f-4199-9060-6cfc7e7d5882"},
		"client_secret": {"H1-8Q~1vD-u3s0rj~Jopw91qS7rxt5zkSitsHdpY"},
		"grant_type":    {"client_credentials"}})
	
	type ResponseData struct {
	    AccessToken string `json:"access_token"`
	}

	// Create a new request
	request, err := http.NewRequest(
		http.MethodPost,
		authendpoint,
		strings.NewReader(body.Encode()))
	if err != nil {
		panic(err)
	}
	// Set the content type to x-www-form-urlencoded
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	cs := &http.Client{}
	resp, err := cs.Do(request)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	        decoder := json.NewDecoder(resp.Body)
        if err != nil {
                panic(err)
        }

        // Create a new ResponseData to hold the decoded data
        data := new(ResponseData)

        // Decode the response body into the ResponseData
        err = decoder.Decode(data)

        //fmt.Printf("Access Token: %v\n", data.AccessToken)
	// Create a new Prometheus API client

	client, err := api.NewClient(api.Config{
		Address: "https://azuremonitor-xyk3.eastus.prometheus.monitor.azure.com",
		RoundTripper: config.NewAuthorizationCredentialsRoundTripper("Bearer", config.Secret(data.AccessToken), api.DefaultRoundTripper),
	})

	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}

	// query the Prometheus API
	v1api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, warnings, err := v1api.Query(ctx, "sum(kube_node_info)", time.Now(),v1.WithTimeout(5*time.Second))

	if err != nil {
		fmt.Printf("Error querying Prometheus: %v\n", err)
		os.Exit(1)
	}
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}
	fmt.Printf("Result:\n%v\n", result)

}

