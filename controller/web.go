package controller

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"

	"project/internal/bootstrap"
	"project/internal/crypto/cert"
	"project/internal/crypto/cert/certutil"
	"project/internal/guid"
	"project/internal/logger"
	"project/internal/messages"
	"project/internal/protocol"
	"project/internal/security"
)

// TODO password need use bCrypt

type hRW = http.ResponseWriter
type hR = http.Request
type hP = httprouter.Params

type web struct {
	ctx *CTRL

	listener net.Listener
	server   *http.Server
	indexFS  http.Handler // index file system

	wg sync.WaitGroup
}

func newWeb(ctx *CTRL, config *Config) (*web, error) {
	cfg := config.Web

	// generate certificate
	certFile, err := ioutil.ReadFile(cfg.CertFile)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	keyFile, err := ioutil.ReadFile(cfg.KeyFile)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	caCert, err := certutil.ParseCertificate(certFile)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	caPri, err := certutil.ParsePrivateKey(keyFile)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// TODO set options
	certOpts := cert.Options{
		DNSNames:    []string{"localhost"},
		IPAddresses: []string{"127.0.0.1"},
	}
	pair, err := cert.Generate(caCert, caPri, &certOpts)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	tlsCert, err := pair.TLSCertificate()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// router
	web := web{
		ctx:      ctx,
		listener: listener,
	}
	router := &httprouter.Router{
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      true,
		HandleMethodNotAllowed: true,
		HandleOPTIONS:          true,
		PanicHandler:           web.handlePanic,
	}
	// resource
	router.ServeFiles("/css/*filepath", http.Dir(cfg.Dir+"/css"))
	router.ServeFiles("/js/*filepath", http.Dir(cfg.Dir+"/js"))
	router.ServeFiles("/img/*filepath", http.Dir(cfg.Dir+"/img"))
	web.indexFS = http.FileServer(http.Dir(cfg.Dir))
	handleFavicon := func(w hRW, r *hR, _ hP) {
		web.indexFS.ServeHTTP(w, r)
	}
	router.GET("/favicon.ico", handleFavicon)
	router.GET("/", web.handleIndex)
	router.GET("/login", web.handleLogin)
	router.POST("/load_keys", web.handleLoadSessionKey)

	// debug api
	router.GET("/api/debug/shutdown", web.handleShutdown)

	// API
	router.GET("/api/boot", web.handleGetBoot)
	router.POST("/api/node/trust", web.handleTrustNode)
	router.GET("/api/node/shell", web.handleShell)

	// HTTPS server
	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{tlsCert},
	}
	web.server = &http.Server{
		TLSConfig:         tlsConfig,
		ReadHeaderTimeout: time.Minute,
		Handler:           router,
		ErrorLog:          logger.Wrap(logger.Warning, "web", ctx.logger),
	}
	return &web, nil
}

func (web *web) Deploy() error {
	errChan := make(chan error, 1)
	serve := func() {
		errChan <- web.server.ServeTLS(web.listener, "", "")
		web.wg.Done()
	}
	web.wg.Add(1)
	go serve()
	select {
	case err := <-errChan:
		return errors.WithStack(err)
	case <-time.After(time.Second):
		return nil
	}
}

func (web *web) Address() string {
	return web.listener.Addr().String()
}

func (web *web) Close() {
	_ = web.server.Close()
	web.ctx = nil
}

func (web *web) handlePanic(w hRW, r *hR, e interface{}) {
	w.WriteHeader(http.StatusInternalServerError)

	// _, _ = io.Copy(w, xpanic.Print(e, "web"))
}

func (web *web) handleLogin(w hRW, r *hR, p hP) {
	_, _ = w.Write([]byte("hello"))
}

func (web *web) handleLoadSessionKey(w hRW, r *hR, p hP) {
	web.wg.Add(1)
	defer web.wg.Done()

	// TODO size, check is load session key
	// if isClosed{
	//  return
	// }

	pwd, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}
	err = web.ctx.LoadSessionKey(pwd, pwd)
	security.CoverBytes(pwd)
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	_, _ = w.Write([]byte("ok"))
}

func (web *web) handleIndex(w hRW, r *hR, p hP) {
	web.indexFS.ServeHTTP(w, r)
}

// ------------------------------debug API----------------------------------

func (web *web) handleShutdown(w hRW, r *hR, p hP) {
	_ = r.ParseForm()
	errStr := r.FormValue("err")
	_, _ = w.Write([]byte("ok"))
	if errStr != "" {
		web.ctx.Exit(errors.New(errStr))
	} else {
		web.ctx.Exit(nil)
	}
}

// ---------------------------------API-------------------------------------

func (web *web) handleGetBoot(w hRW, r *hR, p hP) {
	_, _ = w.Write([]byte("hello"))
}

func (web *web) handleTrustNode(w hRW, r *hR, p hP) {
	m := &mTrustNode{}
	err := json.NewDecoder(r.Body).Decode(m)
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	listener := bootstrap.Listener{
		Mode:    m.Mode,
		Network: m.Network,
		Address: m.Address,
	}
	req, err := web.ctx.TrustNode(context.TODO(), &listener)
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	b, err := json.Marshal(req)
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	_, _ = w.Write(b)
}

func (web *web) handleShell(w hRW, r *hR, p hP) {
	_ = r.ParseForm()
	nodeGUID := guid.GUID{}

	nodeGUIDSlice, err := hex.DecodeString(r.FormValue("guid"))
	if err != nil {
		fmt.Println("1", err)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	fmt.Println(nodeGUIDSlice)
	err = nodeGUID.Write(nodeGUIDSlice)
	if err != nil {
		fmt.Println("2", err)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	shell := messages.Shell{
		Command: r.FormValue("cmd"),
	}

	// TODO check nodeGUID
	err = web.ctx.sender.Send(protocol.Node, &nodeGUID, messages.CMDBShell, &shell)
	if err != nil {
		fmt.Println("2", err)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
}
