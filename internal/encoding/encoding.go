package encoding

import (
	"errors"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"io"
	"net/http"
)

var availableEncodings = map[string]*encoding.Encoder{
	"utf-8":      nil,
	"iso-8859-1": charmap.ISO8859_1.NewEncoder(),
}

var ErrUnknownEncoding = errors.New("unknown encoding")

type ResponseWriter struct {
	rw   http.ResponseWriter
	next io.Writer
}

func NewResponseWriter(rw http.ResponseWriter, encoding string) (*ResponseWriter, error) {
	encoder, ok := availableEncodings[encoding]
	if !ok {
		return nil, ErrUnknownEncoding
	}

	var next io.Writer

	if encoder != nil {
		next = encoder.Writer(rw)
	} else {
		next = rw
	}

	return &ResponseWriter{
		rw:   rw,
		next: next,
	}, nil
}

func (r *ResponseWriter) Header() http.Header {
	return r.rw.Header()
}

func (r *ResponseWriter) Write(b []byte) (int, error) {
	return r.next.Write(b)
}

func (r *ResponseWriter) WriteHeader(statusCode int) {
	r.rw.WriteHeader(statusCode)
}
