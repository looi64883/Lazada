package main

import (
	"encoding/json"
	"fmt"
	"lazada/iop-sdk-go/iop"
	"log"
	"net/http"
)

const (
	clientID     = "131151"                           // Your Lazada App Key
	clientSecret = "pA96smss38jIXWepIxl34VtfVMaDrChx" // Your Lazada App Secret
	redirectURI  = "https://vemp.onrender.com/"       // The redirect URI you specified in Lazada settings
)

func main() {
	http.HandleFunc("/callback", authCallbackHandler)

	// Start the server to handle the callback
	log.Println("Server started on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func authCallbackHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the authorization code from the query parameters
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Authorization code not found", http.StatusBadRequest)
		return
	}
	log.Print(code)

	// Step 2: Exchange the authorization code for an access token using the SDK
	accessToken, err := getAccessToken(code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting access token: %v", err), http.StatusInternalServerError)
		return
	}

	// Now you can use the access token to interact with the Lazada API
	log.Printf("Received access token: %s", accessToken)

	// Example: You can use the access token to fetch data or send the token to the frontend
	w.Write([]byte(fmt.Sprintf("Access Token: %s", accessToken)))
}

// Exchange the authorization code for an access token using Lazada SDK
func getAccessToken(code string) (string, error) {
	// Step 1: Initialize the SDK client with your app credentials
	clientOptions := iop.ClientOptions{
		APIKey:    clientID,
		APISecret: clientSecret,
		Region:    "MY", // Use the region corresponding to your account (e.g., MY for Malaysia)
	}

	client := iop.NewClient(&clientOptions)
	client.AddAPIParam("code", code)

	// Step 2: Call the API to exchange the code for an access token
	getResult, err := client.Execute("/auth/token/create", "GET", nil)
	if err != nil {
		log.Printf("error exchanging code for access token: %v", err)
		return "", err
	}

	// Step 3: Assuming the SDK returns a `Data` field containing the response body
	var result map[string]interface{}
	if err := json.Unmarshal(getResult.Data, &result); err != nil {
		log.Printf("error parsing token response: %v", err)
		return "", err
	}

	// Step 4: Extract the access token from the parsed response
	accessToken, ok := result["access_token"].(string)
	if !ok {
		log.Printf("access token not found in response")
		return "", fmt.Errorf("access token not found in response")
	}

	// Return the access token
	return accessToken, nil
}
