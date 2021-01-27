package namer

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/patch/monkey"
	"project/internal/patch/toml"
	"project/internal/security"
	"project/internal/testsuite"
)

var tests = [...]*struct {
	typ   string
	resFn func(*testing.T) []byte
}{
	{"english", testGenerateEnglishResource},
}

func TestNamers(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	for _, test := range tests {
		t.Run(test.typ, func(t *testing.T) {
			namer, err := Load(test.typ, test.resFn(t))
			require.NoError(t, err)

			for i := 0; i < 10; i++ {
				word, err := namer.Generate(nil)
				require.NoError(t, err)
				t.Log(word)
			}

			fmt.Println(namer.Type())

			testsuite.IsDestroyed(t, namer)
		})
	}
}

func TestLoad(t *testing.T) {
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
		err := Register("namer", regFn)
		require.NoError(t, err)
		defer Unregister("namer")

		res := testGenerateEnglishResource(t)
		namer, err := Load("namer", res)
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
	err := Register("namer", regFn)
	require.NoError(t, err)

	res := testGenerateEnglishResource(t)
	namer, err := Load("namer", res)
	require.NoError(t, err)
	require.NotNil(t, namer)

	Unregister("namer")

	namer, err = Load("namer", res)
	require.Error(t, err)
	require.Nil(t, namer)
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

func TestNamers_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	for _, test := range tests {
		t.Run(test.typ, func(t *testing.T) {
			res := test.resFn(t)

			t.Run("part", func(t *testing.T) {
				namer, err := Load(test.typ, test.resFn(t))
				require.NoError(t, err)

				load := func() {
					err := namer.Load(res)
					require.NoError(t, err)
				}
				gen := func() {
					word, err := namer.Generate(nil)
					require.NoError(t, err)
					require.NotZero(t, word)

					t.Log(word)
				}
				testsuite.RunParallel(100, nil, nil, load, gen, load, gen)

				testsuite.IsDestroyed(t, namer)
			})

			t.Run("whole", func(t *testing.T) {
				var namer Namer

				init := func() {
					var err error
					namer, err = Load(test.typ, test.resFn(t))
					require.NoError(t, err)
				}
				load := func() {
					err := namer.Load(res)
					require.NoError(t, err)
				}
				gen := func() {
					word, err := namer.Generate(nil)
					require.NoError(t, err)
					require.NotZero(t, word)

					t.Log(word)
				}
				testsuite.RunParallel(100, init, nil, load, gen, load, gen)

				testsuite.IsDestroyed(t, namer)
			})
		})
	}
}

func TestLoad_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	res := testGenerateEnglishResource(t)
	regFn := func() Namer { return NewEnglish() }
	err := Register("namer1", regFn)
	require.NoError(t, err)

	init := func() {
		err = Register("namer2", regFn)
		require.NoError(t, err)
	}
	reg1 := func() {
		err := Register("namer3", regFn)
		require.NoError(t, err)
	}
	reg2 := func() {
		err := Register("namer4", regFn)
		require.NoError(t, err)
	}
	reg3 := func() {
		err := Register("namer1", regFn)
		require.Error(t, err)
	}
	un1 := func() {
		Unregister("namer2")
	}
	un2 := func() {
		Unregister("namer5")
	}
	load1 := func() {
		namer, err := Load("english", res)
		require.NoError(t, err)
		require.NotNil(t, namer)
	}
	load2 := func() {
		namer, err := Load("namer1", res)
		require.NoError(t, err)
		require.NotNil(t, namer)
	}
	load3 := func() {
		namer, err := Load("namer5", nil)
		require.Error(t, err)
		require.Nil(t, namer)
	}
	cleanup := func() {
		Unregister("namer3")
		Unregister("namer4")
	}
	fns := []func(){
		reg1, reg2, reg3, un1, un2,
		load1, load2, load3,
	}
	testsuite.RunParallel(100, init, cleanup, fns...)
}

func TestOptions(t *testing.T) {
	data, err := os.ReadFile("testdata/options.toml")
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
