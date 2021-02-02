package certmgr

import (
	"project/internal/cert/certpool"
	"project/internal/security"
)

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
