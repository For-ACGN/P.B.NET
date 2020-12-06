package nmap

import (
	"io/ioutil"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/require"
)

func TestParseOutput(t *testing.T) {
	t.Run("common", func(t *testing.T) {
		data, err := ioutil.ReadFile("testdata/nmap.xml")
		require.NoError(t, err)
		output, err := ParseOutput(data)
		require.NoError(t, err)
		spew.Dump(output)
	})

	t.Run("error", func(t *testing.T) {
		output, err := ParseOutput([]byte("foo"))
		require.Error(t, err)
		require.Nil(t, output)
	})
}
