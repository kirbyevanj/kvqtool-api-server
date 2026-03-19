package integration

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestHTMXWorkflowList(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	mustCreateWorkflow(t, pid)

	status, body, ct, err := doRawWithHeaders("GET", fmt.Sprintf("/htmx/projects/%s/workflows-list", pid))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if status != 200 {
		t.Fatalf("expected 200 got %d: %s", status, body)
	}
	if !strings.Contains(ct, "text/html") {
		t.Fatalf("expected text/html content-type got %q", ct)
	}
	if len(body) == 0 {
		t.Fatalf("body empty")
	}
}

func TestHTMXWorkflowListEmptyProject(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	status, body, ct, err := doRawWithHeaders("GET", fmt.Sprintf("/htmx/projects/%s/workflows-list", pid))
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

func TestHTMXWorkflowListInvalidProject(t *testing.T) {
	status, _, _, err := doRawWithHeaders("GET", "/htmx/projects/00000000-0000-0000-0000-000000000000/workflows-list")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	// Should either return 200 (empty) or 404
	if status != 200 && status != 404 {
		t.Fatalf("expected 200 or 404 got %d", status)
	}
}

func TestHTMXSidebarWithResources(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	// Add a folder and resource to populate sidebar
	doJSON("POST", fmt.Sprintf("/v1/projects/%s/folders", pid), map[string]any{
		"name":      "sidebar-folder",
		"parent_id": nil,
	})
	doJSON("POST", fmt.Sprintf("/v1/projects/%s/resources/upload-url", pid), map[string]string{
		"filename":     "sidebar-test.mp4",
		"content_type": "video/mp4",
	})

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

func TestHTMXCreateProjectEmptyName(t *testing.T) {
	status, _, err := doForm("POST", "/htmx/projects", url.Values{"name": {""}})
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	// Should fail with 4xx
	if status < 400 {
		t.Fatalf("expected 4xx for empty name got %d", status)
	}
}

func TestHTMXCreateProjectAndVerifyInList(t *testing.T) {
	projectName := fmt.Sprintf("htmx-verify-%d", time.Now().UnixNano())

	status, body, err := doForm("POST", "/htmx/projects", url.Values{"name": {projectName}})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if status != 200 {
		t.Fatalf("expected 200 got %d: %s", status, body)
	}

	// Verify it appears in the list
	status2, body2, _, err := doRawWithHeaders("GET", "/htmx/projects")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if status2 != 200 {
		t.Fatalf("list: expected 200 got %d", status2)
	}
	if !strings.Contains(string(body2), projectName) {
		t.Fatalf("project %q not found in list: %s", projectName, body2)
	}
}
