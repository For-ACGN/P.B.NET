package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"

	"project/internal/logger"
	"project/internal/module/filemgr"
	"project/internal/system"

	"project/script/internal/config"
	"project/script/internal/log"
)

const developDir = "temp/develop"

var cfg config.Config

func init() {
	log.SetSource("develop")
}

func main() {
	var path string
	flag.StringVar(&path, "config", "config.json", "configuration file path")
	flag.Parse()
	if !config.Load(path, &cfg) {
		return
	}
	for _, step := range [...]func() bool{
		downloadSourceCode,
		extractSourceCode,
		buildSourceCode,
	} {
		if !step() {
			return
		}
	}
	log.Println(logger.Info, "install development tools successfully")
}

func downloadSourceCode() bool {
	items := [...]*struct {
		name string
		url  string
	}{
		{name: "golint", url: "https://github.com/golang/lint/archive/master.zip"},
		{name: "gocyclo", url: "https://github.com/fzipp/gocyclo/archive/main.zip"},
		{name: "gosec", url: "https://github.com/securego/gosec/archive/master.zip"},
		{name: "golangci-lint", url: "https://github.com/golangci/golangci-lint/archive/master.zip"},
		{name: "go-tools", url: "https://github.com/golang/tools/archive/master.zip"},
	}
	itemsLen := len(items)
	errCh := make(chan error, itemsLen)
	for _, item := range items {
		go func(name, url string) {
			var err error
			defer func() { errCh <- err }()
			resp, err := http.Get(url) // #nosec
			if err != nil {
				return
			}
			defer func() { _ = resp.Body.Close() }()
			// get file size
			size, _ := strconv.Atoi(resp.Header.Get("Content-Length"))
			if size == 0 {
				size = 1024 * 1024
			}
			buf := bytes.NewBuffer(make([]byte, 0, size))
			// download file
			log.Printf(logger.Info, "downloading %s url: %s", name, url)
			_, err = io.Copy(buf, resp.Body)
			if err != nil {
				return
			}
			// write file
			filename := fmt.Sprintf(developDir+"/%s.zip", name)
			err = system.WriteFile(filename, buf.Bytes())
			if err != nil {
				return
			}
			log.Printf(logger.Info, "download %s successfully", name)
		}(item.name, item.url)
	}
	for i := 0; i < itemsLen; i++ {
		err := <-errCh
		if err != nil {
			log.Println(logger.Error, "failed to download source code:", err)
			return false
		}
	}
	log.Println(logger.Info, "download all source code successfully")
	return true
}

func extractSourceCode() bool {
	items := [...]*struct {
		name string
		dir  string
	}{
		{name: "golint", dir: "lint-master"},
		{name: "gocyclo", dir: "gocyclo-main"},
		{name: "gosec", dir: "gosec-master"},
		{name: "golangci-lint", dir: "golangci-lint-master"},
		{name: "go-tools", dir: "tools-master"},
	}
	itemsLen := len(items)
	errCh := make(chan error, itemsLen)
	for _, item := range items {
		go func(name, dir string) {
			var err error
			defer func() { errCh <- err }()
			// clean directory
			dir = filepath.Join(developDir, dir)
			exist, err := system.IsExist(dir)
			if err != nil {
				return
			}
			if exist {
				err = os.RemoveAll(dir)
				if err != nil {
					return
				}
			}
			src := developDir + "/" + name + ".zip"
			// extract files
			var uErr error
			ec := func(_ context.Context, typ uint8, err error, _ *filemgr.SrcDstStat) uint8 {
				if typ == filemgr.ErrCtrlSameFile {
					return filemgr.ErrCtrlOpReplace
				}
				uErr = err
				return filemgr.ErrCtrlOpCancel
			}
			err = filemgr.UnZip(ec, src, developDir)
			if uErr != nil {
				err = uErr
				return
			}
			if err != nil {
				return
			}
			// delete zip file
			err = os.Remove(src)
			if err != nil {
				return
			}
			log.Printf(logger.Info, "extract %s.zip successfully", name)
		}(item.name, item.dir)
	}
	for i := 0; i < itemsLen; i++ {
		err := <-errCh
		if err != nil {
			log.Println(logger.Error, "failed to extract source code:", err)
			return false
		}
	}
	log.Println(logger.Info, "extract all source code successfully")
	return true
}

func buildSourceCode() bool {
	goRoot, err := config.GoRoot()
	if err != nil {
		log.Println(logger.Error, err)
		return false
	}
	goRoot = filepath.Join(goRoot, "bin")
	// start build
	items := [...]*struct {
		name  string
		dir   string
		build string
	}{
		{name: "golint", dir: "lint-master", build: "golint"},
		{name: "gocyclo", dir: "gocyclo-main", build: "cmd/gocyclo"},
		{name: "gosec", dir: "gosec-master", build: "cmd/gosec"},
		{name: "golangci-lint", dir: "golangci-lint-master", build: "cmd/golangci-lint"},
		{name: "goyacc", dir: "tools-master", build: "cmd/goyacc"},
	}
	itemsLen := len(items)
	resultCh := make(chan bool, itemsLen)
	for _, item := range items {
		go func(name, dir, build string) {
			var err error
			defer func() {
				if err == nil {
					log.Printf(logger.Info, "build development tool %s successfully", name)
					resultCh <- true
					return
				}
				log.Printf(logger.Error, "failed to build development tool %s: %s", name, err)
				resultCh <- false
			}()
			var binName string
			switch runtime.GOOS {
			case "windows":
				binName = name + ".exe"
			case "linux":
				binName = name
			default:
				err = errors.New("unsupported platform: " + runtime.GOOS)
				return
			}
			buildPath := filepath.Join(developDir, dir, build)
			// go build -v -i -ldflags "-s -w" -o lint.exe
			args := []string{"build", "-v", "-i", "-ldflags", "-s -w", "-o", binName}
			cmd := exec.Command("go", args...) // #nosec
			cmd.Dir = buildPath
			writer := logger.WrapLogger(logger.Info, "develop", logger.Common)
			cmd.Stdout = writer
			cmd.Stderr = writer
			err = cmd.Run()
			if err != nil {
				return
			}
			// move binary file to GOROOT
			var mvErr error
			ec := func(_ context.Context, typ uint8, err error, _ *filemgr.SrcDstStat) uint8 {
				if typ == filemgr.ErrCtrlSameFile {
					return filemgr.ErrCtrlOpReplace
				}
				mvErr = err
				return filemgr.ErrCtrlOpCancel
			}
			err = filemgr.Move(ec, goRoot, filepath.Join(buildPath, binName))
			if mvErr != nil {
				err = mvErr
				return
			}
			if err != nil {
				return
			}
			// delete source code directory
			err = os.RemoveAll(filepath.Join(developDir, dir))
		}(item.name, item.dir, item.build)
	}
	for i := 0; i < itemsLen; i++ {
		ok := <-resultCh
		if !ok {
			return false
		}
	}
	log.Println(logger.Info, "build all development tools successfully")
	return true
}
