package api

import (
	"log"
	"time"
)

// API 接口定义
type API interface {
	Name() string
	Query(domain string) <-chan string
}

const timeout = 5 * time.Second

var apiList []API

// Query 通过 API 接口查询子域名
func Query(domain string) <-chan string {
	ch := make(chan string)

	go func() {
		apiMap := map[string]<-chan string{}
		for _, api := range apiList {
			apiMap[api.Name()] = api.Query(domain)
		}

		for name, apiRecords := range apiMap {
			counter := 0
			for record := range apiRecords {
				ch <- record
				counter++
			}
			log.Printf("%s: %d records\n", name, counter)
		}

		close(ch)
	}()

	return ch
}

func registerAPI(api API) {
	apiList = append(apiList, api)
}
