package certpool

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSystem(t *testing.T) {
	fn := func() {
		pool, err := System()
		require.NoError(t, err)
		t.Log("the number of the system certificates:", len(pool.Certs()))
		for _, cert := range pool.Certs() {
			t.Log(cert.Subject.CommonName)
		}
	}

	wg := sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fn()
		}()
	}
	wg.Wait()
}
