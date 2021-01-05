package logger

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/testsuite"
)

const (
	testLogPrefixF   = "test format %s %s"
	testLogPrefix    = "test print"
	testLogPrefixLn  = "test println"
	testLogSrc       = "test src"
	testLogText1     = "test-text"
	testLogText2     = "test text2"
	testInvalidLevel = Level(255)
)

var errInvalidLevel = errors.New("invalid logger level: 255")

func TestCommonLogger(t *testing.T) {
	t.Run("Print", func(t *testing.T) {
		Common.Printf(Info, testLogSrc, testLogPrefixF, testLogText1, testLogText2)
		Common.Print(Info, testLogSrc, testLogPrefix, testLogText1, testLogText2)
		Common.Println(Info, testLogSrc, testLogPrefixLn, testLogText1, testLogText2)

		// will discard
		Common.Printf(Debug, testLogSrc, testLogPrefixF, testLogText1, testLogText2)
		Common.Print(Debug, testLogSrc, testLogPrefix, testLogText1, testLogText2)
		Common.Println(Debug, testLogSrc, testLogPrefixLn, testLogText1, testLogText2)
	})

	t.Run("SetLevel", func(t *testing.T) {
		err := Common.SetLevel(Error)
		require.NoError(t, err)

		lv := Common.GetLevel()
		require.Equal(t, Error, lv)

		err = Common.SetLevel(testInvalidLevel)
		require.Equal(t, err, errInvalidLevel)

		err = Common.SetLevel(Info)
		require.NoError(t, err)
	})

	t.Run("NewCommonLogger", func(t *testing.T) {
		lg, err := NewCommonLogger(Warning)
		require.NoError(t, err)
		require.NotNil(t, lg)

		lg, err = NewCommonLogger(testInvalidLevel)
		require.Error(t, err)
		require.Nil(t, lg)
	})
}

func TestTestLogger(t *testing.T) {
	t.Run("Print", func(t *testing.T) {
		Test.Printf(Debug, testLogSrc, testLogPrefixF, testLogText1, testLogText2)
		Test.Print(Debug, testLogSrc, testLogPrefix, testLogText1, testLogText2)
		Test.Println(Debug, testLogSrc, testLogPrefixLn, testLogText1, testLogText2)

		// will discard
		Test.Printf(Trace, testLogSrc, testLogPrefixF, testLogText1, testLogText2)
		Test.Print(Trace, testLogSrc, testLogPrefix, testLogText1, testLogText2)
		Test.Println(Trace, testLogSrc, testLogPrefixLn, testLogText1, testLogText2)
	})

	t.Run("SetLevel", func(t *testing.T) {
		err := Test.SetLevel(Error)
		require.NoError(t, err)

		lv := Test.GetLevel()
		require.Equal(t, Error, lv)

		err = Test.SetLevel(testInvalidLevel)
		require.Equal(t, err, errInvalidLevel)

		err = Test.SetLevel(Info)
		require.NoError(t, err)
	})

	t.Run("NewTestLogger", func(t *testing.T) {
		lg, err := NewTestLogger(Warning)
		require.NoError(t, err)
		require.NotNil(t, lg)

		lg, err = NewTestLogger(testInvalidLevel)
		require.Error(t, err)
		require.Nil(t, lg)
	})
}

func TestDiscardLogger(t *testing.T) {
	Discard.Printf(Debug, testLogSrc, testLogPrefixF, testLogText1, testLogText2)
	Discard.Print(Info, testLogSrc, testLogPrefix, testLogText1, testLogText2)
	Discard.Println(Error, testLogSrc, testLogPrefixLn, testLogText1, testLogText2)

	err := Discard.SetLevel(Info)
	require.NoError(t, err)

	lv := Discard.GetLevel()
	require.Equal(t, Off, lv)
}

func TestMultiLogger(t *testing.T) {
	t.Run("Print", func(t *testing.T) {
		lg, err := NewMultiLogger(Debug, os.Stdout)
		require.NoError(t, err)

		lg.Printf(Debug, testLogSrc, testLogPrefixF, testLogText1, testLogText2)
		lg.Print(Debug, testLogSrc, testLogPrefix, testLogText1, testLogText2)
		lg.Println(Debug, testLogSrc, testLogPrefixLn, testLogText1, testLogText2)

		// will discard
		lg.Printf(Trace, testLogSrc, testLogPrefixF, testLogText1, testLogText2)
		lg.Print(Trace, testLogSrc, testLogPrefix, testLogText1, testLogText2)
		lg.Println(Trace, testLogSrc, testLogPrefixLn, testLogText1, testLogText2)

		testsuite.IsDestroyed(t, lg)
	})

	t.Run("SetLevel", func(t *testing.T) {
		lg, err := NewMultiLogger(Debug, os.Stdout)
		require.NoError(t, err)

		err = lg.SetLevel(Error)
		require.NoError(t, err)

		lv := lg.GetLevel()
		require.Equal(t, Error, lv)

		err = lg.SetLevel(testInvalidLevel)
		require.Equal(t, err, errInvalidLevel)

		err = lg.SetLevel(Info)
		require.NoError(t, err)

		testsuite.IsDestroyed(t, lg)
	})

	t.Run("NewMultiLogger", func(t *testing.T) {
		lg, err := NewMultiLogger(Warning)
		require.NoError(t, err)
		require.NotNil(t, lg)

		testsuite.IsDestroyed(t, lg)

		lg, err = NewMultiLogger(testInvalidLevel)
		require.Error(t, err)
		require.Nil(t, lg)
	})
}
