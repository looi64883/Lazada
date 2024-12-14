package product

import (
	"encoding/json"
	"log"

	"github.com/tidwall/gjson"
)

func ProcessProducts(responseData string) {
	// Parse the response data to get products
	products := gjson.Get(responseData, "products")
	if !products.Exists() {
		log.Println("No products found in response.")
		return
	}

	// Slice to hold the generalized JSON
	var generalizedProducts []map[string]interface{}

	// Loop through each product and extract necessary fields
	products.ForEach(func(_, product gjson.Result) bool {
		generalizedProduct := map[string]interface{}{
			"item_id":       product.Get("item_id").Int(),
			"name":          product.Get("attributes.name").String(),
			"brand":         product.Get("attributes.brand").String(),
			"status":        product.Get("status").String(),
			"created_time":  product.Get("created_time").String(),
			"updated_time":  product.Get("updated_time").String(),
			"price":         product.Get("skus.0.price").Float(),
			"special_price": product.Get("skus.0.special_price").Float(),
			"quantity":      product.Get("skus.0.quantity").Int(),
			"url":           product.Get("skus.0.Url").String(),
			"images":        product.Get("images").Array(),
		}
		generalizedProducts = append(generalizedProducts, generalizedProduct)
		return true
	})

	// Convert to JSON for use or storage
	generalizedJSON, err := json.MarshalIndent(generalizedProducts, "", "  ")
	if err != nil {
		log.Fatalf("Error marshalling generalized products: %v", err)
	}

	log.Println("Generalized JSON:")
	log.Println(string(generalizedJSON))
}
