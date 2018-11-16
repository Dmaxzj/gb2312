package gb2312

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/urfave/negroni"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const (
	charsetGb2312 = "gb2312"

	headerAcceptCharset   = "Accept-Charset"
	headerContentEncoding = "Content-Encoding"
	headerContentLength   = "Content-Length"
	headerContentType     = "Content-Type"
	headerVary            = "Vary"
	headerSecWebSocketKey = "Sec-WebSocket-Key"
)

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
	}
	grw.wroteHeader = true
}

// Write writes bytes to the gzip.Writer. It will also set the Content-Type
// header using the net/http library content type detection if the Content-Type
// header was not set yet.
func (grw *gb2312ResponseWriter) Write(b []byte) (int, error) {
	reader := transform.NewReader(bytes.NewReader(b), simplifiedchinese.HZGB2312.NewEncoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return 0, e
	}
	if len(grw.Header().Get(headerContentType)) == 0 {
		grw.Header().Set(headerContentType, strings.Replace(http.DetectContentType(d), "utf-8", "gb2312", 1))
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
	// fmt.Println("middleware ", r.Method, r.Header.Get(headerAcceptCharset))

	if r.Method == "POST" && strings.Contains(r.Header.Get(headerAcceptCharset), charsetGb2312) {
		// fmt.Println("encode")
		r.ParseForm()
		r.PostForm = gb2312decode(r.PostForm)
		r.Form = gb2312decode(r.Form)
	}

	if !strings.Contains(r.Header.Get(headerAcceptCharset), charsetGb2312) {
		// fmt.Println("no gb2312")
		next(w, r)
		return
	}

	nrw := negroni.NewResponseWriter(w)
	grw := newGb2312ResponseWriter(nrw)

	next(grw, r)
}
