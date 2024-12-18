package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"lazada/iop-sdk-go/iop"
	"lazada/pkg/order"
	"lazada/pkg/product"

	"github.com/labstack/echo/v4"
	"github.com/tidwall/gjson"
)

type RequestPayload struct {
	AccessToken  string `json:"access_token"`
	CreatedAfter string `json:"created_after"`
}

type Task struct {
	Offset int
	Limit  int
}

type WorkerConfig struct {
	ClientOptions iop.ClientOptions
	AccessToken   string
	CreatedAfter  string
	Endpoint      string
	ProcessFunc   func(string) string
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

	// Define the POST endpoints
	e.POST("/process-products", func(c echo.Context) error {
		return handleProcessing(c, "/products/get", "total_products", product.ProcessProducts)
	})
	e.POST("/process-orders", func(c echo.Context) error {
		return handleProcessing(c, "/orders/get", "countTotal", order.ProcessOrders)
	})

	// Start the server
	log.Println("Server started on :8091")
	e.Logger.Fatal(e.Start(":8091"))
}

func handleProcessing(c echo.Context, endpoint, countKey string, processFunc func(string) string) error {
	// Bind the request payload
	payload := new(RequestPayload)
	if err := c.Bind(payload); err != nil {
		log.Printf("Error binding payload: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	// Validate required fields
	if payload.AccessToken == "" {
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

	client := iop.NewClient(&clientOptions)
	client.SetAccessToken(payload.AccessToken)

	// Get total order count
	client.AddAPIParam("created_after", payload.CreatedAfter)

	// Get total count
	totalCount, err := getTotalCount(client, endpoint, countKey)
	if err != nil {
		log.Printf("Error fetching total count: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch count"})
	}

	log.Printf("Payload: %+v", payload)
	log.Printf("Total items to process: %d", totalCount)

	// Worker pool and concurrency setup
	limit := 18
	numWorkers := 5 // Adjust based on system resources
	tasks := make(chan Task, totalCount)
	results := make(chan string, totalCount)
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		config := WorkerConfig{
			ClientOptions: clientOptions,
			AccessToken:   payload.AccessToken,
			CreatedAfter:  payload.CreatedAfter,
			Endpoint:      endpoint,
			ProcessFunc:   processFunc,
		}
		go worker(config, tasks, results, &wg, endpoint)
	}

	// Send tasks to the worker pool
	for offset := 0; offset < totalCount; offset += limit {
		tasks <- Task{Offset: offset, Limit: limit}
	}
	close(tasks)

	// Wait for all workers to complete
	wg.Wait()
	close(results)

	// Process results
	for result := range results {
		log.Printf("Processed data: %s", result)
	}

	// Return success response
	log.Println("Items processed successfully")
	return c.JSON(http.StatusOK, map[string]string{"message": "Items processed successfully"})
}

func getTotalCount(client *iop.IopClient, endpoint, countKey string) (int, error) {
	getResult, err := client.Execute(endpoint, "GET", nil)
	if err != nil {
		return 0, err
	}

	response := string(getResult.Data)
	return int(gjson.Get(response, countKey).Int()), nil
}

func worker(config WorkerConfig, tasks <-chan Task, results chan<- string, wg *sync.WaitGroup, endpoint string) {
	defer wg.Done()
	log.Printf(config.CreatedAfter)

	for task := range tasks {
		client := iop.NewClient(&config.ClientOptions)
		client.SetAccessToken(config.AccessToken)
		client.AddAPIParam("created_after", config.CreatedAfter)
		client.AddAPIParam("offset", fmt.Sprintf("%d", task.Offset))
		client.AddAPIParam("limit", fmt.Sprintf("%d", task.Limit))

		getResult, err := client.Execute(endpoint, "GET", nil)
		if err != nil {
			log.Printf("Error fetching data for offset %d: %v", task.Offset, err)
			continue
		}

		// Process the fetched data using the provided ProcessFunc
		processedData := config.ProcessFunc(string(getResult.Data))

		results <- processedData
	}
}
