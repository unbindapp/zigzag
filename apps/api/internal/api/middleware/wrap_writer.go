package middleware

import (
	"bufio"
	"io"
	"net"
	"net/http"
)

func NewWrapResponseWriter(w http.ResponseWriter, protoMajor int) WrapResponseWriter {
	_, fl := w.(http.Flusher)

	bw := basicWriter{ResponseWriter: w}

	if protoMajor == 2 {
		_, ps := w.(http.Pusher)
		if fl && ps {
			return &http2FancyWriter{bw}
		}
	} else {
		_, hj := w.(http.Hijacker)
		_, rf := w.(io.ReaderFrom)
		if fl && hj && rf {
			return &httpFancyWriter{bw}
		}
		if fl && hj {
			return &flushHijackWriter{bw}
		}
		if hj {
			return &hijackWriter{bw}
		}
	}

	if fl {
		return &flushWriter{bw}
	}

	return &bw
}

type WrapResponseWriter interface {
	http.ResponseWriter
	Status() int
	BytesWritten() int
	Tee(io.Writer)
	Unwrap() http.ResponseWriter
}

type basicWriter struct {
	http.ResponseWriter
	wroteHeader bool
	code        int
	bytes       int
	tee         io.Writer
}

func (b *basicWriter) WriteHeader(code int) {
	if !b.wroteHeader {
		b.code = code
		b.wroteHeader = true
		b.ResponseWriter.WriteHeader(code)
	}
}

func (b *basicWriter) Write(buf []byte) (int, error) {
	b.maybeWriteHeader()
	n, err := b.ResponseWriter.Write(buf)
	if b.tee != nil {
		_, err2 := b.tee.Write(buf[:n])
		if err == nil {
			err = err2
		}
	}
	b.bytes += n
	return n, err
}

func (b *basicWriter) maybeWriteHeader() {
	if !b.wroteHeader {
		b.WriteHeader(http.StatusOK)
	}
}

func (b *basicWriter) Status() int {
	return b.code
}

func (b *basicWriter) BytesWritten() int {
	return b.bytes
}

func (b *basicWriter) Tee(w io.Writer) {
	b.tee = w
}

func (b *basicWriter) Unwrap() http.ResponseWriter {
	return b.ResponseWriter
}

type flushWriter struct {
	basicWriter
}

func (f *flushWriter) Flush() {
	f.wroteHeader = true
	fl := f.ResponseWriter.(http.Flusher)
	fl.Flush()
}

var _ http.Flusher = &flushWriter{}

type hijackWriter struct {
	basicWriter
}

func (f *hijackWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj := f.ResponseWriter.(http.Hijacker)
	return hj.Hijack()
}

var _ http.Hijacker = &hijackWriter{}

type flushHijackWriter struct {
	basicWriter
}

func (f *flushHijackWriter) Flush() {
	f.wroteHeader = true
	fl := f.ResponseWriter.(http.Flusher)
	fl.Flush()
}

func (f *flushHijackWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj := f.ResponseWriter.(http.Hijacker)
	return hj.Hijack()
}

var _ http.Flusher = &flushHijackWriter{}
var _ http.Hijacker = &flushHijackWriter{}

type httpFancyWriter struct {
	basicWriter
}

func (f *httpFancyWriter) Flush() {
	f.wroteHeader = true
	fl := f.ResponseWriter.(http.Flusher)
	fl.Flush()
}

func (f *httpFancyWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj := f.ResponseWriter.(http.Hijacker)
	return hj.Hijack()
}

func (f *http2FancyWriter) Push(target string, opts *http.PushOptions) error {
	return f.ResponseWriter.(http.Pusher).Push(target, opts)
}

func (f *httpFancyWriter) ReadFrom(r io.Reader) (int64, error) {
	if f.tee != nil {
		n, err := io.Copy(&f.basicWriter, r)
		f.bytes += int(n)
		return n, err
	}
	rf := f.ResponseWriter.(io.ReaderFrom)
	f.maybeWriteHeader()
	n, err := rf.ReadFrom(r)
	f.bytes += int(n)
	return n, err
}

var _ http.Flusher = &httpFancyWriter{}
var _ http.Hijacker = &httpFancyWriter{}
var _ http.Pusher = &http2FancyWriter{}
var _ io.ReaderFrom = &httpFancyWriter{}

type http2FancyWriter struct {
	basicWriter
}

func (f *http2FancyWriter) Flush() {
	f.wroteHeader = true
	fl := f.ResponseWriter.(http.Flusher)
	fl.Flush()
}

var _ http.Flusher = &http2FancyWriter{}
