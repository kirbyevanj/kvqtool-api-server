package integration

import (
	"fmt"
	"testing"
	"time"
)

func TestProjectCreate(t *testing.T) {
	name := fmt.Sprintf("test-%d", time.Now().UnixNano())
	status, body, err := doJSON("POST", "/v1/projects", map[string]string{
		"name":        name,
		"description": "integration test",
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if status != 201 {
		t.Fatalf("expected 201 got %d: %s", status, body)
	}
	m := parseJSON(body)
	if _, ok := m["id"]; !ok {
		t.Fatalf("response missing id: %s", body)
	}
}

func TestProjectCreateEmptyName(t *testing.T) {
	status, _, err := doJSON("POST", "/v1/projects", map[string]string{
		"name":        "",
		"description": "desc",
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if status != 400 {
		t.Fatalf("expected 400 got %d", status)
	}
}

func TestProjectCreateMissingBody(t *testing.T) {
	status, _, err := doJSON("POST", "/v1/projects", nil)
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if status != 400 {
		t.Fatalf("expected 400 got %d", status)
	}
}

func TestProjectList(t *testing.T) {
	pid := mustCreateProject(t)
	t.Cleanup(func() { mustDeleteProject(t, pid) })

	status, body, err := doRaw("GET", "/v1/projects")
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}
	if status != 200 {
		t.Fatalf("expected 200 got %d: %s", status, body)
	}
	projects := parseJSONArray(body)
	if projects == nil {
		t.Fatalf("no projects in response: %s", body)
	}
	found := false
	for _, p := range projects {
		pm, ok := p.(map[string]any)
		if !ok {
			continue
		}
		if pm["id"] == pid {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("project %s not in list: %s", pid, body)
	}
}

func TestProjectGetByID(t *testing.T) {
	pid := mustCreateProject(t)
	t.Cleanup(func() { mustDeleteProject(t, pid) })

	status, body, err := doRaw("GET", "/v1/projects/"+pid)
	if err != nil {
		t.Fatalf("get project: %v", err)
	}
	if status != 200 {
		t.Fatalf("expected 200 got %d: %s", status, body)
	}
	m := parseJSON(body)
	if m["id"] != pid {
		t.Fatalf("id mismatch: got %v", m["id"])
	}
	if m["name"] == nil || m["name"] == "" {
		t.Fatalf("name missing: %s", body)
	}
}

func TestProjectGetNotFound(t *testing.T) {
	status, _, err := doRaw("GET", "/v1/projects/00000000-0000-0000-0000-000000000099")
	if err != nil {
		t.Fatalf("get project: %v", err)
	}
	if status >= 200 && status < 300 {
		t.Fatalf("expected non-2xx got %d", status)
	}
}

func TestProjectGetInvalidUUID(t *testing.T) {
	status, _, err := doRaw("GET", "/v1/projects/not-a-uuid")
	if err != nil {
		t.Fatalf("get project: %v", err)
	}
	if status != 400 && status != 500 {
		t.Fatalf("expected 400 or 500 got %d", status)
	}
}

func TestProjectUpdate(t *testing.T) {
	pid := mustCreateProject(t)
	t.Cleanup(func() { mustDeleteProject(t, pid) })

	newName := fmt.Sprintf("updated-%d", time.Now().UnixNano())
	status, body, err := doJSON("PUT", "/v1/projects/"+pid, map[string]string{
		"name":        newName,
		"description": "updated desc",
	})
	if err != nil {
		t.Fatalf("update project: %v", err)
	}
	if status != 200 {
		t.Fatalf("expected 200 got %d: %s", status, body)
	}
	m := parseJSON(body)
	if m["name"] != newName {
		t.Fatalf("name not updated: got %v", m["name"])
	}
}

func TestProjectDelete(t *testing.T) {
	pid := mustCreateProject(t)

	status, _, err := doJSON("DELETE", "/v1/projects/"+pid, nil)
	if err != nil {
		t.Fatalf("delete project: %v", err)
	}
	if status != 204 {
		t.Fatalf("expected 204 got %d", status)
	}

	status, _, _ = doRaw("GET", "/v1/projects/"+pid)
	if status >= 200 && status < 300 {
		t.Fatalf("project should be gone, got %d", status)
	}
}

func TestProjectDeleteCascade(t *testing.T) {
	pid := mustCreateProject(t)

	status, body, err := doJSON("POST", "/v1/projects/"+pid+"/folders", map[string]any{
		"name":      "cascade-folder",
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

	status, body, err = doJSON("POST", "/v1/projects/"+pid+"/workflows", map[string]any{
		"name":         "cascade-workflow",
		"dag_json":     map[string]any{"drawflow": map[string]any{"Home": map[string]any{"data": map[string]any{}}}},
		"input_schema": map[string]any{},
	})
	if err != nil {
		t.Fatalf("create workflow: %v", err)
	}
	if status != 201 {
		t.Fatalf("create workflow: expected 201 got %d: %s", status, body)
	}

	status, _, err = doJSON("DELETE", "/v1/projects/"+pid, nil)
	if err != nil {
		t.Fatalf("delete project: %v", err)
	}
	if status != 204 {
		t.Fatalf("expected 204 got %d", status)
	}

	status, body, _ = doRaw("GET", "/v1/projects/"+pid)
	if status == 200 {
		t.Fatalf("project should be gone, got 200: %s", body)
	}
}
