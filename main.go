package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

const targetIP = "10.49.122.144"
const usernameParam = "Magicred1"
const usernameSecret = "7a6701b14d499914cb69f95f4d64df6523054d343fc4086e7abed8f50284a26c"
const maxConcurrentRequests = 100 // Adjust the maximum number of concurrent requests

func main() {
	var wg sync.WaitGroup
	openPorts := make(chan int) // Buffered channel for open ports

	// Create a worker pool
	for i := 0; i < maxConcurrentRequests; i++ {
		go worker(openPorts, &wg)
	}

	// Continuous port scanning
	for {
		for port := 1; port <= 65535; port++ {
			wg.Add(1)
			go func(port int) {
				defer wg.Done()
				if isOpen(targetIP, port) {
					openPorts <- port
				}
			}(port)
		}

		// Wait for all payloads to be sent
		wg.Wait()
	}
}

func worker(openPorts chan int, wg *sync.WaitGroup) {
	for openPort := range openPorts {
		// send ping GET request
		pingResponse, err := sendGetPayload(targetIP, openPort, "/ping")
		if err != nil && pingResponse["status"] != "success" {
			log.Printf("Error sending ping request: %v", err)
		}

		// send signup POST request
		signupResponse, err := sendPayload(targetIP, openPort, "/signup")
		if err != nil && signupResponse["status"] != "success" {
			log.Printf("Error sending signup request: %v", err)
		}

		fmt.Printf("Signup Response: %+v\n", signupResponse)

		checkResponse, err := sendPayload(targetIP, openPort, "/check")
		if err != nil && checkResponse["status"] != "success" {
			log.Printf("Error sending check request: %v", err)
		}

		fmt.Printf("Check Response: %+v\n", checkResponse)

		// Really don't feel like working today huh..
		getUserSecretResponse, err := sendPayload(targetIP, openPort, "/getUserSecret")
		if err != nil && getUserSecretResponse["status"] != "success" {
			log.Printf("Error sending getUserSecret request: %v", err)
		}

		fmt.Printf("Get User Secret Response: %s", getUserSecretResponse)

		getUserLevelResponse, err := sendPayloadWithSecret(targetIP, openPort, "/getUserLevel", getUserSecretResponse)
		if err != nil {
			log.Printf("Error sending getUserLevel request: %v", err)
		}

		fmt.Printf("Get User Level Response: %+v\n", getUserLevelResponse)

		// fmt.Printf("Get User Level Response: %+v\n", getUserLevelResponse)

		getUserPointsResponse, err := sendPayloadWithSecret(targetIP, openPort, "/getUserPoints", getUserSecretResponse)
		if err != nil {
			log.Printf("Error sending getUserPoints request: %v", err)
		}

		fmt.Printf("Get User Points Response: %+v\n", getUserPointsResponse)

		iNeedAHint, err := sendPayloadWithSecret(targetIP, openPort, "/iNeedAHint", getUserSecretResponse)
		if err != nil {
			log.Printf("Error sending iNeedAHint request: %v", err)
		}

		fmt.Printf("iNeedAHint Response: %+v\n", iNeedAHint)

		enterChallengeResponse, err := sendPayloadWithSecret(targetIP, openPort, "/enterChallenge", getUserSecretResponse)
		if err != nil {
			log.Printf("Error sending enterChallenge request: %v", err)
		}

		fmt.Printf("Enter Challenge Response: %+v\n", enterChallengeResponse)

		// send login POST request
		submitSolutionResponse, err := sendPayloadSolution(targetIP, openPort, "/submitSolution")
		if err != nil {
			log.Printf("Error sending submitSolution request: %v", err)
		}

		fmt.Printf("Submit Solution Response: %+v\n", submitSolutionResponse)

	}
	wg.Done() // Decrement the WaitGroup and exit this goroutine
}

// isOpen checks if a port is open on a target IP address
func isOpen(host string, port int) bool {
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

// sendPayload sends a POST request to a specified endpoint with a JSON body
func sendPayload(host string, port int, endpoint string) (map[string]interface{}, error) {
	url := fmt.Sprintf("http://%s:%d%s", host, port, endpoint)

	// Create the JSON request body
	requestBody := map[string]string{"User": usernameParam}
	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("Failed to marshal JSON request body: %v", err)
		return nil, err
	}

	// Send the POST request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		log.Printf("Failed to send POST request to %s:%d: %v", host, port, err)
		return nil, err
	}
	defer resp.Body.Close()

	fmt.Printf("POST request sent to %s:%d%s - Response Status: %s\n", host, port, endpoint, resp.Status)

	// Check if the response status code is OK (200)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Non-OK status code: %s", resp.Status)
	}

	// Read and parse the JSON response body
	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return nil, err
	}

	fmt.Printf("Response Body: %s\n", string(responseBytes)) // Debugging

	var responseBody map[string]interface{}
	if err := json.Unmarshal(responseBytes, &responseBody); err != nil {
		log.Printf("Failed to unmarshal response body: %v", err)
		return nil, err
	}

	return responseBody, nil
}

func sendGetPayload(host string, port int, endpoint string) (map[string]interface{}, error) {
	url := fmt.Sprintf("http://%s:%d%s", host, port, endpoint)

	// Send the GET request
	resp, err := http.Get(url)
	if err != nil {
		err = fmt.Errorf("Failed to send GET request to %s:%d: %v", host, port, err)
		log.Printf(err.Error())
		return nil, err
	}

	defer resp.Body.Close()

	fmt.Printf("GET request sent to %s:%d%s - Response Status: %s\n", host, port, endpoint, resp.Status)

	// Check if the response status code is OK (200)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Non-OK status code: %s", resp.Status)
	}

	// Read and parse the JSON response body
	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return nil, err
	}

	var responseBody map[string]interface{}
	if err := json.Unmarshal(responseBytes, &responseBody); err != nil {
		log.Printf("Failed to unmarshal response body: %v", err)
		return nil, err
	}

	return responseBody, nil
}

func sendPayloadWithSecret(host string, port int, endpoint string, secretResponse map[string]interface{}) (string, error) {
	url := fmt.Sprintf("http://%s:%d%s", host, port, endpoint)

	requestBody := map[string]interface{}{
		"User":   usernameParam,
		"Secret": usernameSecret,
	}

	fmt.Printf("Request Body: %+v\n", requestBody)

	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("Failed to marshal JSON request body: %v", err)
		return "", err
	}

	// Send the POST request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		log.Printf("Failed to send POST request to %s:%d: %v", host, port, err)
		return "", err
	}
	defer resp.Body.Close()

	fmt.Printf("POST request sent to %s:%d%s - Response Status: %s\n", host, port, endpoint, resp.Status)

	// Check if the response status code is OK (200)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Non-OK status code: %s", resp.Status)
	}

	// Read and parse the JSON response body
	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return "", err
	}

	fmt.Printf("Response Body: %s\n", string(responseBytes)) // Debugging

	// Return the JSON response body as a string
	return string(responseBytes), nil
}

func sendPayloadSolution(host string, port int, endpoint string) (string, error) {
	url := fmt.Sprintf("http://%s:%d%s", host, port, endpoint)

	requestBody := map[string]interface{}{
		"Protocol": "",
	}

	fmt.Printf("Request Body: %+v\n", requestBody)

	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("Failed to marshal JSON request body: %v", err)
		return "", err
	}

	// Send the POST request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBodyBytes))
	if err != nil {
		log.Printf("Failed to send POST request to %s:%d: %v", host, port, err)
		return "", err
	}
	defer resp.Body.Close()

	fmt.Printf("POST request sent to %s:%d%s - Response Status: %s\n", host, port, endpoint, resp.Status)

	// Check if the response status code is OK (200)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Non-OK status code: %s", resp.Status)
	}

	// Read and parse the JSON response body
	responseBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v", err)
		return "", err
	}

	fmt.Printf("Response Body: %s\n", string(responseBytes)) // Debugging

	// Return the JSON response body as a string
	return string(responseBytes), nil
}
