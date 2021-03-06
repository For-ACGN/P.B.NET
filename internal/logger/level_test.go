package logger

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseLevel(t *testing.T) {
	t.Run("valid level", func(t *testing.T) {
		for _, testdata := range [...]*struct {
			name  string
			level Level
		}{
			{"all", All},
			{"trace", Trace},
			{"debug", Debug},
			{"info", Info},
			{"critical", Critical},
			{"warning", Warning},
			{"error", Error},
			{"exploit", Exploit},
			{"fatal", Fatal},
			{"off", Off},
		} {
			t.Run(testdata.name, func(t *testing.T) {
				lv, err := ParseLevel(testdata.name)
				require.NoError(t, err)
				require.Equal(t, lv, testdata.level)
			})
		}
	})

	t.Run("invalid level", func(t *testing.T) {
		lv, err := ParseLevel("invalid level")
		require.Error(t, err)
		require.Equal(t, lv, Level(0))
	})
}

func TestDumpPrefix(t *testing.T) {
	t.Run("valid level", func(t *testing.T) {
		for lv := Level(0); lv < Off; lv++ {
			buf := DumpPrefix(time.Now(), lv, testLogSrc)
			fmt.Println(buf)
		}
	})

	t.Run("invalid level", func(t *testing.T) {
		buf := DumpPrefix(time.Now(), Level(153), testLogSrc)
		fmt.Println(buf)
	})
}
