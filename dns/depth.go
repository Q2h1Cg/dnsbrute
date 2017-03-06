package dns

import "strings"

var (
	rootDomain string
	depthLimit int	= 3
)

// SetRootDomain 设置根域名
func SetRootDomain(root string) {
	rootDomain = root
}


// SetDepthLimit 设置爆破域名深度
func SetDepthLimit(limit int) {
	depthLimit = limit
}

// DomainDepth 获取域名深度
func DomainDepth(subdomain string) int {
	return strings.Count(strings.Trim(subdomain, rootDomain), ".")
}


// UnderLimit 是否小于深度限制
func UnderLimit(subdomain string) bool {
	return DomainDepth(subdomain) <= depthLimit
}
