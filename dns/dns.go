package dns

import (
	"strings"
	"time"

	"github.com/miekg/dns"
)

// Request DNS 请求
type Request struct {
	SentCount int
	Timer     *time.Timer
}

// Record DNS 记录
type Record struct {
	Domain string
	Type   string
	Target string
	IP     []string
}

// CSV 转换为字符串切片供 CSV 输出
func (r Record) CSV() []string {
	return []string{r.Domain, r.Type, r.Target, strings.Join(r.IP, ",")}
}

// NewRecord 新建 DNS 记录
func NewRecord(domain string, response []dns.RR) *Record {
	if len(response) == 0 || isPanDNS(domain, response) {
		return nil
	}

	record := Record{Domain: domain}
	switch firstAnswer := response[0].(type) {
	case *dns.CNAME:
		record.Type = "CNAME"
		record.Target = trimSuffixPoint(firstAnswer.Target)
		response = response[1:]
	case *dns.A:
		record.Type = "A"
	default:
		return nil
	}

	for _, ans := range response {
		if a, ok := ans.(*dns.A); ok {
			record.IP = append(record.IP, a.A.String())
		}
	}

	return &record
}
