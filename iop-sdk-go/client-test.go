package main

import (
	"fmt"
	"log"
	"strconv"

	"iop-go-sdk/iop"

	"github.com/tidwall/gjson"
)

func main() {

	appKey := "131151"
	appSecret := "pA96smss38jIXWepIxl34VtfVMaDrChx"
	var clientOptions = iop.ClientOptions{
		APIKey:    appKey,
		APISecret: appSecret,
		Region:    "MY",
	}

	// client := iop.NewClient(&clientOptions)

	// client.SetAccessToken("50000000341bfGracTOkT7jlPZowBC2HwknqgtxcZCFp3qVVDfKiw1a13d289ZJu")

	// Pagination settings
	limit := 10
	offset := 0
	totalRecords := 0 // Keep track of the total number of records

	for {
		client := iop.NewClient(&clientOptions)

		client.SetAccessToken("50000000341bfGracTOkT7jlPZowBC2HwknqgtxcZCFp3qVVDfKiw1a13d289ZJu")
		// Construct API parameters for pagination
		client.AddAPIParam("offset", fmt.Sprintf("%d", offset))
		client.AddAPIParam("limit", fmt.Sprintf("%d", limit))
		client.AddAPIParam("created_after", "2024-11-10T09:00:00+08:00")

		// Call the API
		getResult, err := client.Execute("/orders/get", "GET", nil)
		if err != nil {
			log.Fatalf("Error calling API: %v", err)
		}
		// Parse the response using gjson
		getResultJson := gjson.Parse(string(getResult.Data))
		// log.Println("Received data:", getResultJson)

		// Extract `countTotal` from the response using gjson
		countTotal := getResultJson.Get("countTotal").Int()

		// Convert countTotal (int64) to string using strconv.FormatInt
		log.Println("**********************************" + strconv.FormatInt(countTotal, 10))
		log.Println(offset)
		if countTotal == 0 {
			log.Println("No records found.")
			break
		}

		totalRecords = int(countTotal)

		// Check for the total number of records
		log.Printf("Total records: %d\n", totalRecords)

		// Check if we have more records to fetch
		if offset+limit >= totalRecords {
			log.Println("All records fetched.")
			break
		}

		// Update the offset for the next batch
		offset += limit

		// Optionally, sleep for a short period between requests to avoid rate limits
		// time.Sleep(6 * time.Second) // Adjust based on the API rate limit policy
	}

	// You can implement similar logic for POST requests, depending on your needs.
}
