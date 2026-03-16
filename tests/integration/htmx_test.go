package integration

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func doRawWithHeaders(method, path string) (int, []byte, string, error) {
	req, err := http.NewRequest(method, baseURL+path, nil)
	if err != nil {
		return 0, nil, "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, body, resp.Header.Get("Content-Type"), nil
}

func TestHTMXListProjects(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	status, body, ct, err := doRawWithHeaders("GET", "/htmx/projects")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if status != 200 {
		t.Fatalf("expected 200 got %d: %s", status, body)
	}
	if !strings.Contains(ct, "text/html") {
		t.Fatalf("expected text/html content-type got %q", ct)
	}
	// Body should contain project info - we created one
	if len(body) == 0 {
		t.Fatalf("body empty")
	}
}

func TestHTMXCreateProject(t *testing.T) {
	status, body, err := doForm("POST", "/htmx/projects", url.Values{"name": {"HTMXTest"}})
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if status != 200 {
		t.Fatalf("expected 200 got %d: %s", status, body)
	}
	if !strings.Contains(string(body), "HTMXTest") {
		t.Fatalf("body should contain HTMXTest: %s", body)
	}
}

func TestHTMXSidebar(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	status, body, ct, err := doRawWithHeaders("GET", fmt.Sprintf("/htmx/projects/%s/sidebar", pid))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if status != 200 {
		t.Fatalf("expected 200 got %d: %s", status, body)
	}
	if !strings.Contains(ct, "text/html") {
		t.Fatalf("expected text/html content-type got %q", ct)
	}
}
