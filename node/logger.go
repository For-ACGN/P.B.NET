package node

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/pkg/errors"

	"project/internal/crypto/aes"
	"project/internal/crypto/rand"
	"project/internal/logger"
	"project/internal/messages"
	"project/internal/random"
	"project/internal/security"
	"project/internal/xpanic"
)

// encLog is encrypted log.
// Node use self session key encrypt it, because Node
// need save it to memory and send to Controller, if
// Controller not connect the Node network, these logs
// will save as plain text, it maybe leak some important
// messages, so we need encrypt these log
type encLog struct {
	time   time.Time
	level  logger.Level
	source string
	log    []byte // encrypted
}

// gLogger is a global logger, all modules's log use it.
// it will send log to Controller and write to writer.
type gLogger struct {
	ctx *Node

	level  logger.Level
	writer io.Writer
	queue  chan *encLog
	rand   *random.Rand
	timer  *time.Timer

	// about encrypt log
	cbc *aes.CBC

	m       sync.Mutex
	context context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

func newLogger(ctx *Node, config *Config) (*gLogger, error) {
	cfg := config.Logger
	lv, err := logger.Parse(cfg.Level)
	if err != nil {
		return nil, err
	}
	if cfg.QueueSize < 512 {
		return nil, errors.New("logger queue size must >= 512")
	}
	// generate self-encrypt key
	aesKeyIV := make([]byte, aes.Key256Bit+aes.IVSize)
	_, err = io.ReadFull(rand.Reader, aesKeyIV)
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate aes key and iv")
	}
	cbc, err := aes.NewCBC(aesKeyIV[:aes.Key256Bit], aesKeyIV[aes.Key256Bit:])
	if err != nil {
		panic("logger internal error: " + err.Error())
	}
	lg := &gLogger{
		ctx:    ctx,
		level:  lv,
		writer: cfg.Writer,
		queue:  make(chan *encLog, cfg.QueueSize),
		rand:   random.New(),
		timer:  time.NewTimer(time.Second),
		cbc:    cbc,
	}
	lg.context, lg.cancel = context.WithCancel(context.Background())
	return lg, nil
}

func (lg *gLogger) Printf(lv logger.Level, src, format string, log ...interface{}) {
	if lv < lg.level {
		return
	}
	lg.m.Lock()
	defer lg.m.Unlock()
	if lg.ctx == nil {
		return
	}
	now := lg.ctx.global.Now()
	buf := logger.Prefix(now, lv, src)
	// log with level and src
	logStr := fmt.Sprintf(format, log...)
	buf.WriteString(logStr)
	buf.WriteString("\n")
	lg.writeLog(now, lv, src, logStr, buf)
}

func (lg *gLogger) Print(lv logger.Level, src string, log ...interface{}) {
	if lv < lg.level {
		return
	}
	lg.m.Lock()
	defer lg.m.Unlock()
	if lg.ctx == nil {
		return
	}
	now := lg.ctx.global.Now()
	buf := logger.Prefix(now, lv, src)
	// log with level and src
	logStr := fmt.Sprint(log...)
	buf.WriteString(logStr)
	buf.WriteString("\n")
	lg.writeLog(now, lv, src, logStr, buf)
}

func (lg *gLogger) Println(lv logger.Level, src string, log ...interface{}) {
	if lv < lg.level {
		return
	}
	lg.m.Lock()
	defer lg.m.Unlock()
	if lg.ctx == nil {
		return
	}
	now := lg.ctx.global.Now()
	buf := logger.Prefix(now, lv, src)
	// log with level and src
	logStr := fmt.Sprintln(log...)
	buf.WriteString(logStr)
	lg.writeLog(now, lv, src, logStr[:len(logStr)-1], buf) // delete "\n"
}

// StartSender is used to start log sender.
func (lg *gLogger) StartSender() {
	lg.wg.Add(1)
	go lg.sender()
}

// Close is used to close log sender and set logger.ctx = nil
func (lg *gLogger) Close() {
	lg.cancel()
	lg.wg.Wait()
	lg.timer.Stop()
	lg.m.Lock()
	defer lg.m.Unlock()
	lg.ctx = nil
}

// string log not include time level src
func (lg *gLogger) writeLog(time time.Time, lv logger.Level, src, log string, b *bytes.Buffer) {
	defer func() {
		if r := recover(); r != nil {
			_, _ = xpanic.Print(r, "gLogger.writeLog").WriteTo(lg.writer)
		}
	}()
	// write to the self writer
	_, _ = b.WriteTo(lg.writer)
	// encrypt log and send to the log queue, then wait sender
	// to send it to the Controller, finally you can receive it.
	cipherData, err := lg.cbc.Encrypt([]byte(log))
	if err != nil {
		panic("logger internal error: " + err.Error())
	}
	ec := encLog{
		time:   time,
		level:  lv,
		source: src,
		log:    cipherData,
	}
	select {
	case lg.queue <- &ec:
	default: // prevent block
	}
}

// sender is used to send logger to Controller.
func (lg *gLogger) sender() {
	defer func() {
		if r := recover(); r != nil {
			lg.Print(logger.Fatal, "logger", xpanic.Print(r, "gLogger.sender"))
			// restart sender
			time.Sleep(time.Second)
			go lg.sender()
		} else {
			lg.wg.Done()
		}
	}()
	var encLog *encLog
	for {
		select {
		case encLog = <-lg.queue:
			lg.send(encLog)
		case <-lg.context.Done():
			return
		}
	}
}

// send will try to send log until Node is exit.
func (lg *gLogger) send(log *encLog) {
	for {
		plainData, err := lg.cbc.Decrypt(log.log)
		if err != nil {
			panic("logger internal error: " + err.Error())
		}
		// decrypt encrypted log
		m := messages.Log{
			Time:   log.time,
			Level:  log.level,
			Source: log.source,
			Log:    plainData,
		}
		err = lg.ctx.sender.Send(messages.CMDBNodeLog, m)
		if err == nil {
			security.CoverBytes(plainData)
			break
		}
		// encrypt log again
		log.log, err = lg.cbc.Encrypt(plainData)
		if err != nil {
			panic("logger internal error: " + err.Error())
		}
		security.CoverBytes(plainData)
		// wait some time and retry
		lg.timer.Reset(time.Duration(1+lg.rand.Int(3)) * time.Second)
		select {
		case <-lg.context.Done():
			return
		case <-lg.timer.C:
		}
	}
}
