package main

import (
	"flag"
	"os"
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
	switch {
	case initMgr:

	case resetPwd:

	default:

	}
}
