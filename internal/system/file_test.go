package system

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/patch/monkey"
	"project/internal/testsuite"
)

func TestOpenFile(t *testing.T) {
	const (
		flag = os.O_WRONLY | os.O_CREATE
		perm = 0600
	)

	t.Run("ok", func(t *testing.T) {
		const name = "testdata/of.dat"

		file, err := OpenFile(name, flag, perm)
		require.NoError(t, err)

		err = file.Close()
		require.NoError(t, err)

		err = os.Remove(name)
		require.NoError(t, err)
	})

	t.Run("failed", func(t *testing.T) {
		file, err := OpenFile("testdata/<</file", flag, perm)
		require.Error(t, err)
		require.Nil(t, file)
	})
}

func TestWriteFile(t *testing.T) {
	testdata := testsuite.Bytes()

	t.Run("ok", func(t *testing.T) {
		const name = "testdata/wf.dat"

		err := WriteFile(name, testdata)
		require.NoError(t, err)

		err = os.Remove(name)
		require.NoError(t, err)
	})

	t.Run("invalid path", func(t *testing.T) {
		err := WriteFile("testdata/<</file", testdata)
		require.Error(t, err)
	})
}

func TestCopyFile(t *testing.T) {
	const (
		src = "testdata/cf_src.dat"
		dst = "testdata/cf_dst.dat"
	)
	err := WriteFile(src, testsuite.Bytes())
	require.NoError(t, err)
	defer func() {
		err = os.Remove(src)
		require.NoError(t, err)
	}()

	t.Run("common", func(t *testing.T) {
		err := CopyFile(dst, src)
		require.NoError(t, err)

		err = os.Remove(dst)
		require.NoError(t, err)
	})

	t.Run("source file is not exist", func(t *testing.T) {
		err := CopyFile(dst, "foo")
		require.Error(t, err)
	})
}

func TestMoveFile(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		const (
			src = "testdata/mf_src.dat"
			dst = "testdata/mf_dst.dat"
		)
		err := WriteFile(src, testsuite.Bytes())
		require.NoError(t, err)
		defer func() {
			err = os.Remove(src)
			require.Error(t, err)
		}()

		err = MoveFile(dst, src)
		require.NoError(t, err)
		defer func() {
			err = os.Remove(dst)
			require.NoError(t, err)
		}()

		exist, err := IsPathExist(dst)
		require.NoError(t, err)
		require.True(t, exist)

		exist, err = IsPathNotExist(src)
		require.NoError(t, err)
		require.True(t, exist)
	})

	t.Run("not exist", func(t *testing.T) {
		err := MoveFile("foo_dst", "foo_src")
		require.Error(t, err)
	})
}

func TestIsSamePath(t *testing.T) {
	t.Run("same", func(t *testing.T) {
		same, err := IsSamePath("a", "a")
		require.NoError(t, err)
		require.True(t, same)
	})

	t.Run("not same", func(t *testing.T) {
		same, err := IsSamePath("a", "b")
		require.NoError(t, err)
		require.False(t, same)
	})

	t.Run("not enough path", func(t *testing.T) {
		same, err := IsSamePath("a")
		require.Error(t, err)
		require.False(t, same)
	})

	t.Run("invalid path", func(t *testing.T) {
		patch := func(string) (string, error) {
			return "", monkey.Error
		}
		pg := monkey.Patch(filepath.Abs, patch)
		defer pg.Unpatch()

		same, err := IsSamePath("a", "b")
		monkey.IsMonkeyError(t, err)
		require.False(t, same)
	})

	t.Run("invalid second path", func(t *testing.T) {
		patch := func(path string) (string, error) {
			if path == "b" {
				return "", monkey.Error
			}
			return "", nil
		}
		pg := monkey.Patch(filepath.Abs, patch)
		defer pg.Unpatch()

		same, err := IsSamePath("a", "b")
		monkey.IsMonkeyError(t, err)
		require.False(t, same)
	})
}

func TestIsFilePath(t *testing.T) {

}

func TestIsDirPath(t *testing.T) {

}

func TestIsPathExist(t *testing.T) {
	t.Run("exist", func(t *testing.T) {
		exist, err := IsPathExist("testdata")
		require.NoError(t, err)
		require.True(t, exist)
	})

	t.Run("is not exist", func(t *testing.T) {
		exist, err := IsPathExist("not")
		require.NoError(t, err)
		require.False(t, exist)
	})

	t.Run("invalid path", func(t *testing.T) {
		exist, err := IsPathExist("testdata/<</file")
		require.Error(t, err)
		require.False(t, exist)
	})
}

func TestIsPathNotExist(t *testing.T) {
	t.Run("is not exist", func(t *testing.T) {
		notExist, err := IsPathNotExist("not")
		require.NoError(t, err)
		require.True(t, notExist)
	})

	t.Run("exist", func(t *testing.T) {
		notExist, err := IsPathNotExist("testdata")
		require.NoError(t, err)
		require.False(t, notExist)
	})

	t.Run("invalid path", func(t *testing.T) {
		notExist, err := IsPathNotExist("testdata/<</file")
		require.Error(t, err)
		require.False(t, notExist)
	})
}
