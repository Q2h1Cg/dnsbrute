package dns

import "strings"

var rootDomain string

// trimSuffixPoint 移除完整域名尾部的 .
func trimSuffixPoint(domain string) string {
	return strings.TrimSuffix(domain, ".")
}

// parentDomain 父域名
func parentDomain(domain string) string {
	if domain == rootDomain {
		return rootDomain
	}

	return strings.SplitN(domain, ".", 2)[1]
}

// strInSlice 判断字符串是否在切片中
func strInSlice(s string, sl []string) bool {
	for _, ss := range sl {
		if s == ss {
			return true
		}
	}
	return false
}
