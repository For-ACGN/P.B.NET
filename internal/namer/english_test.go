package namer

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/patch/monkey"
	"project/internal/security"
	"project/internal/testsuite"
)

func testGenerateEnglishResource(t *testing.T) []byte {
	buf := bytes.NewBuffer(make([]byte, 0, 4096))
	writer := zip.NewWriter(buf)
	for _, name := range [...]string{
		"prefix.txt",
		"stem.txt",
		"suffix.txt",
	} {
		file, err := os.Open("testdata/english/" + name)
		require.NoError(t, err)
		stat, err := file.Stat()
		require.NoError(t, err)

		fileHeader, err := zip.FileInfoHeader(stat)
		require.NoError(t, err)
		w, err := writer.CreateHeader(fileHeader)
		require.NoError(t, err)

		_, err = io.Copy(w, file)
		require.NoError(t, err)
	}
	err := writer.Close()
	require.NoError(t, err)
	return buf.Bytes()
}

func TestEnglish_Load(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("failed to read zip", func(t *testing.T) {
		english := NewEnglish()

		err := english.Load(nil)
		require.Error(t, err)

		testsuite.IsDestroyed(t, english)
	})

	t.Run("read useless file", func(t *testing.T) {
		buf := bytes.NewBuffer(make([]byte, 0, 4096))
		writer := zip.NewWriter(buf)
		_, err := writer.Create("test.txt")
		require.NoError(t, err)
		err = writer.Close()
		require.NoError(t, err)

		english := NewEnglish()

		err = english.Load(buf.Bytes())
		require.Error(t, err)

		testsuite.IsDestroyed(t, english)
	})

	t.Run("failed to load words", func(t *testing.T) {
		patch := func(file *zip.File) (*security.Bytes, error) {
			return nil, monkey.Error
		}
		pg := monkey.Patch(loadWordsFromZipFile, patch)
		defer pg.Unpatch()

		res := testGenerateEnglishResource(t)
		english := NewEnglish()
		err := english.Load(res)
		monkey.IsMonkeyError(t, err)

		testsuite.IsDestroyed(t, english)
	})
}

func TestEnglish_Generate(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("failed to check word number", func(t *testing.T) {
		english := NewEnglish()
		english.prefix = security.NewBytes(nil)
		english.stem = security.NewBytes(nil)
		english.suffix = security.NewBytes(nil)

		word, err := english.Generate(nil)
		require.EqualError(t, err, "empty prefix")
		require.Zero(t, word)

		testsuite.IsDestroyed(t, english)
	})

	t.Run("generate empty word", func(t *testing.T) {
		res := testGenerateEnglishResource(t)
		english := NewEnglish()
		err := english.Load(res)
		require.NoError(t, err)

		opts := Options{
			DisablePrefix: true,
			DisableStem:   true,
			DisableSuffix: true,
		}
		word, err := english.Generate(&opts)
		require.EqualError(t, err, "generated a empty word")
		require.Zero(t, word)

		testsuite.IsDestroyed(t, english)
	})
}

func TestEnglish_checkWordNumber(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	t.Run("empty prefix", func(t *testing.T) {
		english := NewEnglish()
		english.prefix = security.NewBytes(nil)

		err := english.checkWordNumber()
		require.EqualError(t, err, "empty prefix")

		testsuite.IsDestroyed(t, english)
	})

	t.Run("empty stem", func(t *testing.T) {
		english := NewEnglish()
		english.prefix = security.NewBytes([]byte{0})
		english.stem = security.NewBytes(nil)

		err := english.checkWordNumber()
		require.EqualError(t, err, "empty stem")

		testsuite.IsDestroyed(t, english)
	})

	t.Run("empty suffix", func(t *testing.T) {
		english := NewEnglish()
		english.prefix = security.NewBytes([]byte{0})
		english.stem = security.NewBytes([]byte{0})
		english.suffix = security.NewBytes(nil)

		err := english.checkWordNumber()
		require.EqualError(t, err, "empty suffix")

		testsuite.IsDestroyed(t, english)
	})
}

func TestEnglish_Load_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	res := testGenerateEnglishResource(t)

	t.Run("part", func(t *testing.T) {
		english := NewEnglish()

		load := func() {
			err := english.Load(res)
			require.NoError(t, err)
		}
		cleanup := func() {
			err := english.checkWordNumber()
			require.NoError(t, err)
		}
		testsuite.RunParallelTest(100, nil, cleanup, load, load)

		testsuite.IsDestroyed(t, english)
	})

	t.Run("whole", func(t *testing.T) {
		var english *English

		init := func() {
			english = NewEnglish()
		}
		load := func() {
			err := english.Load(res)
			require.NoError(t, err)
		}
		cleanup := func() {
			err := english.checkWordNumber()
			require.NoError(t, err)
		}
		testsuite.RunParallelTest(100, init, cleanup, load, load)

		testsuite.IsDestroyed(t, english)
	})
}

func TestEnglish_Generate_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	res := testGenerateEnglishResource(t)

	t.Run("part", func(t *testing.T) {
		english := NewEnglish()

		err := english.Load(res)
		require.NoError(t, err)

		gen := func() {
			word, err := english.Generate(nil)
			require.NoError(t, err)
			require.NotZero(t, word)

			t.Log(word)
		}
		cleanup := func() {
			err := english.checkWordNumber()
			require.NoError(t, err)
		}
		testsuite.RunParallelTest(100, nil, cleanup, gen, gen)

		testsuite.IsDestroyed(t, english)
	})

	t.Run("whole", func(t *testing.T) {
		var english *English

		init := func() {
			english = NewEnglish()

			err := english.Load(res)
			require.NoError(t, err)
		}
		gen := func() {
			word, err := english.Generate(nil)
			require.NoError(t, err)
			require.NotZero(t, word)

			t.Log(word)
		}
		cleanup := func() {
			err := english.checkWordNumber()
			require.NoError(t, err)
		}
		testsuite.RunParallelTest(100, init, cleanup, gen, gen)

		testsuite.IsDestroyed(t, english)
	})
}

func TestEnglish_Parallel(t *testing.T) {
	gm := testsuite.MarkGoroutines(t)
	defer gm.Compare()

	res := testGenerateEnglishResource(t)

	t.Run("part", func(t *testing.T) {
		english := NewEnglish()

		err := english.Load(res)
		require.NoError(t, err)

		load := func() {
			err := english.Load(res)
			require.NoError(t, err)
		}
		gen := func() {
			word, err := english.Generate(nil)
			require.NoError(t, err)
			require.NotZero(t, word)

			t.Log(word)
		}
		cleanup := func() {
			err := english.checkWordNumber()
			require.NoError(t, err)
		}
		testsuite.RunParallelTest(100, nil, cleanup, load, gen, load, gen)

		testsuite.IsDestroyed(t, english)
	})

	t.Run("whole", func(t *testing.T) {
		var english *English

		init := func() {
			english = NewEnglish()

			err := english.Load(res)
			require.NoError(t, err)
		}
		load := func() {
			err := english.Load(res)
			require.NoError(t, err)
		}
		gen := func() {
			word, err := english.Generate(nil)
			require.NoError(t, err)
			require.NotZero(t, word)

			t.Log(word)
		}
		cleanup := func() {
			err := english.checkWordNumber()
			require.NoError(t, err)
		}
		testsuite.RunParallelTest(100, init, cleanup, load, gen, load, gen)

		testsuite.IsDestroyed(t, english)
	})
}
