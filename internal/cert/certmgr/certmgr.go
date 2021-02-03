package certmgr

import (
	"project/internal/cert/certpool"
	"project/internal/security"
)

// Manager contains raw certificates, it used for Node and Beacon configuration.
type Manager struct {
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

// Load is used to load certificates from certificate pool or other pool.
// Controller or tests will add certificates to the manager.
func (mgr *Manager) Load(pool *certpool.Pool) {
	pubRootCACerts := pool.GetPublicRootCACerts()
	for i := 0; i < len(pubRootCACerts); i++ {
		mgr.PublicRootCACerts = append(mgr.PublicRootCACerts, pubRootCACerts[i].Raw)
	}
	pubClientCACerts := pool.GetPublicClientCACerts()
	for i := 0; i < len(pubClientCACerts); i++ {
		mgr.PublicClientCACerts = append(mgr.PublicClientCACerts, pubClientCACerts[i].Raw)
	}
	pubClientPairs := pool.GetPublicClientPairs()
	for i := 0; i < len(pubClientPairs); i++ {
		c, k := pubClientPairs[i].Encode()
		mgr.PublicClientPairs = append(mgr.PublicClientPairs, struct {
			Cert []byte `msgpack:"a"`
			Key  []byte `msgpack:"b"`
		}{Cert: c, Key: k})
	}
	priRootCACerts := pool.GetPrivateRootCACerts()
	for i := 0; i < len(priRootCACerts); i++ {
		mgr.PrivateRootCACerts = append(mgr.PrivateRootCACerts, priRootCACerts[i].Raw)
	}
	priClientCACerts := pool.GetPrivateClientCACerts()
	for i := 0; i < len(priClientCACerts); i++ {
		mgr.PrivateClientCACerts = append(mgr.PrivateClientCACerts, priClientCACerts[i].Raw)
	}
	priClientPairs := pool.GetPrivateClientPairs()
	for i := 0; i < len(priClientPairs); i++ {
		c, k := priClientPairs[i].Encode()
		mgr.PrivateClientPairs = append(mgr.PrivateClientPairs, struct {
			Cert []byte `msgpack:"a"`
			Key  []byte `msgpack:"b"`
		}{Cert: c, Key: k})
	}
}

// ToCertPool is used to create a certificate pool. Call Clean to cover bytes in pool.
func (mgr *Manager) ToCertPool() (*certpool.Pool, error) {
	memory := security.NewMemory()
	defer memory.Flush()

	pool := certpool.NewPool()
	for i := 0; i < len(mgr.PublicRootCACerts); i++ {
		memory.Padding()
		err := pool.AddPublicRootCACert(mgr.PublicRootCACerts[i])
		if err != nil {
			return nil, err
		}
	}
	for i := 0; i < len(mgr.PublicClientCACerts); i++ {
		memory.Padding()
		err := pool.AddPublicClientCACert(mgr.PublicClientCACerts[i])
		if err != nil {
			return nil, err
		}
	}
	for i := 0; i < len(mgr.PublicClientPairs); i++ {
		memory.Padding()
		pair := mgr.PublicClientPairs[i]
		err := pool.AddPublicClientPair(pair.Cert, pair.Key)
		if err != nil {
			return nil, err
		}
	}
	for i := 0; i < len(mgr.PrivateRootCACerts); i++ {
		memory.Padding()
		err := pool.AddPrivateRootCACert(mgr.PrivateRootCACerts[i])
		if err != nil {
			return nil, err
		}
	}
	for i := 0; i < len(mgr.PrivateClientCACerts); i++ {
		memory.Padding()
		err := pool.AddPrivateClientCACert(mgr.PrivateClientCACerts[i])
		if err != nil {
			return nil, err
		}
	}
	for i := 0; i < len(mgr.PrivateClientPairs); i++ {
		memory.Padding()
		pair := mgr.PrivateClientPairs[i]
		err := pool.AddPrivateClientPair(pair.Cert, pair.Key)
		if err != nil {
			return nil, err
		}
	}
	return pool, nil
}

// Clean is used to clean all data in this certificate manager.
// Usually it will be called after call Manager.ToCertPool.
func (mgr *Manager) Clean() {
	for i := 0; i < len(mgr.PublicRootCACerts); i++ {
		security.CoverBytes(mgr.PublicRootCACerts[i])
	}
	for i := 0; i < len(mgr.PublicClientCACerts); i++ {
		security.CoverBytes(mgr.PublicClientCACerts[i])
	}
	for i := 0; i < len(mgr.PublicClientPairs); i++ {
		pair := mgr.PublicClientPairs[i]
		security.CoverBytes(pair.Cert)
		security.CoverBytes(pair.Key)
	}
	for i := 0; i < len(mgr.PrivateRootCACerts); i++ {
		security.CoverBytes(mgr.PrivateRootCACerts[i])
	}
	for i := 0; i < len(mgr.PrivateClientCACerts); i++ {
		security.CoverBytes(mgr.PrivateClientCACerts[i])
	}
	for i := 0; i < len(mgr.PrivateClientPairs); i++ {
		pair := mgr.PrivateClientPairs[i]
		security.CoverBytes(pair.Cert)
		security.CoverBytes(pair.Key)
	}
}
