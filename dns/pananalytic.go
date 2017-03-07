package dns

import (
	"net"
	"strconv"
)

var PanAnalyticRecord = map[string]struct{}{}

func AnalyzePanAnalytic() {
	for i := 0; i < 5; i++ {
		ipList, _ := net.LookupIP(strconv.Itoa(rand.Int()) + "." + rootDomain)
		for _, ip := range ipList {
			PanAnalyticRecord[ip.String()] = struct{}{}
		}
	}
}
