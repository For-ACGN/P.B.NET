package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/mod/modfile"

	"project/internal/logger"

	"project/script/internal/config"
	"project/script/internal/exec"
	"project/script/internal/log"
)

var cfg config.Config

func init() {
	log.SetSource("sum")
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
		rebuildGoSum,
	} {
		if !step() {
			log.Fatal("rebuild failed")
		}
	}
	log.Println(logger.Info, "rebuild successfully")
}

func setupNetwork() bool {
	log.Println(logger.Info, "setup network")
	return config.SetProxy(cfg.Sum.ProxyURL)
}

func rebuildGoSum() (ok bool) {
	log.Println(logger.Info, "rebuild go sum file")
	data, err := ioutil.ReadFile("go.mod")
	if err != nil {
		log.Println(logger.Error, "failed to read go module file")
		return false
	}
	file, err := modfile.Parse("go.mod", data, nil)
	if err != nil {
		log.Println(logger.Error, "failed to parse go module file")
		return false
	}
	// create backup
	if !config.CreateGoSumBackup() {
		return false
	}
	defer func() {
		if !ok {
			config.RestoreGoSumBackup()
		} else {
			ok = config.RemoveGoSumBackup()
		}
	}()
	// update go sum file
	goSumFile, err := os.OpenFile("go.sum", os.O_WRONLY, 0)
	if err != nil {
		log.Println(logger.Error, err)
		return false
	}
	defer func() { _ = goSumFile.Close() }()
	// clean old file
	err = goSumFile.Truncate(0)
	if err != nil {
		log.Println(logger.Error, "failed to clean go sum file:", err)
		return false
	}
	// call "go get" for build hash
	for _, require := range file.Require {
		pkg := fmt.Sprintf("%s@%s", require.Mod.Path, require.Mod.Version)
		output, code, err := exec.Run("go", "get", "-d", pkg)
		if code != 0 {
			log.Printf(logger.Error, "failed to build hash about %s\n%s", pkg, output)
			if err != nil {
				log.Println(logger.Error, err)
			}
			return false
		}
		log.Println(logger.Info, "build", pkg)
	}
	ok = true
	log.Println(logger.Info, "rebuild go sum file successfully")
	return true
}
