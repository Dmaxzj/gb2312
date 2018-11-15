// Package gzip implements a gzip compression handler middleware for Negroni.
package gb2312

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/urfave/negroni"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// These compression constants are copied from the compress/gzip package.
const (
	charsetGb2312 = "gb2312"

	headerAcceptCharset   = "Accept-Charset"
	headerContentEncoding = "Content-Encoding"
	headerContentLength   = "Content-Length"
	headerContentType     = "Content-Type"
	headerVary            = "Vary"
	headerSecWebSocketKey = "Sec-WebSocket-Key"

	BestCompression    = gzip.BestCompression
	BestSpeed          = gzip.BestSpeed
	DefaultCompression = gzip.DefaultCompression
	NoCompression      = gzip.NoCompression
)

// gzipResponseWriter is the ResponseWriter that negroni.ResponseWriter is
// wrapped in.
type gb2312ResponseWriter struct {
	w *transform.Writer
	negroni.ResponseWriter
	wroteHeader bool
}

// Check whether underlying response is already pre-encoded and disable
// gzipWriter before the body gets written, otherwise encoding headers
func (grw *gb2312ResponseWriter) WriteHeader(code int) {
	headers := grw.ResponseWriter.Header()
	if headers.Get(headerAcceptCharset) == "" {
		headers.Set(headerAcceptCharset, charsetGb2312)
		headers.Add(headerVary, headerAcceptCharset)
	} else {
		grw.w = nil
	}

	// Avoid sending Content-Length header before compression. The length would
	// be invalid, and some browsers like Safari will report
	// "The network connection was lost." errors
	grw.Header().Del(headerContentLength)

	grw.ResponseWriter.WriteHeader(code)
	grw.wroteHeader = true
}

// Write writes bytes to the gzip.Writer. It will also set the Content-Type
// header using the net/http library content type detection if the Content-Type
// header was not set yet.
func (grw *gb2312ResponseWriter) Write(b []byte) (int, error) {
	reader := transform.NewReader(bytes.NewReader(b), simplifiedchinese.HZGB2312.NewEncoder())
	d, e := ioutil.ReadAll(reader)
	if len(grw.Header().Get(headerContentType)) == 0 {
		grw.Header().Set(headerContentType, http.DetectContentType(d))
	}
	if e != nil {
		return 0, e
	}
	_, err := grw.ResponseWriter.Write(d)
	return len(b), err
}

func newGb2312ResponseWriter(rw negroni.ResponseWriter) negroni.ResponseWriter {
	wr := &gb2312ResponseWriter{ResponseWriter: rw}
	return wr
}

type Gb2312Encode struct {
	gb2312ResponseWriter
}

type gb2312RequestReader struct {
}

func copyValues(dst, src url.Values) {
	for k, vs := range src {
		for _, value := range vs {
			dst.Add(k, value)
		}
	}
}

func gb2312decode(src url.Values) url.Values {
	t := make(url.Values)
	for k, vs := range src {
		reader := transform.NewReader(strings.NewReader(k), simplifiedchinese.HZGB2312.NewDecoder())
		s1, err1 := ioutil.ReadAll(reader)
		if err1 != nil {
			panic(err1)
		}

		for _, v := range vs {
			reader = transform.NewReader(strings.NewReader(v), simplifiedchinese.HZGB2312.NewDecoder())
			s2, err2 := ioutil.ReadAll(reader)
			if err2 != nil {
				panic(err2)
			}
			if t.Get(string(s1)) != "" {
				t.Add(string(s1), string(s2))
			} else {
				t.Set(string(s1), string(s2))
			}
		}
	}
	return t
}

// ServeHTTP wraps the http.ResponseWriter with a gzip.Writer.
func (h *Gb2312Encode) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	fmt.Println("middleware ", r.Method, r.Header.Get(headerAcceptCharset))

	if r.Method == "POST" && strings.Contains(r.Header.Get(headerAcceptCharset), charsetGb2312) {
		fmt.Println("encode")
		r.ParseForm()
		r.PostForm = gb2312decode(r.PostForm)
		r.Form = gb2312decode(r.Form)
	}

	// t1, e1 := ioutil.ReadAll(r.Body)
	// if e1 != nil {
	// 	panic(e1)
	// }

	// reader := transform.NewReader(bytes.NewReader(t1), simplifiedchinese.HZGB2312.NewDecoder())
	// b, e := ioutil.ReadAll(reader)
	// // fmt.Println("decode body ", string(b))
	// if e != nil {
	// 	panic(e)
	// }

	// r.Body = ioutil.NopCloser(bytes.NewReader(b))

	// Skip compression if the client doesn't accept gzip encoding.
	if !strings.Contains(r.Header.Get(headerAcceptCharset), charsetGb2312) {
		fmt.Println("no gb2312")
		next(w, r)
		return
	}

	// Wrap the original http.ResponseWriter with negroni.ResponseWriter
	// and create the gzipResponseWriter.
	nrw := negroni.NewResponseWriter(w)
	grw := newGb2312ResponseWriter(nrw)

	// Call the next handler supplying the gzipResponseWriter instead of
	// the original.
	next(grw, r)
}
