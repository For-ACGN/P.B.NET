package cert

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"project/internal/namer"
	"project/internal/patch/monkey"
	"project/internal/patch/toml"
	"project/internal/testsuite"
	"project/internal/testsuite/testnamer"
)

func TestGenerateCertificate(t *testing.T) {
	t.Run("with namer", func(t *testing.T) {
		opts := Options{
			Namer: testnamer.Namer(),
		}
		cert, err := generateCertificate(&opts)
		require.NoError(t, err)
		t.Log("common name:", cert.Subject.CommonName)
		t.Log("organization:", cert.Subject.Organization[0])
	})

	t.Run("failed to generate word", func(t *testing.T) {
		opts := Options{
			Namer: testnamer.WithGenerateFailed(),
		}
		cert, err := generateCertificate(&opts)
		require.Error(t, err)
		require.Nil(t, cert)
	})

	t.Run("failed to set organization", func(t *testing.T) {
		n := testnamer.Namer()
		count := 0
		patch := func(interface{}, *namer.Options) (string, error) {
			count++
			if count == 1 {
				return "foo", nil
			}
			return "", monkey.Error
		}
		pg := monkey.PatchInstanceMethod(n, "Generate", patch)
		defer pg.Unpatch()

		opts := Options{
			Namer: n,
		}
		cert, err := generateCertificate(&opts)
		monkey.IsExistMonkeyError(t, err)
		require.Nil(t, cert)
	})

	t.Run("invalid domain name", func(t *testing.T) {
		opts := Options{
			DNSNames: []string{"foo domain name"},
		}
		cert, err := generateCertificate(&opts)
		require.Error(t, err)
		require.Nil(t, cert)
	})

	t.Run("invalid IP address", func(t *testing.T) {
		opts := Options{
			IPAddresses: []string{"foo ip"},
		}
		cert, err := generateCertificate(&opts)
		require.Error(t, err)
		require.Nil(t, cert)
	})

	t.Run("invalid email address", func(t *testing.T) {
		opts := Options{
			EmailAddresses: []string{"foo email"},
		}
		cert, err := generateCertificate(&opts)
		require.Error(t, err)
		require.Nil(t, cert)
	})

	t.Run("invalid url", func(t *testing.T) {
		opts := Options{
			URLs: []string{"foo url\n"},
		}
		cert, err := generateCertificate(&opts)
		require.Error(t, err)
		require.Nil(t, cert)
	})
}

func TestIsDomainName(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		for _, domain := range [...]string{
			"test.com",
			"Test-sub.com",
			"test-sub2.com",
		} {
			require.True(t, isDomainName(domain))
		}
	})

	t.Run("invalid", func(t *testing.T) {
		for _, domain := range [...]string{
			"",
			string([]byte{255, 254, 12, 35}),
			"test-",
			"Test.-",
			"test..",
			strings.Repeat("a", 64) + ".com",
		} {
			require.False(t, isDomainName(domain))
		}
	})
}

func TestGeneratePrivateKey(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		pri, pub, err := generatePrivateKey("")
		require.NoError(t, err)
		require.NotNil(t, pub)
		require.NotNil(t, pri)
	})

	t.Run("ed25519", func(t *testing.T) {
		pri, pub, err := generatePrivateKey("ed25519")
		require.NoError(t, err)
		require.NotNil(t, pub)
		require.NotNil(t, pri)
	})

	t.Run("rsa", func(t *testing.T) {
		pri, pub, err := generatePrivateKey("rsa|2048")
		require.NoError(t, err)
		require.NotNil(t, pub)
		require.NotNil(t, pri)
	})

	t.Run("ecdsa", func(t *testing.T) {
		pri, pub, err := generatePrivateKey("ecdsa|p256")
		require.NoError(t, err)
		require.NotNil(t, pub)
		require.NotNil(t, pri)
	})

	t.Run("invalid config", func(t *testing.T) {
		pri, pub, err := generatePrivateKey("ecdsa")
		require.Error(t, err)
		require.Nil(t, pub)
		require.Nil(t, pri)
		t.Log(err)
	})

	t.Run("unknown algorithm", func(t *testing.T) {
		pri, pub, err := generatePrivateKey("foo|cfg")
		require.Error(t, err)
		require.Nil(t, pub)
		require.Nil(t, pri)
		t.Log(err)
	})
}

func TestGenerateRSA(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		pri, pub, err := generateRSA("2048")
		require.NoError(t, err)
		require.NotNil(t, pub)
		require.NotNil(t, pri)
	})

	t.Run("invalid bits", func(t *testing.T) {
		pri, pub, err := generateRSA("NaN")
		require.Error(t, err)
		require.Nil(t, pub)
		require.Nil(t, pri)
		t.Log(err)
	})

	t.Run("<1024", func(t *testing.T) {
		pri, pub, err := generateRSA("512")
		require.Error(t, err)
		require.Nil(t, pub)
		require.Nil(t, pri)
		t.Log(err)
	})

	t.Run("failed to generate", func(t *testing.T) {
		patch := func(io.Reader, int) (*rsa.PrivateKey, error) {
			return nil, monkey.Error
		}
		pg := monkey.Patch(rsa.GenerateKey, patch)
		defer pg.Unpatch()

		pri, pub, err := generateRSA("2048")
		monkey.IsMonkeyError(t, err)
		require.Nil(t, pub)
		require.Nil(t, pri)
	})
}

func TestGenerateECDSA(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		pri, pub, err := generateECDSA("p256")
		require.NoError(t, err)
		require.NotNil(t, pub)
		require.NotNil(t, pri)
	})

	for _, curve := range [...]string{
		"p224", "p256", "p384", "p521",
	} {
		t.Run(curve, func(t *testing.T) {
			pri, pub, err := generateECDSA(curve)
			require.NoError(t, err)
			require.NotNil(t, pub)
			require.NotNil(t, pri)
		})
	}

	t.Run("unsupported elliptic curve", func(t *testing.T) {
		pri, pub, err := generateECDSA("foo")
		require.Error(t, err)
		require.Nil(t, pub)
		require.Nil(t, pri)
		t.Log(err)
	})

	t.Run("failed to generate", func(t *testing.T) {
		patch := func(elliptic.Curve, io.Reader) (*ecdsa.PrivateKey, error) {
			return nil, monkey.Error
		}
		pg := monkey.Patch(ecdsa.GenerateKey, patch)
		defer pg.Unpatch()

		pri, pub, err := generateECDSA("p256")
		monkey.IsMonkeyError(t, err)
		require.Nil(t, pub)
		require.Nil(t, pri)
	})
}

func TestGenerateEd25519(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		pri, pub, err := generateEd25519()
		require.NoError(t, err)
		require.NotNil(t, pub)
		require.NotNil(t, pri)
	})

	t.Run("failed to generate", func(t *testing.T) {
		patch := func(io.Reader) (ed25519.PublicKey, ed25519.PrivateKey, error) {
			return nil, nil, monkey.Error
		}
		pg := monkey.Patch(ed25519.GenerateKey, patch)
		defer pg.Unpatch()

		pri, pub, err := generateEd25519()
		monkey.IsMonkeyError(t, err)
		require.Nil(t, pub)
		require.Nil(t, pri)
	})
}

func TestGenerateCA(t *testing.T) {
	t.Run("compare", func(t *testing.T) {
		notBefore := time.Now()
		notAfter := notBefore.AddDate(0, 0, 1)
		opts := &Options{
			Algorithm: "rsa|2048",
			NotBefore: notBefore,
			NotAfter:  notAfter,
		}
		opts.Subject.CommonName = "test common name"
		opts.Subject.Organization = []string{"test organization"}

		ca, err := GenerateCA(opts)
		require.NoError(t, err)

		require.Equal(t, "test common name", ca.Certificate.Subject.CommonName)
		require.Equal(t, []string{"test organization"}, ca.Certificate.Subject.Organization)

		expected := notBefore.Format(timeLayout)
		actual := ca.Certificate.NotBefore.Local().Format(timeLayout)
		require.Equal(t, expected, actual)

		expected = notAfter.Format(timeLayout)
		actual = ca.Certificate.NotAfter.Local().Format(timeLayout)
		require.Equal(t, expected, actual)
	})

	t.Run("failed to generate CA", func(t *testing.T) {
		opts := Options{
			DNSNames: []string{"foo-"},
		}
		pair, err := GenerateCA(&opts)
		require.Error(t, err)
		require.Nil(t, pair)
	})

	t.Run("failed to generate private key", func(t *testing.T) {
		opts := Options{
			Algorithm: "foo",
		}
		_, err := GenerateCA(&opts)
		require.Error(t, err)
	})

	t.Run("failed to create certificate", func(t *testing.T) {
		patch := func(_ io.Reader, _, _ *x509.Certificate, _, _ interface{}) ([]byte, error) {
			return nil, monkey.Error
		}
		pg := monkey.Patch(x509.CreateCertificate, patch)
		defer pg.Unpatch()

		_, err := GenerateCA(nil)
		monkey.IsMonkeyError(t, err)
	})
}

func TestGenerate(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	for _, alg := range [...]string{
		"rsa|2048", "ecdsa|p256", "ed25519",
	} {
		t.Run(alg, func(t *testing.T) {
			opts := Options{Algorithm: alg}
			ca, err := GenerateCA(&opts)
			require.NoError(t, err)
			testGenerate(t, ca)  // CA sign
			testGenerate(t, nil) // self sign
		})
	}

	t.Run("invalid domain name", func(t *testing.T) {
		opts := Options{
			DNSNames: []string{"foo-"},
		}
		_, err := Generate(nil, nil, &opts)
		require.Error(t, err)
	})

	t.Run("failed to generate private key", func(t *testing.T) {
		opts := Options{
			Algorithm: "foo",
		}
		_, err := Generate(new(x509.Certificate), "foo", &opts)
		require.Error(t, err)
	})

	t.Run("invalid private key", func(t *testing.T) {
		_, err := Generate(new(x509.Certificate), "foo", nil)
		require.Error(t, err)
	})
}

func testGenerate(t *testing.T, ca *Pair) {
	opts := &Options{
		Algorithm:   "rsa|2048",
		DNSNames:    []string{"localhost"},
		IPAddresses: []string{"127.0.0.1", "::1"},
	}
	var (
		pair *Pair
		err  error
	)
	if ca != nil {
		pair, err = Generate(ca.Certificate, ca.PrivateKey, opts)
		require.NoError(t, err)
	} else {
		pair, err = Generate(nil, nil, opts)
		require.NoError(t, err)
	}
	require.Equal(t, pair.Certificate.Raw, pair.ASN1())

	// handler
	respData := []byte("hello")
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(respData)
	})

	tlsCert := pair.TLSCertificate()
	// run https servers
	server1 := http.Server{
		Addr:      "localhost:0",
		Handler:   serveMux,
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{tlsCert}},
	}
	port1 := testsuite.RunHTTPServer(t, "tcp", &server1)
	defer func() { _ = server1.Close() }()
	// IPv4-only
	server2 := http.Server{
		Addr:      "127.0.0.1:0",
		Handler:   serveMux,
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{tlsCert}},
	}
	port2 := testsuite.RunHTTPServer(t, "tcp", &server2)
	defer func() { _ = server2.Close() }()
	// IPv6-only
	server3 := http.Server{
		Addr:      "[::1]:0",
		Handler:   serveMux,
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{tlsCert}},
	}
	port3 := testsuite.RunHTTPServer(t, "tcp", &server3)
	defer func() { _ = server3.Close() }()

	// client
	tlsConfig := tls.Config{RootCAs: x509.NewCertPool()}
	if ca != nil {
		tlsConfig.RootCAs.AddCert(ca.Certificate)
	} else {
		tlsConfig.RootCAs.AddCert(pair.Certificate)
	}
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tlsConfig,
		},
	}
	defer client.CloseIdleConnections()

	get := func(hostname, port string) {
		resp, err := client.Get(fmt.Sprintf("https://%s:%s/", hostname, port))
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, respData, b)
	}

	// test
	get("localhost", port1)
	get("127.0.0.1", port2)
	get("[::1]", port3)
}

func TestPair_Encode(t *testing.T) {
	ca, err := GenerateCA(nil)
	require.NoError(t, err)

	pair := &Pair{Certificate: ca.Certificate}

	defer testsuite.DeferForPanic(t)
	pair.Encode()
}

func TestPair_EncodeToPEM(t *testing.T) {
	ca, err := GenerateCA(nil)
	require.NoError(t, err)

	_, err = tls.X509KeyPair(ca.EncodeToPEM())
	require.NoError(t, err)
}

func TestDump(t *testing.T) {
	opts := Options{
		DNSNames:       []string{"test.com", "foo.com"},
		IPAddresses:    []string{"1.1.1.1", "1234::1234"},
		EmailAddresses: []string{"admin@test.com", "user@test.com"},
		URLs:           []string{"https://1.1.1.1/", "http://example.com/"},
	}
	opts.Subject.Organization = []string{"org a", "org b"}

	ca, err := GenerateCA(&opts)
	require.NoError(t, err)

	Dump(ca.Certificate)
}

func TestParseCertificate(t *testing.T) {
	certPEMBlock, err := os.ReadFile("testdata/certs.pem")
	require.NoError(t, err)
	cert, err := ParseCertificate(certPEMBlock)
	require.NoError(t, err)
	t.Log(cert.Issuer)

	// parse invalid PEM data
	_, err = ParseCertificate([]byte{0, 1, 2, 3})
	require.Equal(t, ErrInvalidPEMBlock, err)

	// invalid Type
	certPEMBlock = []byte(`
-----BEGIN INVALID TYPE-----
-----END INVALID TYPE-----
`)
	_, err = ParseCertificate(certPEMBlock)
	require.EqualError(t, err, "invalid PEM block type: INVALID TYPE")

	// invalid certificate data
	certPEMBlock = []byte(`
-----BEGIN CERTIFICATE-----
-----END CERTIFICATE-----
`)
	_, err = ParseCertificate(certPEMBlock)
	require.Error(t, err)
}

func TestParseCertificates(t *testing.T) {
	certPEMBlock, err := os.ReadFile("testdata/certs.pem")
	require.NoError(t, err)
	certs, err := ParseCertificates(certPEMBlock)
	require.NoError(t, err)
	t.Log(certs[0].Issuer)
	t.Log(certs[1].Issuer)

	// parse invalid PEM data
	_, err = ParseCertificates([]byte{0, 1, 2, 3})
	require.Equal(t, ErrInvalidPEMBlock, err)

	// invalid Type
	certPEMBlock = []byte(`
-----BEGIN INVALID TYPE-----
-----END INVALID TYPE-----
`)
	_, err = ParseCertificates(certPEMBlock)
	require.EqualError(t, err, "invalid PEM block type: INVALID TYPE")

	// invalid certificate data
	certPEMBlock = []byte(`
-----BEGIN CERTIFICATE-----
-----END CERTIFICATE-----
`)
	_, err = ParseCertificates(certPEMBlock)
	require.Error(t, err)
}

func TestParsePrivateKey(t *testing.T) {
	for _, file := range [...]string{
		"pkcs1.key", "pkcs8.key", "ecp.key",
	} {
		keyPEMBlock, err := os.ReadFile("testdata/" + file)
		require.NoError(t, err)
		_, err = ParsePrivateKey(keyPEMBlock)
		require.NoError(t, err)
	}

	// parse invalid PEM data
	_, err := ParsePrivateKey([]byte{0, 1, 2, 3})
	require.Equal(t, ErrInvalidPEMBlock, err)

	// invalid certificate data
	keyPEMBlock := []byte(`
-----BEGIN PRIVATE KEY-----
-----END PRIVATE KEY-----
`)
	_, err = ParsePrivateKey(keyPEMBlock)
	require.Error(t, err)
}

func TestParsePrivateKeys(t *testing.T) {
	keyPEMBlock, err := os.ReadFile("testdata/keys.pem")
	require.NoError(t, err)
	keys, err := ParsePrivateKeys(keyPEMBlock)
	require.NoError(t, err)
	require.Len(t, keys, 2)

	// parse invalid PEM data
	_, err = ParsePrivateKeys([]byte{0, 1, 2, 3})
	require.Equal(t, ErrInvalidPEMBlock, err)

	// invalid certificate data
	keyPEMBlock = []byte(`
-----BEGIN CERTIFICATE-----
-----END CERTIFICATE-----
`)
	_, err = ParsePrivateKeys(keyPEMBlock)
	require.Error(t, err)
}

func TestMatch(t *testing.T) {
	cert := new(x509.Certificate)

	t.Run("rsa", func(t *testing.T) {
		t.Run("matched", func(t *testing.T) {
			pri, err := rsa.GenerateKey(rand.Reader, 2048)
			require.NoError(t, err)
			cert.PublicKey = &pri.PublicKey
			require.True(t, IsMatchPrivateKey(cert, pri))
		})

		t.Run("mismatch", func(t *testing.T) {
			pri, err := rsa.GenerateKey(rand.Reader, 2048)
			require.NoError(t, err)
			cert.PublicKey = &pri.PublicKey
			require.False(t, IsMatchPrivateKey(cert, nil))

			pri2, err := rsa.GenerateKey(rand.Reader, 2048)
			require.NoError(t, err)
			require.False(t, IsMatchPrivateKey(cert, pri2))
		})
	})

	t.Run("ecdsa", func(t *testing.T) {
		t.Run("matched", func(t *testing.T) {
			pri, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			require.NoError(t, err)
			cert.PublicKey = &pri.PublicKey
			require.True(t, IsMatchPrivateKey(cert, pri))
		})

		t.Run("mismatch", func(t *testing.T) {
			pri, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			require.NoError(t, err)
			cert.PublicKey = &pri.PublicKey
			require.False(t, IsMatchPrivateKey(cert, nil))

			pri2, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			require.NoError(t, err)
			require.False(t, IsMatchPrivateKey(cert, pri2))
		})
	})

	t.Run("ed25519", func(t *testing.T) {
		t.Run("matched", func(t *testing.T) {
			pub, pri, err := ed25519.GenerateKey(rand.Reader)
			require.NoError(t, err)
			cert.PublicKey = pub
			require.True(t, IsMatchPrivateKey(cert, pri))
		})

		t.Run("mismatched", func(t *testing.T) {
			pub, _, err := ed25519.GenerateKey(rand.Reader)
			require.NoError(t, err)
			cert.PublicKey = pub
			require.False(t, IsMatchPrivateKey(cert, nil))

			_, pri, err := ed25519.GenerateKey(rand.Reader)
			require.NoError(t, err)
			require.False(t, IsMatchPrivateKey(cert, pri))
		})
	})

	t.Run("unknown", func(t *testing.T) {
		cert.PublicKey = []byte{}
		require.False(t, IsMatchPrivateKey(cert, nil))
	})
}

func TestOptions(t *testing.T) {
	data, err := os.ReadFile("testdata/options.toml")
	require.NoError(t, err)

	// check unnecessary field
	opts := Options{}
	err = toml.Unmarshal(data, &opts)
	require.NoError(t, err)

	// check zero value
	testsuite.ContainZeroValue(t, opts)

	// check value is correct
	notBefore := time.Date(2018, 11, 27, 0, 0, 0, 0, time.Local)
	notAfter := time.Date(2028, 11, 27, 0, 0, 0, 0, time.Local)

	for _, testdata := range [...]*struct {
		expected interface{}
		actual   interface{}
	}{
		{expected: "rsa|2048", actual: opts.Algorithm},
		{expected: []string{"localhost", "lo"}, actual: opts.DNSNames},
		{expected: []string{"127.0.0.1", "::1"}, actual: opts.IPAddresses},
		{expected: []string{"admin@test.com", "user@test.com"}, actual: opts.EmailAddresses},
		{expected: []string{"https://1.1.1.1/", "http://example.com/"}, actual: opts.URLs},
		{expected: notBefore, actual: opts.NotBefore},
		{expected: notAfter, actual: opts.NotAfter},
		{expected: "name", actual: opts.Subject.CommonName},
		{expected: "test", actual: opts.Subject.SerialNumber},
		{expected: []string{"test1"}, actual: opts.Subject.Country},
		{expected: []string{"test2"}, actual: opts.Subject.Organization},
		{expected: []string{"test3"}, actual: opts.Subject.OrganizationalUnit},
		{expected: []string{"test4"}, actual: opts.Subject.Locality},
		{expected: []string{"test5"}, actual: opts.Subject.Province},
		{expected: []string{"test6"}, actual: opts.Subject.StreetAddress},
		{expected: []string{"test7"}, actual: opts.Subject.PostalCode},
		{expected: true, actual: opts.NamerOpts.DisablePrefix},
		{expected: true, actual: opts.NamerOpts.DisableStem},
		{expected: true, actual: opts.NamerOpts.DisableSuffix},
	} {
		require.Equal(t, testdata.expected, testdata.actual)
	}
}
