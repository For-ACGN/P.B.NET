package namer

import (
	"archive/zip"
	"bytes"
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/patch/monkey"
	"project/internal/patch/toml"
	"project/internal/security"
	"project/internal/testsuite"
)

func TestLoad(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		res := testGenerateEnglishResource(t)
		namer, err := Load("english", res)
		require.NoError(t, err)

		word, err := namer.Generate(nil)
		require.NoError(t, err)
		t.Log(word)

		testsuite.IsDestroyed(t, namer)
	})

	t.Run("unregistered namer type", func(t *testing.T) {
		namer, err := Load("foo", nil)
		require.EqualError(t, err, "namer \"foo\" is not registered")
		require.Nil(t, namer)
	})

	t.Run("failed to load namer resource", func(t *testing.T) {
		namer, err := Load("english", nil)
		require.Error(t, err)
		require.Nil(t, namer)
	})
}

func TestRegister(t *testing.T) {
	regFn := func() Namer { return NewEnglish() }

	t.Run("common", func(t *testing.T) {
		err := Register("mock", regFn)
		require.NoError(t, err)

		res := testGenerateEnglishResource(t)
		namer, err := Load("mock", res)
		require.NoError(t, err)

		word, err := namer.Generate(nil)
		require.NoError(t, err)
		t.Log(word)

		testsuite.IsDestroyed(t, namer)
	})

	t.Run("already registered", func(t *testing.T) {
		err := Register("english", regFn)
		require.EqualError(t, err, "namer \"english\" is already registered")
	})
}

func TestUnregister(t *testing.T) {
	regFn := func() Namer { return NewEnglish() }
	err := Register("mock", regFn)
	require.NoError(t, err)

	res := testGenerateEnglishResource(t)
	namer, err := Load("mock", res)
	require.NoError(t, err)
	require.NotNil(t, namer)

	Unregister("mock")

	namer, err = Load("mock", res)
	require.Error(t, err)
	require.Nil(t, namer)
}

func TestLoad_Parallel(t *testing.T) {

}

func TestLoadWordsFromZipFile(t *testing.T) {
	// create test zip file
	buf := bytes.NewBuffer(make([]byte, 0, 64))
	writer := zip.NewWriter(buf)
	file, err := writer.Create("test.dat")
	require.NoError(t, err)
	_, err = file.Write([]byte("test data"))
	require.NoError(t, err)
	err = writer.Close()
	require.NoError(t, err)
	reader := bytes.NewReader(buf.Bytes())
	size := int64(buf.Len())

	t.Run("failed to open file", func(t *testing.T) {
		file := new(zip.File)
		patch := func(*zip.File) (io.ReadCloser, error) {
			return nil, monkey.Error
		}
		pg := monkey.PatchInstanceMethod(file, "Open", patch)
		defer pg.Unpatch()

		zipFile, err := zip.NewReader(reader, size)
		require.NoError(t, err)

		sb, err := loadWordsFromZipFile(zipFile.File[0])
		monkey.IsMonkeyError(t, err)
		require.Nil(t, sb)
	})

	t.Run("failed to read file data", func(t *testing.T) {
		patch := func(io.Reader, int64) ([]byte, error) {
			return nil, monkey.Error
		}
		pg := monkey.Patch(security.ReadAll, patch)
		defer pg.Unpatch()

		zipFile, err := zip.NewReader(reader, size)
		require.NoError(t, err)

		sb, err := loadWordsFromZipFile(zipFile.File[0])
		monkey.IsExistMonkeyError(t, err)
		require.Nil(t, sb)
	})
}

func TestOptions(t *testing.T) {
	data, err := ioutil.ReadFile("testdata/options.toml")
	require.NoError(t, err)

	// check unnecessary field
	opts := Options{}
	err = toml.Unmarshal(data, &opts)
	require.NoError(t, err)

	// check zero value
	testsuite.ContainZeroValue(t, opts)

	for _, testdata := range [...]*struct {
		expected interface{}
		actual   interface{}
	}{
		{expected: true, actual: opts.DisablePrefix},
		{expected: true, actual: opts.DisableStem},
		{expected: true, actual: opts.DisableSuffix},
	} {
		require.Equal(t, testdata.expected, testdata.actual)
	}
}
