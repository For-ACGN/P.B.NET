package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"project/internal/logger"
	"project/internal/module/filemgr"
	"project/internal/system"

	"project/script/internal/config"
	"project/script/internal/exec"
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
		setupNetwork,
		downloadSourceCode,
		extractSourceCode,
		downloadModule,
		buildSourceCode,
	} {
		if !step() {
			log.Println(logger.Fatal, "install development tools failed")
			return
		}
	}
	log.Println(logger.Info, "install development tools successfully")
}

func setupNetwork() bool {
	log.Println(logger.Info, "setup network")
	if !config.SetProxy(cfg.Develop.ProxyURL) {
		return false
	}
	if cfg.Develop.Insecure {
		config.SkipTLSVerify()
	}
	return true
}

func downloadSourceCode() bool {
	log.Println(logger.Info, "download source code about development tools")
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
	errCh := make(chan error, len(items))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for _, item := range items {
		go func(name, url string) {
			var err error
			defer func() { errCh <- err }()
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				return
			}
			resp, err := http.DefaultClient.Do(req)
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
	for i := 0; i < len(items); i++ {
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
	log.Println(logger.Info, "extract source code about development tools")
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
	errCh := make(chan error, len(items))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
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
			err = filemgr.UnZipWithContext(ctx, ec, src, developDir)
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
	for i := 0; i < len(items); i++ {
		err := <-errCh
		if err != nil {
			log.Println(logger.Error, "failed to extract source code:", err)
			return false
		}
	}
	log.Println(logger.Info, "extract all source code successfully")
	return true
}

func downloadModule() bool {
	log.Println(logger.Info, "download module if it is not exist")
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
	resultCh := make(chan bool, len(items))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for _, item := range items {
		go func(name, dir string) {
			writer := logger.WrapLogger(logger.Info, "develop", logger.Common)
			cmd := exec.CommandContext(ctx, "go", "get", "-d", "./...") // #nosec
			cmd.Dir = filepath.Join(developDir, dir)
			cmd.Stdout = writer
			cmd.Stderr = writer
			code, err := exec.RunCommand(cmd)
			if code == 0 {
				log.Printf(logger.Info, "download module for %s successfully", name)
				resultCh <- true
				return
			}
			log.Printf(logger.Error, "failed to download module for %s", name)
			if err != nil {
				log.Println(logger.Error, err)
			}
			resultCh <- false
		}(item.name, item.dir)
	}
	for i := 0; i < len(items); i++ {
		ok := <-resultCh
		if !ok {
			return false
		}
	}
	log.Println(logger.Info, "download module successfully")
	return true
}

func buildSourceCode() bool {
	log.Println(logger.Info, "build development tools")
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
	resultCh := make(chan bool, len(items))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
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
			default:
				binName = name
			}
			buildPath := filepath.Join(developDir, dir, build)
			writer := logger.WrapLogger(logger.Info, "develop", logger.Common)
			// go build -v -trimpath -ldflags "-s -w" -o lint.exe
			args := []string{"build", "-v", "-trimpath", "-ldflags", "-s -w", "-o", binName}
			cmd := exec.CommandContext(ctx, "go", args...) // #nosec
			cmd.Dir = buildPath
			cmd.Stdout = writer
			cmd.Stderr = writer
			code, err := exec.RunCommand(cmd)
			if code != 0 {
				if err == nil {
					err = fmt.Errorf("process exit with unexpected code: %d", code)
				}
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
			err = filemgr.MoveWithContext(ctx, ec, goRoot, filepath.Join(buildPath, binName))
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
	for i := 0; i < len(items); i++ {
		ok := <-resultCh
		if !ok {
			return false
		}
	}
	log.Println(logger.Info, "build all development tools successfully")
	return true
}
