package integration

import (
	"fmt"
	"testing"
)

func mustCreateWorkflow(t *testing.T, pid string) string {
	t.Helper()
	dag := map[string]any{
		"drawflow": map[string]any{
			"Home": map[string]any{"data": map[string]any{}},
		},
	}
	status, body, err := doJSON("POST", fmt.Sprintf("/v1/projects/%s/workflows", pid), map[string]any{
		"name":     fmt.Sprintf("wf-%d", 0),
		"dag_json": dag,
	})
	if err != nil {
		t.Fatalf("create workflow: %v", err)
	}
	if status != 201 {
		t.Fatalf("create workflow: expected 201 got %d: %s", status, body)
	}
	m := parseJSON(body)
	id, _ := m["id"].(string)
	if id == "" {
		t.Fatalf("create workflow: no id: %s", body)
	}
	return id
}

func TestWorkflowCreate(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	dag := map[string]any{
		"drawflow": map[string]any{
			"Home": map[string]any{"data": map[string]any{}},
		},
	}
	status, body, err := doJSON("POST", fmt.Sprintf("/v1/projects/%s/workflows", pid), map[string]any{
		"name":     "CRF Encode",
		"dag_json": dag,
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

func TestWorkflowCreateEmptyName(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	dag := map[string]any{
		"drawflow": map[string]any{
			"Home": map[string]any{"data": map[string]any{}},
		},
	}
	status, _, err := doJSON("POST", fmt.Sprintf("/v1/projects/%s/workflows", pid), map[string]any{
		"name":     "",
		"dag_json": dag,
	})
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if status != 400 {
		t.Fatalf("expected 400 got %d", status)
	}
}

func TestWorkflowList(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	mustCreateWorkflow(t, pid)

	status, body, err := doJSON("GET", fmt.Sprintf("/v1/projects/%s/workflows", pid), nil)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if status != 200 {
		t.Fatalf("expected 200 got %d: %s", status, body)
	}
	arr := parseJSONArray(body)
	if arr == nil {
		m := parseJSON(body)
		if wfs, ok := m["workflows"]; ok {
			if _, ok := wfs.([]any); !ok {
				t.Fatalf("workflows not array: %s", body)
			}
		} else {
			t.Fatalf("expected array or workflows key: %s", body)
		}
	}
}

func TestWorkflowGetByID(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	wid := mustCreateWorkflow(t, pid)

	status, body, err := doJSON("GET", fmt.Sprintf("/v1/projects/%s/workflows/%s", pid, wid), nil)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if status != 200 {
		t.Fatalf("expected 200 got %d: %s", status, body)
	}
	m := parseJSON(body)
	if _, ok := m["id"]; !ok {
		t.Fatalf("missing id: %s", body)
	}
}

func TestWorkflowUpdate(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	wid := mustCreateWorkflow(t, pid)

	status, _, err := doJSON("PUT", fmt.Sprintf("/v1/projects/%s/workflows/%s", pid, wid), map[string]string{
		"name": "Updated Workflow",
	})
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if status != 200 {
		t.Fatalf("expected 200 got %d", status)
	}
}

func TestWorkflowDelete(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	wid := mustCreateWorkflow(t, pid)

	status, _, err := doJSON("DELETE", fmt.Sprintf("/v1/projects/%s/workflows/%s", pid, wid), nil)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if status != 204 {
		t.Fatalf("expected 204 got %d", status)
	}
}
