package dns

import (
	"bufio"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/miekg/dns"
)

var dnsServers []string

// VerifyDNSServers 解析 dict/dnsservers.txt 并确认其中存活的 DNS Server
// 注意：调用 VerifyDNSServers 函数确定 DNS Servers 之后再进行其他调用
func VerifyDNSServers() {
	fd, err := os.Open("dict/dnsservers.txt")
	if err != nil {
		log.Fatal("Can't open dict/dnsservers.txt:", err)
	}
	defer fd.Close()

	log.Println("start to verify DNS Servers")
	serversVerified := make(chan string)
	wg := sync.WaitGroup{}
	verifyMsg := &dns.Msg{}
	verifyMsg.SetQuestion("localhost.qq.com.", dns.TypeA)
	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		server := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(server, "#") || server == "" {
			continue
		}
		if !strings.HasSuffix(server, ":53") {
			server += ":53"
		}
		wg.Add(1)
		go func(server string) {
			defer wg.Done()
			if in, err := dns.Exchange(verifyMsg, server); err == nil && len(in.Answer) > 0 && in.Answer[0].(*dns.A).A.String() == "1.1.1.1" {
				serversVerified <- server
				log.Printf("server %s: OK\n", server)
			} else {
				log.Printf("server %s: Failed\n", server)
			}
		}(server)

	}
	if err := scanner.Err(); err != nil {
		log.Fatal("Read dict/dnsservers.txt:", err)
	}

	go func() {
		for server := range serversVerified {
			dnsServers = append(dnsServers, server)
		}
	}()

	wg.Wait()
	close(serversVerified)
	log.Println("available servers:", dnsServers)
}
