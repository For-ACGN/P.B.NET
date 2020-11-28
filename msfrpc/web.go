package msfrpc

import (
	"compress/flate"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/netutil"

	"project/internal/crypto/aes"
	"project/internal/guid"
	"project/internal/httptool"
	"project/internal/logger"
	"project/internal/nettool"
	"project/internal/option"
	"project/internal/patch/json"
	"project/internal/random"
	"project/internal/security"
	"project/internal/virtualconn"
	"project/internal/xpanic"
	"project/internal/xreflect"
	"project/internal/xsync"
)

const (
	defaultAdminUsername    = "admin"
	defaultAdminDisplayName = "Admin"
	defaultServerTimeout    = 2 * time.Minute
	defaultServerMaxConns   = 1000
	minRequestBodySize      = 4 * 1024 * 1024  // 4MB
	minRequestLargeBodySize = 64 * 1024 * 1024 // 64MB
	sessionName             = "Session"
)

// admin user > managers > users > guests
const (
	userGroupAdmins   = 4
	userGroupManagers = 3
	userGroupUsers    = 2
	userGroupGuests   = 1

	UserGroupAdmins   = "admins"   // only admin user
	UserGroupManagers = "managers" // can unlock IOObject force.
	UserGroupUsers    = "users"    // common usage
	UserGroupGuests   = "guests"   // read only
)

var (
	userGroupStr = map[string]int{
		UserGroupManagers: userGroupManagers,
		UserGroupUsers:    userGroupUsers,
		UserGroupGuests:   userGroupGuests,
	}

	userGroupInt = map[int]string{
		userGroupAdmins:   UserGroupAdmins,
		userGroupManagers: UserGroupManagers,
		userGroupUsers:    UserGroupUsers,
		userGroupGuests:   UserGroupGuests,
	}
)

// WebUser contains user information.
type WebUser struct {
	Password    string `toml:"password"`     // "bcrypt"
	UserGroup   string `toml:"user_group"`   // "managers", "users", "guests"
	DisplayName string `toml:"display_name"` // name displayed on UI.
}

// WebOptions contains options about web server.
type WebOptions struct {
	// AdminUsername is the administrator username,
	// if it is empty, use the default admin username.
	AdminUsername string `toml:"admin_username"`

	// AdminPassword is the administrator hashed password(bcrypt),
	// if it is empty, program will generate a random value.
	AdminPassword string `toml:"admin_password"`

	// AdminDisplayName is the administrator name will be show.
	AdminDisplayName string `toml:"admin_display_name"`

	// DisableTLS is used to disable http server use TLS.
	DisableTLS bool `toml:"disable_tls"`

	// MaxConns is the web server maximum connections.
	MaxConns int `toml:"max_conns"`

	// Timeout is the web server timeout.
	Timeout time.Duration `toml:"timeout"`

	// MaxBodySize is the incoming request maximum body size.
	MaxBodySize int64 `toml:"max_body_size"`

	// MaxLargeBodySize is the incoming large request maximum
	// body size, like upload a file, or some else.
	MaxLargeBodySize int64 `toml:"max_large_body_size"`

	// HFS is used to use custom file system.
	HFS http.FileSystem `toml:"-" msgpack:"-"`

	// APIOnly is used to disable Web UI.
	APIOnly bool `toml:"api_only"`

	// Server contains options about http server.
	Server option.HTTPServer `toml:"server" testsuite:"-"`

	// Users contains common users, key is the username.
	Users map[string]*WebUser `toml:"-" testsuite:"-"`
}

// Web is provide a web UI and API server.
type Web struct {
	logger     logger.Logger
	disableTLS bool
	maxConns   int

	srv *http.Server
	api *webAPI
	ui  *webUI

	// listener addresses
	addresses    map[*net.Addr]struct{}
	addressesRWM sync.RWMutex
}

// NewWeb is used to create a web server, password is the common user password.
func NewWeb(msfrpc *MSFRPC, opts *WebOptions) (*Web, error) {
	server, err := opts.Server.Apply()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create web")
	}
	router := mux.NewRouter()
	webAPI, err := newWebAPI(msfrpc, opts, router)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create web api")
	}
	msfrpc.logger.Print(logger.Info, "init", "initialize web api successfully")
	var webUI *webUI
	if !opts.APIOnly {
		webUI, err = newWebUI(opts.HFS, router)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to create web ui")
		}
		msfrpc.logger.Print(logger.Info, "init", "initialize web ui successfully")
	}
	// set http server
	server.Handler = router
	timeout := opts.Timeout
	if timeout < 1 {
		timeout = defaultServerTimeout
	}
	server.ReadTimeout = timeout
	server.WriteTimeout = timeout
	server.ConnState = func(conn net.Conn, state http.ConnState) {
		switch state {
		case http.StateNew:
			webAPI.counter.Add(1)
		case http.StateHijacked, http.StateClosed:
			webAPI.counter.Done()
		}
	}
	server.ErrorLog = logger.Wrap(logger.Warning, "msfrpc-web", msfrpc.logger)
	// set web server
	web := Web{
		logger:     msfrpc.logger,
		disableTLS: opts.DisableTLS,
		maxConns:   opts.MaxConns,
		srv:        server,
		api:        webAPI,
		ui:         webUI,
		addresses:  make(map[*net.Addr]struct{}, 1),
	}
	if web.maxConns < 64 {
		web.maxConns = defaultServerMaxConns
	}
	return &web, nil
}

// MonitorCallbacks is used to return callbacks for Monitor.
func (web *Web) MonitorCallbacks() *MonitorCallbacks {
	return &MonitorCallbacks{
		OnToken:      web.api.onToken,
		OnJob:        web.api.onJob,
		OnSession:    web.api.onSession,
		OnHost:       web.api.onHost,
		OnCredential: web.api.onCredential,
		OnLoot:       web.api.onLoot,
		OnEvent:      web.api.onEvent,
	}
}

// IOEventHandlers is used to return io event handler for IOManager.
func (web *Web) IOEventHandlers() *IOEventHandlers {
	return &IOEventHandlers{
		OnConsoleRead:         web.api.onConsoleRead,
		OnConsoleCleaned:      web.api.onConsoleCleaned,
		OnConsoleClosed:       web.api.onConsoleClosed,
		OnConsoleLocked:       web.api.onConsoleLocked,
		OnConsoleUnlocked:     web.api.onConsoleUnlocked,
		OnShellRead:           web.api.onShellRead,
		OnShellCleaned:        web.api.onShellCleaned,
		OnShellClosed:         web.api.onShellClosed,
		OnShellLocked:         web.api.onShellLocked,
		OnShellUnlocked:       web.api.onShellUnlocked,
		OnMeterpreterRead:     web.api.onMeterpreterRead,
		OnMeterpreterCleaned:  web.api.onMeterpreterCleaned,
		OnMeterpreterClosed:   web.api.onMeterpreterClosed,
		OnMeterpreterLocked:   web.api.onMeterpreterLocked,
		OnMeterpreterUnlocked: web.api.onMeterpreterUnlocked,
	}
}

func (web *Web) logf(lv logger.Level, format string, log ...interface{}) {
	web.logger.Printf(lv, "msfrpc-web", format, log...)
}

func (web *Web) log(lv logger.Level, log ...interface{}) {
	web.logger.Println(lv, "msfrpc-web", log...)
}

func (web *Web) addListenerAddress(addr *net.Addr) {
	web.addressesRWM.Lock()
	defer web.addressesRWM.Unlock()
	web.addresses[addr] = struct{}{}
}

func (web *Web) deleteListenerAddress(addr *net.Addr) {
	web.addressesRWM.Lock()
	defer web.addressesRWM.Unlock()
	delete(web.addresses, addr)
}

// ListenAndServe is used to listen a listener and serve.
func (web *Web) ListenAndServe(network, address string) error {
	err := nettool.IsTCPNetwork(network)
	if err != nil {
		return err
	}
	listener, err := net.Listen(network, address)
	if err != nil {
		return errors.WithStack(err)
	}
	return web.Serve(listener)
}

// Serve accepts incoming connections on the listener.
func (web *Web) Serve(listener net.Listener) (err error) {
	web.api.counter.Add(1)
	defer web.api.counter.Done()

	defer func() {
		if r := recover(); r != nil {
			err = xpanic.Error(r, "Web.Serve")
			web.log(logger.Fatal, err)
		}
	}()

	listener = netutil.LimitListener(listener, web.maxConns)
	defer func() { _ = listener.Close() }()

	address := listener.Addr()
	network := address.Network()
	web.addListenerAddress(&address)
	defer web.deleteListenerAddress(&address)
	web.logf(logger.Info, "serve over listener (%s %s)", network, address)
	defer web.logf(logger.Info, "listener closed (%s %s)", network, address)

	switch listener.(type) {
	case *virtualconn.Listener:
		err = web.srv.Serve(listener)
	default:
		if web.disableTLS {
			err = web.srv.Serve(listener)
		} else {
			err = web.srv.ServeTLS(listener, "", "")
		}
	}
	if nettool.IsNetClosingError(err) || err == http.ErrServerClosed {
		return nil
	}
	return err
}

// Addresses is used to get listener addresses.
func (web *Web) Addresses() []net.Addr {
	web.addressesRWM.RLock()
	defer web.addressesRWM.RUnlock()
	addresses := make([]net.Addr, 0, len(web.addresses))
	for address := range web.addresses {
		addresses = append(addresses, *address)
	}
	return addresses
}

// Close is used to close web server.
func (web *Web) Close() error {
	err := web.srv.Close()
	web.api.Close()
	if err != nil && !nettool.IsNetClosingError(err) {
		return err
	}
	return nil
}

// webUI is used to contains favicon and index data. user can reload it.
type webUI struct {
	hfs     http.FileSystem
	favicon []byte
	index   []byte
	rwm     sync.RWMutex
}

func newWebUI(hfs http.FileSystem, router *mux.Router) (*webUI, error) {
	ui := webUI{hfs: hfs}
	err := ui.Reload()
	if err != nil {
		return nil, err
	}
	router.HandleFunc("/favicon.ico", ui.handleFavicon)
	// set index handler
	for _, name := range [...]string{
		"", "index.html", "index.htm", "index",
	} {
		router.HandleFunc("/"+name, ui.handleIndex)
	}
	// set resource server
	for _, path := range [...]string{
		"css", "js", "img", "fonts",
	} {
		router.NewRoute().PathPrefix("/" + path + "/").Handler(http.FileServer(hfs))
	}
	return &ui, nil
}

func (ui *webUI) Reload() error {
	const maxResourceFileSize = 512 * 1024
	// load favicon.ico and index.html
	res := make(map[string][]byte, 2)
	for _, name := range [...]string{
		"favicon.ico", "index.html",
	} {
		file, err := ui.hfs.Open(name)
		if err != nil {
			return errors.Errorf("failed to open %s: %s", name, err)
		}
		data, err := security.ReadAll(file, maxResourceFileSize)
		if err != nil {
			return errors.Errorf("failed to read %s: %s", name, err)
		}
		res[name] = data
	}
	ui.rwm.Lock()
	defer ui.rwm.Unlock()
	ui.favicon = res["favicon.ico"]
	ui.index = res["index.html"]
	return nil
}

func (ui *webUI) getFavicon() []byte {
	ui.rwm.RLock()
	defer ui.rwm.RUnlock()
	return ui.favicon
}

func (ui *webUI) getIndex() []byte {
	ui.rwm.RLock()
	defer ui.rwm.RUnlock()
	return ui.index
}

func (ui *webUI) handleFavicon(w http.ResponseWriter, _ *http.Request) {
	data := ui.getFavicon()
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (ui *webUI) handleIndex(w http.ResponseWriter, _ *http.Request) {
	data := ui.getIndex()
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

type webUser struct {
	username    *security.String
	password    *security.String
	userGroup   int
	displayName *security.String

	// if user change password, other
	// session will lose efficacy at once
	secret *security.String
}

// webAPI contain the actual handler, login and handle event.
type webAPI struct {
	ctx    *Client
	msfrpc *MSFRPC

	disableTLS          bool
	maxReqBodySize      int64
	maxLargeReqBodySize int64

	encoderPool sync.Pool             // json encoder
	guid        *guid.Generator       // guid generator
	cookieStore *sessions.CookieStore // about cookie
	wsUpgrader  *websocket.Upgrader   // notice event

	// common users, web api can add temporary user.
	users    map[string]*webUser
	usersRWM sync.RWMutex

	// all user websocket connections, key = username
	wsConnGroups    map[string]*wsConnGroup
	wsConnGroupsRWM sync.RWMutex

	inShutdown int32

	// Web use it for prevent cycle reference
	counter xsync.Counter
}

func newWebAPI(msfrpc *MSFRPC, opts *WebOptions, router *mux.Router) (*webAPI, error) {
	api := webAPI{
		ctx:                 msfrpc.client,
		msfrpc:              msfrpc,
		disableTLS:          opts.DisableTLS,
		maxReqBodySize:      opts.MaxBodySize,
		maxLargeReqBodySize: opts.MaxLargeBodySize,
		guid:                guid.New(128, nil),
	}
	err := api.loadUserInfo(opts)
	if err != nil {
		return nil, err
	}
	msfrpc.logger.Print(logger.Info, "init", "load user information successfully")
	// set max body size
	if api.maxReqBodySize < minRequestBodySize {
		api.maxReqBodySize = minRequestBodySize
	}
	if api.maxLargeReqBodySize < minRequestLargeBodySize {
		api.maxLargeReqBodySize = minRequestLargeBodySize
	}
	// json & guid generator
	api.encoderPool.New = func() interface{} {
		return json.NewEncoder(64)
	}
	// set cookie store
	rand := random.NewRand()
	hashKey := rand.Bytes(sha256.Size)
	blockKey := random.Bytes(aes.Key256Bit)
	api.cookieStore = sessions.NewCookieStore(hashKey, blockKey)
	// set websocket upgrader
	api.wsUpgrader = &websocket.Upgrader{
		HandshakeTimeout: time.Minute,
		ReadBufferSize:   4096,
		WriteBufferSize:  4096,
	}
	api.wsConnGroups = make(map[string]*wsConnGroup, 1)
	api.setHandlers(router)
	return &api, nil
}

// when run test with -race, check bcrypt hash
// will spend too much time, so we need skip it.
var testSkipVerifyAdminBcryptHash bool

func (api *webAPI) loadUserInfo(opts *WebOptions) error {
	api.users = make(map[string]*webUser, len(opts.Users)+1) // admin
	// set administrator username
	adminUsername := opts.AdminUsername
	if adminUsername == "" {
		adminUsername = defaultAdminUsername
		const log = "admin username is not set, use the default username:"
		api.log(logger.Warning, log, defaultAdminUsername)
	}
	// set administrator password
	adminPassword := opts.AdminPassword
	if adminPassword == "" { // generate a random password
		password := random.NewRand().String(16)
		defer security.CoverString(password)
		const log = "admin password is not set, use the random password:"
		api.log(logger.Warning, log, password)
		// generate bcrypt hash
		passwordBytes := []byte(password)
		defer security.CoverBytes(passwordBytes)
		hash, err := bcrypt.GenerateFromPassword(passwordBytes, 12)
		if err != nil {
			return errors.Wrap(err, "failed to generate random admin password")
		}
		adminPassword = string(hash)
	} else if !testSkipVerifyAdminBcryptHash {
		// validate bcrypt hash
		err := bcrypt.CompareHashAndPassword([]byte(adminPassword), []byte("123456"))
		if err != nil && err != bcrypt.ErrMismatchedHashAndPassword {
			return errors.New("invalid bcrypt hash about password")
		}
	}
	// set administrator display name
	adminDisplayName := opts.AdminDisplayName
	if adminDisplayName == "" {
		adminDisplayName = defaultAdminDisplayName
		const log = "admin display name is not set, use the default display name:"
		api.log(logger.Warning, log, defaultAdminDisplayName)
	}
	api.users[adminUsername] = &webUser{
		username:    security.NewString(adminUsername),
		password:    security.NewString(adminPassword),
		userGroup:   userGroupAdmins,
		displayName: security.NewString(adminDisplayName),
		secret:      security.NewString(api.guid.Get().Hex()),
	}
	// set common user
	for username, userInfo := range opts.Users {
		// skip check user password bcrypt hash,
		// if it is invalid, admin can change it.

		// check user group
		userGroup, ok := userGroupStr[userInfo.UserGroup]
		if !ok {
			const format = "user: \"%s\" set invalid user group: \"%s\""
			return errors.Errorf(format, username, userInfo.UserGroup)
		}
		// check display name
		if userInfo.DisplayName == "" {
			const format = "user: \"%s\" set empty display name"
			return errors.Errorf(format, username)
		}
		api.users[username] = &webUser{
			username:    security.NewString(username),
			password:    security.NewString(userInfo.Password),
			userGroup:   userGroup,
			displayName: security.NewString(userInfo.DisplayName),
			secret:      security.NewString(api.guid.Get().Hex()),
		}
	}
	return nil
}

func (api *webAPI) setHandlers(router *mux.Router) {
	for path, handler := range map[string]http.HandlerFunc{
		"/api/login":     api.handleLogin,
		"/api/is_online": api.handleIsOnline,
		"/api/websocket": api.handleWebsocket,
		"/api/logoff":    api.handleLogoff,

		"/api/auth/logout":         api.handleAuthenticationLogout,
		"/api/auth/token/list":     api.handleAuthenticationTokenList,
		"/api/auth/token/generate": api.handleAuthenticationTokenGenerate,
		"/api/auth/token/add":      api.handleAuthenticationTokenAdd,
		"/api/auth/token/remove":   api.handleAuthenticationTokenRemove,

		"/api/core/module/status":   api.handleCoreModuleStatus,
		"/api/core/module/add_path": api.handleCoreAddModulePath,
		"/api/core/module/reload":   api.handleCoreReloadModules,
		"/api/core/thread/list":     api.handleCoreThreadList,
		"/api/core/thread/kill":     api.handleCoreThreadKill,
		"/api/core/global/set":      api.handleCoreSetGlobal,
		"/api/core/global/unset":    api.handleCoreUnsetGlobal,
		"/api/core/global/get":      api.handleCoreGetGlobal,
		"/api/core/save":            api.handleCoreSave,
		"/api/core/version":         api.handleCoreVersion,

		"/api/db/status":            api.handleDatabaseStatus,
		"/api/db/host/report":       api.handleDatabaseReportHost,
		"/api/db/host/list":         api.handleDatabaseHosts,
		"/api/db/host/get":          api.handleDatabaseGetHost,
		"/api/db/host/delete":       api.handleDatabaseDeleteHost,
		"/api/db/service/report":    api.handleDatabaseReportService,
		"/api/db/service/list":      api.handleDatabaseServices,
		"/api/db/service/get":       api.handleDatabaseGetService,
		"/api/db/service/delete":    api.handleDatabaseDeleteService,
		"/api/db/client/report":     api.handleDatabaseReportClient,
		"/api/db/client/list":       api.handleDatabaseClients,
		"/api/db/client/get":        api.handleDatabaseGetClient,
		"/api/db/client/delete":     api.handleDatabaseDeleteClient,
		"/api/db/cred/list":         api.handleDatabaseCredentials,
		"/api/db/cred/create":       api.handleDatabaseCreateCredential,
		"/api/db/cred/delete":       api.handleDatabaseDeleteCredentials,
		"/api/db/loot/report":       api.handleDatabaseReportLoot,
		"/api/db/loot/list":         api.handleDatabaseLoots,
		"/api/db/workspace/list":    api.handleDatabaseWorkspaces,
		"/api/db/workspace/get":     api.handleDatabaseGetWorkspace,
		"/api/db/workspace/add":     api.handleDatabaseAddWorkspace,
		"/api/db/workspace/delete":  api.handleDatabaseDeleteWorkspace,
		"/api/db/workspace/set":     api.handleDatabaseSetWorkspace,
		"/api/db/workspace/current": api.handleDatabaseCurrentWorkspace,
		"/api/db/events":            api.handleDatabaseEvents,
		"/api/db/import_data":       api.handleDatabaseImportData,

		"/api/console/list":           api.handleConsoleList,
		"/api/console/create":         api.handleConsoleCreate,
		"/api/console/destroy":        api.handleConsoleDestroy,
		"/api/console/read":           api.handleConsoleRead,
		"/api/console/write":          api.handleConsoleWrite,
		"/api/console/session_detach": api.handleConsoleSessionDetach,
		"/api/console/session_kill":   api.handleConsoleSessionKill,

		"/api/plugin/load":   api.handlePluginLoad,
		"/api/plugin/unload": api.handlePluginUnload,
		"/api/plugin/loaded": api.handlePluginLoaded,

		"/api/module/exploits":                   api.handleModuleExploits,
		"/api/module/auxiliary":                  api.handleModuleAuxiliary,
		"/api/module/post":                       api.handleModulePost,
		"/api/module/payloads":                   api.handleModulePayloads,
		"/api/module/encoders":                   api.handleModuleEncoders,
		"/api/module/nops":                       api.handleModuleNops,
		"/api/module/evasion":                    api.handleModuleEvasion,
		"/api/module/info":                       api.handleModuleInfo,
		"/api/module/options":                    api.handleModuleOptions,
		"/api/module/payloads/compatible":        api.handleModuleCompatiblePayloads,
		"/api/module/payloads/target_compatible": api.handleModuleTargetCompatiblePayloads,
		"/api/module/post/session_compatible":    api.handleModuleCompatibleSessions,
		"/api/module/evasion/compatible":         api.handleModuleCompatibleEvasionPayloads,
		"/api/module/evasion/target_compatible":  api.handleModuleTargetCompatibleEvasionPayloads,
		"/api/module/formats/encode":             api.handleModuleEncodeFormats,
		"/api/module/formats/executable":         api.handleModuleExecutableFormats,
		"/api/module/formats/transform":          api.handleModuleTransformFormats,
		"/api/module/formats/encryption":         api.handleModuleEncryptionFormats,
		"/api/module/platforms":                  api.handleModulePlatforms,
		"/api/module/architectures":              api.handleModuleArchitectures,
		"/api/module/encode":                     api.handleModuleEncode,
		"/api/module/generate_payload":           api.handleModuleGeneratePayload,
		"/api/module/execute":                    api.handleModuleExecute,
		"/api/module/check":                      api.handleModuleCheck,
		"/api/module/running_status":             api.handleModuleRunningStatus,

		"/api/job/list": api.handleJobList,
		"/api/job/info": api.handleJobInfo,
		"/api/job/stop": api.handleJobStop,

		"/api/session/list":                       api.handleSessionList,
		"/api/session/stop":                       api.handleSessionStop,
		"/api/session/shell/read":                 api.handleSessionShellRead,
		"/api/session/shell/write":                api.handleSessionShellWrite,
		"/api/session/upgrade":                    api.handleSessionUpgrade,
		"/api/session/meterpreter/read":           api.handleSessionMeterpreterRead,
		"/api/session/meterpreter/write":          api.handleSessionMeterpreterWrite,
		"/api/session/meterpreter/session_detach": api.handleSessionMeterpreterSessionDetach,
		"/api/session/meterpreter/session_kill":   api.handleSessionMeterpreterSessionKill,
		"/api/session/meterpreter/run_single":     api.handleSessionMeterpreterRunSingle,
		"/api/session/compatible_modules":         api.handleSessionCompatibleModules,
	} {
		router.HandleFunc(path, handler)
	}
}

func (api *webAPI) shuttingDown() bool {
	return atomic.LoadInt32(&api.inShutdown) != 0
}

func (api *webAPI) logf(lv logger.Level, format string, log ...interface{}) {
	if api.shuttingDown() {
		return
	}
	api.msfrpc.logger.Printf(lv, "msfrpc-web api", format, log...)
}

func (api *webAPI) log(lv logger.Level, log ...interface{}) {
	if api.shuttingDown() {
		return
	}
	api.msfrpc.logger.Println(lv, "msfrpc-web api", log...)
}

func (api *webAPI) logfWithReq(lv logger.Level, r *http.Request, format string, log ...interface{}) {
	if api.shuttingDown() {
		return
	}
	buf := httptool.PrintRequest(r)
	format += "\n%s"
	log = append(log, buf)
	api.msfrpc.logger.Printf(lv, "msfrpc-web api", format, log...)
}

func (api *webAPI) logWithReq(lv logger.Level, r *http.Request, log ...interface{}) {
	if api.shuttingDown() {
		return
	}
	buf := httptool.PrintRequest(r)
	log = append(log, "\n", buf)
	api.msfrpc.logger.Println(lv, "msfrpc-web api", log...)
}

func (api *webAPI) Close() {
	api.counter.Wait()
	api.guid.Close()
	api.msfrpc = nil
}

func (api *webAPI) readRequest(r *http.Request, req interface{}) error {
	defer func() { _, _ = io.Copy(ioutil.Discard, r.Body) }()
	err := json.NewDecoder(security.LimitReader(r.Body, api.maxReqBodySize)).Decode(req)

	// TODO start! , copy body
	if err == nil {
		return nil
	}
	if err != io.EOF {
		name := xreflect.GetStructureName(req)
		buf := httptool.PrintRequest(r)
		const format = "failed to read request about %s\nerror: %s\n%s"
		api.logf(logger.Error, format, name, err, buf)
	}
	return err
}

// readLargeRequest is used to request like upload file.
func (api *webAPI) readLargeRequest(r *http.Request, req interface{}) error {
	defer func() { _, _ = io.Copy(ioutil.Discard, r.Body) }()
	err := json.NewDecoder(security.LimitReader(r.Body, api.maxLargeReqBodySize)).Decode(req)
	// copy body
	if err == nil {
		return nil
	}
	if err != io.EOF {
		name := xreflect.GetStructureName(req)
		buf := httptool.PrintRequest(r)
		const format = "failed to read large request about %s\nerror: %s\n%s"
		api.logf(logger.Error, format, name, err, buf)
	}
	return err
}

func (api *webAPI) writeResponse(w http.ResponseWriter, resp interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	encoder := api.encoderPool.Get().(*json.Encoder)
	defer api.encoderPool.Put(encoder)
	data, err := encoder.Encode(resp)
	if err != nil {
		panic(err)
	}
	_, _ = w.Write(data)
}

func (api *webAPI) writeError(w http.ResponseWriter, err error) {
	var str string
	if err != nil {
		str = err.Error()
	}
	api.writeErrorString(w, str)
}

func (api *webAPI) writeErrorString(w http.ResponseWriter, str string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	e := struct {
		Error string `json:"error"`
	}{
		Error: str,
	}
	encoder := api.encoderPool.Get().(*json.Encoder)
	defer api.encoderPool.Put(encoder)
	data, err := encoder.Encode(&e)
	if err != nil {
		panic(err)
	}
	_, _ = w.Write(data)
}

func (api *webAPI) handlePanic(w http.ResponseWriter, _ *http.Request, e interface{}) {
	w.WriteHeader(http.StatusInternalServerError)

	// if is super user return the panic
	_, _ = xpanic.Print(e, "web").WriteTo(w)

	csrf.Protect(nil, nil)
}

// ------------------------------------about web authentication------------------------------------

func (api *webAPI) getUser(username string) *webUser {
	api.usersRWM.RLock()
	defer api.usersRWM.RUnlock()
	return api.users[username]
}

func (api *webAPI) getUserSession(w http.ResponseWriter, r *http.Request) *sessions.Session {
	session, err := api.cookieStore.Get(r, sessionName)
	if err != nil {
		api.log(logger.Debug, "failed to get session:", err)
		api.writeErrorString(w, "invalid session")
		return nil
	}
	// check user is exist
	username := session.Values["username"].(string)
	user := api.getUser(username)
	if user == nil {
		err := fmt.Errorf("user \"%s\" is not exist", username)
		api.log(logger.Debug, err)
		api.writeError(w, err)
		return nil
	}
	// check user password is changed
	secret := session.Values["secret"].(string)
	s := user.secret.Get()
	defer user.secret.Put(s)
	if secret != s {
		err := fmt.Errorf("user \"%s\" changed password", username)
		api.log(logger.Debug, err)
		api.writeError(w, err)
	}
	return session
}

func (api *webAPI) handleLogin(w http.ResponseWriter, r *http.Request) {
	const errInvalidUser = "username or password is incorrect"
	req := struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Insecure bool   `json:"insecure"` // for develop web ui
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	defer security.CoverString(req.Password)
	// get password hash
	user := api.getUser(req.Username)
	if user == nil {
		api.writeErrorString(w, errInvalidUser)
		return
	}
	hash := user.password.GetBytes()
	defer user.password.PutBytes(hash)
	pwd := []byte(req.Password)
	defer security.CoverBytes(pwd)
	// compare password
	err = bcrypt.CompareHashAndPassword(hash, pwd)
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			api.writeErrorString(w, errInvalidUser)
		} else {
			api.writeError(w, err)
		}
		return
	}
	// set session cookie
	session := sessions.NewSession(api.cookieStore, sessionName)
	// set cookie options
	opts := session.Options
	opts.Path = "/"
	opts.Domain = ""
	opts.MaxAge = 0
	if !req.Insecure {
		opts.Secure = true
	}
	opts.HttpOnly = true
	// set cookie value
	username := user.username.Get()
	user.username.Put(username)
	displayName := user.displayName.Get()
	user.displayName.Put(displayName)
	secret := user.secret.Get()
	user.secret.Put(secret)
	session.Values["username"] = username
	session.Values["user_group"] = user.userGroup
	session.Values["display_name"] = displayName
	session.Values["secret"] = secret
	// token is used to prevent users from logging in at the same time
	session.Values["token"] = api.guid.Get().Hex()
	// write session and response
	err = session.Save(r, w)
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Username    string `json:"username"`
		UserGroup   string `json:"user_group"`
		DisplayName string `json:"display_name"`
	}{
		Username:    username,
		UserGroup:   userGroupInt[user.userGroup],
		DisplayName: displayName,
	}
	api.writeResponse(w, &resp)
	api.logf(logger.Debug, "user: \"%s\" is login", username)
}

// handleIsOnline is used to check current user is online.
func (api *webAPI) handleIsOnline(w http.ResponseWriter, r *http.Request) {
	session := api.getUserSession(w, r)
	if session == nil {
		return
	}
	username := session.Values["username"].(string)

	api.wsConnGroupsRWM.RLock()
	defer api.wsConnGroupsRWM.RUnlock()
	_, ok := api.wsConnGroups[username]

	resp := struct {
		Online bool `json:"online"`
	}{
		Online: ok,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleWebsocket(w http.ResponseWriter, r *http.Request) {
	session := api.getUserSession(w, r)
	if session == nil {
		return
	}
	// upgrade to websocket connection, server can push message to client
	conn, err := api.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		api.log(logger.Debug, "failed to upgrade connection", err)
		return
	}
	username := session.Values["username"].(string)
	token := session.Values["token"].(string)
	wsConn := api.newWSConn(username, token, conn)
	if wsConn == nil {
		_ = conn.WriteMessage(websocket.TextMessage, []byte("already login"))
	} else {
		wsConn.Write([]byte("success"))
	}
}

// handleLogoff is used to logoff user force.
func (api *webAPI) handleLogoff(w http.ResponseWriter, r *http.Request) {
	session := api.getUserSession(w, r)
	if session == nil {
		return
	}
	// force close websocket connection

	// check is closed
}

// ----------------------------------------about websocket-----------------------------------------

// a user maybe with multi connections(but the same token).
type wsConnGroup struct {
	ctx      *webAPI
	username string
	token    string

	conns      map[*wsConn]struct{}
	inShutdown int32
	rwm        sync.RWMutex

	// for close all connection
	context context.Context
	cancel  context.CancelFunc
}

func (api *webAPI) getWSConnGroup(username, token string) *wsConnGroup {
	api.wsConnGroupsRWM.Lock()
	defer api.wsConnGroupsRWM.Unlock()
	group, ok := api.wsConnGroups[username]
	if ok {
		if group.token != token {
			return nil
		}
	} else {
		group = &wsConnGroup{
			ctx:      api,
			username: username,
			token:    token,
			conns:    make(map[*wsConn]struct{}, 1),
		}
		group.context, group.cancel = context.WithCancel(context.Background())
		api.wsConnGroups[username] = group
	}
	return group
}

func (group *wsConnGroup) shuttingDown() bool {
	return atomic.LoadInt32(&group.inShutdown) != 0
}

func (group *wsConnGroup) trackConn(conn *wsConn, add bool) bool {
	group.rwm.Lock()
	defer group.rwm.Unlock()
	if add {
		if group.shuttingDown() {
			return false
		}
		group.conns[conn] = struct{}{}
	} else {
		delete(group.conns, conn)
		// delete conn group
		if len(group.conns) == 0 {
			group.close()
			group.ctx.wsConnGroupsRWM.Lock()
			defer group.ctx.wsConnGroupsRWM.Unlock()
			delete(group.ctx.wsConnGroups, group.username)
		}
	}
	return true
}

func (group *wsConnGroup) getConns() []*wsConn {
	group.rwm.RLock()
	defer group.rwm.RUnlock()
	conns := make([]*wsConn, 0, len(group.conns))
	for conn := range group.conns {
		conns = append(conns, conn)
	}
	return conns
}

func (group *wsConnGroup) Write(b []byte) {
	for _, conn := range group.getConns() {
		conn.Write(b)
	}
}

func (group *wsConnGroup) Close() error {
	group.close()
	var err error
	group.rwm.Lock()
	defer group.rwm.Unlock()
	// close all connections
	for conn := range group.conns {
		e := conn.Close()
		if e != nil && err == nil {
			err = e
		}
		delete(group.conns, conn)
	}
	return err
}

func (group *wsConnGroup) close() {
	atomic.StoreInt32(&group.inShutdown, 1)
	group.cancel()
}

type wsConn struct {
	ctx   *webAPI
	group *wsConnGroup

	conn    *websocket.Conn
	writeCh chan []byte

	context context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

func (api *webAPI) newWSConn(username, token string, conn *websocket.Conn) *wsConn {
	group := api.getWSConnGroup(username, token)
	if group == nil {
		return nil
	}
	// enable compress
	conn.EnableWriteCompression(true)
	_ = conn.SetCompressionLevel(flate.BestCompression)
	wsConn := wsConn{
		ctx:     api,
		group:   group,
		conn:    conn,
		writeCh: make(chan []byte, 64),
	}
	wsConn.context, wsConn.cancel = context.WithCancel(group.context)
	wsConn.wg.Add(1)
	go wsConn.readLoop()
	wsConn.wg.Add(1)
	go wsConn.writeLoop()
	return &wsConn
}

func (wsc *wsConn) readLoop() {
	defer wsc.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			b := xpanic.Print(r, "wsConn.readLoop")
			wsc.ctx.log(logger.Fatal, b)
		}
	}()

	wsc.ctx.counter.Add(1)
	defer wsc.ctx.counter.Done()

	defer func() {
		err := wsc.conn.Close()
		if err != nil && !nettool.IsNetClosingError(err) {
			wsc.ctx.log(logger.Error, "failed to close websocket connection:", err)
		}
	}()

	if !wsc.group.trackConn(wsc, true) {
		return
	}
	defer wsc.group.trackConn(wsc, false)

	for {
		_, r, err := wsc.conn.NextReader()
		if err != nil {
			return
		}
		_, _ = io.Copy(ioutil.Discard, r)
	}
}

func (wsc *wsConn) writeLoop() {
	defer wsc.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			b := xpanic.Print(r, "wsConn.writeLoop")
			wsc.ctx.log(logger.Fatal, b)
		}
	}()

	wsc.ctx.counter.Add(1)
	defer wsc.ctx.counter.Done()

	var (
		data []byte
		err  error
	)
	for {
		select {
		case data = <-wsc.writeCh:
			err = wsc.conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				return
			}
		case <-wsc.context.Done():
			return
		}
	}
}

func (wsc *wsConn) Write(b []byte) {
	select {
	case wsc.writeCh <- b:
	case <-wsc.context.Done():
	}
}

func (wsc *wsConn) Close() error {
	err := wsc.conn.Close()
	wsc.cancel()
	wsc.wg.Wait()
	return err
}

// ----------------------------------------monitor callbacks---------------------------------------

func (api *webAPI) onToken(token string, add bool) {
	fmt.Println(token, add)
}

func (api *webAPI) onJob(id, name string, active bool) {
	fmt.Println(id, name, active)
}

func (api *webAPI) onSession(id uint64, info *SessionInfo, opened bool) {
	fmt.Println(id, spew.Sdump(info), opened)
}

func (api *webAPI) onHost(workspace string, host *DBHost, add bool) {
	fmt.Println(workspace, spew.Sdump(host), add)
}

func (api *webAPI) onCredential(workspace string, cred *DBCred, add bool) {
	fmt.Println(workspace, spew.Sdump(cred), add)
}

func (api *webAPI) onLoot(workspace string, loot *DBLoot) {
	fmt.Println(workspace, spew.Sdump(loot))
}

func (api *webAPI) onEvent(event string) {
	fmt.Println("event:", event)
}

// ----------------------------------------IO event handlers---------------------------------------

func (api *webAPI) onConsoleRead(id string) {
	fmt.Println("console id:", id)
}

func (api *webAPI) onConsoleCleaned(id string) {
	fmt.Println("console id:", id)
}

func (api *webAPI) onConsoleClosed(id string) {
	fmt.Println("console id:", id)
}

func (api *webAPI) onConsoleLocked(id, token string) {
	fmt.Println("console id:", id, "token:", token)
}

func (api *webAPI) onConsoleUnlocked(id, token string) {
	fmt.Println("console id:", id, "token:", token)
}

func (api *webAPI) onShellRead(id uint64) {
	fmt.Println("console id:", id)
}

func (api *webAPI) onShellCleaned(id uint64) {
	fmt.Println("console id:", id)
}

func (api *webAPI) onShellClosed(id uint64) {
	fmt.Println("console id:", id)
}

func (api *webAPI) onShellLocked(id uint64, token string) {
	fmt.Println("console id:", id, "token:", token)
}

func (api *webAPI) onShellUnlocked(id uint64, token string) {
	fmt.Println("console id:", id, "token:", token)
}

func (api *webAPI) onMeterpreterRead(id uint64) {
	fmt.Println("console id:", id)
}

func (api *webAPI) onMeterpreterCleaned(id uint64) {
	fmt.Println("console id:", id)
}

func (api *webAPI) onMeterpreterClosed(id uint64) {
	fmt.Println("console id:", id)
}

func (api *webAPI) onMeterpreterLocked(id uint64, token string) {
	fmt.Println("console id:", id, "token:", token)
}

func (api *webAPI) onMeterpreterUnlocked(id uint64, token string) {
	fmt.Println("console id:", id, "token:", token)
}

// ---------------------------------------Metasploit RPC API---------------------------------------

// --------------------------------------about authentication--------------------------------------

func (api *webAPI) handleAuthenticationLogout(w http.ResponseWriter, r *http.Request) {
	req := AuthLogoutRequest{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	err = api.ctx.AuthLogout(req.Token)
	api.writeError(w, err)
}

func (api *webAPI) handleAuthenticationTokenList(w http.ResponseWriter, r *http.Request) {
	tokens, err := api.ctx.AuthTokenList(r.Context())
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Tokens []string `json:"tokens"`
	}{
		Tokens: tokens,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleAuthenticationTokenGenerate(w http.ResponseWriter, r *http.Request) {
	token, err := api.ctx.AuthTokenGenerate(r.Context())
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Token string `json:"token"`
	}{
		Token: token,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleAuthenticationTokenAdd(w http.ResponseWriter, r *http.Request) {
	req := AuthTokenAddRequest{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	err = api.ctx.AuthTokenAdd(r.Context(), req.Token)
	api.writeError(w, err)
}

func (api *webAPI) handleAuthenticationTokenRemove(w http.ResponseWriter, r *http.Request) {
	req := AuthTokenRemoveRequest{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	err = api.ctx.AuthTokenRemove(r.Context(), req.Token)
	api.writeError(w, err)
}

// -------------------------------------------about core-------------------------------------------

func (api *webAPI) handleCoreModuleStatus(w http.ResponseWriter, r *http.Request) {
	status, err := api.ctx.CoreModuleStats(r.Context())
	if err != nil {
		api.writeError(w, err)
		return
	}
	api.writeResponse(w, status)
}

func (api *webAPI) handleCoreAddModulePath(w http.ResponseWriter, r *http.Request) {
	req := CoreAddModulePathRequest{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	status, err := api.ctx.CoreAddModulePath(r.Context(), req.Path)
	if err != nil {
		api.writeError(w, err)
		return
	}
	api.writeResponse(w, status)
}

func (api *webAPI) handleCoreReloadModules(w http.ResponseWriter, r *http.Request) {
	status, err := api.ctx.CoreReloadModules(r.Context())
	if err != nil {
		api.writeError(w, err)
		return
	}
	api.writeResponse(w, status)
}

func (api *webAPI) handleCoreThreadList(w http.ResponseWriter, r *http.Request) {
	list, err := api.ctx.CoreThreadList(r.Context())
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Threads map[uint64]*CoreThreadInfo `json:"threads"`
	}{
		Threads: list,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleCoreThreadKill(w http.ResponseWriter, r *http.Request) {
	req := CoreThreadKillRequest{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	err = api.ctx.CoreThreadKill(r.Context(), req.ID)
	api.writeError(w, err)
}

func (api *webAPI) handleCoreSetGlobal(w http.ResponseWriter, r *http.Request) {
	req := CoreSetGRequest{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	err = api.ctx.CoreSetG(r.Context(), req.Name, req.Value)
	api.writeError(w, err)
}

func (api *webAPI) handleCoreUnsetGlobal(w http.ResponseWriter, r *http.Request) {
	req := CoreUnsetGRequest{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	err = api.ctx.CoreUnsetG(r.Context(), req.Name)
	api.writeError(w, err)
}

func (api *webAPI) handleCoreGetGlobal(w http.ResponseWriter, r *http.Request) {
	req := CoreGetGRequest{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	value, err := api.ctx.CoreGetG(r.Context(), req.Name)
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Value string `json:"value"`
	}{
		Value: value,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleCoreSave(w http.ResponseWriter, r *http.Request) {
	err := api.ctx.CoreSave(r.Context())
	api.writeError(w, err)
}

func (api *webAPI) handleCoreVersion(w http.ResponseWriter, r *http.Request) {
	version, err := api.ctx.CoreVersion(r.Context())
	if err != nil {
		api.writeError(w, err)
		return
	}
	api.writeResponse(w, version)
}

// -----------------------------------------about database-----------------------------------------

func (api *webAPI) handleDatabaseStatus(w http.ResponseWriter, r *http.Request) {
	status, err := api.ctx.DBStatus(r.Context())
	if err != nil {
		api.writeError(w, err)
		return
	}
	api.writeResponse(w, status)
}

func (api *webAPI) handleDatabaseReportHost(w http.ResponseWriter, r *http.Request) {
	req := DBReportHost{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	err = api.ctx.DBReportHost(r.Context(), &req)
	api.writeError(w, err)
}

func (api *webAPI) handleDatabaseHosts(w http.ResponseWriter, r *http.Request) {
	req := struct {
		Workspace string `json:"workspace"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	hosts, err := api.ctx.DBHosts(r.Context(), req.Workspace)
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Hosts []*DBHost `json:"hosts"`
	}{
		Hosts: hosts,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleDatabaseGetHost(w http.ResponseWriter, r *http.Request) {
	req := DBGetHostOptions{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	host, err := api.ctx.DBGetHost(r.Context(), &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	api.writeResponse(w, &host)
}

func (api *webAPI) handleDatabaseDeleteHost(w http.ResponseWriter, r *http.Request) {
	req := DBDelHostOptions{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	_, err = api.ctx.DBDelHost(r.Context(), &req)
	api.writeError(w, err)
}

func (api *webAPI) handleDatabaseReportService(w http.ResponseWriter, r *http.Request) {
	req := DBReportService{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	err = api.ctx.DBReportService(r.Context(), &req)
	api.writeError(w, err)
}

func (api *webAPI) handleDatabaseServices(w http.ResponseWriter, r *http.Request) {
	req := DBServicesOptions{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	services, err := api.ctx.DBServices(r.Context(), &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Services []*DBService `json:"services"`
	}{
		Services: services,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleDatabaseGetService(w http.ResponseWriter, r *http.Request) {
	req := DBGetServiceOptions{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	services, err := api.ctx.DBGetService(r.Context(), &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Services []*DBService `json:"services"`
	}{
		Services: services,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleDatabaseDeleteService(w http.ResponseWriter, r *http.Request) {
	req := DBDelServiceOptions{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	_, err = api.ctx.DBDelService(r.Context(), &req)
	api.writeError(w, err)
}

func (api *webAPI) handleDatabaseReportClient(w http.ResponseWriter, r *http.Request) {
	req := DBReportClient{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	err = api.ctx.DBReportClient(r.Context(), &req)
	api.writeError(w, err)
}

func (api *webAPI) handleDatabaseClients(w http.ResponseWriter, r *http.Request) {
	req := DBClientsOptions{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	clients, err := api.ctx.DBClients(r.Context(), &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Clients []*DBClient `json:"clients"`
	}{
		Clients: clients,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleDatabaseGetClient(w http.ResponseWriter, r *http.Request) {
	req := DBGetClientOptions{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	client, err := api.ctx.DBGetClient(r.Context(), &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	api.writeResponse(w, client)
}

func (api *webAPI) handleDatabaseDeleteClient(w http.ResponseWriter, r *http.Request) {
	req := DBDelClientOptions{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	_, err = api.ctx.DBDelClient(r.Context(), &req)
	api.writeError(w, err)
}

func (api *webAPI) handleDatabaseCredentials(w http.ResponseWriter, r *http.Request) {
	req := struct {
		Workspace string `json:"workspace"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	creds, err := api.ctx.DBCreds(r.Context(), req.Workspace)
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Creds []*DBCred `json:"credentials"`
	}{
		Creds: creds,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleDatabaseCreateCredential(w http.ResponseWriter, r *http.Request) {
	req := DBCreateCredentialOptions{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	_, err = api.ctx.DBCreateCredential(r.Context(), &req)
	api.writeError(w, err)
}

func (api *webAPI) handleDatabaseDeleteCredentials(w http.ResponseWriter, r *http.Request) {
	req := struct {
		Workspace string `json:"workspace"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	_, err = api.ctx.DBDelCreds(r.Context(), req.Workspace)
	api.writeError(w, err)
}

func (api *webAPI) handleDatabaseReportLoot(w http.ResponseWriter, r *http.Request) {
	req := DBReportLoot{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	err = api.ctx.DBReportLoot(r.Context(), &req)
	api.writeError(w, err)
}

func (api *webAPI) handleDatabaseLoots(w http.ResponseWriter, r *http.Request) {
	req := DBLootsOptions{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	loots, err := api.ctx.DBLoots(r.Context(), &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	api.writeResponse(w, loots)
}

func (api *webAPI) handleDatabaseWorkspaces(w http.ResponseWriter, r *http.Request) {
	workspaces, err := api.ctx.DBWorkspaces(r.Context())
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Workspaces []*DBWorkspace `json:"workspaces"`
	}{
		Workspaces: workspaces,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleDatabaseGetWorkspace(w http.ResponseWriter, r *http.Request) {
	req := struct {
		Name string `json:"name"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	workspace, err := api.ctx.DBGetWorkspace(r.Context(), req.Name)
	if err != nil {
		api.writeError(w, err)
		return
	}
	api.writeResponse(w, workspace)
}

func (api *webAPI) handleDatabaseAddWorkspace(w http.ResponseWriter, r *http.Request) {
	req := struct {
		Name string `json:"name"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	err = api.ctx.DBAddWorkspace(r.Context(), req.Name)
	api.writeError(w, err)
}

func (api *webAPI) handleDatabaseDeleteWorkspace(w http.ResponseWriter, r *http.Request) {
	req := struct {
		Name string `json:"name"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	err = api.ctx.DBDelWorkspace(r.Context(), req.Name)
	api.writeError(w, err)
}

func (api *webAPI) handleDatabaseSetWorkspace(w http.ResponseWriter, r *http.Request) {
	req := struct {
		Name string `json:"name"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	err = api.ctx.DBSetWorkspace(r.Context(), req.Name)
	api.writeError(w, err)
}

func (api *webAPI) handleDatabaseCurrentWorkspace(w http.ResponseWriter, r *http.Request) {
	result, err := api.ctx.DBCurrentWorkspace(r.Context())
	if err != nil {
		api.writeError(w, err)
		return
	}
	api.writeResponse(w, result)
}

func (api *webAPI) handleDatabaseEvents(w http.ResponseWriter, r *http.Request) {
	req := DBEventOptions{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	events, err := api.ctx.DBEvent(r.Context(), &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Events []*DBEvent `json:"events"`
	}{
		Events: events,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleDatabaseImportData(w http.ResponseWriter, r *http.Request) {
	req := DBImportDataOptions{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	err = api.ctx.DBImportData(r.Context(), &req)
	api.writeError(w, err)
}

// ------------------------------------------about console-----------------------------------------

func (api *webAPI) handleConsoleList(w http.ResponseWriter, r *http.Request) {
	consoles, err := api.ctx.ConsoleList(r.Context())
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Console []*ConsoleInfo `json:"consoles"`
	}{
		Console: consoles,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleConsoleCreate(w http.ResponseWriter, r *http.Request) {
	req := struct {
		Workspace  string        `json:"workspace"`
		IOInterval time.Duration `json:"io_interval"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	if req.IOInterval < 1 {
		req.IOInterval = minReadInterval
	}

	console, err := api.ctx.NewConsole(r.Context(), req.Workspace, req.IOInterval)
	if err != nil {
		api.writeError(w, err)
		return
	}
	_ = console.Close()
}

func (api *webAPI) handleConsoleDestroy(w http.ResponseWriter, r *http.Request) {
	req := struct {
		ID string `json:"id"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	// first check is in web handler
	err = api.ctx.ConsoleDestroy(r.Context(), req.ID)
	api.writeError(w, err)
}

func (api *webAPI) handleConsoleRead(w http.ResponseWriter, r *http.Request) {
	req := struct {
		ID string `json:"id"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	// first check is in web handler
	err = api.ctx.ConsoleSessionDetach(r.Context(), req.ID)
	api.writeError(w, err)
}

func (api *webAPI) handleConsoleWrite(w http.ResponseWriter, r *http.Request) {
	req := struct {
		ID string `json:"id"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	// first check is in web handler
	err = api.ctx.ConsoleSessionKill(r.Context(), req.ID)
	api.writeError(w, err)
}

func (api *webAPI) handleConsoleSessionDetach(w http.ResponseWriter, r *http.Request) {
	req := struct {
		ID string `json:"id"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	// first check is in web handler
	err = api.ctx.ConsoleSessionDetach(r.Context(), req.ID)
	api.writeError(w, err)
}

func (api *webAPI) handleConsoleSessionKill(w http.ResponseWriter, r *http.Request) {
	req := struct {
		ID string `json:"id"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	// first check is in web handler
	err = api.ctx.ConsoleSessionKill(r.Context(), req.ID)
	api.writeError(w, err)
}

// ------------------------------------------about plugin------------------------------------------

func (api *webAPI) handlePluginLoad(w http.ResponseWriter, r *http.Request) {
	req := struct {
		Name    string            `json:"name"`
		Options map[string]string `json:"options"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	err = api.ctx.PluginLoad(r.Context(), req.Name, req.Options)
	api.writeError(w, err)
}

func (api *webAPI) handlePluginUnload(w http.ResponseWriter, r *http.Request) {
	req := struct {
		Name string `json:"name"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	err = api.ctx.PluginUnload(r.Context(), req.Name)
	api.writeError(w, err)
}

func (api *webAPI) handlePluginLoaded(w http.ResponseWriter, r *http.Request) {
	plugins, err := api.ctx.PluginLoaded(r.Context())
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Plugins []string `json:"plugins"`
	}{
		Plugins: plugins,
	}
	api.writeResponse(w, &resp)
}

// ------------------------------------------about module------------------------------------------

func (api *webAPI) handleModuleExploits(w http.ResponseWriter, r *http.Request) {
	modules, err := api.ctx.ModuleExploits(r.Context())
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Modules []string `json:"modules"`
	}{
		Modules: modules,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleModuleAuxiliary(w http.ResponseWriter, r *http.Request) {
	modules, err := api.ctx.ModuleAuxiliary(r.Context())
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Modules []string `json:"modules"`
	}{
		Modules: modules,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleModulePost(w http.ResponseWriter, r *http.Request) {
	modules, err := api.ctx.ModulePost(r.Context())
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Modules []string `json:"modules"`
	}{
		Modules: modules,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleModulePayloads(w http.ResponseWriter, r *http.Request) {
	modules, err := api.ctx.ModulePayloads(r.Context())
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Modules []string `json:"modules"`
	}{
		Modules: modules,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleModuleEncoders(w http.ResponseWriter, r *http.Request) {
	modules, err := api.ctx.ModuleEncoders(r.Context())
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Modules []string `json:"modules"`
	}{
		Modules: modules,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleModuleNops(w http.ResponseWriter, r *http.Request) {
	modules, err := api.ctx.ModuleNops(r.Context())
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Modules []string `json:"modules"`
	}{
		Modules: modules,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleModuleEvasion(w http.ResponseWriter, r *http.Request) {
	modules, err := api.ctx.ModuleEvasion(r.Context())
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Modules []string `json:"modules"`
	}{
		Modules: modules,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleModuleInfo(w http.ResponseWriter, r *http.Request) {
	req := struct {
		Type string `json:"type"`
		Name string `json:"name"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	info, err := api.ctx.ModuleInfo(r.Context(), req.Type, req.Name)
	if err != nil {
		api.writeError(w, err)
		return
	}
	api.writeResponse(w, info)
}

func (api *webAPI) handleModuleOptions(w http.ResponseWriter, r *http.Request) {
	req := struct {
		Type string `json:"type"`
		Name string `json:"name"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	opts, err := api.ctx.ModuleOptions(r.Context(), req.Type, req.Name)
	if err != nil {
		api.writeError(w, err)
		return
	}
	api.writeResponse(w, opts)
}

func (api *webAPI) handleModuleCompatiblePayloads(w http.ResponseWriter, r *http.Request) {
	req := struct {
		Name string `json:"name"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	payloads, err := api.ctx.ModuleCompatiblePayloads(r.Context(), req.Name)
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Payloads []string `json:"payloads"`
	}{
		Payloads: payloads,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleModuleTargetCompatiblePayloads(w http.ResponseWriter, r *http.Request) {
	req := struct {
		Name   string `json:"name"`
		Target uint64 `json:"target"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	payloads, err := api.ctx.ModuleTargetCompatiblePayloads(r.Context(), req.Name, req.Target)
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Payloads []string `json:"payloads"`
	}{
		Payloads: payloads,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleModuleCompatibleSessions(w http.ResponseWriter, r *http.Request) {
	req := struct {
		Name string `json:"name"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	cSessions, err := api.ctx.ModuleCompatibleSessions(r.Context(), req.Name)
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Sessions []string `json:"sessions"`
	}{
		Sessions: cSessions,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleModuleCompatibleEvasionPayloads(w http.ResponseWriter, r *http.Request) {
	req := struct {
		Name string `json:"name"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	payloads, err := api.ctx.ModuleCompatibleEvasionPayloads(r.Context(), req.Name)
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Payloads []string `json:"payloads"`
	}{
		Payloads: payloads,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleModuleTargetCompatibleEvasionPayloads(w http.ResponseWriter, r *http.Request) {
	req := struct {
		Name   string `json:"name"`
		Target uint64 `json:"target"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	payloads, err := api.ctx.ModuleTargetCompatibleEvasionPayloads(r.Context(), req.Name, req.Target)
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Payloads []string `json:"payloads"`
	}{
		Payloads: payloads,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleModuleEncodeFormats(w http.ResponseWriter, r *http.Request) {
	formats, err := api.ctx.ModuleEncodeFormats(r.Context())
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Formats []string `json:"formats"`
	}{
		Formats: formats,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleModuleExecutableFormats(w http.ResponseWriter, r *http.Request) {
	formats, err := api.ctx.ModuleExecutableFormats(r.Context())
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Formats []string `json:"formats"`
	}{
		Formats: formats,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleModuleTransformFormats(w http.ResponseWriter, r *http.Request) {
	formats, err := api.ctx.ModuleTransformFormats(r.Context())
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Formats []string `json:"formats"`
	}{
		Formats: formats,
	}
	api.writeResponse(w, &resp)
}
func (api *webAPI) handleModuleEncryptionFormats(w http.ResponseWriter, r *http.Request) {
	formats, err := api.ctx.ModuleEncryptionFormats(r.Context())
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Formats []string `json:"formats"`
	}{
		Formats: formats,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleModulePlatforms(w http.ResponseWriter, r *http.Request) {
	platforms, err := api.ctx.ModulePlatforms(r.Context())
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Platforms []string `json:"platforms"`
	}{
		Platforms: platforms,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleModuleArchitectures(w http.ResponseWriter, r *http.Request) {
	architectures, err := api.ctx.ModuleArchitectures(r.Context())
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Architectures []string `json:"architectures"`
	}{
		Architectures: architectures,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleModuleEncode(w http.ResponseWriter, r *http.Request) {
	req := struct {
		Data    string               `json:"data"`
		Encoder string               `json:"encoder"`
		Options *ModuleEncodeOptions `json:"options"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	data, err := api.ctx.ModuleEncode(r.Context(), req.Data, req.Encoder, req.Options)
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Data string `json:"data"`
	}{
		Data: data,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleModuleGeneratePayload(w http.ResponseWriter, r *http.Request) {
	req := struct {
		Name    string                `json:"name"`
		Options *ModuleExecuteOptions `json:"options"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	result, err := api.ctx.ModuleExecute(r.Context(), "payload", req.Name, req.Options)
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Payload string `json:"payload"`
	}{
		Payload: result.Payload,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleModuleExecute(w http.ResponseWriter, r *http.Request) {
	req := struct {
		Type    string                 `json:"type"`
		Name    string                 `json:"name"`
		Options map[string]interface{} `json:"options"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	result, err := api.ctx.ModuleExecute(r.Context(), req.Type, req.Name, req.Options)
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		JobID uint64 `json:"job_id"`
		UUID  string `json:"uuid"`
	}{
		JobID: result.JobID,
		UUID:  result.UUID,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleModuleCheck(w http.ResponseWriter, r *http.Request) {
	req := struct {
		Type    string                 `json:"type"`
		Name    string                 `json:"name"`
		Options map[string]interface{} `json:"options"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	result, err := api.ctx.ModuleCheck(r.Context(), req.Type, req.Name, req.Options)
	if err != nil {
		api.writeError(w, err)
		return
	}
	api.writeResponse(w, result)
}

func (api *webAPI) handleModuleRunningStatus(w http.ResponseWriter, r *http.Request) {
	status, err := api.ctx.ModuleRunningStats(r.Context())
	if err != nil {
		api.writeError(w, err)
		return
	}
	api.writeResponse(w, status)
}

// -------------------------------------------about job--------------------------------------------

func (api *webAPI) handleJobList(w http.ResponseWriter, r *http.Request) {
	jobs, err := api.ctx.JobList(r.Context())
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Jobs map[string]string `json:"jobs"`
	}{
		Jobs: jobs,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleJobInfo(w http.ResponseWriter, r *http.Request) {
	req := struct {
		ID string `json:"id"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	info, err := api.ctx.JobInfo(r.Context(), req.ID)
	if err != nil {
		api.writeError(w, err)
		return
	}
	api.writeResponse(w, info)
}

func (api *webAPI) handleJobStop(w http.ResponseWriter, r *http.Request) {
	req := struct {
		ID string `json:"id"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	err = api.ctx.JobStop(r.Context(), req.ID)
	api.writeError(w, err)
}

// -----------------------------------------about session------------------------------------------

func (api *webAPI) handleSessionList(w http.ResponseWriter, r *http.Request) {
	list, err := api.ctx.SessionList(r.Context())
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Sessions map[uint64]*SessionInfo `json:"sessions"`
	}{
		Sessions: list,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleSessionStop(w http.ResponseWriter, r *http.Request) {
	req := struct {
		ID uint64 `json:"id"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	// first check is in web handler
	err = api.ctx.SessionStop(r.Context(), req.ID)
	api.writeError(w, err)
}

func (api *webAPI) handleSessionShellRead(w http.ResponseWriter, r *http.Request) {
	req := struct {
		ID uint64 `json:"id"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	result, err := api.ctx.SessionShellRead(r.Context(), req.ID)
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Data string `json:"data"`
	}{
		Data: result.Data,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleSessionShellWrite(w http.ResponseWriter, r *http.Request) {
	req := struct {
		ID   uint64 `json:"id"`
		Data string `json:"data"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	// check
	_, err = api.ctx.SessionShellWrite(r.Context(), req.ID, req.Data)
	api.writeError(w, err)
}

func (api *webAPI) handleSessionUpgrade(w http.ResponseWriter, r *http.Request) {
	req := struct {
		ID      uint64                 `json:"id"`
		Host    string                 `json:"host"`
		Port    uint64                 `json:"port"`
		Options map[string]interface{} `json:"options"`
		Wait    time.Duration          `json:"wait"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	result, err := api.ctx.SessionUpgrade(r.Context(),
		req.ID, req.Host, req.Port, req.Options, req.Wait)
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		JobID uint64 `json:"job_id"`
	}{
		JobID: result.JobID,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleSessionMeterpreterRead(w http.ResponseWriter, r *http.Request) {
	req := struct {
		ID uint64 `json:"id"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	data, err := api.ctx.SessionMeterpreterRead(r.Context(), req.ID)
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Data string `json:"data"`
	}{
		Data: data,
	}
	api.writeResponse(w, &resp)
}

func (api *webAPI) handleSessionMeterpreterWrite(w http.ResponseWriter, r *http.Request) {
	req := struct {
		ID   uint64 `json:"id"`
		Data string `json:"data"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	err = api.ctx.SessionMeterpreterWrite(r.Context(), req.ID, req.Data)
	api.writeError(w, err)
}

func (api *webAPI) handleSessionMeterpreterSessionDetach(w http.ResponseWriter, r *http.Request) {
	req := struct {
		ID uint64 `json:"id"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	// check exist
	err = api.ctx.SessionMeterpreterSessionDetach(r.Context(), req.ID)
	api.writeError(w, err)
}

func (api *webAPI) handleSessionMeterpreterSessionKill(w http.ResponseWriter, r *http.Request) {
	req := struct {
		ID uint64 `json:"id"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	// check exist
	err = api.ctx.SessionMeterpreterSessionKill(r.Context(), req.ID)
	api.writeError(w, err)
}

func (api *webAPI) handleSessionMeterpreterRunSingle(w http.ResponseWriter, r *http.Request) {
	req := struct {
		ID      uint64 `json:"id"`
		Command string `json:"command"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	err = api.ctx.SessionMeterpreterRunSingle(r.Context(), req.ID, req.Command)
	api.writeError(w, err)
}

func (api *webAPI) handleSessionCompatibleModules(w http.ResponseWriter, r *http.Request) {
	req := struct {
		ID uint64 `json:"id"`
	}{}
	err := api.readRequest(r, &req)
	if err != nil {
		api.writeError(w, err)
		return
	}
	modules, err := api.ctx.SessionCompatibleModules(r.Context(), req.ID)
	if err != nil {
		api.writeError(w, err)
		return
	}
	resp := struct {
		Modules []string `json:"modules"`
	}{
		Modules: modules,
	}
	api.writeResponse(w, &resp)
}
