package dns

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/miekg/dns"
)

var panDNSBlackList = map[string][]string{}

// queryPanDNS 生成父级域名泛解析黑名单
func queryPanDNS(domain string) (firstTime bool) {
	// 如果父级域名已存在，不再查询
	if _, ok := panDNSBlackList[domain]; ok {
		return
	}

	// md5 域名
	hash := md5.New()
	hash.Write([]byte(domain))
	md5Domain := hex.EncodeToString(hash.Sum(nil))[8:24] + "." + domain

	msg := &dns.Msg{}
	msg.SetQuestion(dns.Fqdn(md5Domain), dns.TypeA)
	in, err := dns.Exchange(msg, dnsServerAddress)
	if err != nil || len(in.Answer) == 0 {
		return
	}

	var rr string
	for _, ans := range in.Answer {
		switch ans := ans.(type) {
		case *dns.CNAME:
			rr = ans.Target
		case *dns.A:
			rr = ans.A.String()
		}
		panDNSBlackList[domain] = append(panDNSBlackList[domain], rr)
	}

	return true
}

// 判断是否是泛解析
func isPanDNS(domain string, response []dns.RR) bool {
	pd := parentDomain(domain)
	firstTime := queryPanDNS(pd)

	// 第一次探测该父级域名，不判定是否是泛解析
	if firstTime {
		return false
	}

	// 无记录，不是泛解析
	records, ok := panDNSBlackList[pd]
	if !ok {
		return false
	}

	// 存在记录，且 CNAME/IP 均存在于黑名单中，是泛解析
	var rr string
	for _, ans := range response {
		switch ans := ans.(type) {
		case *dns.CNAME:
			rr = ans.Target
		case *dns.A:
			rr = ans.A.String()
		}
		if !strInSlice(rr, records) {
			return false
		}
	}

	return true
}
