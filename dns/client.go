package dns

import (
	"bufio"
	_rand "math/rand"
	"net"
	"os"
	"strings"
	"time"

	"github.com/chuhades/dnsbrute/log"

	"github.com/miekg/dns"
)

var dnsServers []string

var rand = _rand.New(_rand.NewSource(time.Now().Unix()))

var (
	Timeout           = time.Second
	RetryLimit   uint = 3
	RequestDelay      = time.Millisecond
	RecvTimeout       = 50 * time.Millisecond
)

type dnsRequest struct {
	counter uint
	domain  string
	timeout <-chan time.Time
	recved  chan struct{}
}

type dnsRetryRequest struct {
	counter uint
	domain  string
}

type dnsRecord struct {
	Domain string
	IP     []string
}

type DNSClient struct {
	Query     chan string
	Record    chan dnsRecord
	chRetry   chan dnsRetryRequest
	chSent    chan dnsRequest
	chTimeout chan dnsRequest
	*dns.Conn
}

func init() {
	// 加载 DNS Server 字典
	fd, err := os.Open("dict/dnsservers.txt")
	if err != nil {
		log.Fatal("Can't open dict/dnsservers.txt:", err)
	}
	defer fd.Close()

	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		server := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(server, "#") || server == "" {
			continue
		}
		if !strings.HasSuffix(server, ":53") {
			server += ":53"
		}
		dnsServers = append(dnsServers, server)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal("Read dict/dnsservers.txt:", err)
	}
}

func NewClient() DNSClient {
	conn, err := dns.DialTimeout("udp", dnsServers[rand.Intn(len(dnsServers))], Timeout)
	if err != nil {
		return NewClient()
	}

	log.Debug("client =>", conn.Conn.RemoteAddr())
	client := DNSClient{
		make(chan string, 1000),
		make(chan dnsRecord, 1000),
		make(chan dnsRetryRequest, 1000),
		make(chan dnsRequest, 1000),
		make(chan dnsRequest, 1000),
		conn,
	}

	go client.send()
	go client.recv()
	go client.retry()

	return client
}

func (client DNSClient) _send(query string, counter uint) {
	query = dns.Fqdn(query)
	msg := &dns.Msg{}
	msg.SetQuestion(query, dns.TypeA)
	client.WriteMsg(msg)
	timer := dnsRequest{counter, query, time.After(Timeout), make(chan struct{})}
	client.chSent <- timer
	client.chTimeout <- timer
	time.Sleep(RequestDelay)
}

func (client DNSClient) send() {
	defer close(client.chTimeout)
	for {
		select {
		case query := <-client.Query:
			client._send(query, 0)
		case retry := <-client.chRetry:
			client._send(retry.domain, retry.counter)
		case <-time.After(time.Second):
			return
		}
	}
}

func (client DNSClient) recv() {
	for timer := range client.chSent {
		client.Conn.SetReadDeadline(time.Now().Add(RecvTimeout))
		msg, err := client.ReadMsg()
		if err != nil {
			// TODO 处理连接关闭的情况
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				log.Debug(nerr)
				continue
			} else {
				log.Fatal(err)
			}
		}

		record := dnsRecord{Domain: msg.Question[0].Name}
		for _, ans := range msg.Answer {
			if a, ok := ans.(*dns.A); ok {
				record.IP = append(record.IP, a.A.String())
			}
		}

		close(timer.recved)

		if len(record.IP) != 0 {
			client.Record <- record
		}
	}
	close(client.Record)
	client.Conn.Close()
}

func (client DNSClient) retry() {
	for timer := range client.chTimeout {
		select {
		case <-timer.recved:
		case <-timer.timeout:
			if timer.counter < RetryLimit {
				log.Debugf("retry %s, round %d\n", timer.domain, timer.counter+1)
				client.chRetry <- dnsRetryRequest{timer.counter + 1, timer.domain}
			}
		}
	}
	close(client.chSent)
}
