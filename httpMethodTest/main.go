package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	urlPtr := flag.String("u", "http://127.0.0.1", "URL to test with the method")
	methodPtr := flag.String("m", "GET", "Method to test")
	flag.Parse()

	url := *urlPtr
	method := *methodPtr
	const blueColor = "\033[1;34m" // Blue and Bold Text
	const resetColor = "\033[0m"

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Create a DELETE request
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}

	// Perform the request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Analyze the response
	if resp.StatusCode == http.StatusMethodNotAllowed {
		fmt.Printf("%s method is not allowed on this endpoint.\n", method)
	} else if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		fmt.Printf("%s%s method is Allowed and Succeeded%s\n\n", blueColor, method, resetColor)
		fmt.Printf("%sResponse Headers%s\n", blueColor, resetColor)
		for key, values := range resp.Header {
			for _, value := range values {
				fmt.Printf("%s: %s\n", key, value)
			}
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading response body: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\n%sResponse Body%s\n%s\n\n", blueColor, resetColor, body)
	} else {
		fmt.Printf("Received HTTP status code %d. %s method behavior may vary.\n", resp.StatusCode, method)
	}
}
