package config

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"

	"project/internal/logger"
	"project/internal/patch/json"
	"project/internal/system"

	"project/script/internal/log"
)

// Config contains configuration about install, build, develop, test and race.
type Config struct {
	Common struct {
		// must be set
		// <security> must be set manually for prevent
		// leak user information.
		GoPath     string `json:"go_path"`
		GoRoot116x string `json:"go_root_1_16_x"`
		GoRoot1108 string `json:"go_root_1_10_8"`

		// options if you need special go version.
		GoRoot11113 string `json:"go_root_1_11_13"`
		GoRoot11217 string `json:"go_root_1_12_17"`
		GoRoot11315 string `json:"go_root_1_13_15"`
		GoRoot11415 string `json:"go_root_1_14_15"`
		GoRoot115x  string `json:"go_root_1_15_x"`

		// options about network
		ProxyURL      string `json:"proxy_url"`
		SkipTLSVerify bool   `json:"skip_tls_verify"`
	} `json:"common"`

	Install struct {
		DownloadAll bool `json:"download_all"`
	} `json:"install"`

	Build struct {
	} `json:"build"`

	Develop struct {
	} `json:"develop"`

	Test struct {
	} `json:"test"`

	Race struct {
	} `json:"race"`
}

// Load is used to load configuration file.
func Load(path string, config *Config) bool {
	// print current directory
	dir, err := os.Getwd()
	if err != nil {
		log.Println(logger.Error, err)
		return false
	}
	log.Println(logger.Info, "current directory:", dir)
	// load config file
	data, err := ioutil.ReadFile(path) // #nosec
	if err != nil {
		log.Println(logger.Error, "failed to load config file:", err)
		return false
	}
	err = json.Unmarshal(data, config)
	if err != nil {
		log.Println(logger.Error, "failed to load config:", err)
		return false
	}
	log.Println(logger.Info, "load configuration file successfully")
	// print and set go path
	if !setGoPath(config.Common.GoPath) {
		return false
	}
	// check go root path, must need go latest and go 1.10.8
	for _, item := range [...]*struct {
		version string
		path    string
	}{
		{version: "1.16.x", path: config.Common.GoRoot116x},
		{version: "1.10.8", path: config.Common.GoRoot1108},
		{version: "1.11.13", path: config.Common.GoRoot11113},
		{version: "1.12.17", path: config.Common.GoRoot11217},
		{version: "1.13.15", path: config.Common.GoRoot11315},
		{version: "1.14.15", path: config.Common.GoRoot11415},
		{version: "1.15.x", path: config.Common.GoRoot115x},
	} {
		if item.path == "" && item.version != "1.16.x" && item.version != "1.10.8" {
			continue
		}
		if !checkGoRoot(item.path) {
			log.Printf(logger.Error, "invalid go %-7s root path: %s", item.version, item.path)
			return false
		}
		log.Printf(logger.Info, "go %-7s root path: %s", item.version, item.path)
	}
	// set proxy and TLS configuration
	tr := http.DefaultTransport.(*http.Transport)
	if !setProxy(tr, config) {
		return false
	}
	if config.Common.SkipTLSVerify {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec
		log.Println(logger.Warning, "skip tls verify")
	}
	return true
}

func setGoPath(goPath string) bool {
	if goPath == "" {
		log.Println(logger.Error, "go path is empty")
		return false
	}
	log.Println(logger.Info, "go path:", goPath)
	// set os environment for build
	err := os.Setenv("GOPATH", goPath)
	if err != nil {
		log.Println(logger.Error, "failed to set environment about GOPATH:", err)
		return false
	}
	return true
}

// checkGoRoot is used to check go root path is valid.
// it will check go.exe, gofmt.exe and src directory.
func checkGoRoot(path string) bool {
	var (
		goFile    string
		goFmtFile string
	)
	switch runtime.GOOS {
	case "windows":
		goFile = "go.exe"
		goFmtFile = "gofmt.exe"
	default:
		goFile = "go"
		goFmtFile = "gofmt"
	}
	goExist, _ := system.IsExist(filepath.Join(path, "bin/"+goFile))
	goFmtExist, _ := system.IsExist(filepath.Join(path, "bin/"+goFmtFile))
	srcExist, _ := system.IsExist(filepath.Join(path, "src"))
	return goExist && goFmtExist && srcExist
}

func setProxy(tr *http.Transport, cfg *Config) bool {
	proxyURL := cfg.Common.ProxyURL
	if proxyURL == "" {
		return true
	}
	URL, err := url.Parse(proxyURL)
	if err != nil {
		log.Println(logger.Error, "invalid proxy url:", err)
		return false
	}
	tr.Proxy = http.ProxyURL(URL)
	// set os environment for build
	err = os.Setenv("HTTP_PROXY", proxyURL)
	if err != nil {
		log.Println(logger.Error, "failed to set environment about HTTP_PROXY:", err)
		return false
	}
	// go1.16, must set HTTPS_PROXY for https URL
	err = os.Setenv("HTTPS_PROXY", proxyURL)
	if err != nil {
		log.Println(logger.Error, "failed to set environment about HTTPS_PROXY:", err)
		return false
	}
	log.Println(logger.Info, "set proxy url:", proxyURL)
	return true
}
