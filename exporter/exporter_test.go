package exporter

import (
	"testing"
)

func TestNewExporterWithoutNamespace(t *testing.T) {
	_, err := NewExporter("", &KibanaCollector{})
	if err == nil {
		t.Errorf("expected error when invalid namespace was provided")
	}
}
