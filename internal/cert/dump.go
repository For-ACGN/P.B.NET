package cert

import (
	"bytes"
	"crypto/x509"
	"fmt"
	"io"
	"os"
	"strings"

	"project/internal/convert"
)

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
		cert.PublicKeyAlgorithm, // TODO print public key
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
		const format = "\nDNS names: [%s]"
		n, err = fmt.Fprintf(w, format, strings.Join(cert.DNSNames, ", "))
		num += n
		if err != nil {
			return num, err
		}
	}
	if len(cert.IPAddresses) != 0 {
		const format = "\nIP addresses: [%s]"
		ip := make([]string, len(cert.IPAddresses))
		for i := 0; i < len(cert.IPAddresses); i++ {
			ip[i] = cert.IPAddresses[i].String()
		}
		n, err = fmt.Fprintf(w, format, strings.Join(ip, ", "))
		num += n
		if err != nil {
			return num, err
		}
	}
	if len(cert.EmailAddresses) != 0 {
		const format = "\nEmail addresses: [%s]"
		n, err = fmt.Fprintf(w, format, strings.Join(cert.EmailAddresses, ", "))
		num += n
		if err != nil {
			return num, err
		}
	}
	if len(cert.URIs) != 0 {
		const format = "\nURLs: [%s]"
		urls := make([]string, len(cert.URIs))
		for i := 0; i < len(cert.URIs); i++ {
			urls[i] = cert.URIs[i].String()
		}
		n, err = fmt.Fprintf(w, format, strings.Join(urls, ", "))
		num += n
		if err != nil {
			return num, err
		}
	}
	return num, nil
}
