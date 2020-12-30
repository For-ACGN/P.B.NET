package node

import (
	"bytes"
	"compress/flate"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"project/internal/cert"
	"project/internal/cert/certmgr"
	"project/internal/crypto/aes"
	"project/internal/dns"
	"project/internal/logger"
	"project/internal/option"
	"project/internal/patch/msgpack"
	"project/internal/proxy"
	"project/internal/random"
	"project/internal/security"
	"project/internal/timesync"
	"project/internal/xpanic"
)

// Config contains configuration about Node.
// use extra msgpack tag to hide raw field name.
type Config struct {
	Logger struct {
		Level     string `toml:"level"      msgpack:"a"`
		QueueSize int    `toml:"queue_size" msgpack:"b"`

		// if false, use ioutil.Discard, if true, use os.Stdout,
		// usually enable it for debug.
		Stdout bool      `toml:"stdout"     msgpack:"c"`
		Writer io.Writer `toml:"-"          msgpack:"-"`
	} `toml:"logger" msgpack:"aa"`

	Global struct {
		DNSCacheExpire      time.Duration `toml:"dns_cache_expire"      msgpack:"a"`
		TimeSyncSleepFixed  uint          `toml:"timesync_sleep_fixed"  msgpack:"b"`
		TimeSyncSleepRandom uint          `toml:"timesync_sleep_random" msgpack:"c"`
		TimeSyncInterval    time.Duration `toml:"timesync_interval"     msgpack:"d"`

		// generate from controller
		CertPool          certmgr.NBCertPool          `toml:"-" msgpack:"w"`
		ProxyClients      []*proxy.Client             `toml:"-" msgpack:"x"`
		DNSServers        map[string]*dns.Server      `toml:"-" msgpack:"y"`
		TimeSyncerClients map[string]*timesync.Client `toml:"-" msgpack:"z"`
	} `toml:"global" msgpack:"bb"`

	Client struct {
		Timeout   time.Duration    `toml:"timeout"   msgpack:"a"`
		ProxyTag  string           `toml:"proxy_tag" msgpack:"b"`
		DNSOpts   dns.Options      `toml:"dns"       msgpack:"c"`
		TLSConfig option.TLSConfig `toml:"tls"       msgpack:"d"`
	} `toml:"client" msgpack:"cc"`

	Register struct {
		SleepFixed  uint `toml:"sleep_fixed"  msgpack:"a"` // about register failed
		SleepRandom uint `toml:"sleep_random" msgpack:"b"`
		Skip        bool `toml:"skip"         msgpack:"c"` // wait controller trust it

		// generate configs from controller
		FirstBoot     []byte `toml:"-" msgpack:"w"` // type: *messages.Bootstrap
		FirstKey      []byte `toml:"-" msgpack:"x"` // decrypt the first bootstrap data, AES CBC
		FollowUpBoots []byte `toml:"-" msgpack:"y"` // type: []*messages.Bootstrap
		FollowUpKey   []byte `toml:"-" msgpack:"z"` // decrypt the follow-up bootstraps data, AES CBC
	} `toml:"register" msgpack:"dd"`

	Forwarder struct {
		MaxClientConns int `toml:"max_client_conns" msgpack:"a"`
		MaxCtrlConns   int `toml:"max_ctrl_conns"   msgpack:"b"`
		MaxNodeConns   int `toml:"max_node_conns"   msgpack:"c"`
		MaxBeaconConns int `toml:"max_beacon_conns" msgpack:"d"`
	} `toml:"forwarder" msgpack:"ee"`

	Sender struct {
		Worker        int           `toml:"worker"          msgpack:"a"`
		Timeout       time.Duration `toml:"timeout"         msgpack:"b"`
		QueueSize     int           `toml:"queue_size"      msgpack:"c"`
		MaxBufferSize int           `toml:"max_buffer_size" msgpack:"d"`
	} `toml:"sender" msgpack:"ff"`

	Syncer struct {
		ExpireTime time.Duration `toml:"expire_time" msgpack:"a"`
	} `toml:"syncer" msgpack:"gg"`

	Worker struct {
		Number        int `toml:"number"          msgpack:"a"`
		QueueSize     int `toml:"queue_size"      msgpack:"b"`
		MaxBufferSize int `toml:"max_buffer_size" msgpack:"c"`
	} `toml:"worker" msgpack:"hh"`

	Server struct {
		MaxConns int           `toml:"max_conns" msgpack:"a"` // each listener
		Timeout  time.Duration `toml:"timeout"   msgpack:"b"` // handshake timeout

		// generate from controller
		Listeners    []byte `toml:"-" msgpack:"y"` // type: []*messages.Listener
		ListenersKey []byte `toml:"-" msgpack:"z"` // decrypt Listeners data, AES CBC
	} `toml:"server" msgpack:"ii"`

	Driver struct {
	} `toml:"driver" msgpack:"jj"`

	// generate from controller
	Ctrl struct {
		KexPublicKey []byte `msgpack:"x"` // key exchange curve25519
		PublicKey    []byte `msgpack:"y"` // verify message ed25519
		BroadcastKey []byte `msgpack:"z"` // decrypt broadcast, key + iv
	} `toml:"-" msgpack:"kk"`

	// about service
	Service struct {
		Name        string `toml:"name"         msgpack:"a"`
		DisplayName string `toml:"display_name" msgpack:"b"`
		Description string `toml:"description"  msgpack:"c"`
	} `toml:"service" msgpack:"ll"`

	Test struct {
		SkipSynchronizeTime bool
	} `toml:"-" msgpack:"-"`
}

// TestOptions include test options.
type TestOptions struct {
	Domain  string        `toml:"domain"` // about Node.global.DNSClient.TestServers()
	DNSOpts dns.Options   `toml:"dns"`
	Timeout time.Duration `toml:"timeout"` // Node running timeout
}

// Run is used to create a Node with current configuration and run it to check error.
func (cfg *Config) Run(ctx context.Context, output io.Writer, opts *TestOptions) (err error) {
	defer func() {
		if err != nil {
			_, _ = fmt.Fprintln(output, "\ntests failed:", err)
		} else {
			_, _ = fmt.Fprintln(output, "\ntests passed")
		}
	}()
	cfg.Logger.Level = "debug"
	cfg.Logger.Writer = output
	node, err := New(cfg)
	if err != nil {
		return
	}
	defer node.Exit(nil)

	line := "------------------------------certificates--------------------------------\n"
	_, _ = output.Write([]byte(line))
	cfg.Certificates(output, node)
	line = "------------------------------proxy clients-------------------------------\n"
	_, _ = output.Write([]byte(line))
	cfg.ProxyClients(output, node)
	line = "-------------------------------DNS servers--------------------------------\n"
	_, _ = output.Write([]byte(line))
	_, err = cfg.DNSServers(ctx, output, node, opts)
	if err != nil {
		return
	}
	line = "---------------------------time syncer clients----------------------------\n"
	_, _ = output.Write([]byte(line))
	_, err = cfg.TimeSyncerClients(ctx, output, node)
	if err != nil {
		return
	}
	line = "------------------------------node running--------------------------------\n"
	_, _ = output.Write([]byte(line))
	err = cfg.wait(ctx, node, opts.Timeout)
	return
}

// Certificates is used to print certificates.
func (cfg *Config) Certificates(writer io.Writer, node *Node) string {
	// set output
	var output io.Writer
	buf := new(bytes.Buffer)
	if writer != nil {
		output = io.MultiWriter(writer, buf)
	} else {
		output = buf
	}

	// print certificates
	_, _ = output.Write([]byte("----------------public root ca----------------\n"))
	for i, c := range node.global.CertPool.GetPublicRootCACerts() {
		_, _ = fmt.Fprintf(output, "ID: %d\n%s\n\n", i+1, cert.Print(c))
	}
	_, _ = output.Write([]byte("---------------public client ca---------------\n"))
	for i, c := range node.global.CertPool.GetPublicClientCACerts() {
		_, _ = fmt.Fprintf(output, "ID: %d\n%s\n\n", i+1, cert.Print(c))
	}
	_, _ = output.Write([]byte("--------------public client cert--------------\n"))
	for i, pair := range node.global.CertPool.GetPublicClientPairs() {
		_, _ = fmt.Fprintf(output, "ID: %d\n%s\n\n", i+1, cert.Print(pair.Certificate))
	}
	_, _ = output.Write([]byte("----------------private root ca---------------\n"))
	for i, c := range node.global.CertPool.GetPrivateRootCACerts() {
		_, _ = fmt.Fprintf(output, "ID: %d\n%s\n\n", i+1, cert.Print(c))
	}
	_, _ = output.Write([]byte("---------------private client ca--------------\n"))
	for i, c := range node.global.CertPool.GetPrivateClientCACerts() {
		_, _ = fmt.Fprintf(output, "ID: %d\n%s\n\n", i+1, cert.Print(c))
	}
	_, _ = output.Write([]byte("--------------private client cert-------------\n"))
	for i, pair := range node.global.CertPool.GetPrivateClientPairs() {
		_, _ = fmt.Fprintf(output, "ID: %d\n%s\n\n", i+1, cert.Print(pair.Certificate))
	}
	return buf.String()
}

// ProxyClients is used to print proxy clients.
func (cfg *Config) ProxyClients(writer io.Writer, node *Node) string {
	// set output
	var output io.Writer
	buf := new(bytes.Buffer)
	if writer != nil {
		output = io.MultiWriter(writer, buf)
	} else {
		output = buf
	}

	// print proxy clients
	for tag, client := range node.global.ProxyPool.Clients() {
		if tag == "" || tag == "direct" { // skip builtin proxy client
			continue
		}
		const format = "tag: %-10s mode: %-7s network: %-3s address: %s\n"
		_, _ = fmt.Fprintf(output, format, tag, client.Mode, client.Network, client.Address)
	}
	return buf.String()
}

// DNSServers is used to print and test DNS servers.
// if tests passed, show resolved IP address.
func (cfg *Config) DNSServers(
	ctx context.Context,
	writer io.Writer,
	node *Node,
	opts *TestOptions,
) (string, error) {
	// set output
	var output io.Writer
	buf := new(bytes.Buffer)
	if writer != nil {
		output = io.MultiWriter(writer, buf)
	} else {
		output = buf
	}
	// print DNS servers
	for tag, server := range node.global.DNSClient.Servers() {
		const format = "tag: %-14s skip: %-5t method: %-3s address: %s\n"
		_, _ = fmt.Fprintf(output, format, tag, server.SkipTest, server.Method, server.Address)
	}
	// resolve domain name
	result, err := node.global.DNSClient.TestServers(ctx, opts.Domain, &opts.DNSOpts)
	if err != nil {
		return buf.String(), err
	}
	const format = "\ntest domain: %s\nresolved ip: %s\n"
	_, _ = fmt.Fprintf(output, format, opts.Domain, strings.Join(result, ", "))
	return buf.String(), nil
}

// TimeSyncerClients is used to print and test time syncer clients.
// if tests passed, show the current time.
func (cfg *Config) TimeSyncerClients(
	ctx context.Context,
	writer io.Writer,
	node *Node,
) (string, error) {
	// set output
	var output io.Writer
	buf := new(bytes.Buffer)
	if writer != nil {
		output = io.MultiWriter(writer, buf)
	} else {
		output = buf
	}

	// print time syncer clients
	for tag, client := range node.global.TimeSyncer.Clients() {
		const format = "tag: %-10s skip: %-5t mode: %-4s\n"
		_, _ = fmt.Fprintf(output, format, tag, client.SkipTest, client.Mode)
	}

	// test syncer clients
	err := node.global.TimeSyncer.Test(ctx)
	if err != nil {
		return buf.String(), err
	}
	now := node.global.Now().Local().Format(logger.TimeLayout)
	_, _ = fmt.Fprintf(output, "\ncurrent time: %s\n", now)
	return buf.String(), nil
}

func (cfg *Config) wait(ctx context.Context, node *Node, timeout time.Duration) error {
	errCh := make(chan error, 1)
	go func() {
		var err error
		defer func() {
			if r := recover(); r != nil {
				err = xpanic.Error(r, "Config.wait")
			}
			errCh <- err
		}()
		err = node.Main()
	}()
	if timeout < 1 {
		timeout = 15 * time.Second
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case err := <-errCh:
		close(errCh)
		return err
	case <-timer.C:
		node.Exit(nil)
		return errors.New("node running timeout")
	case <-ctx.Done():
		node.Exit(nil)
		return ctx.Err()
	}
}

// Build is used to build configuration.
func (cfg *Config) Build() ([]byte, []byte, error) {
	data, err := msgpack.Marshal(cfg)
	if err != nil {
		return nil, nil, err
	}
	defer func() { security.CoverBytes(data) }()
	// compress
	buf := bytes.NewBuffer(make([]byte, 0, len(data)/2))
	defer func() { security.CoverBytes(buf.Bytes()) }()
	writer, _ := flate.NewWriter(buf, flate.BestCompression)
	_, err = writer.Write(data)
	if err != nil {
		return nil, nil, err
	}
	err = writer.Close()
	if err != nil {
		return nil, nil, err
	}
	// encrypt
	rand := random.NewRand()
	aesKey := rand.Bytes(aes.Key256Bit)
	aesIV := rand.Bytes(aes.IVSize)
	cipherData, err := aes.CBCEncrypt(buf.Bytes(), aesKey, aesIV)
	if err != nil {
		return nil, nil, err
	}
	return cipherData, append(aesKey, aesIV...), nil
}

// Load is used to load built configuration.
func (cfg *Config) Load(data, key []byte) error {
	if len(key) != aes.Key256Bit+aes.IVSize {
		return errors.New("invalid key size")
	}
	// decrypt
	aesKey := key[:aes.Key256Bit]
	aesIV := key[aes.Key256Bit:]
	plainData, err := aes.CBCDecrypt(data, aesKey, aesIV)
	if err != nil {
		return err
	}
	defer func() { security.CoverBytes(plainData) }()
	// decompress
	reader := flate.NewReader(bytes.NewReader(plainData))
	defer func() { _ = reader.Close() }()
	return msgpack.NewDecoder(reader).Decode(cfg)
}
