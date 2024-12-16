package main

import (
	"encoding/json"
	"lazada/iop-sdk-go/iop"
	"log"
)

func main() {

	appKey := "your appKey"
	appSecret := "your secret"
	var clientOptions = iop.ClientOptions{
		APIKey:    appKey,
		APISecret: appSecret,
		Region:    "SG",
	}

	client := iop.NewClient(&clientOptions)

	client.SetAccessToken("seller token")

	// DEMO.1 GET Method
	client.AddAPIParam("update_before", "2018-02-10T16:00:00+08:00")
	client.AddAPIParam("sort_direction", "DESC")
	client.AddAPIParam("offset", "0")
	client.AddAPIParam("limit", "10")
	client.AddAPIParam("update_after", "2017-02-10T09:00:00+08:00")
	client.AddAPIParam("sort_by", "updated_at")
	client.AddAPIParam("created_before", "2018-02-10T16:00:00+08:00")
	client.AddAPIParam("created_after", "2017-02-10T09:00:00+08:00")
	client.AddAPIParam("status", "all")

	getResult, _ := client.Execute("/orders/get", "GET", nil)
	getResultJson, _ := json.Marshal(getResult)
	log.Println(string(getResultJson))

	// DEMO.2 POST Method
	params := map[string]string{
		"kolUserId":      "123123123",
		"sort_direction": "DESC",
		"offset":         "0",
		"limit":          "10",
		"update_after":   "2017-02-10T09:00:00+08:00",
		"sort_by":        "updated_at",
		"created_before": "2018-02-10T16:00:00+08:00",
		"created_after":  "2017-02-10T09:00:00+08:00",
		"status":         "all",
	}

	postResult, _ := client.Execute("/orders/get", "POST", params)
	postResultJson, _ := json.Marshal(postResult)
	log.Println(string(postResultJson))

}
