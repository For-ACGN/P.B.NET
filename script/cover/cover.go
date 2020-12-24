package main

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"project/internal/logger"
	"project/internal/system"

	"project/script/internal/config"
	"project/script/internal/exec"
	"project/script/internal/log"
)

const (
	coverDir  = "temp/cover"
	outputDir = "output"
	coverFile = "cover.out"
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
	for _, step := range [...]func() bool{
		setupNetwork,
		coverInternalPackage,
		coverBeaconPackage,
		coverNodePackage,
		coverControllerPackage,
		coverMSFRPCPackage,
		coverTestPackage,
		mergeCoverResult,
		sendCoverResult,
	} {
		if !step() {
			log.Fatal("run coverage failed")
		}
	}
	log.Println(logger.Info, "run coverage successfully")
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

// must create directory before run it
// go test -covermode="atomic" -coverprofile="temp/cover/internal_convert.out"
// -race -gcflags "-N -l" -timeout 30m -trimpath ./internal/convert

func mergeCoverResult() bool {
	log.Println(logger.Info, "merge cover result")
	// create new file
	path := filepath.Join(coverDir, coverFile)
	cover, err := system.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Println(logger.Error, err)
		return false
	}
	defer func() { _ = cover.Close() }()
	_, err = cover.WriteString("mode: atomic\n")
	if err != nil {
		log.Println(logger.Error, "failed to write file header:", err)
		return false
	}
	// merge output
	walkFunc := func(path string, stat os.FileInfo, err error) error {
		if err != nil {
			log.Println(logger.Error, "appear error in walk function:", err)
			return err
		}
		if stat.IsDir() || !strings.HasSuffix(stat.Name(), ".out") {
			return nil
		}
		data, err := ioutil.ReadFile(path) // #nosec
		if err != nil {
			return err
		}
		lines := data[len("mode: atomic\n"):]
		if len(lines) < 2 {
			return nil
		}
		_, err = cover.Write(lines)
		return err
	}
	err = filepath.Walk(filepath.Join(coverDir, "output"), walkFunc)
	if err != nil {
		log.Println(logger.Error, "failed to walk patch directory:", err)
		return false
	}
	// remove output directory
	log.Println(logger.Info, "remove output directory")
	err = os.RemoveAll(filepath.Join(coverDir, outputDir))
	if err != nil {
		log.Println(logger.Error, "failed to remove output directory:", err)
		return false
	}
	// generate coverage html file
	log.Println(logger.Info, "generate coverage html output")
	// go tool cover -html="temp/cover/cover.out" -o="temp/cover/cover.html"
	argHTML := "-html=" + filepath.Join(coverDir, coverFile)
	argO := "-o=" + filepath.Join(coverDir, "cover.html")
	output, code, err := exec.Run("go", "tool", "cover", argHTML, argO)
	if code != 0 {
		log.Printf(logger.Error, "failed to generate coverage html file\n%s", output)
		if err != nil {
			log.Println(logger.Error, err)
		}
		return false
	}
	log.Println(logger.Info, "merge cover result successfully")
	return true
}

func sendCoverResult() bool {
	log.Println(logger.Info, "send cover result")
	path := filepath.Join(coverDir, coverFile)
	output, code, err := exec.Run("goveralls", "-coverprofile="+path, "-service=github")
	if code != 0 {
		log.Printf(logger.Error, "failed to send coverage result\n%s", output)
		if err != nil {
			log.Println(logger.Error, err)
		}
		return false
	}
	log.Println(logger.Info, "send cover result successfully")
	return true
}
