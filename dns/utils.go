package dns

import "strings"

var rootDomain string

// SetRootDomain 设置根域名
func SetRootDomain(domain string) {
	rootDomain = domain
}

// TrimSuffixPoint 去除域名结尾的 .
func TrimSuffixPoint(s string) string {
	return strings.TrimSuffix(s, ".")
}

// IsSubdomain 判断是否是子域名
func IsSubdomain(domain string) bool {
	return strings.HasSuffix(TrimSuffixPoint(domain), "."+rootDomain)
}
