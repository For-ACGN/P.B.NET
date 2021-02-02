package certmgr

import (
	"bytes"
	"compress/flate"
	"crypto/sha256"

	"github.com/pkg/errors"
	"golang.org/x/crypto/pbkdf2"

	"project/internal/cert/certpool"
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
// cert pool data is msgpack.Marshal(ctrlCertPool{})
// Use flate to compress(random + size + cert pool data + random)
// Use AES-CTR to encrypt compressed data
// MAC value is hmac-sha256(IV + AES-CTR(compressed data))

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
func (cp *ctrlCertPool) Load(pool *certpool.Pool) {
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
func (cp *ctrlCertPool) Dump(pool *certpool.Pool) error {
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

// SaveCtrlCertPool is used to compress and encrypt certificate pool.
func SaveCtrlCertPool(pool *certpool.Pool, password []byte) ([]byte, error) {
	certPool := ctrlCertPool{}
	certPool.Load(pool)
	defer certPool.Clean()
	// marshal certificate pool data
	certPoolData, err := msgpack.Marshal(certPool)
	if err != nil {
		return nil, err
	}
	defer security.CoverBytes(certPoolData)
	certPool.Clean()
	// make certificate pool file
	certPoolDataLen := len(certPoolData)
	bufSize := random2018 + convert.Uint32Size + certPoolDataLen + random1127
	buf := bytes.NewBuffer(make([]byte, 0, bufSize))
	defer security.CoverBytes(buf.Bytes())
	// write all data
	buf.Write(random.Bytes(random2018))                         // random data 1
	buf.Write(convert.BEUint32ToBytes(uint32(certPoolDataLen))) // cert pool data size
	buf.Write(certPoolData)                                     // cert pool data
	buf.Write(random.Bytes(random1127 + random.Intn(1024)))     // random data 2
	// cover cert pool data at once
	security.CoverBytes(certPoolData)
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
	// write mac
	hash := hmac.New(sha256.New, aesKey)
	hash.Write(output)
	mac := hash.Sum(nil)
	return append(mac, output...), nil
}

// LoadCtrlCertPool is used to decrypt and decompress certificate pool.
func LoadCtrlCertPool(pool *certpool.Pool, data, password []byte) error {
	if len(data) < sha256.Size+aes.IVSize {
		return errors.New("invalid certificate pool file size")
	}
	memory := security.NewMemory()
	defer memory.Flush()
	// check certificate pool file mac
	aesKey := calculateAESKey(password)
	defer security.CoverBytes(aesKey)
	memory.Padding()
	hash := hmac.New(sha256.New, aesKey)
	hash.Write(data[sha256.Size:])
	mac := hash.Sum(nil)
	if !hmac.Equal(mac, data[:sha256.Size]) {
		return errors.New("incorrect password or certificate pool has been tampered")
	}
	// decrypt data
	compressed, err := aes.CTRDecrypt(data[sha256.Size:], aesKey)
	if err != nil {
		return errors.Wrap(err, "failed to decrypt certificate pool data")
	}
	defer security.CoverBytes(compressed)
	// decompress
	buf := bytes.NewBuffer(make([]byte, 0, len(compressed)*2))
	defer security.CoverBytes(buf.Bytes())
	reader := flate.NewReader(bytes.NewReader(compressed))
	_, err = buf.ReadFrom(reader)
	if err != nil {
		return errors.Wrap(err, "failed to decompress certificate pool data")
	}
	err = reader.Close()
	if err != nil {
		return errors.Wrap(err, "failed to close deflate reader")
	}
	// get cert pool data
	memory.Padding()
	buf.Next(random2018)
	size := int(convert.BEBytesToUint32(buf.Next(convert.Uint32Size)))
	certPoolData := buf.Next(size)
	defer security.CoverBytes(certPoolData)
	// unmarshal
	cp := ctrlCertPool{}
	err = msgpack.Unmarshal(certPoolData, &cp)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal certificate pool")
	}
	defer cp.Clean()
	memory.Padding()
	return cp.Dump(pool)
}

// calculateAESKey is used to generate aes key for encrypt certificate pool.
func calculateAESKey(password []byte) []byte {
	hash := sha256.New()
	hash.Write(password)
	hash.Write([]byte{0x20, 0x17, 0x04, 0x17})
	digest := hash.Sum(nil)
	return pbkdf2.Key(digest, digest[:16], 8192, aes.Key256Bit, sha256.New)
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
func (cp *CertPool) Load(pool *certpool.Pool) {
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
func (cp *CertPool) ToPool() (*certpool.Pool, error) {
	memory := security.NewMemory()
	defer memory.Flush()

	pool := certpool.NewPool()
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
