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
	Timeout           = 200 * time.Millisecond
	RetryLimit   uint = 3
	RequestDelay      = time.Millisecond
	RecvTimeout       = time.Millisecond
	WaitingTime       = time.Second
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

type DNSRecord struct {
	Domain string
	Type   string
	Target string
	IP     []string
}

type DNSClient struct {
	Query     chan string
	Record    chan DNSRecord
	resolved  map[string]struct{}
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
	conn, err := dns.DialTimeout("udp", authoritativeDNSServers[rand.Intn(len(authoritativeDNSServers))], Timeout)
	if err != nil {
		return NewClient()
	}

	log.Debug("client =>", conn.Conn.RemoteAddr())
	client := DNSClient{
		make(chan string, 1000),
		make(chan DNSRecord, 1000),
		make(map[string]struct{}),
		make(chan dnsRetryRequest, 50000),
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
	q := dns.Fqdn(query)
	msg := &dns.Msg{}
	msg.SetQuestion(q, dns.TypeA)
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
			if retry.counter != 233 {
				client._send(retry.domain, retry.counter)
			}
		case <-time.After(WaitingTime):
			return
		}
	}
}

func (client DNSClient) recv() {
	// 泛解析记录
	for record := range chPanAnalyticRecord {
		client.Record <- record
		if record.Type == "CNAME" && IsSubdomain(record.Target) {
			client.Query <- record.Target
		}
	}

	for timer := range client.chSent {
		client.Conn.SetReadDeadline(time.Now().Add(RecvTimeout))
		msg, err := client.ReadMsg()
		if err != nil {
			// TODO 处理连接关闭的情况
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				continue
			} else {
				log.Fatal(err)
			}
		}

		close(timer.recved)
		record := DNSRecord{Domain: TrimSuffixPoint(msg.Question[0].Name)}
		if _, ok := client.resolved[record.Domain]; !ok {
			client.resolved[record.Domain] = struct{}{}

			if len(msg.Answer) > 0 {
				switch firstAnswer := msg.Answer[0].(type) {
				case *dns.CNAME:
					target := TrimSuffixPoint(firstAnswer.Target)
					//if ttl, _ok := panAnalyticRecords[target]; ttl != panAnalyticTtlMagicNum && !(_ok && firstAnswer.Hdr.Ttl == ttl) {
					//if ttl, _ok := panAnalyticRecords[target]; !_ok || (_ok && ttl != panAnalyticTtlMagicNum && ttl != firstAnswer.Hdr.Ttl) {
					if !IsPanAnalytic(target, firstAnswer.Hdr.Ttl) {
						record.Type = "CNAME"
						record.Target = target
						if IsSubdomain(record.Target) {
							go func() {
								client.Query <- record.Target
							}()
						}
					}
				case *dns.A:
					record.Type = "A"
					for _, ans := range msg.Answer {
						if a, ok := ans.(*dns.A); ok {
							//if ttl, _ok := panAnalyticRecords[a.A.String()]; ttl != panAnalyticTtlMagicNum && !(_ok && a.Hdr.Ttl == ttl) {
							//if ttl, _ok := panAnalyticRecords[a.A.String()]; !_ok || (ok && ttl != panAnalyticTtlMagicNum && ttl != a.Hdr.Ttl) {
							if !IsPanAnalytic(a.A.String(), a.Hdr.Ttl) {
								record.IP = append(record.IP, a.A.String())
							}
						}
					}
				}

				if record.Type == "CNAME" || (record.Type == "A" && len(record.IP) > 0) {
					client.Record <- record
				}
			}
		}
	}
	close(client.Record)
	client.Conn.Close()
	log.Debug("close client", client.Conn.RemoteAddr())
}

func (client DNSClient) retry() {
	for timer := range client.chTimeout {
		select {
		case <-timer.recved:
			client.chRetry <- dnsRetryRequest{233, ""}
		case <-timer.timeout:
			select {
			case <-timer.recved:
				client.chRetry <- dnsRetryRequest{233, ""}
			default:
				if timer.counter < RetryLimit {
					log.Debugf("retry %s, round %d\n", timer.domain, timer.counter+1)
					client.chRetry <- dnsRetryRequest{timer.counter + 1, timer.domain}
				}
			}
		}
	}
	close(client.chSent)
}
