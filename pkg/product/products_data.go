package product

import (
	"log"

	"github.com/tidwall/gjson"
)

func ProcessProducts(responseData string) {
	// Parse the response data to get orders
	products := gjson.Get(responseData, "products")
	if !products.Exists() {
		log.Println("No orders found in response.")
		return
	}

}
