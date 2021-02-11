package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"

	"project/internal/security"
	"project/internal/system"

	"project/tool/certificate/manager"
)

var (
	initMgr  bool
	resetPwd bool
	filePath string
)

func init() {
	flag.CommandLine.SetOutput(os.Stdout)

	flag.BoolVar(&initMgr, "init", false, "initialize certificate manager")
	flag.BoolVar(&resetPwd, "reset", false, "reset certificate manager password")
	flag.StringVar(&filePath, "file", "key/certpool.bin", "certificate pool file")
	flag.Parse()
}

func main() {
	mgr := manager.New(os.Stdin, filePath)
	switch {
	case initMgr:
		initialize(mgr)
	case resetPwd:
		resetPassword(mgr)
	default:
		manage(mgr)
	}
}

var stdinFD = int(syscall.Stdin)

func initialize(mgr *manager.Manager) {
	fmt.Print("password: ")
	password, err := term.ReadPassword(stdinFD)
	system.CheckError(err)

	for {
		fmt.Print("\nretype: ")
		retype, err := term.ReadPassword(stdinFD)
		system.CheckError(err)
		if !bytes.Equal(password, retype) {
			fmt.Print("\ndifferent password")
		} else {
			fmt.Println()
			break
		}
	}

	err = mgr.Initialize(password)
	system.CheckError(err)
}

func resetPassword(mgr *manager.Manager) {
	fmt.Print("input old password: ")
	oldPwd, err := term.ReadPassword(stdinFD)
	system.CheckError(err)
	fmt.Println()
	defer security.CoverBytes(oldPwd)

	fmt.Print("input new password: ")
	newPwd, err := term.ReadPassword(stdinFD)
	system.CheckError(err)
	fmt.Println()
	defer security.CoverBytes(newPwd)

	fmt.Print("retype: ")
	rePwd, err := term.ReadPassword(stdinFD)
	system.CheckError(err)
	fmt.Println()
	defer security.CoverBytes(rePwd)

	if !bytes.Equal(newPwd, rePwd) {
		fmt.Println("different password")
		os.Exit(1)
	}
	err = mgr.ResetPassword(oldPwd, newPwd)
	system.CheckError(err)
}

func manage(mgr *manager.Manager) {
	fmt.Println("certificate manager")
	fmt.Println("[*] Remember use \"save\" before exit If you changed")
	fmt.Println()

	fmt.Print("password: ")
	password, err := term.ReadPassword(stdinFD)
	system.CheckError(err)
	fmt.Println()
	defer security.CoverBytes(password)

	// interrupt input
	go func() {
		signalCh := make(chan os.Signal, 1)
		signal.Notify(signalCh, os.Interrupt)
	}()

	err = mgr.Manage(password)
	system.CheckError(err)
}

func checkPassword() bool {
	return true
}
