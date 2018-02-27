package api

import (
	"time"

	"github.com/Q2h1Cg/dnsbrute/log"
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
		defer close(ch)

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
			log.Infof("%s: %d records\n", name, counter)
		}
	}()

	return ch
}

func registerAPI(api API) {
	apiList = append(apiList, api)
}
