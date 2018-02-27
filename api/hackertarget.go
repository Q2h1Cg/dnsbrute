package api

import (
	"bufio"
	"net/http"
	"strings"

	"github.com/Q2h1Cg/dnsbrute/log"
)

type hackertarget struct{}

func init() {
	registerAPI(hackertarget{})
}

// Name 接口名称
func (h hackertarget) Name() string {
	return "hackertarget"
}

// Query 查询接口
func (h hackertarget) Query(domain string) <-chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)

		url := "http://api.hackertarget.com/hostsearch/?q=" + domain
		client := http.Client{Timeout: timeout}
		resp, err := client.Get(url)
		if err != nil {
			log.Info("error while fetching api.hackertarget.com:", err)
			return
		}
		defer resp.Body.Close()

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			record := scanner.Text()
			if record != "" {
				ch <- strings.Split(record, ",")[0]
			}
		}
	}()

	return ch
}
