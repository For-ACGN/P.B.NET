package main

import (
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"

	"project/internal/cert"
	"project/internal/namer"
	"project/internal/patch/toml"
	"project/internal/system"
)

func main() {
	var (
		gc  bool
		gs  bool
		opt string
		nt  string
		np  string
	)
	flag.BoolVar(&gc, "gc", false, "generate CA certificate")
	flag.BoolVar(&gs, "gs", false, "generate certificate and sign it by CA")
	flag.StringVar(&opt, "opt", "options.toml", "options file path")
	flag.StringVar(&nt, "nt", "english", "namer type")
	flag.StringVar(&np, "np", "namer/english.zip", "namer resource path")
	flag.Parse()
	// load options
	options, err := ioutil.ReadFile(opt) // #nosec
	system.CheckError(err)
	opts := new(cert.Options)
	err = toml.Unmarshal(options, opts)
	system.CheckError(err)
	// load namer
	res, err := ioutil.ReadFile(np) // #nosec
	system.CheckError(err)
	opts.Namer, err = namer.Load(nt, res)
	system.CheckError(err)
	// generate certificate
	var certificate *x509.Certificate
	switch {
	case gc:
		ca, err := cert.GenerateCA(opts)
		system.CheckError(err)
		caCert, caKey := ca.EncodeToPEM()
		err = system.WriteFile("ca.crt", caCert)
		system.CheckError(err)
		err = system.WriteFile("ca.key", caKey)
		system.CheckError(err)
		certificate = ca.Certificate
	case gs:
		// load CA certificate
		pemData, err := ioutil.ReadFile("ca.crt")
		system.CheckError(err)
		caCert, err := cert.ParseCertificate(pemData)
		system.CheckError(err)
		// load CA private key
		pemData, err = ioutil.ReadFile("ca.key")
		system.CheckError(err)
		caKey, err := cert.ParsePrivateKey(pemData)
		system.CheckError(err)
		// generate certificate
		kp, err := cert.Generate(caCert, caKey, opts)
		system.CheckError(err)
		crt, key := kp.EncodeToPEM()
		err = system.WriteFile("server.crt", crt)
		system.CheckError(err)
		err = system.WriteFile("server.key", key)
		system.CheckError(err)
		certificate = kp.Certificate
	default:
		flag.PrintDefaults()
		return
	}
	fmt.Println(cert.Print(certificate))
}
