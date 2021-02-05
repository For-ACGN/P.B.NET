package main

import (
	"flag"
	"os"

	"project/internal/cert"
	"project/internal/namer"
	"project/internal/patch/toml"
	"project/internal/system"
)

var (
	gen  bool
	sign bool
	opts string
	typ  string
	res  string
)

func init() {
	flag.CommandLine.SetOutput(os.Stdout)

	flag.BoolVar(&gen, "gen", false, "generate CA certificate")
	flag.BoolVar(&sign, "sign", false, "generate certificate and sign it by CA")
	flag.StringVar(&opts, "opts", "options.toml", "options file path")
	flag.StringVar(&typ, "namer", "english", "namer type")
	flag.StringVar(&res, "res", "namer/english.zip", "namer resource path")
	flag.Parse()
}

func main() {
	// load options
	data, err := os.ReadFile(opts) // #nosec
	system.CheckError(err)
	opts := new(cert.Options)
	err = toml.Unmarshal(data, opts)
	system.CheckError(err)
	// initialize namer
	resource, err := os.ReadFile(res) // #nosec
	system.CheckError(err)
	opts.Namer, err = namer.Load(typ, resource)
	system.CheckError(err)
	switch {
	case gen:
		generateCertificate(opts)
	case sign:
		signCertificate(opts)
	default:
		flag.PrintDefaults()
	}
}

func generateCertificate(opts *cert.Options) {
	ca, err := cert.GenerateCA(opts)
	system.CheckError(err)
	caCert, caKey := ca.EncodeToPEM()
	// save pair
	err = system.WriteFile("ca_cert.pem", caCert)
	system.CheckError(err)
	err = system.WriteFile("ca_key.pem", caKey)
	system.CheckError(err)
	// print information
	cert.Dump(ca.Certificate)
}

func signCertificate(opts *cert.Options) {
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
	// print information
	cert.Dump(child.Certificate)
}
