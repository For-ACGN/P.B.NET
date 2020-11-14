package main

import (
	"flag"
	"fmt"
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

var cfg config.Config

func init() {
	log.SetSource("install")
}

func main() {
	var path string
	flag.StringVar(&path, "config", "config.json", "configuration file path")
	flag.Parse()
	if !config.Load(path, &cfg) {
		return
	}
	for _, step := range [...]func() bool{
		installPatchFiles,
		listModule,
		downloadAllModules,
		verifyModule,
		downloadModule,
	} {
		if !step() {
			return
		}
	}
	log.Println(logger.Info, "install successfully")
}

func installPatchFiles() bool {
	latest := fmt.Sprintf("%s/src", cfg.Common.GoRootLatest)
	go1108 := fmt.Sprintf("%s/src", cfg.Common.GoRoot1108)
	dirs := []string{latest, go1108}
	walkFunc := func(path string, stat os.FileInfo, err error) error {
		if err != nil {
			log.Println(logger.Error, "appear error in walk function:", err)
			return err
		}
		if stat.IsDir() {
			return nil
		}
		for i := 0; i < len(dirs); i++ {
			dst := strings.Replace(path, "patch", dirs[i], 1)
			dst = strings.Replace(dst, ".gop", ".go", 1)
			err = copyFileToGoRoot(path, dst)
			if err != nil {
				return err
			}
		}
		log.Printf(logger.Info, "install patch file: %s", path)
		return nil
	}
	err := filepath.Walk("patch", walkFunc)
	if err != nil {
		log.Println(logger.Error, "failed to walk patch directory:", err)
		return false
	}
	log.Println(logger.Info, "install all patch files to go root path")
	return true
}

func copyFileToGoRoot(src, dst string) error {
	data, err := ioutil.ReadFile(src) // #nosec
	if err != nil {
		return err
	}
	return system.WriteFile(dst, data)
}

func listModule() bool {
	log.Println(logger.Info, "list all modules about project")
	output, code, err := exec.Run("go", "list", "-m", "all")
	if code != 0 {
		log.Println(logger.Error, output)
		if err != nil {
			log.Println(logger.Error, err)
		}
		return false
	}
	output = output[:len(output)-1] // remove the last "\n"
	modules := strings.Split(output, "\n")
	modules = modules[1:] // remove the first module "project"
	for i := 0; i < len(modules); i++ {
		log.Println(logger.Info, modules[i])
	}
	return true
}

func downloadAllModules() bool {
	if !cfg.Install.DownloadAll {
		return true
	}
	log.Println(logger.Info, "download all modules")
	output, code, err := exec.Run("go", "mod", "download", "-x")
	if code != 0 {
		log.Println(logger.Error, output)
		if err != nil {
			log.Println(logger.Error, err)
		}
		return false
	}
	log.Println(logger.Info, "download all modules successfully")
	return true
}

func verifyModule() bool {
	output, code, err := exec.Run("go", "mod", "verify")
	if err != nil {
		log.Println(logger.Error, err)
		return false
	}
	output = output[:len(output)-1] // remove the last "\n"
	log.Println(logger.Info, output)
	if code != 0 {
		return false
	}
	log.Println(logger.Info, "verify module successfully")
	return true
}

func downloadModule() bool {
	log.Println(logger.Info, "download module if it is not exist")
	output, code, err := exec.Run("go", "get", "./...")
	if code != 0 {
		log.Println(logger.Error, output)
		if err != nil {
			log.Println(logger.Error, err)
		}
		return false
	}
	log.Println(logger.Info, "all modules downloaded")
	return true
}
