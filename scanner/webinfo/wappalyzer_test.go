package webinfo

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

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

func TestLoadWappalyzerTechFile(t *testing.T) {
	data, err := ioutil.ReadFile(testTechFilePath)
	require.NoError(t, err)
	techDef, err := LoadWappalyzerTechFile(data)
	require.NoError(t, err)

	fmt.Println("categories:", len(techDef.Cats))
	for _, cat := range techDef.Cats {
		fmt.Println(cat.Name)
	}
	fmt.Println("technologies:", len(techDef.Techs))
	for name, tech := range techDef.Techs {
		fmt.Println(name, tech.CatNames, len(tech.CatNames))
	}
}
