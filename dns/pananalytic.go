package dns

import (
	"net"
	"strconv"

	"github.com/chuhades/dnsbrute/log"
)

var panAnalyticRecord = map[string]bool{}

func AnalyzePanAnalytic() {
	for i := 0; i < 5; i++ {
		ipList, _ := net.LookupIP(strconv.Itoa(rand.Int()) + "." + rootDomain)
		for _, ip := range ipList {
			panAnalyticRecord[ip.String()] = true
		}
	}
	log.Debug("pan analytic record:", panAnalyticRecord)
}
