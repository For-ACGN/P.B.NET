package controller

import (
	"bytes"
	"compress/flate"
	"crypto/sha256"
	"crypto/subtle"

	"github.com/pkg/errors"

	"project/internal/crypto/aes"
	"project/internal/crypto/ed25519"
	"project/internal/random"
	"project/internal/security"
)

// SessionKeyFile is the session key file path.
const SessionKeyFile = "key/session.key"

func calculateAESKeyFromPassword(password []byte) ([]byte, []byte) {
	hash := sha256.New()
	hash.Write(password)
	hash.Write([]byte{20, 18, 11, 27})
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

// --------------------------------session file format--------------------------------
//
// +----------+------------+-------------+------------+---------------+--------------+
// |  SHA256  |   Random   | ED25519 Key |   Random   | Broadcast Key |    Random    |
// +----------+------------+-------------+------------+---------------+--------------+
// | 32 bytes | 7147 bytes |   64 bytes  | 2018 bytes | 32 + 16 bytes | > 1127 bytes |
// +----------+------------+-------------+------------+---------------+--------------+
//
// Hash is used to verify the integrality of the file.
// Hash value is sha256(random + ed25519 key + random + broadcast key + random)
// Random data is not the multiple of the sha256.BlockSize(64 bytes)
//
// Session key is the Private Key + Broadcast Key
// Private Key = ed25519 Private Key(64 Bytes)
// Broadcast Key = AES Key(256 Bit) + AES IV (32 Bytes + 16 Bytes, AES CBC)

// GenerateSessionKey is used to generate session key.
func GenerateSessionKey(password []byte) ([]byte, error) {
	keys, err := generateSessionKey()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return encryptSessionKey(keys, password)
}

func generateSessionKey() ([3][]byte, error) {
	var keys [3][]byte
	// generate ed25519 private key(for sign message)
	privateKey, err := ed25519.GenerateKey()
	if err != nil {
		return keys, errors.WithStack(err)
	}
	// generate aes key & iv(for broadcast message)
	aesKey := random.Bytes(aes.Key256Bit)
	aesIV := random.Bytes(aes.IVSize)
	keys[0] = privateKey
	keys[1] = aesKey
	keys[2] = aesIV
	return keys, nil
}

func encryptSessionKey(keys [3][]byte, password []byte) ([]byte, error) {
	// make session key file
	buf := bytes.NewBuffer(make([]byte, 0, 10240))
	buf.Write(random.Bytes(7147)) // random data 1
	buf.Write(keys[0])            // private key
	buf.Write(random.Bytes(2018)) // random data 2
	buf.Write(keys[1])            // aes key
	buf.Write(keys[2])            // aes iv
	thirdSize := 1127 + random.Int(1024)
	buf.Write(random.Bytes(thirdSize)) // random data 3
	// compress
	compressed := bytes.NewBuffer(make([]byte, 0, 5120))
	writer, err := flate.NewWriter(compressed, flate.BestCompression)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	_, err = writer.Write(buf.Bytes())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	err = writer.Close()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// encrypt file
	aesKey, aesIV := calculateAESKeyFromPassword(password)
	fileEnc, err := aes.CBCEncrypt(compressed.Bytes(), aesKey, aesIV)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// calculate file hash
	hash := sha256.New()
	hash.Write(buf.Bytes())
	fileHash := hash.Sum(nil)
	return append(fileHash, fileEnc...), nil
}

// LoadSessionKey is used to decrypt session key file and
// return session key (private key, aes key, aes iv).
func LoadSessionKey(sessionKey, password []byte) ([3][]byte, error) {
	var keys [3][]byte
	if len(sessionKey) < sha256.Size+aes.BlockSize {
		return keys, errors.New("invalid session key size")
	}
	memory := security.NewMemory()
	defer memory.Flush()
	// decrypt session key file
	aesKey, aesIV := calculateAESKeyFromPassword(password)
	fileDec, err := aes.CBCDecrypt(sessionKey[sha256.Size:], aesKey, aesIV)
	if err != nil {
		return keys, errors.WithStack(err)
	}
	// decompress
	buf := bytes.NewBuffer(make([]byte, 0, len(fileDec)*2))
	reader := flate.NewReader(bytes.NewReader(fileDec))
	_, err = buf.ReadFrom(reader)
	if err != nil {
		return keys, errors.WithStack(err)
	}
	file := buf.Bytes()
	// compare file hash
	fileHash := sha256.Sum256(file)
	if subtle.ConstantTimeCompare(sessionKey[:sha256.Size], fileHash[:]) != 1 {
		return keys, errors.New("incorrect password")
	}
	// ed25519 private key
	memory.Padding()
	offset := 7147
	privateKey := file[offset : offset+ed25519.PrivateKeySize]
	// broadcast key: aes key
	memory.Padding()
	offset += ed25519.PrivateKeySize + 2018
	aesKey = file[offset : offset+aes.Key256Bit]
	// broadcast key: aes iv
	memory.Padding()
	offset += aes.Key256Bit
	aesIV = file[offset : offset+aes.IVSize]
	// return
	keys[0] = privateKey
	keys[1] = aesKey
	keys[2] = aesIV
	return keys, nil
}

// ResetPassword is used to use new password to encrypt session key.
func ResetPassword(sessionKey, old, new []byte) ([]byte, error) {
	keys, err := LoadSessionKey(sessionKey, old)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to load session key")
	}
	return encryptSessionKey(keys, new)
}
