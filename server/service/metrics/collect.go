package metrics

import (
	"bytes"
	"io"

	log "github.com/sirupsen/logrus"
)

// section is one part of the scrape. Its families belong to it alone.
type section struct {
	name    string
	collect func(*Writer)
}

// sections run in this order on every scrape.
var sections = []section{
	{name: "memory", collect: collectMemory},
	{name: "ion", collect: collectIon},
	{name: "streams", collect: collectStreams},
	{name: "usb", collect: collectUSB},
	{name: "server", collect: collectServer},
	{name: "watchdog", collect: collectWatchdog},
	{name: "node_system", collect: collectNodeSystem},
	{name: "node_memory", collect: collectNodeMemory},
	{name: "node_pressure", collect: collectNodePressure},
	{name: "node_filesystem", collect: collectNodeFilesystem},
	{name: "node_network", collect: collectNodeNetwork},
	{name: "node_disk", collect: collectNodeDisk},
	{name: "node_thermal", collect: collectNodeThermal},
}

// Collect renders every section to out.
//
// Each section renders into its own buffer and is copied to out only when it
// finished cleanly. A section that panics, or writes a family another section
// already wrote, is logged and left out, and the scrape serves the rest. A
// missing source is not an error at all: the section simply writes less.
func Collect(out io.Writer) error {
	seen := make(map[string]bool)

	for _, s := range sections {
		var buf bytes.Buffer
		if !runSection(s, newWriter(&buf, seen)) {
			continue
		}
		if _, err := out.Write(buf.Bytes()); err != nil {
			return err
		}
	}

	return nil
}

func runSection(s section, w *Writer) (ok bool) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("metrics: the %s section panicked and was left out: %v", s.name, r)
			ok = false
		}
	}()

	s.collect(w)
	if err := w.Err(); err != nil {
		log.Errorf("metrics: the %s section was left out: %s", s.name, err)
		return false
	}

	return true
}
