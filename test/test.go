package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"lazada/iop-sdk-go/iop"
	"lazada/pkg/order"

	"github.com/labstack/echo/v4"
	"github.com/tidwall/gjson"
)

type RequestPayload struct {
	AccessToken  string `json:"access_token"`
	CreatedAfter string `json:"created_after"`
}

type OrderTask struct {
	Offset int
	Limit  int
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

	// Validate required fields
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

	// Lazada client configuration
	client := iop.NewClient(&clientOptions)
	client.SetAccessToken(payload.AccessToken)

	// Get total order count
	client.AddAPIParam("created_after", payload.CreatedAfter)
	totalOrders, err := getTotalOrderCount(client)
	if err != nil {
		log.Printf("Error fetching total orders: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch order count"})
	}

	log.Printf("Payload: %+v", payload)

	log.Printf("Total orders to process: %d", totalOrders)

	// Worker pool and concurrency setup
	limit := 18
	numWorkers := 5 // Adjust based on system resources
	tasks := make(chan OrderTask, totalOrders)
	results := make(chan string, totalOrders)
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(clientOptions, payload.AccessToken, payload.CreatedAfter, tasks, results, &wg)
	}

	// Send tasks to the worker pool
	for offset := 0; offset < totalOrders; offset += limit {
		tasks <- OrderTask{Offset: offset, Limit: limit}
	}
	close(tasks)

	// Wait for all workers to complete
	wg.Wait()
	close(results)

	// Process results
	for result := range results {
		order.ProcessOrders(result)
	}

	// Return success response
	log.Println("Orders processed successfully")
	return c.JSON(http.StatusOK, map[string]string{"message": "Orders processed successfully"})
}

func getTotalOrderCount(client *iop.IopClient) (int, error) {
	getResult, err := client.Execute("/orders/get", "GET", nil)
	if err != nil {
		return 0, err
	}

	response := string(getResult.Data)
	return int(gjson.Get(response, "countTotal").Int()), nil
}

func worker(clientOptions iop.ClientOptions, accessToken string, CreatedAfter string, tasks <-chan OrderTask, results chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	for task := range tasks {
		client := iop.NewClient(&clientOptions)
		client.SetAccessToken(accessToken)
		client.AddAPIParam("created_after", CreatedAfter)
		client.AddAPIParam("offset", fmt.Sprintf("%d", task.Offset))
		client.AddAPIParam("limit", fmt.Sprintf("%d", task.Limit))

		getResult, err := client.Execute("/orders/get", "GET", nil)
		if err != nil {
			log.Printf("Error fetching orders for offset %d: %v", task.Offset, err)
			continue
		}

		results <- string(getResult.Data)
	}
}
