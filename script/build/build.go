package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"

	"project/internal/logger"
	"project/internal/system"

	"project/script/internal/config"
	"project/script/internal/log"
)

var cfg config.Config

func init() {
	log.SetSource("build")
}

func main() {
	var (
		path          string
		installHook   bool
		uninstallHook bool
	)
	flag.StringVar(&path, "config", "config.json", "configuration file path")
	flag.BoolVar(&installHook, "install-hook", false, "only install runtime patch for hook package")
	flag.BoolVar(&uninstallHook, "uninstall-hook", false, "only uninstall runtime patch for hook package")
	flag.Parse()
	if !config.Load(path, &cfg) {
		return
	}
	switch {
	case installHook:
		installRuntimePatch()
		log.Println(logger.Info, "install runtime patch for hook package successfully")
		return
	case uninstallHook:
		uninstallRuntimePatch()
		log.Println(logger.Info, "uninstall runtime patch for hook package successfully")
		return
	}
	for _, step := range [...]func() bool{
		installRuntimePatch,
		uninstallRuntimePatch,
	} {
		if !step() {
			return
		}
	}
	log.Println(logger.Info, "build successfully")
}

func createBackup(path string) ([]byte, error) {
	data, err := ioutil.ReadFile(path) // #nosec
	if err != nil {
		return nil, err
	}
	err = os.Rename(path, path+".bak")
	if err != nil {
		return nil, err
	}
	err = system.WriteFile(path, data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func restoreBackup(path string) error {
	return os.Rename(path+".bak", path)
}

func installRuntimePatch() bool {
	// maybe change to string slice
	path := filepath.Join(cfg.Common.GoRootLatest, "src/runtime/os_windows.go")
	data, err := createBackup(path)
	if err != nil {
		log.Printf(logger.Error, "failed to create backup about %s: %s", path, err)
		return false
	}
	// replace code
	originCode := []byte("stdcall1(_CloseHandle, mp.thread)\n\t\tmp.thread = 0")
	patchCode := []byte("stdcall1withlasterror(_CloseHandle, mp.thread)\n\t\tmp.thread = 0")
	data = bytes.Replace(data, originCode, patchCode, 1)
	// save file
	err = system.WriteFile(path, data)
	if err != nil {
		log.Printf(logger.Error, "failed to save patch file %s: %s", path, err)
		return false
	}
	return true
}

func uninstallRuntimePatch() bool {
	path := filepath.Join(cfg.Common.GoRootLatest, "src/runtime/os_windows.go")
	err := restoreBackup(path)
	if err != nil {
		log.Printf(logger.Error, "failed to restore backup about %s: %s", path, err)
		return false
	}
	return true
}
