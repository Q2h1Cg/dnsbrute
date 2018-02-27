package dns

import (
	"sync"
	"time"

	"github.com/miekg/dns"

	"github.com/Q2h1Cg/dnsbrute/log"
)

const timeout = time.Second

var (
	// Queries 输入
	Queries = make(chan string)
	// Records 输出
	Records = make(chan *Record)

	dnsServerAddress string
	client           *dns.Conn
	transmitRate     int
	retryLimit       int
	requests         = sync.Map{}
	received         = map[string]struct{}{}
	noMoreQueries    = make(chan struct{})
)

// Configure 设置发包速率、DNS 服务器地址
func Configure(domain, server string, rate, retry int) (err error) {
	rootDomain = domain
	dnsServerAddress = server
	transmitRate = rate
	retryLimit = retry

	if client, err = dns.DialTimeout("udp", server, timeout); err != nil {
		return err
	}

	go send()
	go receive()

	return nil
}

// send 发送查询
func send() {
	delay := time.Second / time.Nanosecond / time.Duration(transmitRate)

	for {
		select {
		case domain := <-Queries:
			time.Sleep(delay)

			req, _ := requests.LoadOrStore(domain, &Request{0, nil})
			request := req.(*Request)

			// 超出重试次数
			if request.SentCount >= retryLimit {
				requests.Delete(domain)
				continue
			}

			// 发送 DNS 查询
			msg := &dns.Msg{}
			msg.SetQuestion(dns.Fqdn(domain), dns.TypeA)
			if err := client.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
				requests.Delete(domain)
				continue
			}
			if err := client.WriteMsg(msg); err != nil {
				requests.Delete(domain)
				continue
			}

			// 超时重发请求
			request.SentCount++
			request.Timer = time.AfterFunc(timeout, func() {
				log.Debug("retry", domain, request.SentCount)
				Queries <- domain
			})
		case <-time.After(3 * timeout):
			log.Debug("no more queries")
			close(noMoreQueries)
			return
		}
	}
}

// receive 接收 DNS 响应
func receive() {
	for {
		select {
		case <-noMoreQueries:
			close(Records)
			return
		default:
		}

		// 接收 DNS 响应
		var msg *dns.Msg
		var err error
		if err = client.SetReadDeadline(time.Now().Add(timeout)); err != nil {
			continue
		}
		if msg, err = client.ReadMsg(); err != nil {
			continue
		}

		domain := trimSuffixPoint(msg.Question[0].Name)
		if request, ok := requests.Load(domain); ok {
			if timer := request.(*Request).Timer; timer != nil {
				request.(*Request).Timer.Stop()
			}
			requests.Delete(domain)
		}
		// 标记响应
		if _, ok := received[domain]; ok {
			continue
		}
		received[domain] = struct{}{}

		if record := NewRecord(domain, msg.Answer); record != nil {
			Records <- record
		}
	}
}
