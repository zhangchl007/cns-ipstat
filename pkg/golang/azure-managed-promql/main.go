package main

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "os"
    "strings"
    "time"

    "github.com/prometheus/client_golang/api"
    v1 "github.com/prometheus/client_golang/api/prometheus/v1"
    "github.com/prometheus/common/config"
)

const subscriptionID = "ffa067fd-36c1-4774-a161-7ebdac9a934f"

var httpClient = &http.Client{}

func handleError(err error) {
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
}

func main() {
    authendpoint := os.Getenv("AUTH_ENDPOINT")
    clientID := os.Getenv("CLIENT_ID")
    clientSecret := os.Getenv("CLIENT_SECRET")

    body := url.Values(map[string][]string{
        "resource":      {"https://prometheus.monitor.azure.com"},
        "client_id":     {clientID},
        "client_secret": {clientSecret},
        "grant_type":    {"client_credentials"}})

    type ResponseData struct {
        AccessToken string `json:"access_token"`
    }

    request, err := http.NewRequest(
        http.MethodPost,
        authendpoint,
        strings.NewReader(body.Encode()))
    handleError(err)

    request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    resp, err := httpClient.Do(request)
    handleError(err)
    defer resp.Body.Close()

    data := new(ResponseData)
    err = json.NewDecoder(resp.Body).Decode(data)
    handleError(err)

    client, err := api.NewClient(api.Config{
        Address:      "https://azuremonitor-xyk3.eastus.prometheus.monitor.azure.com",
        RoundTripper: config.NewAuthorizationCredentialsRoundTripper("Bearer", config.Secret(data.AccessToken), api.DefaultRoundTripper),
    })
    handleError(err)

    v1api := v1.NewAPI(client)
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    result, warnings, err := v1api.Query(ctx, "sum(kube_node_info)", time.Now(), v1.WithTimeout(5*time.Second))
    handleError(err)

    if len(warnings) > 0 {
        fmt.Printf("Warnings: %v\n", warnings)
    }
    fmt.Printf("Result:\n%v\n", result)
}
