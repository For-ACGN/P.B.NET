package certmgr

import (
	"bytes"
	"compress/flate"
	"crypto/sha256"
	"crypto/subtle"

	"github.com/pkg/errors"

	"project/internal/convert"
	"project/internal/crypto/aes"
	"project/internal/crypto/cert"
	"project/internal/patch/msgpack"
	"project/internal/random"
	"project/internal/security"
	"project/internal/system"
)

// ---------------------------certificate pool file format--------------------------
//
// +----------+------------+--------------+-------------------------+--------------+
// |  SHA256  |   Random   | size(uint32) | msgpack(ctrlCertPool{}) |    Random    |
// +----------+------------+--------------+-------------------------+--------------+
// | 32 bytes | 2019 bytes |   4 bytes    |        var bytes        | > 1127 bytes |
// +----------+------------+--------------+-------------------------+--------------+
//
// Hash is used to verify the integrality of the file.
// Hash value is sha256(random + size + data + random)
// Random data is not the multiple of the sha256.BlockSize(64 bytes)
//
// use flate to compress(random + size + data + random)

// CertPoolFilePath is the certificate pool file path.
const CertPoolFilePath = "key/cert.pool"

const (
	randomSize2019 = 2019
	randomSize1127 = 1127
)

// ctrlCertPool include bytes about certificates and private keys.
// package controller and tool/certificate/manager will use it.
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

// SaveCtrlCertPool is used to compress and encrypt certificate pool.
func SaveCtrlCertPool(pool *cert.Pool, password []byte) error {
	cp := new(ctrlCertPool)
	// clean private key at once
	defer func() {
		for i := 0; i < len(cp.PublicClientPairs); i++ {
			security.CoverBytes(cp.PublicClientPairs[i].Key)
		}
		for i := 0; i < len(cp.PrivateRootCAPairs); i++ {
			security.CoverBytes(cp.PrivateRootCAPairs[i].Key)
		}
		for i := 0; i < len(cp.PrivateClientCAPairs); i++ {
			security.CoverBytes(cp.PrivateClientCAPairs[i].Key)
		}
		for i := 0; i < len(cp.PrivateClientPairs); i++ {
			security.CoverBytes(cp.PrivateClientPairs[i].Key)
		}
	}()
	getCertsFromPool(pool, cp)
	// marshal certificate pool
	certPool, err := msgpack.Marshal(cp)
	if err != nil {
		return err
	}
	defer security.CoverBytes(certPool)
	certPoolLen := len(certPool)
	// make certificate pool file
	buf := bytes.NewBuffer(make([]byte, 0, randomSize2019+4+certPoolLen+randomSize1127))
	buf.Write(random.Bytes(randomSize2019))               // random data 1
	buf.Write(convert.Uint32ToBytes(uint32(certPoolLen))) // msgpack data size
	buf.Write(certPool)                                   // msgpack data
	secondSize := randomSize1127 + random.Int(1024)
	buf.Write(random.Bytes(secondSize)) // random data 2
	// compress
	compressed := bytes.NewBuffer(make([]byte, 0, buf.Len()/2))
	writer, err := flate.NewWriter(compressed, flate.BestCompression)
	if err != nil {
		return errors.Wrap(err, "failed to create deflate writer")
	}
	_, err = writer.Write(buf.Bytes())
	if err != nil {
		return errors.Wrap(err, "failed to compress certificate data")
	}
	err = writer.Close()
	if err != nil {
		return errors.Wrap(err, "failed to close deflate writer")
	}
	// encrypt file
	aesKey, aesIV := calculateAESKeyFromPassword(password)
	defer func() {
		security.CoverBytes(aesKey)
		security.CoverBytes(aesIV)
	}()
	fileEnc, err := aes.CBCEncrypt(compressed.Bytes(), aesKey, aesIV)
	if err != nil {
		return errors.Wrap(err, "failed to encrypt certificate data")
	}
	// calculate file hash
	hash := sha256.New()
	hash.Write(buf.Bytes())
	fileHash := hash.Sum(nil)
	return system.WriteFile(CertPoolFilePath, append(fileHash, fileEnc...))
}

func getCertsFromPool(pool *cert.Pool, cp *ctrlCertPool) {
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

// calculateAESKeyFromPassword is used to generate aes key for encrypt certificate pool.
func calculateAESKeyFromPassword(password []byte) ([]byte, []byte) {
	hash := sha256.New()
	hash.Write(password)
	hash.Write([]byte{20, 17, 4, 17})
	hashed := hash.Sum(nil)
	// calculate more count
	var n int
	for i := 0; i < sha256.Size; i++ {
		n += int(hashed[i])
	}
	for i := 0; i < 10000+n; i++ {
		hash.Write(hashed)
		hashed = hash.Sum(nil)
	}
	keyIV := hash.Sum(nil)
	return keyIV, keyIV[:aes.IVSize]
}

// LoadCtrlCertPool is used to decrypt and decompress certificate pool.
func LoadCtrlCertPool(pool *cert.Pool, certPool, password []byte) error {
	if len(certPool) < sha256.Size+aes.BlockSize {
		return errors.New("invalid certificate pool file size")
	}
	memory := security.NewMemory()
	defer memory.Flush()
	// decrypt certificate pool file
	aesKey, aesIV := calculateAESKeyFromPassword(password)
	defer func() {
		security.CoverBytes(aesKey)
		security.CoverBytes(aesIV)
	}()
	compressed, err := aes.CBCDecrypt(certPool[sha256.Size:], aesKey, aesIV)
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
	offset := randomSize2019
	size := int(convert.BytesToUint32(file[offset : offset+4]))
	memory.Padding()
	offset += 4
	// unmarshal
	cp := ctrlCertPool{}
	err = msgpack.Unmarshal(file[offset:offset+size], &cp)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal certificate pool")
	}
	memory.Padding()
	return addCertsToPool(pool, &cp)
}

func addCertsToPool(pool *cert.Pool, cp *ctrlCertPool) error {
	memory := security.NewMemory()
	defer memory.Flush()

	var err error
	for i := 0; i < len(cp.PublicRootCACerts); i++ {
		err = pool.AddPublicRootCACert(cp.PublicRootCACerts[i])
		if err != nil {
			return err
		}
	}
	for i := 0; i < len(cp.PublicClientCACerts); i++ {
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

// NBCertPool contains raw certificates, it used for Node and Beacon configuration.
// package node and beacon will use it.
type NBCertPool struct {
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

// GetCertsFromPool is used to add certificates to NBCertPool from certificate pool.
// controller will add certificates to NBCertPool.
func (cp *NBCertPool) GetCertsFromPool(pool *cert.Pool) {
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

// ToPool is used to create a certificate pool from NBCertPool.
func (cp *NBCertPool) ToPool() (*cert.Pool, error) {
	memory := security.NewMemory()
	defer memory.Flush()

	pool := cert.NewPool()
	for i := 0; i < len(cp.PublicRootCACerts); i++ {
		err := pool.AddPublicRootCACert(cp.PublicRootCACerts[i])
		if err != nil {
			return nil, err
		}
	}
	for i := 0; i < len(cp.PublicClientCACerts); i++ {
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
		err := pool.AddPrivateRootCACert(cp.PrivateRootCACerts[i])
		if err != nil {
			return nil, err
		}
	}
	for i := 0; i < len(cp.PrivateClientCACerts); i++ {
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
