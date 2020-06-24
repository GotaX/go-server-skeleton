package tracing

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"net/textproto"
	"strings"

	"go.opencensus.io/trace"
)

const (
	traceHeader = "uber-trace-id"
)

// JaegerFormat implements the TraceContext trace propagation format.
type JaegerFormat struct{}

func (f *JaegerFormat) SpanContextFromRequest(req *http.Request) (sc trace.SpanContext, ok bool) {
	v, ok := getRequestHeader(req, traceHeader, false)
	if !ok {
		return trace.SpanContext{}, false
	}

	sections := strings.Split(v, ":")
	if len(sections) < 4 {
		return trace.SpanContext{}, false
	}

	if len(sections[0]) < 16 {
		return trace.SpanContext{}, false
	}
	tid, err := hex.DecodeString(sections[0])
	if err != nil {
		return trace.SpanContext{}, false
	}
	if len(tid) == 8 {
		copy(sc.TraceID[8:], tid)
	} else {
		copy(sc.TraceID[:], tid)
	}

	if len(sections[1]) != 16 {
		return trace.SpanContext{}, false
	}
	sid, err := hex.DecodeString(sections[1])
	if err != nil {
		return trace.SpanContext{}, false
	}
	copy(sc.SpanID[:], sid)

	if len(sections[3]) == 1 {
		sections[3] = "0" + sections[3]
	}
	opts, err := hex.DecodeString(sections[3])
	if err != nil || len(opts) < 1 {
		return trace.SpanContext{}, false
	}
	sc.TraceOptions = trace.TraceOptions(opts[0])

	// Don't allow all zero trace or span ID.
	if sc.TraceID == [16]byte{} || sc.SpanID == [8]byte{} {
		return trace.SpanContext{}, false
	}
	return sc, true
}

func getRequestHeader(req *http.Request, name string, commaSeparated bool) (hdr string, ok bool) {
	v := req.Header[textproto.CanonicalMIMEHeaderKey(name)]
	switch len(v) {
	case 0:
		return "", false
	case 1:
		return v[0], true
	default:
		return strings.Join(v, ","), commaSeparated
	}
}

func (f *JaegerFormat) SpanContextToRequest(sc trace.SpanContext, req *http.Request) {
	h := fmt.Sprintf("%x:%x:0:%x",
		sc.TraceID[:],
		sc.SpanID[:],
		sc.TraceOptions)
	h = strings.TrimLeft(h, "0000000000000000")
	req.Header.Set(traceHeader, h)
}
