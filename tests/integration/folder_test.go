package integration

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestFolderCreateRoot(t *testing.T) {
	pid := mustCreateProject(t)
	t.Cleanup(func() { mustDeleteProject(t, pid) })

	status, body, err := doJSON("POST", "/v1/projects/"+pid+"/folders", map[string]any{
		"name":      fmt.Sprintf("root-%d", time.Now().UnixNano()),
		"parent_id": nil,
	})
	if err != nil {
		t.Fatalf("create folder: %v", err)
	}
	if status != 201 {
		t.Fatalf("expected 201 got %d: %s", status, body)
	}
	m := parseJSON(body)
	if _, ok := m["id"]; !ok {
		t.Fatalf("response missing id: %s", body)
	}
}

func TestFolderCreateNested(t *testing.T) {
	pid := mustCreateProject(t)
	t.Cleanup(func() { mustDeleteProject(t, pid) })

	status, body, err := doJSON("POST", "/v1/projects/"+pid+"/folders", map[string]any{
		"name":      "parent",
		"parent_id": nil,
	})
	if err != nil {
		t.Fatalf("create parent: %v", err)
	}
	if status != 201 {
		t.Fatalf("create parent: expected 201 got %d: %s", status, body)
	}
	pm := parseJSON(body)
	parentID, _ := pm["id"].(string)
	if parentID == "" {
		t.Fatalf("parent missing id: %s", body)
	}

	status, body, err = doJSON("POST", "/v1/projects/"+pid+"/folders", map[string]any{
		"name":      "child",
		"parent_id": parentID,
	})
	if err != nil {
		t.Fatalf("create child: %v", err)
	}
	if status != 201 {
		t.Fatalf("expected 201 got %d: %s", status, body)
	}
	cm := parseJSON(body)
	if _, ok := cm["id"]; !ok {
		t.Fatalf("child missing id: %s", body)
	}
}

func TestFolderCreateEmptyName(t *testing.T) {
	pid := mustCreateProject(t)
	t.Cleanup(func() { mustDeleteProject(t, pid) })

	status, _, err := doJSON("POST", "/v1/projects/"+pid+"/folders", map[string]any{
		"name":      "",
		"parent_id": nil,
	})
	if err != nil {
		t.Fatalf("create folder: %v", err)
	}
	if status != 400 {
		t.Fatalf("expected 400 got %d", status)
	}
}

func TestFolderTree(t *testing.T) {
	pid := mustCreateProject(t)
	t.Cleanup(func() { mustDeleteProject(t, pid) })

	status, body, err := doJSON("POST", "/v1/projects/"+pid+"/folders", map[string]any{
		"name":      "root",
		"parent_id": nil,
	})
	if err != nil {
		t.Fatalf("create root: %v", err)
	}
	if status != 201 {
		t.Fatalf("create root: expected 201 got %d: %s", status, body)
	}
	rm := parseJSON(body)
	rootID, _ := rm["id"].(string)
	if rootID == "" {
		t.Fatalf("root missing id: %s", body)
	}

	status, body, err = doJSON("POST", "/v1/projects/"+pid+"/folders", map[string]any{
		"name":      "nested",
		"parent_id": rootID,
	})
	if err != nil {
		t.Fatalf("create nested: %v", err)
	}
	if status != 201 {
		t.Fatalf("create nested: expected 201 got %d: %s", status, body)
	}

	status, body, err = doRaw("GET", "/v1/projects/"+pid+"/folders")
	if err != nil {
		t.Fatalf("get folders: %v", err)
	}
	if status != 200 {
		t.Fatalf("expected 200 got %d: %s", status, body)
	}
	folders := parseJSONArray(body)
	if folders == nil {
		t.Fatalf("no folders in response: %s", body)
	}
	if !strings.Contains(string(body), "root") || !strings.Contains(string(body), "nested") {
		t.Fatalf("expected nested structure with root and nested: %s", body)
	}
}

func TestFolderUpdate(t *testing.T) {
	pid := mustCreateProject(t)
	t.Cleanup(func() { mustDeleteProject(t, pid) })

	status, body, err := doJSON("POST", "/v1/projects/"+pid+"/folders", map[string]any{
		"name":      "original",
		"parent_id": nil,
	})
	if err != nil {
		t.Fatalf("create folder: %v", err)
	}
	if status != 201 {
		t.Fatalf("create folder: expected 201 got %d: %s", status, body)
	}
	m := parseJSON(body)
	fid, _ := m["id"].(string)
	if fid == "" {
		t.Fatalf("folder missing id: %s", body)
	}

	newName := "renamed"
	status, body, err = doJSON("PUT", "/v1/projects/"+pid+"/folders/"+fid, map[string]any{
		"name":      newName,
		"parent_id": nil,
	})
	if err != nil {
		t.Fatalf("update folder: %v", err)
	}
	if status != 200 {
		t.Fatalf("expected 200 got %d: %s", status, body)
	}
	um := parseJSON(body)
	if um["name"] != newName {
		t.Fatalf("name not updated: got %v", um["name"])
	}
}

func TestFolderDelete(t *testing.T) {
	pid := mustCreateProject(t)
	t.Cleanup(func() { mustDeleteProject(t, pid) })

	status, body, err := doJSON("POST", "/v1/projects/"+pid+"/folders", map[string]any{
		"name":      "to-delete",
		"parent_id": nil,
	})
	if err != nil {
		t.Fatalf("create folder: %v", err)
	}
	if status != 201 {
		t.Fatalf("create folder: expected 201 got %d: %s", status, body)
	}
	m := parseJSON(body)
	fid, _ := m["id"].(string)
	if fid == "" {
		t.Fatalf("folder missing id: %s", body)
	}

	status, _, err = doJSON("DELETE", "/v1/projects/"+pid+"/folders/"+fid, nil)
	if err != nil {
		t.Fatalf("delete folder: %v", err)
	}
	if status != 204 {
		t.Fatalf("expected 204 got %d", status)
	}
}

func TestFolderDeleteResourcesOrphaned(t *testing.T) {
	pid := mustCreateProject(t)
	t.Cleanup(func() { mustDeleteProject(t, pid) })

	status, body, err := doJSON("POST", "/v1/projects/"+pid+"/folders", map[string]any{
		"name":      "resource-folder",
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

	status, body, err = doJSON("POST", "/v1/projects/"+pid+"/resources/upload-url", map[string]any{
		"filename":     "test.txt",
		"content_type": "text/plain",
		"folder_id":    fid,
	})
	if err != nil {
		t.Fatalf("upload-url: %v", err)
	}
	if status != 200 {
		t.Fatalf("upload-url: expected 200 got %d: %s", status, body)
	}
	rm := parseJSON(body)
	rid, _ := rm["resource_id"].(string)
	if rid == "" {
		t.Fatalf("upload-url missing resource_id: %s", body)
	}

	status, _, err = doJSON("DELETE", "/v1/projects/"+pid+"/folders/"+fid, nil)
	if err != nil {
		t.Fatalf("delete folder: %v", err)
	}
	if status != 204 {
		t.Fatalf("expected 204 got %d", status)
	}

	status, body, err = doRaw("GET", "/v1/projects/"+pid+"/resources")
	if err != nil {
		t.Fatalf("get resources: %v", err)
	}
	if status != 200 {
		t.Fatalf("expected 200 got %d: %s", status, body)
	}
	var resources []struct {
		ID       string  `json:"id"`
		FolderID *string `json:"folder_id"`
	}
	if err := json.Unmarshal(body, &resources); err != nil {
		t.Fatalf("parse resources: %v", err)
	}
	var found bool
	for _, r := range resources {
		if r.ID == rid {
			found = true
			if r.FolderID != nil {
				t.Fatalf("resource folder_id should be null after folder delete, got %v", *r.FolderID)
			}
			break
		}
	}
	if !found {
		t.Fatalf("resource %s not found after folder delete: %s", rid, body)
	}
}
