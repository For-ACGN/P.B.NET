package main

import (
	"crypto/x509"
	"flag"
	"os"

	"project/internal/cert"
	"project/internal/namer"
	"project/internal/patch/toml"
	"project/internal/system"
)

func main() {
	var (
		gen  bool
		sign bool
		opt  string
		nt   string
		np   string
	)
	flag.BoolVar(&gen, "gen", false, "generate CA certificate")
	flag.BoolVar(&sign, "sign", false, "generate certificate and sign it by CA")
	flag.StringVar(&opt, "opt", "options.toml", "options file path")
	flag.StringVar(&nt, "nt", "english", "namer type")
	flag.StringVar(&np, "np", "namer/english.zip", "namer resource path")
	flag.Parse()
	// load options
	options, err := os.ReadFile(opt) // #nosec
	system.CheckError(err)
	opts := new(cert.Options)
	err = toml.Unmarshal(options, opts)
	system.CheckError(err)
	// load namer
	res, err := os.ReadFile(np) // #nosec
	system.CheckError(err)
	opts.Namer, err = namer.Load(nt, res)
	system.CheckError(err)
	// generate or sign certificate
	var crt *x509.Certificate
	switch {
	case gen:
		ca, err := cert.GenerateCA(opts)
		system.CheckError(err)
		caCert, caKey := ca.EncodeToPEM()
		// save pair
		err = system.WriteFile("ca_cert.pem", caCert)
		system.CheckError(err)
		err = system.WriteFile("ca_key.pem", caKey)
		system.CheckError(err)
		crt = ca.Certificate
	case sign:
		// load CA certificate
		pemData, err := os.ReadFile("ca_cert.pem")
		system.CheckError(err)
		caCert, err := cert.ParseCertificatePEM(pemData)
		system.CheckError(err)
		// load CA private key
		pemData, err = os.ReadFile("ca_key.pem")
		system.CheckError(err)
		caKey, err := cert.ParsePrivateKeyPEM(pemData)
		system.CheckError(err)
		// generate certificate
		child, err := cert.Generate(caCert, caKey, opts)
		system.CheckError(err)
		childCert, childKey := child.EncodeToPEM()
		// save pair
		err = system.WriteFile("cert.pem", childCert)
		system.CheckError(err)
		err = system.WriteFile("key.pem", childKey)
		system.CheckError(err)
		crt = child.Certificate
	default:
		flag.PrintDefaults()
		return
	}
	cert.Dump(crt)
}
