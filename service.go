package main

import (
	"fmt"
	"log"
	"net/http"

	"lazada/iop-sdk-go/iop"
	"lazada/pkg/order"

	"github.com/labstack/echo/v4"
	"github.com/tidwall/gjson"
)

type RequestPayload struct {
	AccessToken  string `json:"access_token"`
	CreatedAfter string `json:"created_after"`
}

func main() {
	e := echo.New()

	// Middleware for logging requests
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			log.Println("Processing request...")
			return next(c)
		}
	})

	// Define the POST endpoint
	e.POST("/process-orders", handleOrderProcessing)

	// Start the server
	log.Println("Server started on :8091")
	e.Logger.Fatal(e.Start(":8091"))
}

func handleOrderProcessing(c echo.Context) error {
	// Bind the request payload
	payload := new(RequestPayload)
	if err := c.Bind(payload); err != nil {
		log.Printf("Error binding payload: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	// Validate required fields (manually for now)
	if payload.AccessToken == "" || payload.CreatedAfter == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing or invalid fields"})
	}

	// Setup the Lazada client
	appKey := "131151"
	appSecret := "pA96smss38jIXWepIxl34VtfVMaDrChx"
	clientOptions := iop.ClientOptions{
		APIKey:    appKey,
		APISecret: appSecret,
		Region:    "MY",
	}
	// client := iop.NewClient(&clientOptions)
	// client.SetAccessToken(payload.AccessToken)

	// Pagination logic
	limit := 18
	offset := 0
	totalRecords := 0

	for {
		// Add API params for pagination
		client := iop.NewClient(&clientOptions)
		client.SetAccessToken(payload.AccessToken)
		client.AddAPIParam("offset", fmt.Sprintf("%d", offset))
		client.AddAPIParam("limit", fmt.Sprintf("%d", limit))
		client.AddAPIParam("created_after", payload.CreatedAfter)

		// Execute API call
		getResult, err := client.Execute("/orders/get", "GET", nil)
		if err != nil {
			log.Printf("Error calling API: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch orders"})
		}

		// Parse the response using gjson
		response := string(getResult.Data)
		getResultJson := gjson.Parse(response)

		// Extract countTotal
		countTotal := getResultJson.Get("countTotal").Int()
		log.Printf("Total orders count: %d", countTotal)

		// Check if there are any orders
		if countTotal == 0 {
			log.Println("No records found.")
			break
		}

		// Process the orders
		order.ProcessOrders(response)

		// Update the total records and check if we fetched all records
		totalRecords = int(countTotal)
		if offset+limit >= totalRecords {
			log.Println("All records fetched.")
			break
		}

		// Increment the offset for the next batch
		offset += limit
	}

	// Return success response
	log.Println("Orders processed successfully")
	return c.JSON(http.StatusOK, map[string]string{"message": "Orders processed successfully"})
}
