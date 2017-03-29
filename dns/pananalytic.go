package dns

import (
	"crypto/md5"
	"encoding/hex"
	"strings"

	"github.com/chuhades/dnsbrute/log"

	"github.com/miekg/dns"
)

var (
	panAnalyticRecord   = map[string]struct{}{}
	chPanAnalyticRecord = make(chan DNSRecord, 1)
)

func query(domain string) (record DNSRecord) {
	msg := &dns.Msg{}
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeA)
	in, err := dns.Exchange(msg, dnsServers[0])
	if err == nil {
		if len(in.Answer) > 0 {
			record.Domain = domain
			if c, ok := in.Answer[0].(*dns.CNAME); ok {
				record.Type = "CNAME"
				record.Target = strings.TrimSuffix(c.Target, ".")
			} else if _, ok := in.Answer[0].(*dns.A); ok {
				record.Type = "A"
				for _, ans := range in.Answer {
					if a, ok := ans.(*dns.A); ok {
						record.IP = append(record.IP, a.A.String())
					}
				}
			}
		}
	}

	return record
}

// FIXME 子域名也有可能存在泛解析
// FIXME 某真实存在的域名可能指向泛解析记录
func AnalyzePanAnalytic() {
	hash := md5.New()
	hash.Write([]byte(rootDomain))
	domain := hex.EncodeToString(hash.Sum(nil)) + "." + rootDomain
	record := query(domain)
	if record.Type == "CNAME" {
		// TODO cname 泛解析的情况下，是否把 IP 也加入黑名单
		panAnalyticRecord[record.Target] = struct{}{}
	} else if record.Type == "A" {
		for _, ip := range record.IP {
			panAnalyticRecord[ip] = struct{}{}
		}
	}
	chPanAnalyticRecord <- record
	close(chPanAnalyticRecord)
	log.Debugf("pan analytic record: %v\n", panAnalyticRecord)
}
