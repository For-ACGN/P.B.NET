package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"

	"project/internal/interpreter/anko"

	_ "project/internal/interpreter/anko/goroot"
	_ "project/internal/interpreter/anko/project"
	_ "project/internal/interpreter/anko/thirdparty"
)

func main() {
	args := getArgs()

	data, err := ioutil.ReadFile(os.Args[1])
	checkError(err)
	stmt, err := anko.ParseSrc(string(data))
	checkError(err)

	env := anko.NewEnv()
	_ = env.Define("args", args)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)
	go func() {
		<-ch
		cancel()
	}()

	_, err = anko.RunContext(ctx, env, stmt)
	checkError(err)
}

func getArgs() []string {
	args := os.Args
	switch len(args) {
	case 1:
		fmt.Println("no input file")
		os.Exit(1)
	case 2:
	default:
		return args[2:]
	}
	return nil
}

func checkError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
