package main

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

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

var showModules bool

func main() {
	var (
		configPath     string
		installPatch   bool
		uninstallPatch bool
		verifyPatch    bool
	)
	flag.StringVar(&configPath, "config", "config.json", "configuration file path")
	flag.BoolVar(&showModules, "show-modules", false, "show modules when list module")
	flag.BoolVar(&installPatch, "install-patch", false, "only install patch files")
	flag.BoolVar(&uninstallPatch, "uninstall-patch", false, "only uninstall patch files")
	flag.BoolVar(&verifyPatch, "verify-patch", false, "only verify patch files")
	flag.Parse()
	if !config.Load(configPath, &cfg) {
		return
	}
	switch {
	case installPatch:
		installPatchFiles()
	case uninstallPatch:
		uninstallPatchFiles()
	case verifyPatch:
		verifyPatchFiles()
	default:
		installDefault()
	}
}

func installDefault() {
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

func getGoRootPaths() []string {
	list := []string{
		cfg.Common.GoRoot116x,
		cfg.Common.GoRoot1108,
		cfg.Common.GoRoot11113,
		cfg.Common.GoRoot11217,
		cfg.Common.GoRoot11315,
		cfg.Common.GoRoot11415,
		cfg.Common.GoRoot115x,
	}
	paths := make([]string, 0, len(list))
	for i := 0; i < len(list); i++ {
		if list[i] == "" {
			continue
		}
		paths = append(paths, list[i]+"/src")
	}
	return paths
}

func installPatchFiles() bool {
	log.Println(logger.Info, "install patch files")
	paths := getGoRootPaths()
	var errs []error
	walkFunc := func(path string, stat os.FileInfo, err error) error {
		if err != nil {
			log.Println(logger.Error, "appear error in walk function:", err)
			return err
		}
		if stat.IsDir() {
			return nil
		}
		var appearErr bool
		for i := 0; i < len(paths); i++ {
			dst := strings.Replace(path, "patch", paths[i], 1)
			dst = strings.Replace(dst, ".gop", ".go", 1)
			err = copyFileToGoRoot(path, dst)
			if err != nil {
				errs = append(errs, err)
				appearErr = true
			}
		}
		if !appearErr {
			log.Printf(logger.Info, "install patch file: %s", path)
		}
		return nil
	}
	err := filepath.Walk("patch", walkFunc)
	if err != nil {
		log.Println(logger.Error, "failed to walk patch directory:", err)
		return false
	}
	if len(errs) == 0 {
		log.Println(logger.Info, "install patch files successfully")
		return true
	}
	log.Println(logger.Error, "appear error when install patch file")
	for i := 0; i < len(errs); i++ {
		log.Println(logger.Error, errs[i])
	}
	return false
}

func copyFileToGoRoot(src, dst string) error {
	data, err := ioutil.ReadFile(src) // #nosec
	if err != nil {
		return err
	}
	return system.WriteFile(dst, data)
}

func uninstallPatchFiles() bool {
	log.Println(logger.Info, "uninstall patch files")
	paths := getGoRootPaths()
	var errs []error
	walkFunc := func(path string, stat os.FileInfo, err error) error {
		if err != nil {
			log.Println(logger.Error, "appear error in walk function:", err)
			return err
		}
		if stat.IsDir() {
			return nil
		}
		var appearErr bool
		for i := 0; i < len(paths); i++ {
			dst := strings.Replace(path, "patch", paths[i], 1)
			dst = strings.Replace(dst, ".gop", ".go", 1)
			err = os.Remove(dst)
			if err != nil {
				errs = append(errs, err)
				appearErr = true
			}
		}
		if !appearErr {
			log.Printf(logger.Info, "uninstall patch file: %s", path)
		}
		return nil
	}
	err := filepath.Walk("patch", walkFunc)
	if err != nil {
		log.Println(logger.Error, "failed to walk patch directory:", err)
		return false
	}
	if len(errs) == 0 {
		log.Println(logger.Info, "uninstall patch files successfully")
		return true
	}
	log.Println(logger.Error, "appear error when uninstall patch file")
	for i := 0; i < len(errs); i++ {
		log.Println(logger.Error, errs[i])
	}
	return false
}

func verifyPatchFiles() (ok bool) {
	log.Println(logger.Info, "verify patch files")
	if !config.CreateGoModBackup() {
		return false
	}
	defer func() {
		if !config.RestoreGoModBackup() {
			ok = false
			return
		}
		log.Println(logger.Info, "verify patch files successfully")
	}()
	list := []string{
		cfg.Common.GoRoot116x,
		cfg.Common.GoRoot1108,
		cfg.Common.GoRoot11113,
		cfg.Common.GoRoot11217,
		cfg.Common.GoRoot11315,
		cfg.Common.GoRoot11415,
		cfg.Common.GoRoot115x,
	}
	paths := make([]string, 0, len(list))
	for i := 0; i < len(list); i++ {
		if list[i] == "" {
			continue
		}
		paths = append(paths, list[i]+"/bin/go")
	}
	pathsLen := len(paths)
	errCh := make(chan error, pathsLen)
	for _, path := range paths {
		go func(path string) {
			var err error
			defer func() { errCh <- err }()
			output, code, err := exec.Run(path, "run", "script/install/patch/verify.go")
			if err != nil {
				return
			}
			output = output[:len(output)-1] // remove the last "\n"
			if code != 0 {
				err = errors.New(output)
				return
			}
			log.Println(logger.Info, "go run output:\n"+output)
		}(path)
	}
	for i := 0; i < pathsLen; i++ {
		err := <-errCh
		if err != nil {
			log.Println(logger.Error, err)
			return false
		}
	}
	return true
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
	if !showModules {
		return true
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
	log.Println(logger.Info, "verify modules")
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
	log.Println(logger.Info, "verify modules successfully")
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
