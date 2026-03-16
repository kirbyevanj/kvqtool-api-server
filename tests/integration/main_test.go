package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

var baseURL string
var client *http.Client

func TestMain(m *testing.M) {
	baseURL = envOr("KVQ_TEST_API_URL", "http://localhost:8080")
	client = &http.Client{Timeout: 10 * time.Second}

	if err := waitForAPI(30); err != nil {
		fmt.Fprintf(os.Stderr, "API not reachable at %s: %v\n", baseURL, err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func waitForAPI(maxSeconds int) error {
	for i := range maxSeconds {
		resp, err := client.Get(baseURL + "/healthz")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				return nil
			}
		}
		if i < maxSeconds-1 {
			time.Sleep(1 * time.Second)
		}
	}
	return fmt.Errorf("timeout after %ds", maxSeconds)
}

func doJSON(method, path string, body any) (int, []byte, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return 0, nil, err
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, baseURL+path, reqBody)
	if err != nil {
		return 0, nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, respBody, nil
}

func doForm(method, path string, values url.Values) (int, []byte, error) {
	req, err := http.NewRequest(method, baseURL+path, strings.NewReader(values.Encode()))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, respBody, nil
}

func doRaw(method, path string) (int, []byte, error) {
	req, err := http.NewRequest(method, baseURL+path, nil)
	if err != nil {
		return 0, nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, respBody, nil
}

func parseJSON(data []byte) map[string]any {
	var m map[string]any
	json.Unmarshal(data, &m)
	return m
}

func parseJSONArray(data []byte) []any {
	var a []any
	json.Unmarshal(data, &a)
	return a
}

func mustCreateProject(t *testing.T) string {
	t.Helper()
	status, body, err := doJSON("POST", "/v1/projects", map[string]string{
		"name":        fmt.Sprintf("test-%d", time.Now().UnixNano()),
		"description": "integration test",
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if status != 201 {
		t.Fatalf("create project: expected 201 got %d: %s", status, body)
	}
	m := parseJSON(body)
	id, _ := m["id"].(string)
	if id == "" {
		t.Fatalf("create project: no id in response: %s", body)
	}
	return id
}

func mustDeleteProject(t *testing.T, id string) {
	t.Helper()
	doJSON("DELETE", "/v1/projects/"+id, nil)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
