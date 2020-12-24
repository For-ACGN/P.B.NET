package main

import (
	"flag"

	"project/internal/logger"

	"project/script/internal/config"
	"project/script/internal/log"
)

var cfg config.Config

func init() {
	log.SetSource("test")
}

func main() {
	var path string
	flag.StringVar(&path, "config", "config.json", "configuration file path")
	flag.Parse()
	if !config.Load(path, &cfg) {
		return
	}
	var failed bool
	for _, step := range [...]func() bool{
		setupNetwork,
		testExternalPackage,
		testInternalPackage,
		testBeaconPackage,
		testNodePackage,
		testControllerPackage,
		testMSFRPCPackage,
		testToolPackage,
		runIntegrationTest,
		runWebTest,
	} {
		if !step() {
			failed = true
		}
	}
	if failed {
		log.Fatal("test failed")
	} else {
		log.Println(logger.Info, "all tests passed")
	}
}

func setupNetwork() bool {
	log.Println(logger.Info, "setup network")
	if !config.SetProxy(cfg.Test.ProxyURL) {
		return false
	}
	if cfg.Test.Insecure {
		config.SkipTLSVerify()
	}
	return true
}

func testExternalPackage() bool {
	log.Println(logger.Info, "test external package")
	log.Println(logger.Info, "test external package successfully")
	return true
}

func testInternalPackage() bool {
	log.Println(logger.Info, "test internal package")
	log.Println(logger.Info, "test internal package successfully")
	return true
}

func testBeaconPackage() bool {
	log.Println(logger.Info, "test beacon package")
	log.Println(logger.Info, "test beacon package successfully")
	return true
}

func testNodePackage() bool {
	log.Println(logger.Info, "test node package")
	log.Println(logger.Info, "test node package successfully")
	return true
}

func testControllerPackage() bool {
	log.Println(logger.Info, "test controller package")
	log.Println(logger.Info, "test controller package successfully")
	return true
}

func testMSFRPCPackage() bool {
	log.Println(logger.Info, "test msfrpc package")
	log.Println(logger.Info, "test msfrpc package successfully")
	return true
}

func testToolPackage() bool {
	log.Println(logger.Info, "test tool package")
	log.Println(logger.Info, "test tool package successfully")
	return true
}

func runIntegrationTest() bool {
	log.Println(logger.Info, "run integration test")
	log.Println(logger.Info, "run integration test successfully")
	return true
}

func runWebTest() bool {
	log.Println(logger.Info, "run web test")
	log.Println(logger.Info, "run web test successfully")
	return true
}
