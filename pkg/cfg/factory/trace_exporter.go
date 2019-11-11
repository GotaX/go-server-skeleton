// Copyright 2017, OpenCensus Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package factory // import "go.opencensus.io/examples/exporter"

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
)

// indent these many spaces
const indent = "  "

// reZero provides a simple way to detect an empty ID
var reZero = regexp.MustCompile(`^0+$`)

// PrintExporter is a stats and trace exporter that logs
// the exported data to the console.
//
// The intent is help new users familiarize themselves with the
// capabilities of opencensus.
//
// This should NOT be used for production workloads.
type PrintExporter struct {
	logger *logrus.Logger
}

// ExportView logs the view data.
func (e *PrintExporter) ExportView(vd *view.Data) {
	sb := &strings.Builder{}
	for _, row := range vd.Rows {
		fmt.Fprintf(sb, "%v %-45s", vd.End.Format("15:04:05"), vd.View.Name)

		switch v := row.Data.(type) {
		case *view.DistributionData:
			fmt.Fprintf(sb, "distribution: min=%.1f max=%.1f mean=%.1f", v.Min, v.Max, v.Mean)
		case *view.CountData:
			fmt.Fprintf(sb, "count:        value=%v", v.Value)
		case *view.SumData:
			fmt.Fprintf(sb, "sum:          value=%v", v.Value)
		case *view.LastValueData:
			fmt.Fprintf(sb, "last:         value=%v", v.Value)
		}
		fmt.Fprintln(sb)

		for _, tag := range row.Tags {
			fmt.Fprintf(sb, "%v- %v=%v\n", indent, tag.Key.Name(), tag.Value)
		}
	}
	e.logger.Traceln(sb)
}

// ExportSpan logs the trace span.
func (e *PrintExporter) ExportSpan(vd *trace.SpanData) {
	var (
		traceID      = hex.EncodeToString(vd.SpanContext.TraceID[:])
		spanID       = hex.EncodeToString(vd.SpanContext.SpanID[:])
		parentSpanID = hex.EncodeToString(vd.ParentSpanID[:])
	)
	sb := &strings.Builder{}
	fmt.Fprintln(sb)
	fmt.Fprintln(sb, "#----------------------------------------------")
	fmt.Fprintln(sb)
	fmt.Fprintln(sb, "TraceID:     ", traceID)
	fmt.Fprintln(sb, "SpanID:      ", spanID)
	if !reZero.MatchString(parentSpanID) {
		fmt.Fprintln(sb, "ParentSpanID:", parentSpanID)
	}

	fmt.Fprintln(sb)
	fmt.Fprintf(sb, "Span:    %v\n", vd.Name)
	fmt.Fprintf(sb, "Status:  %v [%v]\n", vd.Status.Message, vd.Status.Code)
	fmt.Fprintf(sb, "Elapsed: %v\n", vd.EndTime.Sub(vd.StartTime).Round(time.Millisecond))

	if len(vd.Annotations) > 0 {
		fmt.Fprintln(sb)
		fmt.Fprintln(sb, "Annotations:")
		for _, item := range vd.Annotations {
			fmt.Fprint(sb, indent, item.Message)
			for k, v := range item.Attributes {
				fmt.Fprintf(sb, " %v=%v", k, v)
			}
			fmt.Fprintln(sb)
		}
	}

	if len(vd.Attributes) > 0 {
		fmt.Fprintln(sb)
		fmt.Fprintln(sb, "Attributes:")
		for k, v := range vd.Attributes {
			fmt.Fprintf(sb, "%v- %v=%v\n", indent, k, v)
		}
	}
	logrus.Traceln(sb)
}
