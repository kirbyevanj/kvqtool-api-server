package integration

import (
	"strings"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	status, body, err := doRaw("GET", "/healthz")
	if err != nil {
		t.Fatalf("health check: %v", err)
	}
	if status != 200 {
		t.Fatalf("expected 200 got %d: %s", status, body)
	}
	if !strings.Contains(string(body), "ok") {
		t.Fatalf("body should contain 'ok': %s", body)
	}
}
