package msfrpc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"

	"project/internal/logger"
	"project/internal/option"
	"project/internal/patch/msgpack"
	"project/internal/xsync"
)

// ClientOptions contains options about NewClient().
type ClientOptions struct {
	// DisableTLS is used to "http" scheme
	DisableTLS bool `toml:"disable_tls"`

	// TLSVerify is used to enable TLS verify
	TLSVerify bool `toml:"tls_verify"`

	// Handler is the MSFRPCD web url, if it is empty, use "api"
	Handler string `toml:"handler"`

	// Timeout is the request timeout
	Timeout time.Duration `toml:"timeout"`

	// Token is a permanent token, if use it, client not need to login.
	Token string `toml:"token"`

	// Transport contains options about http transport.
	Transport option.HTTPTransport `toml:"transport" testsuite:"-"`
}

// Client is used to connect metasploit-framework RPC service.
// It provide a lot of API for interactive with MSF.
type Client struct {
	username string
	password string
	logger   logger.Logger

	url    string
	client *http.Client

	// sync pool about send request
	encoderPool   sync.Pool
	decoderPool   sync.Pool
	rdDecoderPool sync.Pool

	// for authorization
	token    string
	tokenRWM sync.RWMutex

	// key = console id
	consoles map[string]*Console
	// key = shell session id
	shells map[uint64]*Shell
	// key = meterpreter session id
	meterpreters map[uint64]*Meterpreter
	// client status
	inShutdown int32
	// protect above fields
	rwm sync.RWMutex

	// io object counter
	ioObjCtr xsync.Counter

	ctx    context.Context
	cancel context.CancelFunc
}

type bufEncoder struct {
	buf     *bytes.Buffer
	encoder *msgpack.Encoder
}

type bufDecoder struct {
	buf     *bytes.Buffer
	decoder *msgpack.Decoder
}

type readerDecoder struct {
	reader  *bytes.Reader
	decoder *msgpack.Decoder
}

// NewClient is used to create a new metasploit-framework RPC client.
func NewClient(address, username, password string, lg logger.Logger, opts *ClientOptions) (*Client, error) {
	if opts == nil {
		opts = new(ClientOptions)
	}
	// make http client
	tr, err := opts.Transport.Apply()
	if err != nil {
		return nil, err
	}
	// [note] cover options about max connection
	if opts.Transport.MaxIdleConns < 1 {
		tr.MaxIdleConns = 32
	}
	if opts.Transport.MaxIdleConnsPerHost < 1 {
		tr.MaxIdleConnsPerHost = 32
	}
	if opts.Transport.MaxConnsPerHost < 1 {
		tr.MaxConnsPerHost = 32
	}
	// tls
	if !opts.TLSVerify {
		tr.TLSClientConfig.InsecureSkipVerify = true
	}
	// make http client
	jar, _ := cookiejar.New(nil)
	timeout := opts.Timeout
	if timeout < 1 {
		timeout = 30 * time.Second
	}
	httpClient := http.Client{
		Transport: tr,
		Jar:       jar,
		Timeout:   timeout,
	}
	client := Client{
		username:     username,
		password:     password,
		logger:       lg,
		client:       &httpClient,
		token:        opts.Token,
		consoles:     make(map[string]*Console),
		shells:       make(map[uint64]*Shell),
		meterpreters: make(map[uint64]*Meterpreter),
	}
	// url
	var scheme string
	if opts.DisableTLS {
		scheme = "http"
	} else {
		scheme = "https"
	}
	var handler string
	if opts.Handler == "" {
		handler = "api"
	} else {
		handler = opts.Handler
	}
	client.url = fmt.Sprintf("%s://%s/%s", scheme, address, handler)
	// sync pool
	client.encoderPool.New = func() interface{} {
		buf := bytes.NewBuffer(make([]byte, 0, 64))
		encoder := msgpack.NewEncoder(buf)
		encoder.UseArrayEncodedStructs(true)
		return &bufEncoder{
			buf:     buf,
			encoder: encoder,
		}
	}
	client.decoderPool.New = func() interface{} {
		buf := bytes.NewBuffer(make([]byte, 0, 64))
		decoder := msgpack.NewDecoder(buf)
		return &bufDecoder{
			buf:     buf,
			decoder: decoder,
		}
	}
	client.rdDecoderPool.New = func() interface{} {
		reader := bytes.NewReader(nil)
		decoder := msgpack.NewDecoder(reader)
		return &readerDecoder{
			reader:  reader,
			decoder: decoder,
		}
	}
	client.ctx, client.cancel = context.WithCancel(context.Background())
	return &client, nil
}

// SetToken is used to set token to current client.
func (client *Client) SetToken(token string) {
	client.tokenRWM.Lock()
	defer client.tokenRWM.Unlock()
	client.token = token
}

// GetToken is used to get token from current client.
func (client *Client) GetToken() string {
	client.tokenRWM.RLock()
	defer client.tokenRWM.RUnlock()
	return client.token
}

func (client *Client) send(ctx context.Context, req, resp interface{}) error {
	return client.sendWithReplace(ctx, req, resp, nil)
}

// sendWithReplace is used to replace response to another response like
// CoreThreadList and MSFError if decode failed(return a MSFError).
func (client *Client) sendWithReplace(ctx context.Context, req, resp, replace interface{}) error {
	// pack request
	be := client.encoderPool.Get().(*bufEncoder)
	defer client.encoderPool.Put(be)
	be.buf.Reset()
	err := be.encoder.Encode(req)
	if err != nil {
		return errors.WithStack(err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, client.url, be.buf)
	if err != nil {
		return errors.WithStack(err)
	}
	header := request.Header
	header.Set("Content-Type", "binary/message-pack")
	header.Set("Accept", "binary/message-pack")
	header.Set("Accept-Charset", "utf-8")
	header.Set("Connection", "keep-alive")
	// send request
	response, err := client.client.Do(request)
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		_, _ = io.Copy(ioutil.Discard, response.Body)
		_ = response.Body.Close()
	}()
	// read response body
	switch response.StatusCode {
	case http.StatusOK:
		bd := client.decoderPool.Get().(*bufDecoder)
		defer client.decoderPool.Put(bd)
		bd.buf.Reset()
		_, err = bd.buf.ReadFrom(response.Body)
		if err != nil {
			return errors.WithStack(err)
		}
		if replace == nil {
			return bd.decoder.Decode(resp)
		}
		// first try to decode to response
		rd := client.rdDecoderPool.Get().(*readerDecoder)
		defer client.rdDecoderPool.Put(rd)
		rd.reader.Reset(bd.buf.Bytes())
		err = rd.decoder.Decode(resp)
		if err == nil {
			return nil
		}
		// then try to decode to replace
		return bd.decoder.Decode(replace)
	case http.StatusInternalServerError:
		bd := client.decoderPool.Get().(*bufDecoder)
		defer client.decoderPool.Put(bd)
		bd.buf.Reset()
		_, err = bd.buf.ReadFrom(response.Body)
		if err != nil {
			return errors.WithStack(err)
		}
		msfErr := new(MSFError)
		err = bd.decoder.Decode(msfErr)
		if err != nil {
			return errors.WithStack(err)
		}
		return errors.WithStack(msfErr)
	case http.StatusUnauthorized:
		err = errors.New("invalid token")
	case http.StatusForbidden:
		err = errors.New("this token is not granted access to the resource")
	case http.StatusNotFound:
		err = errors.New("the request was sent to an invalid URL")
	default:
		err = errors.Errorf("unexpected http status code: %d", response.StatusCode)
	}
	return err
}

func (client *Client) shuttingDown() bool {
	return atomic.LoadInt32(&client.inShutdown) != 0
}

func (client *Client) logf(lv logger.Level, format string, log ...interface{}) {
	if client.shuttingDown() {
		return
	}
	client.logger.Printf(lv, "msfrpc-client", format, log...)
}

func (client *Client) log(lv logger.Level, log ...interface{}) {
	if client.shuttingDown() {
		return
	}
	client.logger.Println(lv, "msfrpc-client", log...)
}

func (client *Client) increaseIOObjCount(delta int) {
	client.ioObjCtr.Add(delta)
}

func (client *Client) decreaseIOObjCount(delta int) {
	client.ioObjCtr.Add(-delta)
}

func (client *Client) trackConsole(console *Console, add bool) bool {
	client.rwm.Lock()
	defer client.rwm.Unlock()
	if add {
		if client.shuttingDown() {
			return false
		}
		client.consoles[console.id] = console
	} else {
		delete(client.consoles, console.id)
	}
	return true
}

func (client *Client) trackShell(shell *Shell, add bool) bool {
	client.rwm.Lock()
	defer client.rwm.Unlock()
	if add {
		if client.shuttingDown() {
			return false
		}
		client.shells[shell.id] = shell
	} else {
		delete(client.shells, shell.id)
	}
	return true
}

func (client *Client) trackMeterpreter(mp *Meterpreter, add bool) bool {
	client.rwm.Lock()
	defer client.rwm.Unlock()
	if add {
		if client.shuttingDown() {
			return false
		}
		client.meterpreters[mp.id] = mp
	} else {
		delete(client.meterpreters, mp.id)
	}
	return true
}

// GetConsole is used to get console by id.
func (client *Client) GetConsole(id string) (*Console, error) {
	client.rwm.RLock()
	defer client.rwm.RUnlock()
	if console, ok := client.consoles[id]; ok {
		return console, nil
	}
	return nil, errors.Errorf("console %s is not exist", id)
}

// GetShell is used to get shell by id.
func (client *Client) GetShell(id uint64) (*Shell, error) {
	client.rwm.RLock()
	defer client.rwm.RUnlock()
	if shell, ok := client.shells[id]; ok {
		return shell, nil
	}
	return nil, errors.Errorf("shell session %d is not exist", id)
}

// GetMeterpreter is used to get meterpreter by id.
func (client *Client) GetMeterpreter(id uint64) (*Meterpreter, error) {
	client.rwm.RLock()
	defer client.rwm.RUnlock()
	if meterpreter, ok := client.meterpreters[id]; ok {
		return meterpreter, nil
	}
	return nil, errors.Errorf("meterpreter session %d is not exist", id)
}

// Close is used to logout metasploit RPC and destroy all objects.
func (client *Client) Close() error {
	err := client.AuthLogout(client.GetToken())
	if err != nil {
		return err
	}
	client.close()
	client.wait()
	return nil
}

// Kill is used to logout metasploit RPC when can't connect target.
func (client *Client) Kill() {
	err := client.AuthLogout(client.GetToken())
	if err != nil {
		client.log(logger.Warning, "appear error when kill msfrpc client:", err)
	}
	client.close()
	client.wait()
}

func (client *Client) close() {
	atomic.StoreInt32(&client.inShutdown, 1)
	client.rwm.Lock()
	defer client.rwm.Unlock()
	// close all consoles
	for _, console := range client.consoles {
		_ = console.Close()
	}
	// close all shells
	for _, shell := range client.shells {
		_ = shell.Close()
	}
	// close all meterpreters
	for _, meterpreter := range client.meterpreters {
		_ = meterpreter.Close()
	}
	client.cancel()
}

func (client *Client) wait() {
	client.ioObjCtr.Wait()
	client.client.CloseIdleConnections()
}
