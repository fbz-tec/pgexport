package exporters

import (
	"fmt"
	"sort"
	"strings"
)

type ExporterFactory func() Exporter

var exportersRegistry = map[string]ExporterFactory{}

func RegisterExporter(format string, factory ExporterFactory) error {
	format = strings.ToLower(strings.TrimSpace(format))
	if _, exists := exportersRegistry[format]; exists {
		return fmt.Errorf("exporter: format %q already registered", format)
	}
	exportersRegistry[format] = factory
	return nil
}

func GetExporter(format string) (Exporter, error) {
	factory, ok := exportersRegistry[format]
	if !ok {
		return nil, fmt.Errorf("unsupported format: %q (available: %s)",
			format, strings.Join(ListExporters(), ", "))
	}
	return factory(), nil
}

func ListExporters() []string {
	formats := make([]string, 0, len(exportersRegistry))
	for name := range exportersRegistry {
		formats = append(formats, name)
	}
	sort.Strings(formats)
	return formats
}

func MustRegisterExporter(format string, factory ExporterFactory) {
	if err := RegisterExporter(format, factory); err != nil {
		panic(err)
	}
}
