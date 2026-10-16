package middleware

import (
	"bytes"
	"context"
	"net/http"
	"runtime"
	"time"

	"github.com/zigzagapp-org/zigzag-api/internal/common/log"
	"github.com/zigzagapp-org/zigzag-api/internal/common/utils"
)

type contextKey struct {
	name string
}

var (
	LogEntryCtxKey = &contextKey{"LogEntry"}

	DefaultLogger func(next http.Handler) http.Handler
)

func Logger(next http.Handler) http.Handler {
	return DefaultLogger(next)
}

func RequestLogger(f LogFormatter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			entry := f.NewLogEntry(r)
			ww := NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				entry.Write(ww.Status(), ww.BytesWritten(), ww.Header(), time.Since(t1), nil)
			}()

			next.ServeHTTP(ww, WithLogEntry(r, entry))
		}
		return http.HandlerFunc(fn)
	}
}

type LogFormatter interface {
	NewLogEntry(r *http.Request) LogEntry
}

type LogEntry interface {
	Write(status, bytes int, header http.Header, elapsed time.Duration, extra any)
}

func WithLogEntry(r *http.Request, entry LogEntry) *http.Request {
	r = r.WithContext(context.WithValue(r.Context(), LogEntryCtxKey, entry))
	return r
}

type DefaultLogFormatter struct {
	NoColor bool
}

func (l *DefaultLogFormatter) NewLogEntry(r *http.Request) LogEntry {
	useColor := !l.NoColor
	entry := &defaultLogEntry{
		DefaultLogFormatter: l,
		request:             r,
		buf:                 &bytes.Buffer{},
		useColor:            useColor,
	}

	reqID := GetReqID(r.Context())
	if reqID != "" {
		cW(entry.buf, useColor, nYellow, "[%s] ", reqID)
	}
	cW(entry.buf, useColor, nCyan, "\"")
	cW(entry.buf, useColor, bMagenta, "%s ", r.Method)

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	cW(entry.buf, useColor, nCyan, "%s://%s%s %s\" ", scheme, r.Host, r.RequestURI, r.Proto)

	entry.buf.WriteString("from ")
	entry.buf.WriteString(utils.GetIPAddress(r))
	entry.buf.WriteString(" - ")

	return entry
}

type defaultLogEntry struct {
	*DefaultLogFormatter
	request  *http.Request
	buf      *bytes.Buffer
	useColor bool
}

func (l *defaultLogEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra any) {
	switch {
	case status < 200:
		cW(l.buf, l.useColor, bBlue, "%03d", status)
	case status < 300:
		cW(l.buf, l.useColor, bGreen, "%03d", status)
	case status < 400:
		cW(l.buf, l.useColor, bCyan, "%03d", status)
	case status < 500:
		cW(l.buf, l.useColor, bYellow, "%03d", status)
	default:
		cW(l.buf, l.useColor, bRed, "%03d", status)
	}

	cW(l.buf, l.useColor, bBlue, " %dB", bytes)

	l.buf.WriteString(" in ")

	elapsedTrunc := elapsed.Truncate(time.Millisecond)

	if elapsed < 500*time.Millisecond {
		cW(l.buf, l.useColor, nGreen, "%s", elapsedTrunc)
	} else if elapsed < 5*time.Second {
		cW(l.buf, l.useColor, nYellow, "%s", elapsedTrunc)
	} else {
		cW(l.buf, l.useColor, nRed, "%s", elapsedTrunc)
	}

	log.Infof("%s", l.buf.String())
}

func init() {
	useColor := runtime.GOOS != "windows"
	DefaultLogger = RequestLogger(&DefaultLogFormatter{NoColor: !useColor})
}
