package logger

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

const (
	testPrefixF  = "test format %s %s"
	testPrefix   = "test print"
	testPrefixLn = "test println"
	testSrc      = "test src"
	testLog1     = "test"
	testLog2     = "log"
)

func TestLogger(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		Common.Printf(Info, testSrc, testPrefixF, testLog1, testLog2)
		Common.Print(Info, testSrc, testPrefix, testLog1, testLog2)
		Common.Println(Info, testSrc, testPrefixLn, testLog1, testLog2)

		// will not display
		Common.Printf(Debug, testSrc, testPrefixF, testLog1, testLog2)
		Common.Print(Debug, testSrc, testPrefix, testLog1, testLog2)
		Common.Println(Debug, testSrc, testPrefixLn, testLog1, testLog2)
	})

	t.Run("test", func(t *testing.T) {
		Test.Printf(Debug, testSrc, testPrefixF, testLog1, testLog2)
		Test.Print(Debug, testSrc, testPrefix, testLog1, testLog2)
		Test.Println(Debug, testSrc, testPrefixLn, testLog1, testLog2)
	})

	t.Run("discard", func(t *testing.T) {
		Discard.Printf(Debug, testSrc, testPrefixF, testLog1, testLog2)
		Discard.Print(Debug, testSrc, testPrefix, testLog1, testLog2)
		Discard.Println(Debug, testSrc, testPrefixLn, testLog1, testLog2)
	})
}

func TestMultiLogger(t *testing.T) {
	logger, err := NewMultiLogger(Debug, os.Stdout)
	require.NoError(t, err)

	t.Run("common", func(t *testing.T) {
		logger.Printf(Debug, testSrc, testPrefixF, testLog1, testLog2)
		logger.Print(Debug, testSrc, testPrefix, testLog1, testLog2)
		logger.Println(Debug, testSrc, testPrefixLn, testLog1, testLog2)
	})

	t.Run("low level", func(t *testing.T) {
		err := logger.SetLevel(Info)
		require.NoError(t, err)

		logger.Printf(Debug, testSrc, testPrefixF, testLog1, testLog2)
		logger.Print(Debug, testSrc, testPrefix, testLog1, testLog2)
		logger.Println(Debug, testSrc, testPrefixLn, testLog1, testLog2)
	})

	t.Run("invalid level", func(t *testing.T) {
		err := logger.SetLevel(Level(123))
		require.EqualError(t, err, "invalid logger level: 123")
	})

	testsuite.IsDestroyed(t, logger)
}
