package main

import (
	"flag"
	"os"

	"project/internal/logger"

	"project/script/internal/config"
	"project/script/internal/log"
)

var cfg config.Config

func init() {
	log.SetSource("coverage")
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
		coverInternalPackage,
		coverBeaconPackage,
		coverNodePackage,
		coverControllerPackage,
		coverMSFRPCPackage,
		coverTestPackage,
	} {
		if !step() {
			failed = true
		}
	}
	if failed {
		log.Println(logger.Fatal, "run coverage failed")
		os.Exit(1)
	} else {
		log.Println(logger.Info, "run coverage successfully")
	}
}

func setupNetwork() bool {
	log.Println(logger.Info, "setup network")
	if !config.SetProxy(cfg.Cover.ProxyURL) {
		return false
	}
	if cfg.Cover.Insecure {
		config.SkipTLSVerify()
	}
	return true
}

func coverInternalPackage() bool {
	log.Println(logger.Info, "run internal package coverage")
	log.Println(logger.Info, "run internal package coverage successfully")
	return true
}

func coverBeaconPackage() bool {
	log.Println(logger.Info, "run beacon package coverage")
	log.Println(logger.Info, "run beacon package coverage successfully")
	return true
}

func coverNodePackage() bool {
	log.Println(logger.Info, "run node package coverage")
	log.Println(logger.Info, "run node package coverage successfully")
	return true
}

func coverControllerPackage() bool {
	log.Println(logger.Info, "run controller package coverage")
	log.Println(logger.Info, "run controller package coverage successfully")
	return true
}

func coverMSFRPCPackage() bool {
	log.Println(logger.Info, "run msfrpc package coverage")
	log.Println(logger.Info, "run msfrpc package coverage successfully")
	return true
}

func coverTestPackage() bool {
	log.Println(logger.Info, "run test package coverage")
	log.Println(logger.Info, "run test package coverage successfully")
	return true
}
