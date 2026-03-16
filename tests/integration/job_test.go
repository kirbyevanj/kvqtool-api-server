package integration

import (
	"fmt"
	"testing"
)

func mustCreateJob(t *testing.T, pid, wid string) string {
	t.Helper()
	status, body, err := doJSON("POST", fmt.Sprintf("/v1/projects/%s/jobs", pid), map[string]string{
		"workflow_id": wid,
	})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}
	if status != 202 {
		t.Fatalf("create job: expected 202 got %d: %s", status, body)
	}
	m := parseJSON(body)
	id, _ := m["job_id"].(string)
	if id == "" {
		id, _ = m["id"].(string)
	}
	if id == "" {
		t.Fatalf("create job: no job_id: %s", body)
	}
	return id
}

func TestJobCreate(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	wid := mustCreateWorkflow(t, pid)

	status, body, err := doJSON("POST", fmt.Sprintf("/v1/projects/%s/jobs", pid), map[string]string{
		"workflow_id": wid,
	})
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if status != 202 {
		t.Fatalf("expected 202 got %d: %s", status, body)
	}
	m := parseJSON(body)
	if _, ok := m["job_id"]; !ok {
		if _, ok = m["id"]; !ok {
			t.Fatalf("missing job_id/id: %s", body)
		}
	}
	if _, ok := m["status"]; !ok {
		t.Fatalf("missing status: %s", body)
	}
}

func TestJobCreateInvalidWorkflow(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	fakeWid := "00000000-0000-0000-0000-000000000000"
	status, _, err := doJSON("POST", fmt.Sprintf("/v1/projects/%s/jobs", pid), map[string]string{
		"workflow_id": fakeWid,
	})
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if status >= 200 && status < 300 {
		t.Fatalf("expected non-2xx got %d", status)
	}
}

func TestJobList(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	wid := mustCreateWorkflow(t, pid)
	mustCreateJob(t, pid, wid)

	status, body, err := doJSON("GET", fmt.Sprintf("/v1/projects/%s/jobs", pid), nil)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if status != 200 {
		t.Fatalf("expected 200 got %d: %s", status, body)
	}
	arr := parseJSONArray(body)
	if arr == nil {
		m := parseJSON(body)
		if _, ok := m["jobs"]; !ok {
			t.Fatalf("expected array or jobs key: %s", body)
		}
	}
}

func TestJobGetByID(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	wid := mustCreateWorkflow(t, pid)
	jid := mustCreateJob(t, pid, wid)

	status, body, err := doJSON("GET", fmt.Sprintf("/v1/projects/%s/jobs/%s", pid, jid), nil)
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

func TestJobCancel(t *testing.T) {
	pid := mustCreateProject(t)
	defer mustDeleteProject(t, pid)

	wid := mustCreateWorkflow(t, pid)
	jid := mustCreateJob(t, pid, wid)

	status, body, err := doJSON("POST", fmt.Sprintf("/v1/projects/%s/jobs/%s/cancel", pid, jid), nil)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if status != 200 {
		t.Fatalf("expected 200 got %d: %s", status, body)
	}
	m := parseJSON(body)
	st, _ := m["status"].(string)
	if st != "cancelled" {
		t.Fatalf("expected status cancelled got %q", st)
	}
}
