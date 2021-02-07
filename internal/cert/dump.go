package cert

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
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

[Key usage]
  %s

[Serial number]
%s

[Subject key ID]
%s
  
[Authority key ID]
%s

[Subject]
%s

[Issuer]
%s

[Public key]
  Algo: %s
  Size: %s bits
  Data: %s

[Signature]
  Algo: %s
  Size: %d bits
  Data: %s

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
	str := dumpHexBytes(cert.SerialNumber.Bytes())
	serialNum := convert.SdumpStringWithPL(str, "  ", 48)
	if serialNum == "" {
		serialNum = "  [nil]"
	}
	subjectKeyID := "  [nil]"
	if len(cert.SubjectKeyId) > 0 {
		str = dumpHexBytes(cert.SubjectKeyId)
		subjectKeyID = convert.SdumpStringWithPL(str, "  ", 48)
	}
	authorityKeyID := "  [nil]"
	if len(cert.AuthorityKeyId) > 0 {
		str = dumpHexBytes(cert.AuthorityKeyId)
		authorityKeyID = convert.SdumpStringWithPL(str, "  ", 48)
	}
	prefix := strings.Repeat(" ", len("  data: "))
	str = dumpHexBytes(pub)
	publicKey := convert.SdumpStringWithPL(str, prefix, 48)
	publicKey = convert.RemoveFirstPrefix(publicKey, prefix)
	str = dumpHexBytes(cert.Signature)
	signature := convert.SdumpStringWithPL(str, prefix, 48)
	signature = convert.RemoveFirstPrefix(signature, prefix)
	var num int
	n, err := fmt.Fprintf(w, dumpTemplate[1:],
		cert.Version, cert.IsCA, dumpKeyUsage(cert.KeyUsage),
		serialNum, subjectKeyID, authorityKeyID,
		dumpPKIXName(cert.Subject), dumpPKIXName(cert.Issuer),
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

// "0x89, 0x5E, 0x75," -> "89:5E:75"
func dumpHexBytes(b []byte) string {
	l := len(b)
	if l == 0 {
		return ""
	}
	builder := strings.Builder{}
	builder.Grow(l*3 - 1)
	buf := make([]byte, 2)
	hex.Encode(buf, []byte{b[0]})
	buf = bytes.ToUpper(buf)
	builder.Write(buf)
	for i := 1; i < l; i++ {
		hex.Encode(buf, []byte{b[i]})
		buf = bytes.ToUpper(buf)
		builder.WriteString(":")
		builder.Write(buf)
	}
	return builder.String()
}

var keyUsage = map[x509.KeyUsage]string{
	x509.KeyUsageDigitalSignature:  "Digital Signature",
	x509.KeyUsageContentCommitment: "Content Commitment",
	x509.KeyUsageKeyEncipherment:   "Key Encipherment",
	x509.KeyUsageDataEncipherment:  "Data Encipherment",
	x509.KeyUsageKeyAgreement:      "Key Agreement",
	x509.KeyUsageCertSign:          "Certificate Signing",
	x509.KeyUsageCRLSign:           "CRL Signing",
	x509.KeyUsageEncipherOnly:      "Encipher Only",
	x509.KeyUsageDecipherOnly:      "Decipher Only",
}

func dumpKeyUsage(usage x509.KeyUsage) string {
	if usage == 0 {
		return "[nil]"
	}
	var usages []string
	for i := 0; i < len(keyUsage); i++ {
		ku := x509.KeyUsage(1 << i)
		if (usage & (ku)) != 0 {
			usages = append(usages, keyUsage[ku])
		}
	}
	return strings.Join(usages, "\n  ")
}

func dumpPKIXName(name pkix.Name) string {
	const format = `
  Common name:    %s
  Organization:   %s
  Unit:           %s
  Country:        %s
  Locality:       %s
  Province:       %s
  Street address: %s
  Postal code:    %s
  Serial number:  %s`
	const sep = "\n                  "
	builder := strings.Builder{}
	builder.Grow(256)
	_, _ = fmt.Fprintf(&builder, format[1:],
		name.CommonName,
		strings.Join(name.Organization, sep),
		strings.Join(name.OrganizationalUnit, sep),
		strings.Join(name.Country, sep),
		strings.Join(name.Locality, sep),
		strings.Join(name.Province, sep),
		strings.Join(name.StreetAddress, sep),
		strings.Join(name.PostalCode, sep),
		name.SerialNumber,
	)
	return builder.String()
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
