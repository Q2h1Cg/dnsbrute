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

// 解析 dict/dnsservers.txt 确认存活的 DNS Server
// 注意：调用 VerifyDNSServers 函数确定 DNS Servers 之后再进行其他调用
func VerifyDNSServers() {
	fd, err := os.Open("dict/dnsservers.txt")
	if err != nil {
		log.Fatal("Can't open dict/dnsservers.txt:", err)
	}
	defer fd.Close()

	log.Println("start to verify DNS Servers")
	servers := func() <-chan string {
		in := make(chan string)

		go func() {
			scanner := bufio.NewScanner(fd)
			for scanner.Scan() {
				server := strings.TrimSpace(scanner.Text())
				if strings.HasPrefix(server, "#") || server == "" {
					continue
				}
				if !strings.HasSuffix(server, ":53") {
					server += ":53"
				}
				in <- server
			}
			if err := scanner.Err(); err != nil {
				log.Fatal("Read dict/dnsservers.txt:", err)
			}
			close(in)
		}()

		return in
	}()

	serversVerified := make(chan string)
	wg := sync.WaitGroup{}
	verifyMsg := &dns.Msg{}
	verifyMsg.SetQuestion("localhost.qq.com.", dns.TypeA)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for server := range servers {
				if in, err := dns.Exchange(verifyMsg, server); err == nil && len(in.Answer) > 0 && in.Answer[0].(*dns.A).A.String() == "1.1.1.1" {
					serversVerified <- server
					log.Printf("server %s: OK\n", server)
				} else {
					log.Printf("server %s: Failed\n", server)
				}
			}
		}()
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
