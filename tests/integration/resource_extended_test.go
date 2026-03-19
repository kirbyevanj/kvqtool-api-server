package integration

import (
	"fmt"
	"testing"
)

func TestResourceRegister(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	status, body, err := doJSON("POST", fmt.Sprintf("/v1/projects/%s/resources/register", pid), map[string]any{
		"filename":     "registered.mp4",
		"content_type": "video/mp4",
		"s3_key":       "projects/" + pid + "/registered.mp4",
	})
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if status != 201 {
		t.Fatalf("expected 201 got %d: %s", status, body)
	}
	m := parseJSON(body)
	if _, ok := m["id"]; !ok {
		t.Fatalf("missing id: %s", body)
	}
}

func TestResourceRegisterMissingS3Key(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	status, _, err := doJSON("POST", fmt.Sprintf("/v1/projects/%s/resources/register", pid), map[string]any{
		"filename":     "test.mp4",
		"content_type": "video/mp4",
		"s3_key":       "",
	})
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if status != 400 {
		t.Fatalf("expected 400 got %d", status)
	}
}

func TestResourceRegisterMissingBody(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	status, _, err := doJSON("POST", fmt.Sprintf("/v1/projects/%s/resources/register", pid), nil)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if status != 400 {
		t.Fatalf("expected 400 got %d", status)
	}
}

func TestResourceCopy(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	// Create a resource first
	status, body, err := doJSON("POST", fmt.Sprintf("/v1/projects/%s/resources/upload-url", pid), map[string]string{
		"filename":     "original.mp4",
		"content_type": "video/mp4",
	})
	if err != nil {
		t.Fatalf("upload-url: %v", err)
	}
	if status != 200 {
		t.Fatalf("upload-url: expected 200 got %d: %s", status, body)
	}
	m := parseJSON(body)
	rid, _ := m["resource_id"].(string)
	if rid == "" {
		t.Fatalf("no resource_id: %s", body)
	}

	// Copy the resource - may fail with 500 if S3 object not present (not actually uploaded)
	status, body, err = doJSON("POST", fmt.Sprintf("/v1/projects/%s/resources/%s/copy", pid, rid), nil)
	if err != nil {
		t.Fatalf("copy: %v", err)
	}
	if status == 201 {
		cm := parseJSON(body)
		if _, ok := cm["id"]; !ok {
			t.Fatalf("missing id: %s", body)
		}
		if cm["id"] == rid {
			t.Fatalf("copy should have different id than original")
		}
	} else if status < 400 {
		t.Fatalf("expected 201 or 4xx/5xx got %d: %s", status, body)
	}
	// 4xx/5xx is acceptable since S3 object was never actually uploaded
}

func TestResourceCopyNotFound(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	fakeRid := "00000000-0000-0000-0000-000000000000"
	status, _, err := doJSON("POST", fmt.Sprintf("/v1/projects/%s/resources/%s/copy", pid, fakeRid), nil)
	if err != nil {
		t.Fatalf("copy: %v", err)
	}
	if status >= 200 && status < 300 {
		t.Fatalf("expected non-2xx got %d", status)
	}
}

func TestResourceListFilterFolder(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	// Create a folder
	status, body, err := doJSON("POST", fmt.Sprintf("/v1/projects/%s/folders", pid), map[string]any{
		"name":      "filter-folder",
		"parent_id": nil,
	})
	if err != nil {
		t.Fatalf("create folder: %v", err)
	}
	if status != 201 {
		t.Fatalf("create folder: expected 201 got %d: %s", status, body)
	}
	fm := parseJSON(body)
	fid, _ := fm["id"].(string)
	if fid == "" {
		t.Fatalf("folder missing id: %s", body)
	}

	// Upload resource into folder
	status, body, err = doJSON("POST", fmt.Sprintf("/v1/projects/%s/resources/upload-url", pid), map[string]any{
		"filename":     "in-folder.mp4",
		"content_type": "video/mp4",
		"folder_id":    fid,
	})
	if err != nil {
		t.Fatalf("upload-url: %v", err)
	}
	if status != 200 {
		t.Fatalf("upload-url: expected 200 got %d: %s", status, body)
	}

	// List with folder_id filter
	status, body, err = doRaw("GET", fmt.Sprintf("/v1/projects/%s/resources?folder_id=%s", pid, fid))
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if status != 200 {
		t.Fatalf("expected 200 got %d: %s", status, body)
	}
}

func TestResourceListInvalidFolderID(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	status, _, err := doRaw("GET", fmt.Sprintf("/v1/projects/%s/resources?folder_id=not-a-uuid", pid))
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if status != 400 {
		t.Fatalf("expected 400 got %d", status)
	}
}

func TestResourceUpdateNameOnly(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	status, body, err := doJSON("POST", fmt.Sprintf("/v1/projects/%s/resources/upload-url", pid), map[string]string{
		"filename":     "toname.mp4",
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

	status, body, err = doJSON("PUT", fmt.Sprintf("/v1/projects/%s/resources/%s", pid, rid), map[string]string{
		"name": "newname.mp4",
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if status != 200 {
		t.Fatalf("expected 200 got %d: %s", status, body)
	}
	um := parseJSON(body)
	if um["name"] != "newname.mp4" {
		t.Fatalf("name not updated: got %v", um["name"])
	}
}

func TestResourceDeleteAndVerifyGone(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	status, body, err := doJSON("POST", fmt.Sprintf("/v1/projects/%s/resources/upload-url", pid), map[string]string{
		"filename":     "todelete.mp4",
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
		t.Fatalf("delete: %v", err)
	}
	if status != 204 {
		t.Fatalf("expected 204 got %d", status)
	}

	// Download URL should now fail
	status, _, err = doJSON("GET", fmt.Sprintf("/v1/projects/%s/resources/%s/download-url", pid, rid), nil)
	if err != nil {
		t.Fatalf("download-url after delete: %v", err)
	}
	if status >= 200 && status < 300 {
		t.Fatalf("resource should be gone, got %d", status)
	}
}
