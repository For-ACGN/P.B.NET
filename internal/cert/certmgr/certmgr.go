package certmgr

import (
	"bytes"
	"compress/flate"
	"crypto/sha256"
	"crypto/subtle"

	"github.com/pkg/errors"
	"golang.org/x/crypto/pbkdf2"

	"project/internal/cert"
	"project/internal/convert"
	"project/internal/crypto/aes"
	"project/internal/crypto/hmac"
	"project/internal/patch/msgpack"
	"project/internal/random"
	"project/internal/security"
)

// -----------------------------certificate pool file format-----------------------------
//
// +-------------+----------+------------+--------------+----------------+--------------+
// | HMAC-SHA256 |    IV    |   random   | size(uint32) | cert pool data |    random    |
// +-------------+----------+------------+--------------+----------------+--------------+
// |  32 bytes   | 16 bytes | 2018 bytes |   4 bytes    |   var bytes    | > 1127 bytes |
// +-------------+----------+------------+--------------+----------------+--------------+
//
// Hash is used to verify the integrality of the file.
// Hash value is hmac-sha256(random + size + cert pool data + random)
// Use flate to compress(random + size + data + random)
// Use AES-CTR to encrypt compressed data

const (
	random2018 = 2018
	random1127 = 1127
)

// ctrlCertPool include bytes about certificates and private keys.
// Controller and tool/certificate/manager will use it.
type ctrlCertPool struct {
	PublicRootCACerts   [][]byte `msgpack:"a"`
	PublicClientCACerts [][]byte `msgpack:"b"`
	PublicClientPairs   []struct {
		Cert []byte `msgpack:"a"`
		Key  []byte `msgpack:"b"`
	} `msgpack:"c"`
	PrivateRootCAPairs []struct {
		Cert []byte `msgpack:"a"`
		Key  []byte `msgpack:"b"`
	} `msgpack:"d"`
	PrivateClientCAPairs []struct {
		Cert []byte `msgpack:"a"`
		Key  []byte `msgpack:"b"`
	} `msgpack:"e"`
	PrivateClientPairs []struct {
		Cert []byte `msgpack:"a"`
		Key  []byte `msgpack:"b"`
	} `msgpack:"f"`
}

// Load is used to load certificates from certificate pool.
func (cp *ctrlCertPool) Load(pool *cert.Pool) {
	pubRootCACerts := pool.GetPublicRootCACerts()
	for i := 0; i < len(pubRootCACerts); i++ {
		cp.PublicRootCACerts = append(cp.PublicRootCACerts, pubRootCACerts[i].Raw)
	}
	pubClientCACerts := pool.GetPublicClientCACerts()
	for i := 0; i < len(pubClientCACerts); i++ {
		cp.PublicClientCACerts = append(cp.PublicClientCACerts, pubClientCACerts[i].Raw)
	}
	pubClientPairs := pool.GetPublicClientPairs()
	for i := 0; i < len(pubClientPairs); i++ {
		c, k := pubClientPairs[i].Encode()
		cp.PublicClientPairs = append(cp.PublicClientPairs, struct {
			Cert []byte `msgpack:"a"`
			Key  []byte `msgpack:"b"`
		}{Cert: c, Key: k})
	}
	priRootCAPairs := pool.GetPrivateRootCAPairs()
	for i := 0; i < len(priRootCAPairs); i++ {
		c, k := priRootCAPairs[i].Encode()
		cp.PrivateRootCAPairs = append(cp.PrivateRootCAPairs, struct {
			Cert []byte `msgpack:"a"`
			Key  []byte `msgpack:"b"`
		}{Cert: c, Key: k})
	}
	priClientCAPairs := pool.GetPrivateClientCAPairs()
	for i := 0; i < len(priClientCAPairs); i++ {
		c, k := priClientCAPairs[i].Encode()
		cp.PrivateClientCAPairs = append(cp.PrivateClientCAPairs, struct {
			Cert []byte `msgpack:"a"`
			Key  []byte `msgpack:"b"`
		}{Cert: c, Key: k})
	}
	priClientPairs := pool.GetPrivateClientPairs()
	for i := 0; i < len(priClientPairs); i++ {
		c, k := priClientPairs[i].Encode()
		cp.PrivateClientPairs = append(cp.PrivateClientPairs, struct {
			Cert []byte `msgpack:"a"`
			Key  []byte `msgpack:"b"`
		}{Cert: c, Key: k})
	}
}

// Dump is used to dump certificates to the certificate pool.
func (cp *ctrlCertPool) Dump(pool *cert.Pool) error {
	memory := security.NewMemory()
	defer memory.Flush()

	var err error
	for i := 0; i < len(cp.PublicRootCACerts); i++ {
		memory.Padding()
		err = pool.AddPublicRootCACert(cp.PublicRootCACerts[i])
		if err != nil {
			return err
		}
	}
	for i := 0; i < len(cp.PublicClientCACerts); i++ {
		memory.Padding()
		err = pool.AddPublicClientCACert(cp.PublicClientCACerts[i])
		if err != nil {
			return err
		}
	}
	for i := 0; i < len(cp.PublicClientPairs); i++ {
		memory.Padding()
		pair := cp.PublicClientPairs[i]
		err = pool.AddPublicClientPair(pair.Cert, pair.Key)
		if err != nil {
			return err
		}
	}
	for i := 0; i < len(cp.PrivateRootCAPairs); i++ {
		memory.Padding()
		pair := cp.PrivateRootCAPairs[i]
		err = pool.AddPrivateRootCAPair(pair.Cert, pair.Key)
		if err != nil {
			return err
		}
	}
	for i := 0; i < len(cp.PrivateClientCAPairs); i++ {
		memory.Padding()
		pair := cp.PrivateClientCAPairs[i]
		err = pool.AddPrivateClientCAPair(pair.Cert, pair.Key)
		if err != nil {
			return err
		}
	}
	for i := 0; i < len(cp.PrivateClientPairs); i++ {
		memory.Padding()
		pair := cp.PrivateClientPairs[i]
		err = pool.AddPrivateClientPair(pair.Cert, pair.Key)
		if err != nil {
			return err
		}
	}
	return nil
}

// Clean is used to clean all data in this certificate pool.
func (cp *ctrlCertPool) Clean() {
	for i := 0; i < len(cp.PublicRootCACerts); i++ {
		security.CoverBytes(cp.PublicRootCACerts[i])
	}
	for i := 0; i < len(cp.PublicClientCACerts); i++ {
		security.CoverBytes(cp.PublicClientCACerts[i])
	}
	for i := 0; i < len(cp.PublicClientPairs); i++ {
		pair := cp.PublicClientPairs[i]
		security.CoverBytes(pair.Cert)
		security.CoverBytes(pair.Key)
	}
	for i := 0; i < len(cp.PrivateRootCAPairs); i++ {
		pair := cp.PrivateRootCAPairs[i]
		security.CoverBytes(pair.Cert)
		security.CoverBytes(pair.Key)
	}
	for i := 0; i < len(cp.PrivateClientCAPairs); i++ {
		pair := cp.PrivateClientCAPairs[i]
		security.CoverBytes(pair.Cert)
		security.CoverBytes(pair.Key)
	}
	for i := 0; i < len(cp.PrivateClientPairs); i++ {
		pair := cp.PrivateClientPairs[i]
		security.CoverBytes(pair.Cert)
		security.CoverBytes(pair.Key)
	}
}

// calculateAESKey is used to generate aes key for encrypt certificate pool.
func calculateAESKey(password []byte) []byte {
	hash := sha256.New()
	hash.Write(password)
	hash.Write([]byte{0x20, 0x17, 0x04, 0x17})
	digest := hash.Sum(nil)
	return pbkdf2.Key(digest, digest[:16], 8192, aes.Key256Bit, sha256.New)
}

// SaveCtrlCertPool is used to compress and encrypt certificate pool.
func SaveCtrlCertPool(pool *cert.Pool, password []byte) ([]byte, error) {
	certPool := ctrlCertPool{}
	certPool.Load(pool)
	defer certPool.Clean()
	// marshal certificate pool data
	certData, err := msgpack.Marshal(certPool)
	if err != nil {
		return nil, err
	}
	defer security.CoverBytes(certData)
	certPool.Clean()
	// make certificate pool file
	certPoolLen := len(certData)
	bufSize := random2018 + convert.Uint32Size + certPoolLen + random1127
	buf := bytes.NewBuffer(make([]byte, 0, bufSize))
	defer security.CoverBytes(buf.Bytes())
	// write all data
	buf.Write(random.Bytes(random2018))                     // random data 1
	buf.Write(convert.BEUint32ToBytes(uint32(certPoolLen))) // cert pool data size
	buf.Write(certData)                                     // cert pool data
	buf.Write(random.Bytes(random1127 + random.Intn(1024))) // random data 2
	// cover cert pool data at once
	security.CoverBytes(certData)
	// compress cert pool data
	flateBuf := bytes.NewBuffer(make([]byte, 0, buf.Len()/2))
	defer security.CoverBytes(flateBuf.Bytes())
	writer, err := flate.NewWriter(flateBuf, flate.BestCompression)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create deflate writer")
	}
	_, err = buf.WriteTo(writer)
	if err != nil {
		return nil, errors.Wrap(err, "failed to compress certificate pool data")
	}
	err = writer.Close()
	if err != nil {
		return nil, errors.Wrap(err, "failed to close deflate writer")
	}
	// encrypt compressed data
	aesKey := calculateAESKey(password)
	defer security.CoverBytes(aesKey)
	output, err := aes.CTREncrypt(flateBuf.Bytes(), aesKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt certificate pool data")
	}
	// write hash
	hash := hmac.New(sha256.New, aesKey)
	hash.Write(output)
	digest := hash.Sum(nil)
	return append(digest, output...), nil
}

// LoadCtrlCertPool is used to decrypt and decompress certificate pool.
func LoadCtrlCertPool(pool *cert.Pool, certPool, password []byte) error {
	if len(certPool) < sha256.Size+aes.BlockSize {
		return errors.New("invalid certificate pool file size")
	}
	memory := security.NewMemory()
	defer memory.Flush()
	// decrypt certificate pool file
	aesKey, aesIV := calculateAESKey(password)
	defer func() {
		security.CoverBytes(aesKey)
		security.CoverBytes(aesIV)
	}()
	compressed, err := aes.CBCDecrypt(certPool[sha256.Size:], aesKey)
	if err != nil {
		return errors.Wrap(err, "failed to decrypt certificate pool file")
	}
	defer security.CoverBytes(compressed)
	// decompress
	buf := bytes.NewBuffer(make([]byte, 0, len(compressed)*2))
	reader := flate.NewReader(bytes.NewReader(compressed))
	_, err = buf.ReadFrom(reader)
	if err != nil {
		return errors.Wrap(err, "failed to decompress certificate pool file")
	}
	err = reader.Close()
	if err != nil {
		return errors.Wrap(err, "failed to close deflate reader")
	}
	file := buf.Bytes()
	// compare file hash
	fileHash := sha256.Sum256(file)
	if subtle.ConstantTimeCompare(certPool[:sha256.Size], fileHash[:]) != 1 {
		return errors.New("incorrect password or certificate pool has been tampered")
	}
	memory.Padding()
	offset := random2018
	size := int(convert.BEBytesToUint32(file[offset : offset+4]))
	memory.Padding()
	offset += 4
	// unmarshal
	cp := ctrlCertPool{}
	err = msgpack.Unmarshal(file[offset:offset+size], &cp)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal certificate pool")
	}
	memory.Padding()
	return cp.Dump(pool)
}

// CertPool contains raw certificates, it used for Node and Beacon configuration.
type CertPool struct {
	PublicRootCACerts   [][]byte `msgpack:"a"`
	PublicClientCACerts [][]byte `msgpack:"b"`
	PublicClientPairs   []struct {
		Cert []byte `msgpack:"a"`
		Key  []byte `msgpack:"b"`
	} `msgpack:"c"`
	PrivateRootCACerts   [][]byte `msgpack:"d"`
	PrivateClientCACerts [][]byte `msgpack:"e"`
	PrivateClientPairs   []struct {
		Cert []byte `msgpack:"a"`
		Key  []byte `msgpack:"b"`
	} `msgpack:"f"`
}

// Load is used to load certificates from certificate pool or other pool,
// Controller or tests will add certificates to CertPool.
func (cp *CertPool) Load(pool *cert.Pool) {
	pubRootCACerts := pool.GetPublicRootCACerts()
	for i := 0; i < len(pubRootCACerts); i++ {
		cp.PublicRootCACerts = append(cp.PublicRootCACerts, pubRootCACerts[i].Raw)
	}
	pubClientCACerts := pool.GetPublicClientCACerts()
	for i := 0; i < len(pubClientCACerts); i++ {
		cp.PublicClientCACerts = append(cp.PublicClientCACerts, pubClientCACerts[i].Raw)
	}
	pubClientPairs := pool.GetPublicClientPairs()
	for i := 0; i < len(pubClientPairs); i++ {
		c, k := pubClientPairs[i].Encode()
		cp.PublicClientPairs = append(cp.PublicClientPairs, struct {
			Cert []byte `msgpack:"a"`
			Key  []byte `msgpack:"b"`
		}{Cert: c, Key: k})
	}
	priRootCACerts := pool.GetPrivateRootCACerts()
	for i := 0; i < len(priRootCACerts); i++ {
		cp.PrivateRootCACerts = append(cp.PrivateRootCACerts, priRootCACerts[i].Raw)
	}
	priClientCACerts := pool.GetPrivateClientCACerts()
	for i := 0; i < len(priClientCACerts); i++ {
		cp.PrivateClientCACerts = append(cp.PrivateClientCACerts, priClientCACerts[i].Raw)
	}
	priClientPairs := pool.GetPrivateClientPairs()
	for i := 0; i < len(priClientPairs); i++ {
		c, k := priClientPairs[i].Encode()
		cp.PrivateClientPairs = append(cp.PrivateClientPairs, struct {
			Cert []byte `msgpack:"a"`
			Key  []byte `msgpack:"b"`
		}{Cert: c, Key: k})
	}
}

// ToPool is used to create a certificate pool. Call Clean to cover bytes in pool.
func (cp *CertPool) ToPool() (*cert.Pool, error) {
	memory := security.NewMemory()
	defer memory.Flush()

	pool := cert.NewPool()
	for i := 0; i < len(cp.PublicRootCACerts); i++ {
		memory.Padding()
		err := pool.AddPublicRootCACert(cp.PublicRootCACerts[i])
		if err != nil {
			return nil, err
		}
	}
	for i := 0; i < len(cp.PublicClientCACerts); i++ {
		memory.Padding()
		err := pool.AddPublicClientCACert(cp.PublicClientCACerts[i])
		if err != nil {
			return nil, err
		}
	}
	for i := 0; i < len(cp.PublicClientPairs); i++ {
		memory.Padding()
		pair := cp.PublicClientPairs[i]
		err := pool.AddPublicClientPair(pair.Cert, pair.Key)
		if err != nil {
			return nil, err
		}
	}
	for i := 0; i < len(cp.PrivateRootCACerts); i++ {
		memory.Padding()
		err := pool.AddPrivateRootCACert(cp.PrivateRootCACerts[i])
		if err != nil {
			return nil, err
		}
	}
	for i := 0; i < len(cp.PrivateClientCACerts); i++ {
		memory.Padding()
		err := pool.AddPrivateClientCACert(cp.PrivateClientCACerts[i])
		if err != nil {
			return nil, err
		}
	}
	for i := 0; i < len(cp.PrivateClientPairs); i++ {
		memory.Padding()
		pair := cp.PrivateClientPairs[i]
		err := pool.AddPrivateClientPair(pair.Cert, pair.Key)
		if err != nil {
			return nil, err
		}
	}
	return pool, nil
}

// Clean is used to clean all data in this certificate pool.
func (cp *CertPool) Clean() {
	for i := 0; i < len(cp.PublicRootCACerts); i++ {
		security.CoverBytes(cp.PublicRootCACerts[i])
	}
	for i := 0; i < len(cp.PublicClientCACerts); i++ {
		security.CoverBytes(cp.PublicClientCACerts[i])
	}
	for i := 0; i < len(cp.PublicClientPairs); i++ {
		pair := cp.PublicClientPairs[i]
		security.CoverBytes(pair.Cert)
		security.CoverBytes(pair.Key)
	}
	for i := 0; i < len(cp.PrivateRootCACerts); i++ {
		security.CoverBytes(cp.PrivateRootCACerts[i])
	}
	for i := 0; i < len(cp.PrivateClientCACerts); i++ {
		security.CoverBytes(cp.PrivateClientCACerts[i])
	}
	for i := 0; i < len(cp.PrivateClientPairs); i++ {
		pair := cp.PrivateClientPairs[i]
		security.CoverBytes(pair.Cert)
		security.CoverBytes(pair.Key)
	}
}
