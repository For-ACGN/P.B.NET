package system

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/patch/monkey"
)

func TestExecutableName(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		name, err := ExecutableName()
		require.NoError(t, err)
		t.Log(name)
	})

	t.Run("failed to get executable path", func(t *testing.T) {
		patch := func() (string, error) {
			return "", monkey.Error
		}
		pg := monkey.Patch(os.Executable, patch)
		defer pg.Unpatch()

		name, err := ExecutableName()
		monkey.IsMonkeyError(t, err)
		require.Empty(t, name)
	})
}

func TestChdirToExe(t *testing.T) {
	cd, err := os.Getwd()
	require.NoError(t, err)
	t.Log("current directory:", cd)
	defer func() {
		err = os.Chdir(cd)
		require.NoError(t, err)
	}()

	t.Run("common", func(t *testing.T) {
		err = ChdirToExe()
		require.NoError(t, err)

		dd, err := os.Getwd()
		require.NoError(t, err)
		t.Log("now directory:", dd)

		require.NotEqual(t, cd, dd)
	})

	t.Run("failed to get executable path", func(t *testing.T) {
		patch := func() (string, error) {
			return "", monkey.Error
		}
		pg := monkey.Patch(os.Executable, patch)
		defer pg.Unpatch()

		err = ChdirToExe()
		monkey.IsMonkeyError(t, err)
	})
}

func TestCheckError(t *testing.T) {
	t.Run("not nil", func(t *testing.T) {
		patch := func(int) {}
		pg := monkey.Patch(os.Exit, patch)
		defer pg.Unpatch()

		CheckError(errors.New("test error"))
	})

	t.Run("nil", func(t *testing.T) {
		CheckError(nil)
	})
}

func TestCheckErrorf(t *testing.T) {
	t.Run("not nil", func(t *testing.T) {
		patch := func(int) {}
		pg := monkey.Patch(os.Exit, patch)
		defer pg.Unpatch()

		CheckErrorf("error: %s\n", errors.New("test error"))
	})

	t.Run("nil", func(t *testing.T) {
		CheckErrorf("error: %s\n", nil)
	})
}

func TestPrintError(t *testing.T) {
	patch := func(int) {}
	pg := monkey.Patch(os.Exit, patch)
	defer pg.Unpatch()

	PrintError("test error")
}

func TestPrintErrorf(t *testing.T) {
	patch := func(int) {}
	pg := monkey.Patch(os.Exit, patch)
	defer pg.Unpatch()

	PrintErrorf("error: %s\n", "test error")
}
