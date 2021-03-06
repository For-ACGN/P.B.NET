package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"project/internal/logger"
	"project/internal/patch/json"
	"project/internal/system"

	"project/script/internal/log"
)

// Config contains configuration about install, build, develop and test.
type Config struct {
	// Common contains common configuration about script, these field
	// must be set except GoProxy and GoSumDB.
	// <security> must be set manually for prevent leak user information.
	Common struct {
		Go116x  string `json:"go_1_16_x"`
		Go1108  string `json:"go_1_10_8"`
		GoPath  string `json:"go_path"`
		GoProxy string `json:"go_proxy"`
		GoSumDB string `json:"go_sum_db"`
	} `json:"common"`

	// Specific contains go root path that if you need specific go version.
	Specific struct {
		Go11113 string `json:"go_1_11_13"`
		Go11217 string `json:"go_1_12_17"`
		Go11315 string `json:"go_1_13_15"`
		Go11415 string `json:"go_1_14_15"`
		Go115x  string `json:"go_1_15_x"`
	} `json:"specific"`

	// Environment contains environments about go.
	Environment struct {
		GoPrivate  string `json:"go_private"`
		GoInsecure string `json:"go_insecure"`
		GoNoProxy  string `json:"go_no_proxy"`
		GoNoSumDB  string `json:"go_no_sum_db"`
	} `json:"environment"`

	// Install contains options about install script.
	Install struct {
		ProxyURL    string `json:"proxy_url"`
		Insecure    bool   `json:"insecure"`
		ShowModules bool   `json:"show_modules"`
		DownloadAll bool   `json:"download_all"`
	} `json:"install"`

	// Build contains options about build script.
	Build struct {
	} `json:"build"`

	// Develop contains options about develop script.
	Develop struct {
		ProxyURL string `json:"proxy_url"`
		Insecure bool   `json:"insecure"`
	} `json:"develop"`

	// Sum contains options about sum script.
	Sum struct {
		ProxyURL string `json:"proxy_url"`
	} `json:"sum"`

	// Test contains options about test script.
	Test struct {
		ProxyURL    string `json:"proxy_url"`
		Insecure    bool   `json:"insecure"`
		DisableRace bool   `json:"disable_race"`
	} `json:"test"`

	// Cover contains options about cover script.
	Cover struct {
		ProxyURL    string `json:"proxy_url"`
		Insecure    bool   `json:"insecure"`
		DisableRace bool   `json:"disable_race"`
	} `json:"cover"`
}

// Load is used to load and verify configuration file.
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
	if !setGoEnv(config) {
		return false
	}
	if !verifyGoRoot(config) {
		return false
	}
	log.Println(logger.Info, "load configuration file successfully")
	return true
}

// setGoEnv is used to print and set go environment.
func setGoEnv(config *Config) bool {
	for _, item := range [...]*struct {
		name    string
		value   string
		mustSet bool
	}{
		{"GOPATH", config.Common.GoPath, true},
		{"GOPROXY", config.Common.GoProxy, false},
		{"GOSUMDB", config.Common.GoSumDB, false},
		{"GOPRIVATE", config.Environment.GoPrivate, false},
		{"GOINSECURE", config.Environment.GoInsecure, false},
		{"GONOPROXY", config.Environment.GoNoProxy, false},
		{"GONOSUMDB", config.Environment.GoNoSumDB, false},
	} {
		if item.value == "" {
			if item.mustSet {
				log.Printf(logger.Error, "%s is not set", item.name)
				return false
			}
			continue
		}
		log.Printf(logger.Info, "%s: %s", item.name, item.value)
		err := os.Setenv(item.name, item.value)
		if err != nil {
			log.Printf(logger.Error, "failed to set env %s: %s", item.name, err)
			return false
		}
	}
	return true
}

// verifyGoRoot is used to check go root path is valid, it will check
// go.exe, gofmt.exe and src directory, go latest and go 1.10.8 must be set.
func verifyGoRoot(config *Config) bool {
	for _, item := range [...]*struct {
		version string
		path    string
	}{
		// common (must set)
		{version: "1.16.x", path: config.Common.Go116x},
		{version: "1.10.8", path: config.Common.Go1108},
		// specific
		{version: "1.11.13", path: config.Specific.Go11113},
		{version: "1.12.17", path: config.Specific.Go11217},
		{version: "1.13.15", path: config.Specific.Go11315},
		{version: "1.14.15", path: config.Specific.Go11415},
		{version: "1.15.x", path: config.Specific.Go115x},
	} {
		// skip empty go root path in specific
		if item.path == "" {
			// check this go version can be skipped
			var notSkip bool
			for _, version := range []string{
				"1.16.x", "1.16.x_original", "1.10.8", "1.10.8_original",
			} {
				if item.version == version {
					notSkip = true
					break
				}
			}
			if !notSkip {
				continue
			}
			log.Printf(logger.Error, "go %-7s must be set", item.version)
			return false
		}
		// verify go root path
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
		goExist, _ := system.IsExist(filepath.Join(item.path, "bin/"+goFile))
		goFmtExist, _ := system.IsExist(filepath.Join(item.path, "bin/"+goFmtFile))
		srcExist, _ := system.IsExist(filepath.Join(item.path, "src"))
		if !(goExist && goFmtExist && srcExist) {
			log.Printf(logger.Error, "invalid go %-7s root path: %s", item.version, item.path)
			return false
		}
		log.Printf(logger.Info, "go %-7s root path: %s", item.version, item.path)
	}
	return true
}
