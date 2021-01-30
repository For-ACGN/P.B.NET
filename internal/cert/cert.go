package cert

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"project/internal/convert"
	"project/internal/crypto/rand"
	"project/internal/namer"
	"project/internal/random"
	"project/internal/security"
)

// Options contains options about generate certificate.
type Options struct {
	// Algorithm is the certificate signer algorithm
	// like "rsa|2048", "ecdsa|p256", "ed25519"
	Algorithm string `toml:"algorithm"`

	// These values may not be valid if invalid values
	// were contained within a parsed certificate. For
	// example, an element of DNSNames may not be a valid
	// DNS domain name. IPAddresses is used to IP SANs.
	DNSNames       []string `toml:"dns_names"`
	IPAddresses    []string `toml:"ip_addresses"`
	EmailAddresses []string `toml:"email_addresses"`
	URLs           []string `toml:"urls"`

	// Valid time range of certificate.
	NotBefore time.Time `toml:"not_before"`
	NotAfter  time.Time `toml:"not_after"`

	// subject information.
	Subject Subject `toml:"subject"`

	// for generate random name like Subject.CommonName.
	NamerOpts *namer.Options `toml:"namer"`
	Namer     namer.Namer    `toml:"-" msgpack:"-"`
}

// Subject contains certificate subject information.
type Subject struct {
	CommonName         string   `toml:"common_name"`
	SerialNumber       string   `toml:"serial_number"`
	Country            []string `toml:"country"`
	Organization       []string `toml:"organization"`
	OrganizationalUnit []string `toml:"organizational_unit"`
	Locality           []string `toml:"locality"`
	Province           []string `toml:"province"`
	StreetAddress      []string `toml:"street_address"`
	PostalCode         []string `toml:"postal_code"`
}

// Pair contains certificate and private key.
type Pair struct {
	Certificate *x509.Certificate
	PrivateKey  interface{}
}

// ASN1 is used to get certificate ASN1 data.
func (p *Pair) ASN1() []byte {
	asn1Data := make([]byte, len(p.Certificate.Raw))
	copy(asn1Data, p.Certificate.Raw)
	return asn1Data
}

// Encode is used to get certificate ASN1 data and encode private key to PKCS8.
func (p *Pair) Encode() ([]byte, []byte) {
	cert := p.ASN1()
	key, err := x509.MarshalPKCS8PrivateKey(p.PrivateKey)
	if err != nil {
		panic(fmt.Sprintf("cert: internal error: %s", err))
	}
	return cert, key
}

// EncodeToPEM is used to encode certificate and private key to PEM data.
func (p *Pair) EncodeToPEM() ([]byte, []byte) {
	cert, key := p.Encode()
	defer func() {
		security.CoverBytes(cert)
		security.CoverBytes(key)
	}()
	certBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert,
	}
	keyBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: key,
	}
	return pem.EncodeToMemory(certBlock), pem.EncodeToMemory(keyBlock)
}

// TLSCertificate is used to generate tls certificate.
func (p *Pair) TLSCertificate() tls.Certificate {
	var cert tls.Certificate
	cert.Certificate = make([][]byte, 1)
	cert.Certificate[0] = make([]byte, len(p.Certificate.Raw))
	copy(cert.Certificate[0], p.Certificate.Raw)
	cert.PrivateKey = p.PrivateKey
	return cert
}

// GenerateCA is used to generate a CA certificate from Options.
func GenerateCA(opts *Options) (*Pair, error) {
	if opts == nil {
		opts = new(Options)
	}
	ca, err := generateCertificate(opts)
	if err != nil {
		return nil, err
	}
	ca.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageCRLSign
	ca.BasicConstraintsValid = true
	ca.IsCA = true
	privateKey, publicKey, err := generatePrivateKey(opts.Algorithm)
	if err != nil {
		return nil, err
	}
	asn1Data, err := x509.CreateCertificate(rand.Reader, ca, ca, publicKey, privateKey)
	if err != nil {
		return nil, err
	}
	ca, _ = x509.ParseCertificate(asn1Data)
	return &Pair{Certificate: ca, PrivateKey: privateKey}, nil
}

// Generate is used to generate a signed certificate by CA or self-signed certificate.
func Generate(parent *x509.Certificate, pri interface{}, opts *Options) (*Pair, error) {
	if opts == nil {
		opts = new(Options)
	}
	cert, err := generateCertificate(opts)
	if err != nil {
		return nil, err
	}
	cert.KeyUsage = x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment
	// generate certificate
	privateKey, publicKey, err := generatePrivateKey(opts.Algorithm)
	if err != nil {
		return nil, err
	}
	var asn1Data []byte
	if parent != nil && pri != nil { // by CA
		asn1Data, err = x509.CreateCertificate(rand.Reader, cert, parent, publicKey, pri)
	} else { // self-sign
		asn1Data, err = x509.CreateCertificate(rand.Reader, cert, cert, publicKey, privateKey)
	}
	if err != nil {
		return nil, err
	}
	cert, _ = x509.ParseCertificate(asn1Data)
	return &Pair{Certificate: cert, PrivateKey: privateKey}, nil
}

func generateCertificate(opts *Options) (*x509.Certificate, error) {
	r := random.NewRand()
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(r.Int63()),
		SubjectKeyId: r.Bytes(4),
	}
	err := setCertCommonName(r, cert, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to set common name: %s", err)
	}
	err = setCertOrganization(r, cert, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to set organization: %s", err)
	}
	// check and set domain names
	dn := opts.DNSNames
	for i := 0; i < len(dn); i++ {
		if !isDomainName(dn[i]) {
			return nil, fmt.Errorf("%s is not a valid domain name", dn[i])
		}
		cert.DNSNames = append(cert.DNSNames, dn[i])
	}
	// check and set IP addresses
	ips := opts.IPAddresses
	for i := 0; i < len(ips); i++ {
		ip := net.ParseIP(ips[i])
		if ip == nil {
			return nil, fmt.Errorf("%s is not a valid IP address", ips[i])
		}
		cert.IPAddresses = append(cert.IPAddresses, ip)
	}
	// check and set email addresses
	emails := opts.EmailAddresses
	for i := 0; i < len(emails); i++ {
		if !isEmailAddress(emails[i]) {
			return nil, fmt.Errorf("%s is not a valid email address", emails[i])
		}
		cert.EmailAddresses = append(cert.EmailAddresses, emails[i])
	}
	// check and set URLs
	urls := opts.URLs
	for i := 0; i < len(urls); i++ {
		URL, err := url.Parse(urls[i])
		if err != nil {
			return nil, fmt.Errorf("%s is not a valid url", urls[i])
		}
		cert.URIs = append(cert.URIs, URL)
	}
	setCertTime(r, cert, opts)
	// copy []string in Subject
	cert.Subject.Country = make([]string, len(opts.Subject.Country))
	copy(cert.Subject.Country, opts.Subject.Country)
	cert.Subject.OrganizationalUnit = make([]string, len(opts.Subject.OrganizationalUnit))
	copy(cert.Subject.OrganizationalUnit, opts.Subject.OrganizationalUnit)
	cert.Subject.Locality = make([]string, len(opts.Subject.Locality))
	copy(cert.Subject.Locality, opts.Subject.Locality)
	cert.Subject.Province = make([]string, len(opts.Subject.Province))
	copy(cert.Subject.Province, opts.Subject.Province)
	cert.Subject.StreetAddress = make([]string, len(opts.Subject.StreetAddress))
	copy(cert.Subject.StreetAddress, opts.Subject.StreetAddress)
	cert.Subject.PostalCode = make([]string, len(opts.Subject.PostalCode))
	copy(cert.Subject.PostalCode, opts.Subject.PostalCode)
	cert.Subject.SerialNumber = opts.Subject.SerialNumber
	return cert, nil
}

func setCertCommonName(r *random.Rand, cert *x509.Certificate, opts *Options) error {
	if opts.Subject.CommonName != "" {
		cert.Subject.CommonName = opts.Subject.CommonName
		return nil
	}
	// generate random common name
	if opts.Namer == nil {
		cert.Subject.CommonName = r.String(6 + r.Intn(8))
		return nil
	}
	name, err := opts.Namer.Generate(opts.NamerOpts)
	if err != nil {
		return err
	}
	cert.Subject.CommonName = name
	return nil
}

func setCertOrganization(r *random.Rand, cert *x509.Certificate, opts *Options) error {
	if len(opts.Subject.Organization) != 0 {
		cert.Subject.Organization = make([]string, len(opts.Subject.Organization))
		copy(cert.Subject.Organization, opts.Subject.Organization)
		return nil
	}
	// generate random organization
	if opts.Namer == nil {
		cert.Subject.Organization = []string{r.String(6 + r.Intn(8))}
		return nil
	}
	name, err := opts.Namer.Generate(opts.NamerOpts)
	if err != nil {
		return err
	}
	cert.Subject.Organization = []string{name}
	return nil
}

func setCertTime(r *random.Rand, cert *x509.Certificate, opts *Options) {
	now := time.Date(2018, 11, 27, 0, 0, 0, 0, time.UTC)
	if opts.NotBefore.IsZero() {
		years := r.Intn(10)
		months := r.Intn(12)
		days := r.Intn(31)
		cert.NotBefore = now.AddDate(-years, -months, -days)
	} else {
		cert.NotBefore = opts.NotBefore
	}
	if opts.NotAfter.IsZero() {
		years := 10 + r.Intn(10)
		months := r.Intn(12)
		days := r.Intn(31)
		cert.NotAfter = now.AddDate(years, months, days)
	} else {
		cert.NotAfter = opts.NotAfter
	}
}

// copy from internal/dns/protocol.go
func isDomainName(s string) bool {
	l := len(s)
	if l == 0 || l > 254 || l == 254 && s[l-1] != '.' {
		return false
	}
	last := byte('.')
	nonNumeric := false // true once we've seen a letter or hyphen
	partLen := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		ok := false
		checkChar(c, last, &nonNumeric, &partLen, &ok)
		if !ok {
			return false
		}
		last = c
	}
	if last == '-' || partLen > 63 {
		return false
	}
	return nonNumeric
}

func checkChar(c byte, last byte, nonNumeric *bool, partLen *int, ok *bool) {
	switch {
	case 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || c == '_':
		*nonNumeric = true
		*partLen++
		*ok = true
	case '0' <= c && c <= '9':
		// fine
		*partLen++
		*ok = true
	case c == '-':
		// Byte before dash cannot be dot.
		if last == '.' {
			return
		}
		*partLen++
		*nonNumeric = true
		*ok = true
	case c == '.':
		// Byte before dot cannot be dot, dash.
		if last == '.' || last == '-' {
			return
		}
		if *partLen > 63 || *partLen == 0 {
			return
		}
		*partLen = 0
		*ok = true
	}
}

var emailAddressReg = regexp.MustCompile(`\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*`)

func isEmailAddress(email string) bool {
	return len(emailAddressReg.FindAllString(email, -1)) == 1
}

func generatePrivateKey(algorithm string) (interface{}, interface{}, error) {
	alg := strings.ToLower(algorithm)
	switch alg {
	case "":
		return generateRSA("2048")
	case "ed25519":
		return generateEd25519()
	}
	configs := strings.Split(alg, "|")
	if len(configs) != 2 {
		return nil, nil, fmt.Errorf("invalid algorithm configs: %s", algorithm)
	}
	switch configs[0] {
	case "rsa":
		return generateRSA(configs[1])
	case "ecdsa":
		return generateECDSA(configs[1])
	}
	return nil, nil, fmt.Errorf("unknown algorithm: %s", configs[0])
}

func generateRSA(bits string) (interface{}, interface{}, error) {
	n, err := strconv.Atoi(bits)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid RSA bits: %s %s", bits, err)
	}
	if n < 1024 {
		return nil, nil, errors.New("RSA bits must >= 1024")
	}
	privateKey, err := rsa.GenerateKey(rand.Reader, n)
	if err != nil {
		return nil, nil, err
	}
	return privateKey, &privateKey.PublicKey, nil
}

func generateECDSA(curve string) (interface{}, interface{}, error) {
	var (
		privateKey *ecdsa.PrivateKey
		err        error
	)
	switch curve {
	case "p224":
		privateKey, err = ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	case "p256":
		privateKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case "p384":
		privateKey, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case "p521":
		privateKey, err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	default:
		return nil, nil, fmt.Errorf("unsupported elliptic curve: %s", curve)
	}
	if err != nil {
		return nil, nil, err
	}
	return privateKey, &privateKey.PublicKey, nil
}

func generateEd25519() (interface{}, interface{}, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return privateKey, publicKey, nil
}

const timeLayout = "2006-01-02 15:04:05 Z07:00"

// Dump is used to dump certificate information to os.Stdout.
func Dump(cert *x509.Certificate) {
	buf := bytes.NewBuffer(make([]byte, 0, 512))
	_, _ = Fdump(buf, cert)
	buf.WriteString("\n")
	_, _ = buf.WriteTo(os.Stdout)
}

// Sdump is used to dump certificate information to a string.
func Sdump(cert *x509.Certificate) string {
	builder := strings.Builder{}
	builder.Grow(512)
	_, _ = Fdump(&builder, cert)
	return builder.String()
}

// Fdump is used to dump certificate information to a io.Writer.
func Fdump(w io.Writer, cert *x509.Certificate) (int, error) {
	const format = `
Version: %d
Serial number: [%s]

[Subject]
  Common name:  %s
  Organization: %s

[Issuer]
  Common name:  %s
  Organization: %s

Public key algorithm: %s
Public key: [%s]

Signature algorithm: %s
Signature: [%s]

Not before: %s
Not after:  %s
`
	var num int
	n, err := fmt.Fprintf(w, format[1:],
		cert.Version,
		strings.TrimSuffix(convert.SdumpBytesWithPL(cert.SerialNumber.Bytes(), "", 16), ","),
		cert.Subject.CommonName, strings.Join(cert.Subject.Organization, ", "),
		cert.Issuer.CommonName, strings.Join(cert.Issuer.Organization, ", "),
		cert.PublicKeyAlgorithm,
		strings.TrimSuffix(convert.SdumpBytesWithPL(cert.Signature[:8], "", 8), ","),
		cert.SignatureAlgorithm,
		strings.TrimSuffix(convert.SdumpBytesWithPL(cert.Signature[:8], "", 8), ","),
		cert.NotBefore.Local().Format(timeLayout),
		cert.NotAfter.Local().Format(timeLayout),
	)
	num += n
	if err != nil {
		return num, err
	}
	if len(cert.DNSNames) != 0 {
		n, err = fmt.Fprintf(w, "\nDNS names: [%s]", strings.Join(cert.DNSNames, ", "))
		num += n
		if err != nil {
			return num, err
		}
	}
	if len(cert.IPAddresses) != 0 {
		ip := make([]string, len(cert.IPAddresses))
		for i := 0; i < len(cert.IPAddresses); i++ {
			ip[i] = cert.IPAddresses[i].String()
		}
		n, err = fmt.Fprintf(w, "\nIP addresses: [%s]", strings.Join(ip, ", "))
		num += n
		if err != nil {
			return num, err
		}
	}
	if len(cert.EmailAddresses) != 0 {
		n, err = fmt.Fprintf(w, "\nEmail addresses: [%s]", strings.Join(cert.EmailAddresses, ", "))
		num += n
		if err != nil {
			return num, err
		}
	}
	if len(cert.URIs) != 0 {
		urls := make([]string, len(cert.URIs))
		for i := 0; i < len(cert.URIs); i++ {
			urls[i] = cert.URIs[i].String()
		}
		n, err = fmt.Fprintf(w, "\nURLs: [%s]", strings.Join(urls, ", "))
		num += n
		if err != nil {
			return num, err
		}
	}
	return num, nil
}

// ErrInvalidPEMBlock is the error about the PEM block.
var ErrInvalidPEMBlock = errors.New("invalid PEM block")

// ParseCertificate is used to parse certificate from PEM.
func ParseCertificate(pemBlock []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(pemBlock)
	if block == nil {
		return nil, ErrInvalidPEMBlock
	}
	if block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("invalid PEM block type: %s", block.Type)
	}
	return x509.ParseCertificate(block.Bytes)
}

// ParseCertificates is used to parse certificates from PEM.
func ParseCertificates(pemBlock []byte) ([]*x509.Certificate, error) {
	var (
		certs []*x509.Certificate
		block *pem.Block
	)
	for {
		block, pemBlock = pem.Decode(pemBlock)
		if block == nil {
			return nil, ErrInvalidPEMBlock
		}
		if block.Type != "CERTIFICATE" {
			return nil, fmt.Errorf("invalid PEM block type: %s", block.Type)
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}
		certs = append(certs, cert)
		if len(pemBlock) == 0 {
			break
		}
	}
	return certs, nil
}

// ParsePrivateKey is used to parse private key from PEM.
func ParsePrivateKey(pemBlock []byte) (interface{}, error) {
	block, _ := pem.Decode(pemBlock)
	if block == nil {
		return nil, ErrInvalidPEMBlock
	}
	return ParsePrivateKeyBytes(block.Bytes)
}

// ParsePrivateKeys is used to parse private keys from PEM.
func ParsePrivateKeys(pemBlock []byte) ([]interface{}, error) {
	var (
		keys  []interface{}
		block *pem.Block
	)
	for {
		block, pemBlock = pem.Decode(pemBlock)
		if block == nil {
			return nil, ErrInvalidPEMBlock
		}
		key, err := ParsePrivateKeyBytes(block.Bytes)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
		if len(pemBlock) == 0 {
			break
		}
	}
	return keys, nil
}

// ParsePrivateKeyBytes is used to parse private key from bytes.
func ParsePrivateKeyBytes(bytes []byte) (interface{}, error) {
	if key, err := x509.ParsePKCS1PrivateKey(bytes); err == nil {
		return key, nil
	}
	if key, err := x509.ParsePKCS8PrivateKey(bytes); err == nil {
		return key, nil
	}
	if key, err := x509.ParseECPrivateKey(bytes); err == nil {
		return key, nil
	}
	return nil, errors.New("failed to parse private key")
}

// IsMatchPrivateKey is used to check the private key is match the public key in the certificate.
func IsMatchPrivateKey(cert *x509.Certificate, pri interface{}) bool {
	switch pub := cert.PublicKey.(type) {
	case *rsa.PublicKey:
		pri, ok := pri.(*rsa.PrivateKey)
		if !ok {
			return false
		}
		if pub.N.Cmp(pri.N) != 0 {
			return false
		}
	case *ecdsa.PublicKey:
		pri, ok := pri.(*ecdsa.PrivateKey)
		if !ok {
			return false
		}
		if pub.X.Cmp(pri.X) != 0 || pub.Y.Cmp(pri.Y) != 0 {
			return false
		}
	case ed25519.PublicKey:
		pri, ok := pri.(ed25519.PrivateKey)
		if !ok {
			return false
		}
		if !bytes.Equal(pri.Public().(ed25519.PublicKey), pub) {
			return false
		}
	default:
		return false
	}
	return true
}
