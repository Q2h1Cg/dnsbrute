package dns

import (
	"strconv"

	"github.com/chuhades/dnsbrute/log"

	"github.com/miekg/dns"
)

var panAnalyticRecord = map[string]bool{}

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
func AnalyzePanAnalytic(rootDomain string) {
	for i := 0; i < 5; i++ {
		for _, ip := range query(strconv.Itoa(rand.Int()) + "." + rootDomain) {
			panAnalyticRecord[ip] = true
		}
	}
	log.Debugf("pan analytic record: %v\n", panAnalyticRecord)
}
