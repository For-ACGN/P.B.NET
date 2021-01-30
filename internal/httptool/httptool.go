package httptool

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const (
	defaultBodyLineLength = 64   // default post data length in one line.
	defaultMaxBodyLength  = 1024 // default maximum body length.
)

// DumpRequest is used to dump http request to os.Stdout.
func DumpRequest(r *http.Request) {
	DumpRequestWithBM(r, defaultBodyLineLength, defaultMaxBodyLength)
}

// SdumpRequest is used to dump http request to a string.
func SdumpRequest(r *http.Request) string {
	return SdumpRequestWithBM(r, defaultBodyLineLength, defaultMaxBodyLength)
}

// FdumpRequest is used to dump http request to a io.Writer.
func FdumpRequest(w io.Writer, r *http.Request) (int, error) {
	return FdumpRequestWithBM(w, r, defaultBodyLineLength, defaultMaxBodyLength)
}

// DumpRequestWithBM is used to dump http request to os.Stdout.
// bll is the body line length, mbl is the max body length.
func DumpRequestWithBM(r *http.Request, bll, mbl int) {
	buf := bytes.NewBuffer(make([]byte, 0, 512))
	_, _ = FdumpRequestWithBM(buf, r, bll, mbl)
	buf.WriteString("\n")
	_, _ = buf.WriteTo(os.Stdout)
}

// SdumpRequestWithBM is used to dump http request to a string.
// bll is the body line length, mbl is the max body length.
func SdumpRequestWithBM(r *http.Request, bll, mbl int) string {
	builder := strings.Builder{}
	builder.Grow(512)
	_, _ = FdumpRequestWithBM(&builder, r, bll, mbl)
	return builder.String()
}

// FdumpRequestWithBM is used to dump http request to a io.Writer.
// bll is the body line length, mbl is the max body length.
//
// Output:
// Remote: 127.0.0.1:1234
// POST /index HTTP/1.1
// Host: github.com
// Accept: text/html
// Connection: keep-alive
// User-Agent: Mozilla
//
// post data...
// post data...
func FdumpRequestWithBM(w io.Writer, r *http.Request, bll, mbl int) (int, error) {
	var num int
	n, err := fmt.Fprintf(w, "Remote: %s\n", r.RemoteAddr)
	num += n
	if err != nil {
		return num, err
	}
	// header
	n, err = fmt.Fprintf(w, "%s %s %s", r.Method, r.RequestURI, r.Proto)
	num += n
	if err != nil {
		return num, err
	}
	// dump host
	n, err = fmt.Fprintf(w, "\nHost: %s", r.Host)
	num += n
	if err != nil {
		return num, err
	}
	// dump header
	for k, v := range r.Header {
		n, err = fmt.Fprintf(w, "\n%s: %s", k, v[0])
		num += n
		if err != nil {
			return num, err
		}
	}
	if r.Body == nil {
		return n, nil
	}
	n, err = fDumpBody(w, r, bll, mbl)
	num += n
	return num, err
}

// fDumpBody is used to dump http response body to io.Writer.
func fDumpBody(w io.Writer, r *http.Request, bll, mbl int) (int, error) {
	rawBody := new(bytes.Buffer)
	defer func() {
		// for recover already read body data
		r.Body = io.NopCloser(io.MultiReader(rawBody, r.Body))
	}()
	// check body
	buffer := make([]byte, bll)
	n, err := io.ReadFull(r.Body, buffer)
	if err != nil {
		if n == 0 { // no body
			return 0, nil
		}
		// 0 < data size < bodyLineLength
		nn, err := fmt.Fprintf(w, "\n\n%s", buffer[:n])
		if err != nil {
			return nn, err
		}
		rawBody.Write(buffer[:n])
		return n, nil
	}
	var num int
	// write new line and data
	n, err = fmt.Fprintf(w, "\n\n%s", buffer)
	num += n
	if err != nil {
		return num, err
	}
	rawBody.Write(buffer)
	for {
		// <security> prevent too large resp.Body
		if rawBody.Len() > mbl {
			break
		}
		n, err = io.ReadFull(r.Body, buffer)
		if err != nil {
			// write last line
			if n != 0 {
				nn, err := fmt.Fprintf(w, "\n%s", buffer[:n])
				num += nn
				if err != nil {
					return num, err
				}
				rawBody.Write(buffer[:n])
			}
			break
		}
		n, err = fmt.Fprintf(w, "\n%s", buffer)
		num += n
		if err != nil {
			return num, err
		}
		rawBody.Write(buffer)
	}
	return num, nil
}

// subHTTPFileSystem is used to open sub directory for http file server.
type subHTTPFileSystem struct {
	hfs  http.FileSystem
	path string
}

// NewSubHTTPFileSystem is used to create a new sub http file system.
func NewSubHTTPFileSystem(hfs http.FileSystem, path string) http.FileSystem {
	return &subHTTPFileSystem{hfs: hfs, path: path + "/"}
}

func (sfs *subHTTPFileSystem) Open(name string) (http.File, error) {
	return sfs.hfs.Open(sfs.path + name)
}
