package webinfo

import (
	"io"
	"net/http"
	"os"

	"project/internal/system"
	"project/internal/testsuite"
)

const testTechFilePath = "testdata/tech.json"

// download tech.json file if it is not exist in the testdata directory.
func testDownloadWappalyzerData() {
	exist, err := system.IsExist(testTechFilePath)
	testsuite.TestMainCheckError(err)
	if exist {
		return
	}
	file, err := os.Create(testTechFilePath)
	testsuite.TestMainCheckError(err)

	client := http.Client{}
	defer client.CloseIdleConnections()
	resp, err := client.Get(WappalyzerURL)
	testsuite.TestMainCheckError(err)
	defer func() { _ = resp.Body.Close() }()
	_, err = io.Copy(file, resp.Body)
	testsuite.TestMainCheckError(err)
}
