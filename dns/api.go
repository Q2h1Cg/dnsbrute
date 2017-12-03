package dns

import (
	"bufio"
	"fmt"
	"strings"
	"time"

	"github.com/Q2h1Cg/dnsbrute/log"

	"github.com/astaxie/beego/httplib"
)

const apiTimeout = 3 * time.Second

var apiList = []API{hackertarget{}, passiveDNS{}}

type API interface {
	Name() string
	Query(rootDomain string, subDomains chan<- string, message chan<- string)
}

func QueryOverAPI(rootDomain string) <-chan string {
	subDomains := make(chan string)
	message := make(chan string)

	for _, api := range apiList {
		go api.Query(rootDomain, subDomains, message)
	}

	go func() {
		for range apiList {
			log.Debug(<-message)
		}
		close(subDomains)
	}()

	return subDomains
}

type hackertarget struct{}

func (h hackertarget) Name() string {
	return "www.hackertarget.com"
}

func (h hackertarget) Query(rootDomain string, subDomains chan<- string, message chan<- string) {
	counter := 0
	urlSearch := "http://api.hackertarget.com/hostsearch/?q=" + rootDomain
	resp, err := httplib.Get(urlSearch).SetTimeout(apiTimeout, apiTimeout).Response()
	if err != nil {
		message <- fmt.Sprintf("API %s error: %v", h.Name(), err)
		return
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		record := strings.TrimSpace(scanner.Text())
		if record != "" {
			subDomains <- strings.Split(record, ",")[0]
			counter++
		}
	}
	if err := scanner.Err(); err != nil {
		message <- fmt.Sprintf("API %s error: %v", h.Name(), err)
		return
	}
	message <- fmt.Sprintf("got %d domains from %s", counter, h.Name())
}

type passiveDNS struct{}

func (p passiveDNS) Name() string {
	return "ptrarchive.com"
}

func (p passiveDNS) Query(rootDomain string, subDomains chan<- string, message chan<- string) {
	counter := 0
	urlSearch := "http://ptrarchive.com/tools/search.htm?label=" + rootDomain
	resp, err := httplib.Get(urlSearch).SetTimeout(apiTimeout, apiTimeout).String()
	if err != nil {
		message <- fmt.Sprintf("API %s error: %v", p.Name(), err)
		return
	}

	for _, i := range strings.Split(resp, "</td><td>") {
		domain := strings.TrimSpace(strings.Split(i, " ")[0])
		if IsSubdomain(domain) {
			subDomains <- domain
			counter++
		}
	}

	message <- fmt.Sprintf("got %d domains from %s", counter, p.Name())
}
