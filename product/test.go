package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"lazada/iop-sdk-go/iop"

	"github.com/labstack/echo/v4"
	"github.com/tidwall/gjson"
)

type RequestPayload struct {
	AccessToken  string `json:"access_token"`
	CreatedAfter string `json:"created_after"`
}

type ProductTask struct {
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
	e.POST("/process-products", handleProductProcessing)

	// Start the server
	log.Println("Server started on :8091")
	e.Logger.Fatal(e.Start(":8091"))
}

func handleProductProcessing(c echo.Context) error {
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

	// Lazada client configuration
	client := iop.NewClient(&clientOptions)
	client.SetAccessToken(payload.AccessToken)
	// Get total order count
	client.AddAPIParam("created_after", payload.CreatedAfter)

	// Get total product count
	totalProducts, err := getTotalProductCount(client)
	if err != nil {
		log.Printf("Error fetching total products: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch product count"})
	}

	log.Printf("Payload: %+v", payload)
	log.Printf("Total products to process: %d", totalProducts)

	// Worker pool and concurrency setup
	limit := 18
	numWorkers := 5 // Adjust based on system resources
	tasks := make(chan ProductTask, totalProducts)
	results := make(chan string, totalProducts)
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go productWorker(clientOptions, payload.AccessToken, tasks, results, &wg)
	}

	// Send tasks to the worker pool
	for offset := 0; offset < totalProducts; offset += limit {
		tasks <- ProductTask{Offset: offset, Limit: limit}
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
	log.Println("Products processed successfully")
	return c.JSON(http.StatusOK, map[string]string{"message": "Products processed successfully"})
}

func getTotalProductCount(client *iop.IopClient) (int, error) {
	getResult, err := client.Execute("/products/get", "GET", nil)
	if err != nil {
		return 0, err
	}

	response := string(getResult.Data)
	return int(gjson.Get(response, "data.total_products").Int()), nil
}

func productWorker(clientOptions iop.ClientOptions, accessToken string, tasks <-chan ProductTask, results chan<- string, wg *sync.WaitGroup) {
	defer wg.Done()

	for task := range tasks {
		client := iop.NewClient(&clientOptions)
		client.SetAccessToken(accessToken)
		client.AddAPIParam("offset", fmt.Sprintf("%d", task.Offset))
		client.AddAPIParam("limit", fmt.Sprintf("%d", task.Limit))

		getResult, err := client.Execute("/products/get", "GET", nil)
		if err != nil {
			log.Printf("Error fetching products for offset %d: %v", task.Offset, err)
			continue
		}

		results <- string(getResult.Data)
	}
}
