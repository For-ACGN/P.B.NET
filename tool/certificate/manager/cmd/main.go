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

func banner() {
	fmt.Println()
	fmt.Println("  ::::::::  :::::::::: :::::::::  ::::::::::: ::::    ::::   ::::::::  ::::::::: ")
	fmt.Println(" :+:    :+: :+:        :+:    :+:     :+:     +:+:+: :+:+:+ :+:    :+: :+:    :+:")
	fmt.Println(" +:+        +:+        +:+    +:+     +:+     +:+ +:+:+ +:+ +:+        +:+    +:+")
	fmt.Println(" +#+        +#++:++#   +#++:++#:      +#+     +#+  +:+  +#+ :#:        +#++:++#: ")
	fmt.Println(" +#+        +#+        +#+    +#+     +#+     +#+       +#+ +#+   +#+# +#+    +#+")
	fmt.Println(" #+#    #+# #+#        #+#    #+#     #+#     #+#       #+# #+#    #+# #+#    #+#")
	fmt.Println("  ########  ########## ###    ###     ###     ###       ###  ########  ###    ###")
	fmt.Println()
	fmt.Println("[*] Remember use \"save\" before exit if you changed")
	fmt.Println("[*] Use \"reload\" If you accidentally delete certificate")
	fmt.Println("[*] You can find the backup file in the destination path")
	fmt.Println()
}

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

	banner()
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
	fmt.Println()

	err = security.CheckPasswordStrength(password)
	system.CheckError(err)

	for {
		fmt.Print("retype: ")
		retype, err := term.ReadPassword(stdinFD)
		system.CheckError(err)
		fmt.Println()

		if !bytes.Equal(password, retype) {
			fmt.Println("different password")
		} else {
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
	defer security.CoverBytes(oldPwd)
	fmt.Println()

	fmt.Print("input new password: ")
	newPwd, err := term.ReadPassword(stdinFD)
	system.CheckError(err)
	defer security.CoverBytes(newPwd)
	fmt.Println()

	err = security.CheckPasswordStrength(newPwd)
	system.CheckError(err)

	if bytes.Equal(oldPwd, newPwd) {
		system.PrintError("as same as the old password")
	}

	fmt.Print("retype: ")
	rePwd, err := term.ReadPassword(stdinFD)
	system.CheckError(err)
	defer security.CoverBytes(rePwd)
	fmt.Println()

	if !bytes.Equal(newPwd, rePwd) {
		system.PrintError("different password")
	}

	err = mgr.ResetPassword(oldPwd, newPwd)
	system.CheckError(err)
}

func manage(mgr *manager.Manager) {
	fmt.Print("password: ")
	password, err := term.ReadPassword(stdinFD)
	system.CheckError(err)
	defer security.CoverBytes(password)
	fmt.Println()

	// interrupt input
	go func() {
		signalCh := make(chan os.Signal, 1)
		signal.Notify(signalCh, os.Interrupt)
	}()

	err = mgr.Manage(password)
	system.CheckError(err)
}
