package tracing

import (
	"context"
	"fmt"

	"github.com/rs/xid"
	"go.opencensus.io/trace"
)

type RequestInfo struct {
	RootID string
	ID     string
}

func (info RequestInfo) String() string {
	return fmt.Sprintf("%s-%s", info.RootID, info.ID)
}

func GetRequestInfo(ctx context.Context) (info RequestInfo) {
	if span := trace.FromContext(ctx); span != nil {
		sc := span.SpanContext()
		info.RootID = sc.TraceID.String()
		info.ID = sc.SpanID.String()
	} else {
		id := xid.New().String()
		info.RootID = id
		info.ID = id
	}
	return
}
