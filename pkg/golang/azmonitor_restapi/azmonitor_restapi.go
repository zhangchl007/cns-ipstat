package azmonitor_restapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type AuthToken struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	ExtExpiresIn int    `json:"ext_expires_in"`
	TokenType    string `json:"token_type"`
}

var httpClient = &http.Client{}

func main() {
	// get the auth token
	authToken, err := getAuthToken()
	if err != nil {
		log.Fatalf("Failed to get auth token: %v", err)
	}

	// get the metrics
	err = getMetrics(authToken)
	if err != nil {
		log.Fatalf("Failed to get metrics: %v", err)
	}
}

func getAuthToken() (string, error) {
	// get the auth token
	authToken := AuthToken{}
	url := fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/token", os.Getenv("TENANT_ID"))
	payload := bytes.NewBufferString(fmt.Sprintf("grant_type=client_credentials&client_id=%s&client_secret=%s&resource=https://monitoring", os.Getenv("CLIENT_ID"), os.Getenv("CLIENT_SECRET")))

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(body, &authToken)
	if err != nil {
		return "", err
	}

	return authToken.AccessToken, nil
}

// Assuming getMetrics function
func getMetrics(authToken string) error {
	// Implementation here
	return nil
}
