package yaegi

import (
	"bytes"
	"strings"

	"github.com/traefik/yaegi/extract"
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
	code = strings.Replace(code, "init", "init"+init, 1)
	return code + "\n", nil
}
