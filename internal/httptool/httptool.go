package httptool

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

const (
	defaultBodyLineLength = 64   // default post data length in one line.
	defaultMaxBodyLength  = 1024 // default maximum body length.
)

// FdumpRequest is used to dump http request to a io.Writer.
func FdumpRequest(w io.Writer, r *http.Request) (int, error) {
	return FdumpRequestWithBM(w, r, defaultBodyLineLength, defaultMaxBodyLength)
}

// SdumpRequest is used to dump http request to a string.
func SdumpRequest(r *http.Request) string {
	return SdumpRequestWithBM(r, defaultBodyLineLength, defaultMaxBodyLength)
}

// DumpRequest is used to dump http request to os.Stdout.
func DumpRequest(r *http.Request) {
	DumpRequestWithBM(r, defaultBodyLineLength, defaultMaxBodyLength)
}

// FdumpRequestWithBM is used to dump http request to a io.Writer.
// bll is the body line length, mbl is the max body length.
func FdumpRequestWithBM(w io.Writer, r *http.Request, bll, mbl int) (int, error) {
	return dumpRequest(w, r, bll, mbl)
}

// SdumpRequestWithBM is used to dump http request to a string.
// bll is the body line length, mbl is the max body length.
func SdumpRequestWithBM(r *http.Request, bll, mbl int) string {
	buf := bytes.NewBuffer(make([]byte, 0, 512))
	_, _ = dumpRequest(buf, r, bll, mbl)
	return buf.String()
}

// DumpRequestWithBM is used to dump http request to os.Stdout.
// bll is the body line length, mbl is the max body length.
func DumpRequestWithBM(r *http.Request, bll, mbl int) {
	buf := bytes.NewBuffer(make([]byte, 0, 512))
	_, _ = dumpRequest(buf, r, bll, mbl)
	buf.WriteString("\n")
	_, _ = os.Stdout.Write(buf.Bytes())
}

// dumpRequest is used to dump http request to io.Writer.
// bll is the body line length, mbl is the max body size.
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
func dumpRequest(w io.Writer, r *http.Request, bll, mbl int) (int, error) {
	n, err := fmt.Fprintf(w, "Remote: %s\n", r.RemoteAddr)
	if err != nil {
		return n, err
	}
	var nn int
	// request
	nn, err = fmt.Fprintf(w, "%s %s %s", r.Method, r.RequestURI, r.Proto)
	if err != nil {
		return n + nn, err
	}
	n += nn
	// dump host
	nn, err = fmt.Fprintf(w, "\nHost: %s", r.Host)
	if err != nil {
		return n + nn, err
	}
	n += nn
	// dump header
	for k, v := range r.Header {
		nn, err = fmt.Fprintf(w, "\n%s: %s", k, v[0])
		if err != nil {
			return n + nn, err
		}
		n += nn
	}
	if r.Body == nil {
		return n, nil
	}
	nn, err = dumpBody(w, r, bll, mbl)
	return n + nn, err
}

// dumpBody is used to dump http response body to io.Writer.
func dumpBody(w io.Writer, r *http.Request, bll, mbl int) (int, error) {
	rawBody := new(bytes.Buffer)
	defer func() { r.Body = ioutil.NopCloser(io.MultiReader(rawBody, r.Body)) }()
	var (
		total int
		err   error
	)
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
	// new line and write data
	n, err = fmt.Fprintf(w, "\n\n%s", buffer)
	if err != nil {
		return n, err
	}
	total += n
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
				if err != nil {
					return total + nn, err
				}
				rawBody.Write(buffer[:n])
			}
			break
		}
		n, err = fmt.Fprintf(w, "\n%s", buffer)
		if err != nil {
			return total + n, err
		}
		total += n
		rawBody.Write(buffer)
	}
	return total, nil
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
