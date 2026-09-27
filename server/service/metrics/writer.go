// Package metrics serves the board's own state in the Prometheus text
// exposition format, version 0.0.4.
//
// It does not use prometheus/client_golang. The format is a few lines of
// text, and the library brings its own collectors, a registry and several MB
// of binary to a board that has about 110 MB free after boot.
package metrics

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Label is one name="value" pair on a sample.
type Label struct {
	Name  string
	Value string
}

// L builds a Label. It keeps the collectors' call sites on one line.
func L(name, value string) Label {
	return Label{Name: name, Value: value}
}

const (
	typeGauge   = "gauge"
	typeCounter = "counter"
)

var (
	helpEscaper  = strings.NewReplacer(`\`, `\\`, "\n", `\n`)
	labelEscaper = strings.NewReplacer(`\`, `\\`, "\n", `\n`, `"`, `\"`)
)

// Writer renders samples in the text format.
//
// It writes a family's HELP and TYPE lines in front of the family's first
// sample and never again. The format requires a family's lines to be
// contiguous, and a scrape that breaks the rule is rejected whole, so a sample
// for a family that another family has already followed is refused. After the
// first error the writer writes nothing more.
type Writer struct {
	out     io.Writer
	seen    map[string]bool
	current string
	err     error
}

// NewWriter returns a writer that renders to out.
func NewWriter(out io.Writer) *Writer {
	return newWriter(out, make(map[string]bool))
}

// newWriter shares seen with other writers, so a family stays unique across
// sections that Collect renders into separate buffers.
func newWriter(out io.Writer, seen map[string]bool) *Writer {
	return &Writer{out: out, seen: seen}
}

// Gauge writes one sample of a gauge family.
func (w *Writer) Gauge(name, help string, value float64, labels ...Label) {
	w.sample(name, typeGauge, help, value, labels)
}

// Counter writes one sample of a counter family. The name carries its _total.
func (w *Writer) Counter(name, help string, value float64, labels ...Label) {
	w.sample(name, typeCounter, help, value, labels)
}

// Err reports the first error: a family split in two, or a failed write.
func (w *Writer) Err() error {
	return w.err
}

func (w *Writer) sample(name, typ, help string, value float64, labels []Label) {
	if w.err != nil {
		return
	}

	var b strings.Builder

	if name != w.current {
		if w.seen[name] {
			w.err = fmt.Errorf("metrics: family %s is written in two places", name)
			return
		}
		w.seen[name] = true
		w.current = name

		b.WriteString("# HELP " + name + " " + helpEscaper.Replace(help) + "\n")
		b.WriteString("# TYPE " + name + " " + typ + "\n")
	}

	b.WriteString(name)
	if len(labels) > 0 {
		b.WriteByte('{')
		for i, label := range labels {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteString(label.Name)
			b.WriteString(`="`)
			b.WriteString(labelEscaper.Replace(label.Value))
			b.WriteByte('"')
		}
		b.WriteByte('}')
	}
	b.WriteByte(' ')
	// 'f' with -1 precision gives the shortest exact decimal: 0.25, 1099511627776.
	// It also spells the special values the way the format wants: NaN, +Inf, -Inf.
	b.WriteString(strconv.FormatFloat(value, 'f', -1, 64))
	b.WriteByte('\n')

	_, w.err = io.WriteString(w.out, b.String())
}
