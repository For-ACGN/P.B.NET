package controller

import (
	"io"
	"time"

	"github.com/pkg/errors"

	"project/internal/crypto/aes"
	"project/internal/crypto/cert"
	"project/internal/dns"
	"project/internal/messages"
	"project/internal/option"
	"project/internal/patch/msgpack"
	"project/internal/random"
)

// Config include configuration about Controller.
type Config struct {
	Test struct {
		SkipTestClientDNS   bool
		SkipSynchronizeTime bool
	} `toml:"-"`

	Database struct {
		Dialect         string    `toml:"dialect"` // "mysql"
		DSN             string    `toml:"dsn"`
		MaxOpenConns    int       `toml:"max_open_conns"`
		MaxIdleConns    int       `toml:"max_idle_conns"`
		LogFile         string    `toml:"log_file"`
		GORMLogFile     string    `toml:"gorm_log_file"`
		GORMDetailedLog bool      `toml:"gorm_detailed_log"`
		LogWriter       io.Writer `toml:"-"`
	} `toml:"database"`

	Logger struct {
		Level  string    `toml:"level"`
		File   string    `toml:"file"`
		Writer io.Writer `toml:"-"`
	} `toml:"logger"`

	Global struct {
		DNSCacheExpire      time.Duration `toml:"dns_cache_expire"`
		TimeSyncSleepFixed  uint          `toml:"timesync_sleep_fixed"`
		TimeSyncSleepRandom uint          `toml:"timesync_sleep_random"`
		TimeSyncInterval    time.Duration `toml:"timesync_interval"`
	} `toml:"global"`

	Client struct {
		Timeout   time.Duration    `toml:"timeout"`
		ProxyTag  string           `toml:"proxy_tag"`
		DNSOpts   dns.Options      `toml:"dns"`
		TLSConfig option.TLSConfig `toml:"tls"`
	} `toml:"client"`

	Sender struct {
		MaxConns      int           `toml:"max_conns"`
		Worker        int           `toml:"worker"`
		Timeout       time.Duration `toml:"timeout"`
		QueueSize     int           `toml:"queue_size"`
		MaxBufferSize int           `toml:"max_buffer_size"`
	} `toml:"sender"`

	Syncer struct {
		ExpireTime time.Duration `toml:"expire_time"`
	} `toml:"syncer"`

	Worker struct {
		Number        int `toml:"number"`
		QueueSize     int `toml:"queue_size"`
		MaxBufferSize int `toml:"max_buffer_size"`
	} `toml:"worker"`

	Web struct {
		Dir      string       `toml:"dir"`
		CertFile string       `toml:"cert_file"`
		KeyFile  string       `toml:"key_file"`
		CertOpts cert.Options `toml:"cert"`
		Network  string       `toml:"network"`
		Address  string       `toml:"address"`
		Username string       `toml:"username"` // super user
		Password string       `toml:"password"`
	} `toml:"web"`
}

// GenerateRoleConfigAboutTheFirstBootstrap is used to generate the first bootstrap.
func GenerateRoleConfigAboutTheFirstBootstrap(b *messages.Bootstrap) ([]byte, []byte, error) {
	return generateRoleConfigAboutBootstraps(b)
}

// GenerateRoleConfigAboutRestBootstraps is used to generate role rest bootstraps.
func GenerateRoleConfigAboutRestBootstraps(b ...*messages.Bootstrap) ([]byte, []byte, error) {
	if len(b) == 0 {
		return nil, nil, nil
	}
	return generateRoleConfigAboutBootstraps(b)
}

func generateRoleConfigAboutBootstraps(b interface{}) ([]byte, []byte, error) {
	data, _ := msgpack.Marshal(b)
	rand := random.New()
	aesKey := rand.Bytes(aes.Key256Bit)
	aesIV := rand.Bytes(aes.IVSize)
	enc, err := aes.CBCEncrypt(data, aesKey, aesIV)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	return enc, append(aesKey, aesIV...), nil
}

// GenerateNodeConfigAboutListeners is used to generate node listener and encrypt it.
func GenerateNodeConfigAboutListeners(l ...*messages.Listener) ([]byte, []byte, error) {
	if len(l) == 0 {
		return nil, nil, errors.New("no listeners")
	}
	data, _ := msgpack.Marshal(l)
	rand := random.New()
	aesKey := rand.Bytes(aes.Key256Bit)
	aesIV := rand.Bytes(aes.IVSize)
	enc, err := aes.CBCEncrypt(data, aesKey, aesIV)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	return enc, append(aesKey, aesIV...), nil
}
