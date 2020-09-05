// Package project generate by script/code/anko/package.go, don't edit it.
package project

import (
	"reflect"

	"github.com/mattn/anko/env"

	"project/internal/cert"
	"project/internal/convert"
	"project/internal/crypto/aes"
	"project/internal/crypto/curve25519"
	"project/internal/crypto/ed25519"
	"project/internal/crypto/hmac"
	"project/internal/crypto/lsb"
	"project/internal/crypto/rand"
	"project/internal/dns"
	"project/internal/guid"
	"project/internal/httptool"
	"project/internal/logger"
	"project/internal/namer"
	"project/internal/nettool"
	"project/internal/option"
	"project/internal/patch/json"
	"project/internal/patch/msgpack"
	"project/internal/patch/toml"
	"project/internal/proxy"
	"project/internal/proxy/direct"
	"project/internal/proxy/http"
	"project/internal/proxy/socks"
	"project/internal/random"
	"project/internal/security"
	"project/internal/system"
	"project/internal/timesync"
	"project/internal/xpanic"
	"project/internal/xreflect"
	"project/internal/xsync"
)

func init() {
	initInternalCert()
	initInternalConvert()
	initInternalCryptoAES()
	initInternalCryptoCurve25519()
	initInternalCryptoED25519()
	initInternalCryptoHMAC()
	initInternalCryptoLSB()
	initInternalCryptoRand()
	initInternalDNS()
	initInternalGUID()
	initInternalHTTPTool()
	initInternalLogger()
	initInternalNamer()
	initInternalNetTool()
	initInternalOption()
	initInternalPatchJSON()
	initInternalPatchMsgpack()
	initInternalPatchToml()
	initInternalProxy()
	initInternalProxyDirect()
	initInternalProxyHTTP()
	initInternalProxySocks()
	initInternalRandom()
	initInternalSecurity()
	initInternalSystem()
	initInternalTimeSync()
	initInternalXPanic()
	initInternalXReflect()
	initInternalXSync()
}

func initInternalCert() {
	env.Packages["project/internal/cert"] = map[string]reflect.Value{
		// define constants

		// define variables
		"ErrInvalidPEMBlock": reflect.ValueOf(cert.ErrInvalidPEMBlock),

		// define functions
		"Generate":               reflect.ValueOf(cert.Generate),
		"GenerateCA":             reflect.ValueOf(cert.GenerateCA),
		"Match":                  reflect.ValueOf(cert.Match),
		"NewPool":                reflect.ValueOf(cert.NewPool),
		"NewPoolWithSystemCerts": reflect.ValueOf(cert.NewPoolWithSystemCerts),
		"ParseCertificate":       reflect.ValueOf(cert.ParseCertificate),
		"ParseCertificates":      reflect.ValueOf(cert.ParseCertificates),
		"ParsePrivateKey":        reflect.ValueOf(cert.ParsePrivateKey),
		"ParsePrivateKeyBytes":   reflect.ValueOf(cert.ParsePrivateKeyBytes),
		"ParsePrivateKeys":       reflect.ValueOf(cert.ParsePrivateKeys),
		"Print":                  reflect.ValueOf(cert.Print),
	}
	var (
		options cert.Options
		pair    cert.Pair
		pool    cert.Pool
		subject cert.Subject
	)
	env.PackageTypes["project/internal/cert"] = map[string]reflect.Type{
		"Options": reflect.TypeOf(&options).Elem(),
		"Pair":    reflect.TypeOf(&pair).Elem(),
		"Pool":    reflect.TypeOf(&pool).Elem(),
		"Subject": reflect.TypeOf(&subject).Elem(),
	}
}

func initInternalConvert() {
	env.Packages["project/internal/convert"] = map[string]reflect.Value{
		// define constants
		"Byte": reflect.ValueOf(convert.Byte),
		"EB":   reflect.ValueOf(convert.EB),
		"GB":   reflect.ValueOf(convert.GB),
		"KB":   reflect.ValueOf(convert.KB),
		"MB":   reflect.ValueOf(convert.MB),
		"PB":   reflect.ValueOf(convert.PB),
		"TB":   reflect.ValueOf(convert.TB),

		// define variables

		// define functions
		"AbsInt64":            reflect.ValueOf(convert.AbsInt64),
		"BEBytesToFloat32":    reflect.ValueOf(convert.BEBytesToFloat32),
		"BEBytesToFloat64":    reflect.ValueOf(convert.BEBytesToFloat64),
		"BEBytesToInt16":      reflect.ValueOf(convert.BEBytesToInt16),
		"BEBytesToInt32":      reflect.ValueOf(convert.BEBytesToInt32),
		"BEBytesToInt64":      reflect.ValueOf(convert.BEBytesToInt64),
		"BEBytesToUint16":     reflect.ValueOf(convert.BEBytesToUint16),
		"BEBytesToUint32":     reflect.ValueOf(convert.BEBytesToUint32),
		"BEBytesToUint64":     reflect.ValueOf(convert.BEBytesToUint64),
		"BEFloat32ToBytes":    reflect.ValueOf(convert.BEFloat32ToBytes),
		"BEFloat64ToBytes":    reflect.ValueOf(convert.BEFloat64ToBytes),
		"BEInt16ToBytes":      reflect.ValueOf(convert.BEInt16ToBytes),
		"BEInt32ToBytes":      reflect.ValueOf(convert.BEInt32ToBytes),
		"BEInt64ToBytes":      reflect.ValueOf(convert.BEInt64ToBytes),
		"BEUint16ToBytes":     reflect.ValueOf(convert.BEUint16ToBytes),
		"BEUint32ToBytes":     reflect.ValueOf(convert.BEUint32ToBytes),
		"BEUint64ToBytes":     reflect.ValueOf(convert.BEUint64ToBytes),
		"FormatByte":          reflect.ValueOf(convert.FormatByte),
		"FormatNumber":        reflect.ValueOf(convert.FormatNumber),
		"LEBytesToFloat32":    reflect.ValueOf(convert.LEBytesToFloat32),
		"LEBytesToFloat64":    reflect.ValueOf(convert.LEBytesToFloat64),
		"LEBytesToInt16":      reflect.ValueOf(convert.LEBytesToInt16),
		"LEBytesToInt32":      reflect.ValueOf(convert.LEBytesToInt32),
		"LEBytesToInt64":      reflect.ValueOf(convert.LEBytesToInt64),
		"LEBytesToUint16":     reflect.ValueOf(convert.LEBytesToUint16),
		"LEBytesToUint32":     reflect.ValueOf(convert.LEBytesToUint32),
		"LEBytesToUint64":     reflect.ValueOf(convert.LEBytesToUint64),
		"LEFloat32ToBytes":    reflect.ValueOf(convert.LEFloat32ToBytes),
		"LEFloat64ToBytes":    reflect.ValueOf(convert.LEFloat64ToBytes),
		"LEInt16ToBytes":      reflect.ValueOf(convert.LEInt16ToBytes),
		"LEInt32ToBytes":      reflect.ValueOf(convert.LEInt32ToBytes),
		"LEInt64ToBytes":      reflect.ValueOf(convert.LEInt64ToBytes),
		"LEUint16ToBytes":     reflect.ValueOf(convert.LEUint16ToBytes),
		"LEUint32ToBytes":     reflect.ValueOf(convert.LEUint32ToBytes),
		"LEUint64ToBytes":     reflect.ValueOf(convert.LEUint64ToBytes),
		"OutputBytes":         reflect.ValueOf(convert.OutputBytes),
		"OutputBytesWithSize": reflect.ValueOf(convert.OutputBytesWithSize),
	}
	var ()
	env.PackageTypes["project/internal/convert"] = map[string]reflect.Type{}
}

func initInternalCryptoAES() {
	env.Packages["project/internal/crypto/aes"] = map[string]reflect.Value{
		// define constants
		"BlockSize": reflect.ValueOf(aes.BlockSize),
		"IVSize":    reflect.ValueOf(aes.IVSize),
		"Key128Bit": reflect.ValueOf(aes.Key128Bit),
		"Key192Bit": reflect.ValueOf(aes.Key192Bit),
		"Key256Bit": reflect.ValueOf(aes.Key256Bit),

		// define variables
		"ErrEmptyData":          reflect.ValueOf(aes.ErrEmptyData),
		"ErrInvalidCipherData":  reflect.ValueOf(aes.ErrInvalidCipherData),
		"ErrInvalidIVSize":      reflect.ValueOf(aes.ErrInvalidIVSize),
		"ErrInvalidPaddingSize": reflect.ValueOf(aes.ErrInvalidPaddingSize),

		// define functions
		"CBCDecrypt": reflect.ValueOf(aes.CBCDecrypt),
		"CBCEncrypt": reflect.ValueOf(aes.CBCEncrypt),
		"NewCBC":     reflect.ValueOf(aes.NewCBC),
	}
	var (
		cBC aes.CBC
	)
	env.PackageTypes["project/internal/crypto/aes"] = map[string]reflect.Type{
		"CBC": reflect.TypeOf(&cBC).Elem(),
	}
}

func initInternalCryptoCurve25519() {
	env.Packages["project/internal/crypto/curve25519"] = map[string]reflect.Value{
		// define constants
		"ScalarSize": reflect.ValueOf(curve25519.ScalarSize),

		// define variables

		// define functions
		"ScalarBaseMult": reflect.ValueOf(curve25519.ScalarBaseMult),
		"ScalarMult":     reflect.ValueOf(curve25519.ScalarMult),
	}
	var ()
	env.PackageTypes["project/internal/crypto/curve25519"] = map[string]reflect.Type{}
}

func initInternalCryptoED25519() {
	env.Packages["project/internal/crypto/ed25519"] = map[string]reflect.Value{
		// define constants
		"PrivateKeySize": reflect.ValueOf(ed25519.PrivateKeySize),
		"PublicKeySize":  reflect.ValueOf(ed25519.PublicKeySize),
		"SeedSize":       reflect.ValueOf(ed25519.SeedSize),
		"SignatureSize":  reflect.ValueOf(ed25519.SignatureSize),

		// define variables
		"ErrInvalidPrivateKey": reflect.ValueOf(ed25519.ErrInvalidPrivateKey),
		"ErrInvalidPublicKey":  reflect.ValueOf(ed25519.ErrInvalidPublicKey),

		// define functions
		"GenerateKey":      reflect.ValueOf(ed25519.GenerateKey),
		"ImportPrivateKey": reflect.ValueOf(ed25519.ImportPrivateKey),
		"ImportPublicKey":  reflect.ValueOf(ed25519.ImportPublicKey),
		"NewKeyFromSeed":   reflect.ValueOf(ed25519.NewKeyFromSeed),
		"Sign":             reflect.ValueOf(ed25519.Sign),
		"Verify":           reflect.ValueOf(ed25519.Verify),
	}
	var (
		privateKey ed25519.PrivateKey
		publicKey  ed25519.PublicKey
	)
	env.PackageTypes["project/internal/crypto/ed25519"] = map[string]reflect.Type{
		"PrivateKey": reflect.TypeOf(&privateKey).Elem(),
		"PublicKey":  reflect.TypeOf(&publicKey).Elem(),
	}
}

func initInternalCryptoHMAC() {
	env.Packages["project/internal/crypto/hmac"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"Equal": reflect.ValueOf(hmac.Equal),
		"New":   reflect.ValueOf(hmac.New),
	}
	var ()
	env.PackageTypes["project/internal/crypto/hmac"] = map[string]reflect.Type{}
}

func initInternalCryptoLSB() {
	env.Packages["project/internal/crypto/lsb"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"CalculateStorageSize": reflect.ValueOf(lsb.CalculateStorageSize),
		"Decrypt":              reflect.ValueOf(lsb.Decrypt),
		"DecryptFromPNG":       reflect.ValueOf(lsb.DecryptFromPNG),
		"Encrypt":              reflect.ValueOf(lsb.Encrypt),
		"EncryptToPNG":         reflect.ValueOf(lsb.EncryptToPNG),
	}
	var ()
	env.PackageTypes["project/internal/crypto/lsb"] = map[string]reflect.Type{}
}

func initInternalCryptoRand() {
	env.Packages["project/internal/crypto/rand"] = map[string]reflect.Value{
		// define constants

		// define variables
		"Reader": reflect.ValueOf(rand.Reader),

		// define functions
	}
	var ()
	env.PackageTypes["project/internal/crypto/rand"] = map[string]reflect.Type{}
}

func initInternalDNS() {
	env.Packages["project/internal/dns"] = map[string]reflect.Value{
		// define constants
		"MethodDoH":  reflect.ValueOf(dns.MethodDoH),
		"MethodDoT":  reflect.ValueOf(dns.MethodDoT),
		"MethodTCP":  reflect.ValueOf(dns.MethodTCP),
		"MethodUDP":  reflect.ValueOf(dns.MethodUDP),
		"ModeCustom": reflect.ValueOf(dns.ModeCustom),
		"ModeSystem": reflect.ValueOf(dns.ModeSystem),
		"TypeIPv4":   reflect.ValueOf(dns.TypeIPv4),
		"TypeIPv6":   reflect.ValueOf(dns.TypeIPv6),

		// define variables
		"ErrInvalidExpireTime": reflect.ValueOf(dns.ErrInvalidExpireTime),
		"ErrNoConnection":      reflect.ValueOf(dns.ErrNoConnection),
		"ErrNoDNSServers":      reflect.ValueOf(dns.ErrNoDNSServers),
		"ErrNoResolveResult":   reflect.ValueOf(dns.ErrNoResolveResult),

		// define functions
		"IsDomainName": reflect.ValueOf(dns.IsDomainName),
		"NewClient":    reflect.ValueOf(dns.NewClient),
	}
	var (
		client             dns.Client
		options            dns.Options
		server             dns.Server
		unknownMethodError dns.UnknownMethodError
		unknownTypeError   dns.UnknownTypeError
	)
	env.PackageTypes["project/internal/dns"] = map[string]reflect.Type{
		"Client":             reflect.TypeOf(&client).Elem(),
		"Options":            reflect.TypeOf(&options).Elem(),
		"Server":             reflect.TypeOf(&server).Elem(),
		"UnknownMethodError": reflect.TypeOf(&unknownMethodError).Elem(),
		"UnknownTypeError":   reflect.TypeOf(&unknownTypeError).Elem(),
	}
}

func initInternalGUID() {
	env.Packages["project/internal/guid"] = map[string]reflect.Value{
		// define constants
		"Size": reflect.ValueOf(guid.Size),

		// define variables

		// define functions
		"New": reflect.ValueOf(guid.New),
	}
	var (
		gUID      guid.GUID
		generator guid.Generator
	)
	env.PackageTypes["project/internal/guid"] = map[string]reflect.Type{
		"GUID":      reflect.TypeOf(&gUID).Elem(),
		"Generator": reflect.TypeOf(&generator).Elem(),
	}
}

func initInternalHTTPTool() {
	env.Packages["project/internal/httptool"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"FprintRequest":        reflect.ValueOf(httptool.FprintRequest),
		"NewSubHTTPFileSystem": reflect.ValueOf(httptool.NewSubHTTPFileSystem),
		"PrintRequest":         reflect.ValueOf(httptool.PrintRequest),
	}
	var ()
	env.PackageTypes["project/internal/httptool"] = map[string]reflect.Type{}
}

func initInternalLogger() {
	env.Packages["project/internal/logger"] = map[string]reflect.Value{
		// define constants
		"Debug":      reflect.ValueOf(logger.Debug),
		"Error":      reflect.ValueOf(logger.Error),
		"Exploit":    reflect.ValueOf(logger.Exploit),
		"Fatal":      reflect.ValueOf(logger.Fatal),
		"Info":       reflect.ValueOf(logger.Info),
		"Off":        reflect.ValueOf(logger.Off),
		"TimeLayout": reflect.ValueOf(logger.TimeLayout),
		"Warning":    reflect.ValueOf(logger.Warning),

		// define variables
		"Common":  reflect.ValueOf(logger.Common),
		"Discard": reflect.ValueOf(logger.Discard),
		"Test":    reflect.ValueOf(logger.Test),

		// define functions
		"Conn":                reflect.ValueOf(logger.Conn),
		"HijackLogWriter":     reflect.ValueOf(logger.HijackLogWriter),
		"NewMultiLogger":      reflect.ValueOf(logger.NewMultiLogger),
		"NewWriterWithPrefix": reflect.ValueOf(logger.NewWriterWithPrefix),
		"Parse":               reflect.ValueOf(logger.Parse),
		"Prefix":              reflect.ValueOf(logger.Prefix),
		"SetErrorLogger":      reflect.ValueOf(logger.SetErrorLogger),
		"Wrap":                reflect.ValueOf(logger.Wrap),
		"WrapLogger":          reflect.ValueOf(logger.WrapLogger),
	}
	var (
		level       logger.Level
		levelSetter logger.LevelSetter
		lg          logger.Logger
		multiLogger logger.MultiLogger
	)
	env.PackageTypes["project/internal/logger"] = map[string]reflect.Type{
		"Level":       reflect.TypeOf(&level).Elem(),
		"LevelSetter": reflect.TypeOf(&levelSetter).Elem(),
		"Logger":      reflect.TypeOf(&lg).Elem(),
		"MultiLogger": reflect.TypeOf(&multiLogger).Elem(),
	}
}

func initInternalNamer() {
	env.Packages["project/internal/namer"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"NewEnglish": reflect.ValueOf(namer.NewEnglish),
	}
	var (
		english namer.English
		n       namer.Namer
		options namer.Options
	)
	env.PackageTypes["project/internal/namer"] = map[string]reflect.Type{
		"English": reflect.TypeOf(&english).Elem(),
		"Namer":   reflect.TypeOf(&n).Elem(),
		"Options": reflect.TypeOf(&options).Elem(),
	}
}

func initInternalNetTool() {
	env.Packages["project/internal/nettool"] = map[string]reflect.Value{
		// define constants

		// define variables
		"ErrEmptyPort": reflect.ValueOf(nettool.ErrEmptyPort),

		// define functions
		"CheckPort":             reflect.ValueOf(nettool.CheckPort),
		"CheckPortString":       reflect.ValueOf(nettool.CheckPortString),
		"DeadlineConn":          reflect.ValueOf(nettool.DeadlineConn),
		"DecodeExternalAddress": reflect.ValueOf(nettool.DecodeExternalAddress),
		"EncodeExternalAddress": reflect.ValueOf(nettool.EncodeExternalAddress),
		"IPEnabled":             reflect.ValueOf(nettool.IPEnabled),
		"IPToHost":              reflect.ValueOf(nettool.IPToHost),
		"IsNetClosingError":     reflect.ValueOf(nettool.IsNetClosingError),
		"JoinHostPort":          reflect.ValueOf(nettool.JoinHostPort),
		"SplitHostPort":         reflect.ValueOf(nettool.SplitHostPort),
	}
	var ()
	env.PackageTypes["project/internal/nettool"] = map[string]reflect.Type{}
}

func initInternalOption() {
	env.Packages["project/internal/option"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
	}
	var (
		hTTPRequest   option.HTTPRequest
		hTTPServer    option.HTTPServer
		hTTPTransport option.HTTPTransport
		tLSConfig     option.TLSConfig
		x509KeyPair   option.X509KeyPair
	)
	env.PackageTypes["project/internal/option"] = map[string]reflect.Type{
		"HTTPRequest":   reflect.TypeOf(&hTTPRequest).Elem(),
		"HTTPServer":    reflect.TypeOf(&hTTPServer).Elem(),
		"HTTPTransport": reflect.TypeOf(&hTTPTransport).Elem(),
		"TLSConfig":     reflect.TypeOf(&tLSConfig).Elem(),
		"X509KeyPair":   reflect.TypeOf(&x509KeyPair).Elem(),
	}
}

func initInternalPatchJSON() {
	env.Packages["project/internal/patch/json"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"Marshal":    reflect.ValueOf(json.Marshal),
		"NewDecoder": reflect.ValueOf(json.NewDecoder),
		"NewEncoder": reflect.ValueOf(json.NewEncoder),
		"Unmarshal":  reflect.ValueOf(json.Unmarshal),
	}
	var (
		decoder json.Decoder
		encoder json.Encoder
	)
	env.PackageTypes["project/internal/patch/json"] = map[string]reflect.Type{
		"Decoder": reflect.TypeOf(&decoder).Elem(),
		"Encoder": reflect.TypeOf(&encoder).Elem(),
	}
}

func initInternalPatchMsgpack() {
	env.Packages["project/internal/patch/msgpack"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"Marshal":    reflect.ValueOf(msgpack.Marshal),
		"NewDecoder": reflect.ValueOf(msgpack.NewDecoder),
		"NewEncoder": reflect.ValueOf(msgpack.NewEncoder),
		"Unmarshal":  reflect.ValueOf(msgpack.Unmarshal),
	}
	var (
		decoder msgpack.Decoder
		encoder msgpack.Encoder
	)
	env.PackageTypes["project/internal/patch/msgpack"] = map[string]reflect.Type{
		"Decoder": reflect.TypeOf(&decoder).Elem(),
		"Encoder": reflect.TypeOf(&encoder).Elem(),
	}
}

func initInternalPatchToml() {
	env.Packages["project/internal/patch/toml"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"Marshal":   reflect.ValueOf(toml.Marshal),
		"Unmarshal": reflect.ValueOf(toml.Unmarshal),
	}
	var ()
	env.PackageTypes["project/internal/patch/toml"] = map[string]reflect.Type{}
}

func initInternalProxy() {
	env.Packages["project/internal/proxy"] = map[string]reflect.Value{
		// define constants
		"ModeBalance": reflect.ValueOf(proxy.ModeBalance),
		"ModeChain":   reflect.ValueOf(proxy.ModeChain),
		"ModeDirect":  reflect.ValueOf(proxy.ModeDirect),
		"ModeHTTP":    reflect.ValueOf(proxy.ModeHTTP),
		"ModeHTTPS":   reflect.ValueOf(proxy.ModeHTTPS),
		"ModeSocks4":  reflect.ValueOf(proxy.ModeSocks4),
		"ModeSocks4a": reflect.ValueOf(proxy.ModeSocks4a),
		"ModeSocks5":  reflect.ValueOf(proxy.ModeSocks5),

		// define variables

		// define functions
		"NewBalance": reflect.ValueOf(proxy.NewBalance),
		"NewChain":   reflect.ValueOf(proxy.NewChain),
		"NewManager": reflect.ValueOf(proxy.NewManager),
		"NewPool":    reflect.ValueOf(proxy.NewPool),
	}
	var (
		balance proxy.Balance
		chain   proxy.Chain
		client  proxy.Client
		manager proxy.Manager
		pool    proxy.Pool
		server  proxy.Server
	)
	env.PackageTypes["project/internal/proxy"] = map[string]reflect.Type{
		"Balance": reflect.TypeOf(&balance).Elem(),
		"Chain":   reflect.TypeOf(&chain).Elem(),
		"Client":  reflect.TypeOf(&client).Elem(),
		"Manager": reflect.TypeOf(&manager).Elem(),
		"Pool":    reflect.TypeOf(&pool).Elem(),
		"Server":  reflect.TypeOf(&server).Elem(),
	}
}

func initInternalProxyDirect() {
	env.Packages["project/internal/proxy/direct"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
	}
	var (
		d direct.Direct
	)
	env.PackageTypes["project/internal/proxy/direct"] = map[string]reflect.Type{
		"Direct": reflect.TypeOf(&d).Elem(),
	}
}

func initInternalProxyHTTP() {
	env.Packages["project/internal/proxy/http"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"CheckNetwork":   reflect.ValueOf(http.CheckNetwork),
		"NewHTTPClient":  reflect.ValueOf(http.NewHTTPClient),
		"NewHTTPSClient": reflect.ValueOf(http.NewHTTPSClient),
		"NewHTTPSServer": reflect.ValueOf(http.NewHTTPSServer),
		"NewHTTPServer":  reflect.ValueOf(http.NewHTTPServer),
	}
	var (
		client  http.Client
		options http.Options
		server  http.Server
	)
	env.PackageTypes["project/internal/proxy/http"] = map[string]reflect.Type{
		"Client":  reflect.TypeOf(&client).Elem(),
		"Options": reflect.TypeOf(&options).Elem(),
		"Server":  reflect.TypeOf(&server).Elem(),
	}
}

func initInternalProxySocks() {
	env.Packages["project/internal/proxy/socks"] = map[string]reflect.Value{
		// define constants

		// define variables
		"ErrServerClosed": reflect.ValueOf(socks.ErrServerClosed),

		// define functions
		"CheckNetwork":     reflect.ValueOf(socks.CheckNetwork),
		"NewSocks4Client":  reflect.ValueOf(socks.NewSocks4Client),
		"NewSocks4Server":  reflect.ValueOf(socks.NewSocks4Server),
		"NewSocks4aClient": reflect.ValueOf(socks.NewSocks4aClient),
		"NewSocks4aServer": reflect.ValueOf(socks.NewSocks4aServer),
		"NewSocks5Client":  reflect.ValueOf(socks.NewSocks5Client),
		"NewSocks5Server":  reflect.ValueOf(socks.NewSocks5Server),
	}
	var (
		client  socks.Client
		options socks.Options
		server  socks.Server
	)
	env.PackageTypes["project/internal/proxy/socks"] = map[string]reflect.Type{
		"Client":  reflect.TypeOf(&client).Elem(),
		"Options": reflect.TypeOf(&options).Elem(),
		"Server":  reflect.TypeOf(&server).Elem(),
	}
}

func initInternalRandom() {
	env.Packages["project/internal/random"] = map[string]reflect.Value{
		// define constants
		"MaxSleepTime": reflect.ValueOf(random.MaxSleepTime),

		// define variables

		// define functions
		"Bytes":      reflect.ValueOf(random.Bytes),
		"Cookie":     reflect.ValueOf(random.Cookie),
		"Int":        reflect.ValueOf(random.Int),
		"Int64":      reflect.ValueOf(random.Int64),
		"NewRand":    reflect.ValueOf(random.NewRand),
		"NewSleeper": reflect.ValueOf(random.NewSleeper),
		"Sleep":      reflect.ValueOf(random.Sleep),
		"String":     reflect.ValueOf(random.String),
		"Uint64":     reflect.ValueOf(random.Uint64),
	}
	var (
		r       random.Rand
		sleeper random.Sleeper
	)
	env.PackageTypes["project/internal/random"] = map[string]reflect.Type{
		"Rand":    reflect.TypeOf(&r).Elem(),
		"Sleeper": reflect.TypeOf(&sleeper).Elem(),
	}
}

func initInternalSecurity() {
	env.Packages["project/internal/security"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"CoverBytes":            reflect.ValueOf(security.CoverBytes),
		"CoverString":           reflect.ValueOf(security.CoverString),
		"FlushMemory":           reflect.ValueOf(security.FlushMemory),
		"NewBogo":               reflect.ValueOf(security.NewBogo),
		"NewBytes":              reflect.ValueOf(security.NewBytes),
		"NewMemory":             reflect.ValueOf(security.NewMemory),
		"PaddingMemory":         reflect.ValueOf(security.PaddingMemory),
		"SwitchThread":          reflect.ValueOf(security.SwitchThread),
		"SwitchThreadAsync":     reflect.ValueOf(security.SwitchThreadAsync),
		"WaitSwitchThreadAsync": reflect.ValueOf(security.WaitSwitchThreadAsync),
	}
	var (
		bogo   security.Bogo
		bytes  security.Bytes
		memory security.Memory
	)
	env.PackageTypes["project/internal/security"] = map[string]reflect.Type{
		"Bogo":   reflect.TypeOf(&bogo).Elem(),
		"Bytes":  reflect.TypeOf(&bytes).Elem(),
		"Memory": reflect.TypeOf(&memory).Elem(),
	}
}

func initInternalSystem() {
	env.Packages["project/internal/system"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"ChangeCurrentDirectory": reflect.ValueOf(system.ChangeCurrentDirectory),
		"CheckError":             reflect.ValueOf(system.CheckError),
		"ExecutableName":         reflect.ValueOf(system.ExecutableName),
		"GetConnHandle":          reflect.ValueOf(system.GetConnHandle),
		"IsExist":                reflect.ValueOf(system.IsExist),
		"IsNotExist":             reflect.ValueOf(system.IsNotExist),
		"OpenFile":               reflect.ValueOf(system.OpenFile),
		"WriteFile":              reflect.ValueOf(system.WriteFile),
	}
	var ()
	env.PackageTypes["project/internal/system"] = map[string]reflect.Type{}
}

func initInternalTimeSync() {
	env.Packages["project/internal/timesync"] = map[string]reflect.Value{
		// define constants
		"ModeHTTP": reflect.ValueOf(timesync.ModeHTTP),
		"ModeNTP":  reflect.ValueOf(timesync.ModeNTP),

		// define variables
		"ErrAllClientsFailed": reflect.ValueOf(timesync.ErrAllClientsFailed),
		"ErrNoClients":        reflect.ValueOf(timesync.ErrNoClients),

		// define functions
		"NewHTTP":   reflect.ValueOf(timesync.NewHTTP),
		"NewNTP":    reflect.ValueOf(timesync.NewNTP),
		"NewSyncer": reflect.ValueOf(timesync.NewSyncer),
		"TestHTTP":  reflect.ValueOf(timesync.TestHTTP),
		"TestNTP":   reflect.ValueOf(timesync.TestNTP),
	}
	var (
		client timesync.Client
		hTTP   timesync.HTTP
		nTP    timesync.NTP
		syncer timesync.Syncer
	)
	env.PackageTypes["project/internal/timesync"] = map[string]reflect.Type{
		"Client": reflect.TypeOf(&client).Elem(),
		"HTTP":   reflect.TypeOf(&hTTP).Elem(),
		"NTP":    reflect.TypeOf(&nTP).Elem(),
		"Syncer": reflect.TypeOf(&syncer).Elem(),
	}
}

func initInternalXPanic() {
	env.Packages["project/internal/xpanic"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"Error":      reflect.ValueOf(xpanic.Error),
		"Log":        reflect.ValueOf(xpanic.Log),
		"Print":      reflect.ValueOf(xpanic.Print),
		"PrintPanic": reflect.ValueOf(xpanic.PrintPanic),
		"PrintStack": reflect.ValueOf(xpanic.PrintStack),
	}
	var ()
	env.PackageTypes["project/internal/xpanic"] = map[string]reflect.Type{}
}

func initInternalXReflect() {
	env.Packages["project/internal/xreflect"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
		"GetStructureName":          reflect.ValueOf(xreflect.GetStructureName),
		"StructureToMap":            reflect.ValueOf(xreflect.StructureToMap),
		"StructureToMapWithoutZero": reflect.ValueOf(xreflect.StructureToMapWithoutZero),
	}
	var ()
	env.PackageTypes["project/internal/xreflect"] = map[string]reflect.Type{}
}

func initInternalXSync() {
	env.Packages["project/internal/xsync"] = map[string]reflect.Value{
		// define constants

		// define variables

		// define functions
	}
	var (
		counter xsync.Counter
	)
	env.PackageTypes["project/internal/xsync"] = map[string]reflect.Type{
		"Counter": reflect.TypeOf(&counter).Elem(),
	}
}
