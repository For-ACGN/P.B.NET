package testnamer

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

func TestNamer(t *testing.T) {
	namer := Namer()

	err := namer.Load(nil)
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		word, err := namer.Generate(nil)
		require.NoError(t, err)
		t.Log(word)
	}

	fmt.Println(namer.Type())

	testsuite.IsDestroyed(t, namer)
}

func TestWithLoadFailed(t *testing.T) {
	namer := WithLoadFailed()

	err := namer.Load(nil)
	require.Error(t, err)

	testsuite.IsDestroyed(t, namer)
}

func TestWithGenerateFailed(t *testing.T) {
	namer := WithGenerateFailed()

	err := namer.Load(nil)
	require.NoError(t, err)

	word, err := namer.Generate(nil)
	require.Error(t, err)
	require.Zero(t, word)

	testsuite.IsDestroyed(t, namer)
}
