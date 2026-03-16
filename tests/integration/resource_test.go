package integration

import (
	"fmt"
	"testing"
)

func TestResourceUploadURL(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	status, body, err := doJSON("POST", fmt.Sprintf("/v1/projects/%s/resources/upload-url", pid), map[string]string{
		"filename":     "test.mp4",
		"content_type": "video/mp4",
	})
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if status != 200 {
		t.Fatalf("expected 200 got %d: %s", status, body)
	}
	m := parseJSON(body)
	if _, ok := m["upload_url"]; !ok {
		t.Fatalf("missing upload_url: %s", body)
	}
	if _, ok := m["resource_id"]; !ok {
		t.Fatalf("missing resource_id: %s", body)
	}
}

func TestResourceUploadURLMissingFilename(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	status, _, err := doJSON("POST", fmt.Sprintf("/v1/projects/%s/resources/upload-url", pid), map[string]string{
		"filename":     "",
		"content_type": "video/mp4",
	})
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if status != 400 {
		t.Fatalf("expected 400 got %d", status)
	}
}

func TestResourceList(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	_, _, err := doJSON("POST", fmt.Sprintf("/v1/projects/%s/resources/upload-url", pid), map[string]string{
		"filename":     "list.mp4",
		"content_type": "video/mp4",
	})
	if err != nil {
		t.Fatalf("upload-url: %v", err)
	}

	status, body, err := doJSON("GET", fmt.Sprintf("/v1/projects/%s/resources", pid), nil)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if status != 200 {
		t.Fatalf("expected 200 got %d: %s", status, body)
	}
	if status != 200 {
		t.Fatalf("expected 200 for filtered list: %s", body)
	}
}

func TestResourceListFilterType(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	status, body, err := doJSON("GET", fmt.Sprintf("/v1/projects/%s/resources?type=media", pid), nil)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if status != 200 {
		t.Fatalf("expected 200 got %d: %s", status, body)
	}
	if status != 200 {
		t.Fatalf("expected 200 for filtered list: %s", body)
	}
}

func TestResourceDownloadURL(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	status, body, err := doJSON("POST", fmt.Sprintf("/v1/projects/%s/resources/upload-url", pid), map[string]string{
		"filename":     "dl.mp4",
		"content_type": "video/mp4",
	})
	if err != nil {
		t.Fatalf("upload-url: %v", err)
	}
	if status != 200 {
		t.Fatalf("upload-url: %d: %s", status, body)
	}
	m := parseJSON(body)
	rid, _ := m["resource_id"].(string)
	if rid == "" {
		t.Fatalf("no resource_id: %s", body)
	}

	status, body, err = doJSON("GET", fmt.Sprintf("/v1/projects/%s/resources/%s/download-url", pid, rid), nil)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if status != 200 {
		t.Fatalf("expected 200 got %d: %s", status, body)
	}
	m = parseJSON(body)
	if _, ok := m["download_url"]; !ok {
		t.Fatalf("missing download_url: %s", body)
	}
}

func TestResourceUpdate(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	status, body, err := doJSON("POST", fmt.Sprintf("/v1/projects/%s/resources/upload-url", pid), map[string]string{
		"filename":     "orig.mp4",
		"content_type": "video/mp4",
	})
	if err != nil {
		t.Fatalf("upload-url: %v", err)
	}
	if status != 200 {
		t.Fatalf("upload-url: %d: %s", status, body)
	}
	m := parseJSON(body)
	rid, _ := m["resource_id"].(string)
	if rid == "" {
		t.Fatalf("no resource_id: %s", body)
	}

	status, _, err = doJSON("PUT", fmt.Sprintf("/v1/projects/%s/resources/%s", pid, rid), map[string]string{
		"name": "renamed.mp4",
	})
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if status != 200 {
		t.Fatalf("expected 200 got %d", status)
	}
}

func TestResourceDelete(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	status, body, err := doJSON("POST", fmt.Sprintf("/v1/projects/%s/resources/upload-url", pid), map[string]string{
		"filename":     "del.mp4",
		"content_type": "video/mp4",
	})
	if err != nil {
		t.Fatalf("upload-url: %v", err)
	}
	if status != 200 {
		t.Fatalf("upload-url: %d: %s", status, body)
	}
	m := parseJSON(body)
	rid, _ := m["resource_id"].(string)
	if rid == "" {
		t.Fatalf("no resource_id: %s", body)
	}

	status, _, err = doJSON("DELETE", fmt.Sprintf("/v1/projects/%s/resources/%s", pid, rid), nil)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if status != 204 {
		t.Fatalf("expected 204 got %d", status)
	}
}

func TestResourceConfirmUploadNotFound(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	fakeRid := "00000000-0000-0000-0000-000000000000"
	status, _, err := doJSON("POST", fmt.Sprintf("/v1/projects/%s/resources/%s/confirm-upload", pid, fakeRid), nil)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if status >= 200 && status < 300 {
		t.Fatalf("expected non-2xx got %d", status)
	}
}
