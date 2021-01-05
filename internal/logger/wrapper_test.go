package logger

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewWriterWithPrefix(t *testing.T) {
	w := NewWriterWithPrefix(os.Stdout, "prefix")
	_, err := w.Write([]byte("test\n"))
	require.NoError(t, err)
}

func TestWrapLogger(t *testing.T) {
	w := WrapLogger(Error, "test wrap", Test)
	_, err := w.Write([]byte("test data\n"))
	require.NoError(t, err)
	_, err = w.Write([]byte("test data"))
	require.NoError(t, err)
}

func TestWrap(t *testing.T) {
	l := Wrap(Error, "test wrap", Test)
	l.Printf("Printf")
	l.Print("Print")
	l.Println("Println")
}

func TestHijackLogWriter(t *testing.T) {
	HijackLogWriter(Error, "test", Test)
	log.Printf("Printf")
	log.Print("Print")
	log.Println("Println")
}

func TestSetErrorLogger(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		const name = "testdata/test.err"

		file, err := SetErrorLogger(name)
		require.NoError(t, err)

		log.Println("test log")

		err = file.Close()
		require.NoError(t, err)
		err = os.Remove(name)
		require.NoError(t, err)
	})

	t.Run("fail", func(t *testing.T) {
		file, err := SetErrorLogger("testdata/<</file")
		require.Error(t, err)
		require.Nil(t, file)
	})
}
