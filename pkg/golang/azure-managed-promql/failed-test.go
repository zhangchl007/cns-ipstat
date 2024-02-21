package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/config"
)

const subscriptionID = "ffa067fd-36c1-4774-a161-7ebdac9a934f"

func main() {

	// Define Azure service principal credentials
	clientID := "xxxx"
	clientsecret := "xxxx"
	tenantID := "xxxxx"

	// Create a new service principal credential
	cred, err := azidentity.NewClientSecretCredential(tenantID, clientID, clientsecret, nil)

	if err != nil {
		panic(err)
	}

	// Get aad oauth token
	token, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{Scopes: []string{"https://management.core.windows.net//.default"}})
	if err != nil {
		fmt.Printf("Error getting token: %v\n", err)
		os.Exit(1)
	}

	// Print the token
	fmt.Printf("Token: %s\n", token.Token)

	// Create a new Prometheus API client with the token

	client, err := api.NewClient(api.Config{
		Address:      "http://localhost:9090",
		RoundTripper: config.NewAuthorizationCredentialsRoundTripper("Bearer", config.Secret(token.Token), api.DefaultRoundTripper),
	})

	if err != nil {
		fmt.Printf("Error creating client: %v\n", err)
		os.Exit(1)
	}

	// query the Prometheus API
	v1api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, warnings, err := v1api.Query(ctx, "up", time.Now())

	if err != nil {
		fmt.Printf("Error querying Prometheus: %v\n", err)
		os.Exit(1)
	}
	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}
	fmt.Printf("Result:\n%v\n", result)

}
