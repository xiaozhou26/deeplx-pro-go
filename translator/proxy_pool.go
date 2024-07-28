package translator

import (
	"fmt"
	"log"
	"os"
	"strings"
)

var proxies []string
var invalidProxies []string
var currentProxyIndex int

func validateProxies() {
	proxyStr := os.Getenv("PROXY_LIST")
	if proxyStr == "" {
		log.Fatal("No proxies provided. Please check your PROXY_LIST environment variable.")
	}
	proxyList := strings.Split(proxyStr, ",")
	for _, proxy := range proxyList {
		proxy = strings.TrimSpace(proxy)
		proxies = append(proxies, proxy)
	}
	if len(proxies) == 0 {
		log.Fatal("No valid proxies provided. Please check your PROXY_LIST environment variable.")
	}
}

func GetNextProxy() (string, error) {
	attempts := 0
	for attempts < len(proxies) {
		proxy := proxies[currentProxyIndex]
		if !stringSliceContains(invalidProxies, proxy) {
			currentProxyIndex = (currentProxyIndex + 1) % len(proxies)
			return proxy, nil
		}
		currentProxyIndex = (currentProxyIndex + 1) % len(proxies)
		attempts++
	}
	return "", fmt.Errorf("no valid proxies available")
}

func MarkProxyInvalid(proxy string) {
	if !stringSliceContains(invalidProxies, proxy) {
		invalidProxies = append(invalidProxies, proxy)
	}
}

func initProxies() {
	validateProxies()
}
