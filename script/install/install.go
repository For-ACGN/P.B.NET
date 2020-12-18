package main

import (
	"context"
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

var cfg config.Config

func init() {
	log.SetSource("install")
}

func main() {
	var (
		configPath     string
		installPatch   bool
		uninstallPatch bool
		verifyPatch    bool
	)
	usage := "configuration file path"
	flag.StringVar(&configPath, "config", "config.json", usage)
	usage = "only install patch files"
	flag.BoolVar(&installPatch, "install-patch", false, usage)
	usage = "only uninstall patch files"
	flag.BoolVar(&uninstallPatch, "uninstall-patch", false, usage)
	usage = "only verify patch files"
	flag.BoolVar(&verifyPatch, "verify-patch", false, usage)
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
		setupNetwork,
		listModule,
		downloadAllModules,
		downloadModule,
		verifyModule,
	} {
		if !step() {
			log.Println(logger.Fatal, "install failed")
			return
		}
	}
	log.Println(logger.Info, "install successfully")
}

func getGoRootPaths(suffix string) []string {
	list := []string{
		cfg.Common.Go116x,
		cfg.Common.Go1108,
		cfg.Special.Go11113,
		cfg.Special.Go11217,
		cfg.Special.Go11315,
		cfg.Special.Go11415,
		cfg.Special.Go115x,
	}
	paths := make([]string, 0, len(list))
	for i := 0; i < len(list); i++ {
		if list[i] == "" {
			continue
		}
		paths = append(paths, list[i]+suffix)
	}
	return paths
}

func installPatchFiles() bool {
	log.Println(logger.Info, "install patch files")
	paths := getGoRootPaths("/src")
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
	paths := getGoRootPaths("/src")
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

func verifyPatchFiles() bool {
	log.Println(logger.Info, "verify patch files")
	// prevent change go.mod file
	err := os.Setenv("GO111MODULE", "off")
	if err != nil {
		log.Println(logger.Error, "failed to disable go module:", err)
		return false
	}
	paths := getGoRootPaths("/bin/go")
	errCh := make(chan error, len(paths))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for _, path := range paths {
		go func(path string) {
			const file = "script/install/patch/verify.go"
			var err error
			defer func() { errCh <- err }()
			output, _, err := exec.RunContext(ctx, path, "run", file)
			output = output[:len(output)-1] // remove the last "\n"
			log.Printf(logger.Info, "go run output:\n%s", output)
		}(path)
	}
	for i := 0; i < len(paths); i++ {
		err := <-errCh
		if err != nil {
			log.Println(logger.Error, err)
			return false
		}
	}
	// recover go environment
	err = os.Setenv("GO111MODULE", "on")
	if err != nil {
		log.Println(logger.Error, "failed to enable go module:", err)
		return false
	}
	log.Println(logger.Info, "verify patch files successfully")
	return true
}

func setupNetwork() bool {
	if !config.SetProxy(cfg.Install.ProxyURL) {
		return false
	}
	if cfg.Install.Insecure {
		config.SkipTLSVerify()
	}
	return true
}

func listModule() bool {
	log.Println(logger.Info, "list all modules about project")
	output, code, err := exec.Run("go", "list", "-m", "all")
	if code != 0 {
		log.Printf(logger.Error, "failed to list module\n%s", output)
		if err != nil {
			log.Println(logger.Error, err)
		}
		return false
	}
	if !cfg.Install.ShowModules {
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
		log.Printf(logger.Error, "failed to download all modules\n%s", output)
		if err != nil {
			log.Println(logger.Error, err)
		}
		return false
	}
	log.Println(logger.Info, "download all modules successfully")
	return true
}

func downloadModule() bool {
	log.Println(logger.Info, "download module if it is not exist")
	args := []string{"get", "-d"}
	if cfg.Install.Insecure {
		args = append(args, "-insecure")
	}
	args = append(args, "./...")
	output, code, err := exec.Run("go", args...)
	if code != 0 {
		log.Printf(logger.Error, "failed to download module\n%s", output)
		if err != nil {
			log.Println(logger.Error, err)
		}
		return false
	}
	log.Println(logger.Info, "all modules downloaded")
	return true
}

func verifyModule() bool {
	log.Println(logger.Info, "verify modules")
	output, code, err := exec.Run("go", "mod", "verify")
	output = output[:len(output)-1] // remove the last "\n"
	if code != 0 {
		log.Printf(logger.Error, "some modules has been modified\n%s", output)
		if err != nil {
			log.Println(logger.Error, err)
		}
		return false
	}
	log.Println(logger.Info, output)
	log.Println(logger.Info, "verify modules successfully")
	return true
}
