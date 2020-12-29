package yaegi

import (
	"bytes"
	"fmt"
	"go/format"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/traefik/yaegi/extract"

	"project/internal/system"
)

func generateCode(pkg, init string) (string, error) {
	ext := extract.Extractor{
		Dest: "yaegi", // will be removed
	}
	buf := bytes.NewBuffer(make([]byte, 0, 10240))
	_, err := ext.Extract(pkg, "", buf)
	if err != nil {
		return "", err
	}
	code := buf.String()
	// remove all code before func init()
	idx := strings.Index(code, "func init()")
	code = code[idx:]
	// rename init like initString() for search easily
	code = strings.Replace(code, "init", "init_"+init, 1)
	return code + "\n", nil
}

func formatCodeAndSave(t *testing.T, code, path string) {
	data, err := format.Source([]byte(code))
	require.NoError(t, err)
	fmt.Println(string(data))
	err = system.WriteFile(path, data)
	require.NoError(t, err)
}
