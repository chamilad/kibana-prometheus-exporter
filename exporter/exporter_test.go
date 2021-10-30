package exporter

import (
	"testing"
)

func TestNewExporterUnauthenticated(t *testing.T) {
	exporter, err := NewExporter("http://localhost:5601", "", "", "kibana_test", false)
	if err != nil {
		t.Errorf("NewExporter failed with valid input")
	}

	if exporter.collector.authHeader != "" {
		t.Errorf("collector.authHeader found when no auth details were provided: %s", exporter.collector.authHeader)
	}
}

func TestNewExporterUnauthenticatedWithOnlyUsername(t *testing.T) {
	exporter, err := NewExporter("http://localhjost:5601", "someusername", "", "kibana_test", false)
	if err != nil {
		t.Errorf("NewExporter failed with valid input")
	}

	if exporter.collector.authHeader != "" {
		t.Errorf("collector.authHeader found when only username was provided: %s", exporter.collector.authHeader)
	}
}

func TestNewExporterUnauthenticatedWithOnlyPassword(t *testing.T) {
	exporter, err := NewExporter("http://localhjost:5601", "", "somepassword", "kibana_test", false)
	if err != nil {
		t.Errorf("NewExporter failed with valid input")
	}

	if exporter.collector.authHeader != "" {
		t.Errorf("collector.authHeader found when only password was provided: %s", exporter.collector.authHeader)
	}
}

func TestNewExporterWithoutNamespace(t *testing.T) {
	_, err := NewExporter("http://localhost:5601", "", "", "", false)
	if err == nil {
		t.Errorf("expected error when invalid namespace was provided")
	}
}
