package order

import (
	"encoding/json"
	"log"

	"github.com/tidwall/gjson"
)

func ProcessOrders(responseData string) {
	// Parse the response data to get orders
	orders := gjson.Get(responseData, "orders")
	if !orders.Exists() {
		log.Println("No orders found in response.")
		return
	}

	// Slice to hold the generalized JSON
	var generalizedOrders []map[string]interface{}

	// Loop through each order and extract necessary fields
	orders.ForEach(func(_, order gjson.Result) bool {
		generalizedOrder := map[string]interface{}{
			"order_id":              order.Get("order_number").String(),
			"created_at":            order.Get("created_at").String(),
			"updated_at":            order.Get("updated_at").String(),
			"price":                 order.Get("price").Float(),
			"voucher_platform":      order.Get("voucher_platform").Float(),
			"voucher_seller":        order.Get("voucher_seller").Float(),
			"shipping_fee_discount": order.Get("shipping_fee_discount_platform").Float(),
			"warehouse_code":        order.Get("warehouse_code").String(),
			"shipping_fee":          order.Get("shipping_fee_original").Float(),
			"items_count":           order.Get("items_count").Int(),
		}
		generalizedOrders = append(generalizedOrders, generalizedOrder)
		return true
	})

	// Convert to JSON for use or storage
	_, err := json.MarshalIndent(generalizedOrders, "", "  ")
	if err != nil {
		log.Fatalf("Error marshalling generalized orders: %v", err)
	}

	// log.Println("Generalized JSON:")
	// log.Println(string(generalizedJSON))
}
