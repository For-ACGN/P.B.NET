package main

import (
	"bytes"
	"flag"
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
		configPath    string
		installHook   bool
		uninstallHook bool
	)
	usage := "configuration file path"
	flag.StringVar(&configPath, "config", "config.json", usage)
	usage = "only install runtime patch for hook package"
	flag.BoolVar(&installHook, "install-hook", false, usage)
	usage = "only uninstall runtime patch for hook package"
	flag.BoolVar(&uninstallHook, "uninstall-hook", false, usage)
	flag.Parse()
	if !config.Load(configPath, &cfg) {
		return
	}
	switch {
	case installHook:
		installRuntimePatch()
	case uninstallHook:
		uninstallRuntimePatch()
	default:
		buildStandard()
	}
}

func buildStandard() {
	for _, step := range [...]func() bool{
		installRuntimePatch,
		// then
		uninstallRuntimePatch,
	} {
		if !step() {
			log.Fatal("build failed")
		}
	}
	log.Println(logger.Info, "build successfully")
}

func installRuntimePatch() bool {
	// maybe change to string slice, when new go released.
	path := filepath.Join(cfg.Common.Go116x, "src/runtime/os_windows.go")
	data, err := config.CreateBackup(path)
	if err != nil {
		log.Printf(logger.Error, "failed to create backup about %s: %s", path, err)
		return false
	}
	// replace code
	originCode := []byte("stdcall1(_CloseHandle, mp.thread)\n\t\tmp.thread = 0")
	patchCode := []byte("stdcall1withlasterror(_CloseHandle, mp.thread)\n\t\tmp.thread = 0")
	data = bytes.Replace(data, originCode, patchCode, 1)
	// save changed file
	err = system.WriteFile(path, data)
	if err != nil {
		log.Printf(logger.Error, "failed to save runtime patch file %s: %s", path, err)
		return false
	}
	log.Println(logger.Info, "install runtime patch successfully")
	return true
}

func uninstallRuntimePatch() bool {
	path := filepath.Join(cfg.Common.Go116x, "src/runtime/os_windows.go")
	err := config.RestoreBackup(path)
	if err != nil {
		log.Printf(logger.Error, "failed to restore backup about %s: %s", path, err)
		return false
	}
	log.Println(logger.Info, "uninstall runtime patch successfully")
	return true
}
