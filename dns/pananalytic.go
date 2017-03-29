package dns

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/chuhades/dnsbrute/log"

	"github.com/miekg/dns"
)

var (
	panAnalyticRecord   = map[string]struct{}{}
	chPanAnalyticRecord = make(chan DNSRecord, 1)
)

func query(domain string) (IP []string) {
	msg := &dns.Msg{}
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeA)
	in, err := dns.Exchange(msg, dnsServers[0])
	if err == nil {
		for _, ans := range in.Answer {
			if a, ok := ans.(*dns.A); ok {
				IP = append(IP, a.A.String())
			}
		}
	}

	return IP
}

// FIXME 子域名也有可能存在泛解析
// FIXME 某真实存在的域名可能指向泛解析记录
func AnalyzePanAnalytic() {
	hash := md5.New()
	hash.Write([]byte(rootDomain))
	domain := hex.EncodeToString(hash.Sum(nil)) + "." + rootDomain
	for i := 0; i < 3; i++ {
		for _, ip := range query(domain) {
			panAnalyticRecord[ip] = struct{}{}
		}
	}
	if len(panAnalyticRecord) > 0 {
		ipList := []string{}
		for ip := range panAnalyticRecord {
			ipList = append(ipList, ip)
		}
		chPanAnalyticRecord <- DNSRecord{Domain: domain, IP: ipList}
	}
	close(chPanAnalyticRecord)
	log.Debugf("pan analytic record: %v\n", panAnalyticRecord)
}
