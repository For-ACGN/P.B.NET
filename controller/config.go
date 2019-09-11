package controller

import (
	"time"
)

type Debug struct {
	SkipTimeSyncer bool
}

type Config struct {
	Debug Debug `toml:"-"`

	// logger
	LogLevel string `toml:"log_level"`

	// global
	BuiltinDir         string        `toml:"builtin_dir"`
	KeyDir             string        `toml:"key_dir"`
	DNSCacheDeadline   time.Duration `toml:"dns_cache_deadline"`
	TimeSyncerInterval time.Duration `toml:"time_syncer_interval"`

	// database
	Dialect         string `toml:"dialect"` // "mysql"
	DSN             string `toml:"dsn"`
	DBLogFile       string `toml:"db_log_file"`
	DBMaxOpenConns  int    `toml:"db_max_open_conns"`
	DBMaxIdleConns  int    `toml:"db_max_idle_conns"`
	GORMLogFile     string `toml:"gorm_log_file"`
	GORMDetailedLog bool   `toml:"gorm_detailed_log"`

	// sender
	MaxBufferSize   int `toml:"max_buffer_size"` // syncer also use it
	SenderNumber    int `toml:"sender_number"`
	SenderQueueSize int `toml:"sender_queue_size"`

	// syncer
	MaxSyncer        int           `toml:"max_syncer"`
	WorkerNumber     int           `toml:"worker_number"`
	WorkerQueueSize  int           `toml:"worker_queue_size"`
	ReserveWorker    int           `toml:"reserve_worker"`
	RetryTimes       int           `toml:"retry_times"`
	RetryInterval    time.Duration `toml:"retry_interval"`
	BroadcastTimeout time.Duration `toml:"broadcast_timeout"`
	ReceiveTimeout   time.Duration `toml:"receive_timeout"`
	DBSyncInterval   time.Duration `toml:"db_sync_interval"`

	// web server
	HTTPSAddress  string `toml:"https_address"`
	HTTPSCertFile string `toml:"https_cert_file"`
	HTTPSKeyFile  string `toml:"https_key_file"`
	HTTPSWebDir   string `toml:"https_web_dir"`
	HTTPSUsername string `toml:"https_username"`
	HTTPSPassword string `toml:"https_password"`
}
