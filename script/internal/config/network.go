package config

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"os"

	"project/internal/logger"

	"project/script/internal/log"
)

// SetProxy is used to set proxy about http.DefaultClient and set os environment.
func SetProxy(proxyURL string) bool {
	if proxyURL == "" {
		return true
	}
	URL, err := url.Parse(proxyURL)
	if err != nil {
		log.Println(logger.Error, "invalid proxy url:", err)
		return false
	}
	tr := http.DefaultTransport.(*http.Transport)
	tr.Proxy = http.ProxyURL(URL)
	// set os environment for build
	err = os.Setenv("HTTP_PROXY", proxyURL)
	if err != nil {
		log.Println(logger.Error, "failed to set HTTP_PROXY:", err)
		return false
	}
	// go1.16, must set HTTPS_PROXY for https URL
	err = os.Setenv("HTTPS_PROXY", proxyURL)
	if err != nil {
		log.Println(logger.Error, "failed to set HTTPS_PROXY:", err)
		return false
	}
	log.Println(logger.Info, "Proxy:", proxyURL)
	return true
}

// SkipTLSVerify is used to skip TLS verify about http.DefaultClient.
func SkipTLSVerify() {
	tr := http.DefaultTransport.(*http.Transport)
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec
	log.Println(logger.Warning, "skip tls verify")
}
