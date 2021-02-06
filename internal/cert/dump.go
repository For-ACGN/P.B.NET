package cert

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"project/internal/convert"
)

const timeLayout = "2006-01-02 15:04:05 Z07:00"

const dumpTemplate = `
[Basic]
  Version: %d
  Is CA: %t
  Key usage: %d

[Subject key ID]
%s
  
[Authority key ID]
%s

[Serial number]
%s

[Subject]
  Common name:  %s
  Organization: %s

[Issuer]
  Common name:  %s
  Organization: %s

[Public key]
  algo: %s
  size: %s bits
  data: %s

[Signature]
  algo: %s
  size: %d bits
  data: %s

[Valid time]
  Not before: %s
  Not after:  %s
`

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
	pub, pubSize, err := dumpPublicKey(cert.PublicKey)
	if err != nil {
		_, _ = w.Write([]byte("[error]: " + err.Error()))
		return 0, err
	}
	subjectKeyID := "  [nil]"
	if len(cert.SubjectKeyId) > 0 {
		subjectKeyID = convert.SdumpBytesWithPL(cert.SubjectKeyId, "  ", 8)
		subjectKeyID = strings.TrimSuffix(subjectKeyID, ",")
	}
	authorityKeyID := "  [nil]"
	if len(cert.AuthorityKeyId) > 0 {
		authorityKeyID = convert.SdumpBytesWithPL(cert.AuthorityKeyId, "  ", 8)
		authorityKeyID = strings.TrimSuffix(authorityKeyID, ",")
	}
	serialNum := convert.SdumpBytesWithPL(cert.SerialNumber.Bytes(), "  ", 8)
	serialNum = strings.TrimSuffix(serialNum, ",")
	subjectOrg := strings.Join(cert.Subject.Organization, ", ")
	issuerOrg := strings.Join(cert.Issuer.Organization, ", ")
	prefix := strings.Repeat(" ", len("  data: "))
	publicKey := convert.SdumpBytesWithPL(pub, prefix, 8)
	publicKey = convert.RemoveFirstPrefix(publicKey, prefix)
	publicKey = strings.TrimSuffix(publicKey, ",")
	signature := convert.SdumpBytesWithPL(cert.Signature, prefix, 8)
	signature = convert.RemoveFirstPrefix(signature, prefix)
	signature = strings.TrimSuffix(signature, ",")
	var num int
	n, err := fmt.Fprintf(w, dumpTemplate[1:],
		cert.Version, cert.IsCA, cert.KeyUsage,
		subjectKeyID, authorityKeyID, serialNum,
		cert.Subject.CommonName, subjectOrg,
		cert.Issuer.CommonName, issuerOrg,
		cert.PublicKeyAlgorithm, pubSize, publicKey,
		cert.SignatureAlgorithm, len(cert.Signature)*8, signature,
		cert.NotBefore.Local().Format(timeLayout),
		cert.NotAfter.Local().Format(timeLayout),
	)
	num += n
	if err != nil {
		return num, err
	}
	n, err = dumpAlternate(w, cert)
	num += n
	return num, err
}

func dumpAlternate(w io.Writer, cert *x509.Certificate) (int, error) {
	maxPaddingLen := calcMaxPaddingLen(cert)
	if maxPaddingLen == 0 {
		return 0, nil
	}
	var num int
	n, err := fmt.Fprint(w, "\n[Alternate]")
	num += n
	if err != nil {
		return num, err
	}
	if len(cert.DNSNames) > 0 {
		const format = "\n  DNS names: %s[%s]"
		padding := strings.Repeat(" ", maxPaddingLen-len("DNS names"))
		n, err = fmt.Fprintf(w, format, padding, strings.Join(cert.DNSNames, ", "))
		num += n
		if err != nil {
			return num, err
		}
	}
	if len(cert.IPAddresses) > 0 {
		const format = "\n  IP addresses: %s[%s]"
		padding := strings.Repeat(" ", maxPaddingLen-len("IP addresses"))
		ip := make([]string, len(cert.IPAddresses))
		for i := 0; i < len(cert.IPAddresses); i++ {
			ip[i] = cert.IPAddresses[i].String()
		}
		n, err = fmt.Fprintf(w, format, padding, strings.Join(ip, ", "))
		num += n
		if err != nil {
			return num, err
		}
	}
	if len(cert.EmailAddresses) > 0 {
		const format = "\n  Email addresses: %s[%s]"
		padding := strings.Repeat(" ", maxPaddingLen-len("Email addresses"))
		n, err = fmt.Fprintf(w, format, padding, strings.Join(cert.EmailAddresses, ", "))
		num += n
		if err != nil {
			return num, err
		}
	}
	if len(cert.URIs) > 0 {
		const format = "\n  URIs: %s[%s]"
		padding := strings.Repeat(" ", maxPaddingLen-len("URIs"))
		urls := make([]string, len(cert.URIs))
		for i := 0; i < len(cert.URIs); i++ {
			urls[i] = cert.URIs[i].String()
		}
		n, err = fmt.Fprintf(w, format, padding, strings.Join(urls, ", "))
		num += n
		if err != nil {
			return num, err
		}
	}
	return num, err
}

// dumpPublicKey is used to dump a part information about public key.
func dumpPublicKey(publicKey interface{}) ([]byte, string, error) {
	switch pub := publicKey.(type) {
	case *rsa.PublicKey:
		size := pub.Size() * 8
		return pub.N.Bytes(), strconv.Itoa(size), nil
	case *ecdsa.PublicKey:
		size := pub.Curve.Params().BitSize
		return pub.X.Bytes(), strconv.Itoa(size), nil
	case ed25519.PublicKey:
		return pub, "256", nil
	default:
		return nil, "", errors.Errorf("unsupported public key: %T", pub)
	}
}

func calcMaxPaddingLen(cert *x509.Certificate) int {
	var max int
	if len(cert.DNSNames) > 0 {
		max = len("DNS names")
	}
	if len(cert.IPAddresses) > 0 {
		l := len("IP addresses")
		if l > max {
			max = l
		}
	}
	if len(cert.EmailAddresses) > 0 {
		l := len("Email addresses")
		if l > max {
			max = l
		}
	}
	if len(cert.URIs) > 0 {
		l := len("URIs")
		if l > max {
			max = l
		}
	}
	return max
}
