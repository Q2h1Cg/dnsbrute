package dns

import "strings"

// ParentDomain 获取父域名
func ParentDomain(domain string) string {
	idx := strings.Index(domain, ".")
	return string([]byte(domain)[idx+1:])
}
