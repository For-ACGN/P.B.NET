package httptool

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"project/internal/patch/monkey"
)

func testGenerateRequest(t *testing.T) *http.Request {
	req, err := http.NewRequest(http.MethodGet, "https://github.com/", nil)
	require.NoError(t, err)

	req.RemoteAddr = "127.0.0.1:1234"
	req.RequestURI = "/index"
	req.Header.Set("User-Agent", "Mozilla")
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Connection", "keep-alive")
	return req
}

func TestDumpRequest(t *testing.T) {
	req := testGenerateRequest(t)

	t.Run("GET and no body", func(t *testing.T) {
		fmt.Println("-----begin-----")
		fmt.Println(SdumpRequest(req))
		fmt.Println("-----end-----")
	})

	equalBody := func(b1, b2 io.Reader) {
		d1, err := io.ReadAll(b1)
		require.NoError(t, err)
		d2, err := io.ReadAll(b2)
		require.NoError(t, err)
		require.Equal(t, d1, d2)
	}
	body := new(bytes.Buffer)
	rawBody := bytes.NewReader(body.Bytes())

	t.Run("GET with body but no data", func(t *testing.T) {
		req.Body = io.NopCloser(body)

		fmt.Println("-----begin-----")
		DumpRequest(req)
		fmt.Println("-----end-----")

		equalBody(rawBody, req.Body)
	})

	t.Run("POST with data < bodyLineLength", func(t *testing.T) {
		body.Reset()
		body.WriteString(strings.Repeat("a", defaultBodyLineLength-10))
		rawBody.Reset(body.Bytes())
		req.Body = io.NopCloser(body)

		fmt.Println("-----begin-----")
		DumpRequest(req)
		fmt.Println("-----end-----")

		equalBody(rawBody, req.Body)
	})

	t.Run("POST with data = bodyLineLength", func(t *testing.T) {
		body.Reset()
		body.WriteString(strings.Repeat("a", defaultBodyLineLength))
		rawBody.Reset(body.Bytes())
		req.Body = io.NopCloser(body)

		fmt.Println("-----begin-----")
		DumpRequest(req)
		fmt.Println("-----end-----")

		equalBody(rawBody, req.Body)
	})

	t.Run("POST with data 3*bodyLineLength-1", func(t *testing.T) {
		body.Reset()
		body.WriteString(strings.Repeat("a", 3*defaultBodyLineLength-1))
		rawBody = bytes.NewReader(body.Bytes())
		req.Body = io.NopCloser(body)

		fmt.Println("-----begin-----")
		DumpRequest(req)
		fmt.Println("-----end-----")

		equalBody(rawBody, req.Body)
	})

	t.Run("POST with data 100*bodyLineLength-1", func(t *testing.T) {
		body.Reset()
		body.WriteString(strings.Repeat("a", 100*defaultBodyLineLength-1))
		rawBody = bytes.NewReader(body.Bytes())
		req.Body = io.NopCloser(body)

		fmt.Println("-----begin-----")
		DumpRequest(req)
		fmt.Println("-----end-----")

		equalBody(rawBody, req.Body)
	})
}

func TestDumpRequestWithError(t *testing.T) {
	req := testGenerateRequest(t)

	for _, testdata := range [...]*struct {
		name   string
		format string
	}{
		{"remote", "Remote: %s\n"},
		{"request", "%s %s %s"},
		{"host", "\nHost: %s"},
		{"header", "\n%s: %s"},
	} {
		t.Run(testdata.name, func(t *testing.T) {
			patch := func(w io.Writer, format string, a ...interface{}) (int, error) {
				if format == testdata.format {
					return 0, monkey.Error
				}
				return w.Write([]byte(fmt.Sprintf(format, a...)))
			}
			pg := monkey.Patch(fmt.Fprintf, patch)
			defer pg.Unpatch()

			_, err := FdumpRequest(os.Stdout, req)
			monkey.IsMonkeyError(t, err)

			// fix goland new line bug
			fmt.Println()
		})
	}

	t.Run("SdumpRequestWithBM", func(t *testing.T) {
		patch := func(w io.Writer, format string, a ...interface{}) (int, error) {
			return 0, monkey.Error
		}
		pg := monkey.Patch(fmt.Fprintf, patch)
		defer pg.Unpatch()

		fmt.Println(SdumpRequest(req))
	})

	t.Run("DumpRequestWithBM", func(t *testing.T) {
		patch := func(w io.Writer, format string, a ...interface{}) (int, error) {
			return 0, monkey.Error
		}
		pg := monkey.Patch(fmt.Fprintf, patch)
		defer pg.Unpatch()

		DumpRequest(req)
	})
}

func TestDumpBodyWithError(t *testing.T) {
	req := testGenerateRequest(t)

	t.Run("size < bodyLineLength", func(t *testing.T) {
		patch := func(w io.Writer, format string, a ...interface{}) (int, error) {
			if format == "\n\n%s" {
				return 0, monkey.Error
			}
			return w.Write([]byte(fmt.Sprintf(format, a...)))
		}
		pg := monkey.Patch(fmt.Fprintf, patch)
		defer pg.Unpatch()

		req.Body = io.NopCloser(strings.NewReader("test"))

		_, err := FdumpRequest(os.Stdout, req)
		monkey.IsMonkeyError(t, err)

		// fix goland new line bug
		fmt.Println()
	})

	t.Run("bodyLineLength < size < 2x", func(t *testing.T) {
		patch := func(w io.Writer, format string, a ...interface{}) (int, error) {
			if format == "\n\n%s" {
				return 0, monkey.Error
			}
			return w.Write([]byte(fmt.Sprintf(format, a...)))
		}
		pg := monkey.Patch(fmt.Fprintf, patch)
		defer pg.Unpatch()

		testdata := "test" + strings.Repeat("a", defaultBodyLineLength)
		req.Body = io.NopCloser(strings.NewReader(testdata))

		_, err := FdumpRequest(os.Stdout, req)
		monkey.IsMonkeyError(t, err)

		// fix goland new line bug
		fmt.Println()
	})

	t.Run("1.5 x bodyLineLength", func(t *testing.T) {
		patch := func(w io.Writer, format string, a ...interface{}) (int, error) {
			if format == "\n%s" {
				return 0, monkey.Error
			}
			return w.Write([]byte(fmt.Sprintf(format, a...)))
		}
		pg := monkey.Patch(fmt.Fprintf, patch)
		defer pg.Unpatch()

		testdata := "test" + strings.Repeat("a", defaultBodyLineLength)
		req.Body = io.NopCloser(strings.NewReader(testdata))

		_, err := FdumpRequest(os.Stdout, req)
		monkey.IsMonkeyError(t, err)

		// fix goland new line bug
		fmt.Println()
	})

	t.Run("2.5 x bodyLineLength", func(t *testing.T) {
		patch := func(w io.Writer, format string, a ...interface{}) (int, error) {
			if format == "\n%s" {
				return 0, monkey.Error
			}
			return w.Write([]byte(fmt.Sprintf(format, a...)))
		}
		pg := monkey.Patch(fmt.Fprintf, patch)
		defer pg.Unpatch()

		testdata := "test" + strings.Repeat("a", 2*defaultBodyLineLength)
		req.Body = io.NopCloser(strings.NewReader(testdata))

		_, err := FdumpRequest(os.Stdout, req)
		monkey.IsMonkeyError(t, err)

		// fix goland new line bug
		fmt.Println()
	})
}

func TestSubHTTPFileSystem_Open(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	fs := http.Dir(wd)

	sfs := NewSubHTTPFileSystem(fs, "testdata")
	file, err := sfs.Open("data.txt")
	require.NoError(t, err)
	data, err := io.ReadAll(file)
	require.NoError(t, err)
	require.Equal(t, "hello", string(data))

	file, err = sfs.Open("foo")
	require.Error(t, err)
	require.Nil(t, file)
}
