package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	var token string

	flag.StringVar(&token, "token", "", "--token <token>")

	flag.Parse()

	if token == "" {
		log.Fatalln("please add token")
		os.Exit(1)
	}

	// 1. Create a new GET request
	url := "https://api.track.toggl.com/api/v9/me"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}

	// 2. Set Basic Auth (API Token as username, "api_token" as password)
	// Equivalent to curl -u <token>:api_token
	req.SetBasicAuth(token, "api_token")

	// 3. Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	// 4. Read and print the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}

	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Body: %s\n", string(body))
}
