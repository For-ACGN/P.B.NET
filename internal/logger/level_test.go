package logger

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
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
			l, err := Parse(testdata.name)
			require.NoError(t, err)
			require.Equal(t, l, testdata.level)
		})
	}

	t.Run("invalid level", func(t *testing.T) {
		lv, err := Parse("invalid level")
		require.Error(t, err)
		require.Equal(t, lv, Level(0))
	})
}

func TestPrefix(t *testing.T) {
	for lv := Level(0); lv < Off; lv++ {
		fmt.Println(Prefix(time.Now(), lv, testSrc).String())
	}
	// unknown level
	fmt.Println(Prefix(time.Now(), Level(153), testSrc).String())
}
