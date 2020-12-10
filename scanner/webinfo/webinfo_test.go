package webinfo

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	testDownloadWappalyzerData()
	os.Exit(m.Run())
}
