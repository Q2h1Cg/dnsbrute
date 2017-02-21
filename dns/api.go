package dns

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"strings"
)

var apiList []API = []API{hackertarget{}}

type API interface {
	Name() string
	Query(rootDomain string, subDomains chan<- string, message chan<- string)
}

func QueryOnAPI(rootDomain string) {
	subDomains := make(chan string)
	message := make(chan string)

	for _, api := range apiList {
		go api.Query(rootDomain, subDomains, message)
	}

	go func() {
		for subDomain := range subDomains {
			fmt.Println(subDomain)
		}
	}()

	for _ = range apiList {
		log.Println(<-message)
	}
	close(subDomains)
}

type hackertarget struct{}

func (h hackertarget) Name() string {
	return "http://www.hackertarget.com/"
}

func (h hackertarget) Query(rootDomain string, subDomains chan<- string, message chan<- string) {
	n := 0
	url := "http://api.hackertarget.com/hostsearch/?q=" + rootDomain
	resp, err := http.Get(url)
	if err != nil {
		message <- fmt.Sprintf("API %s error: %v\n", h.Name(), err)
		return
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		record := strings.TrimSpace(scanner.Text())
		if record != "" {
			subDomains <- strings.Split(record, ",")[0]
			n++
		}
	}
	if err := scanner.Err(); err != nil {
		message <- fmt.Sprintf("API %s error: %v\n", h.Name(), err)
		return
	}
	message <- fmt.Sprintf("got %d domains from %s\n", n, h.Name())
}
