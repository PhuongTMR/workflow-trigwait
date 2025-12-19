package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestLoadConfig_Defaults(t *testing.T) {
	// Set required env vars
	os.Setenv("INPUT_OWNER", "test-owner")
	os.Setenv("INPUT_REPO", "test-repo")
	os.Setenv("INPUT_GITHUB_TOKEN", "test-token")
	os.Setenv("INPUT_WORKFLOW_FILE_NAME", "test.yml")
	defer func() {
		os.Unsetenv("INPUT_OWNER")
		os.Unsetenv("INPUT_REPO")
		os.Unsetenv("INPUT_GITHUB_TOKEN")
		os.Unsetenv("INPUT_WORKFLOW_FILE_NAME")
	}()

	config, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig failed: %v", err)
	}

	if config.Ref != "main" {
		t.Errorf("expected ref 'main', got '%s'", config.Ref)
	}
	if config.WaitInterval != 10*time.Second {
		t.Errorf("expected wait_interval 10s, got %v", config.WaitInterval)
	}
	if config.TriggerTimeout != 120*time.Second {
		t.Errorf("expected trigger_timeout 120s, got %v", config.TriggerTimeout)
	}
	if !config.PropagateFailure {
		t.Error("expected propagate_failure true by default")
	}
	if !config.TriggerWorkflow {
		t.Error("expected trigger_workflow true by default")
	}
	if !config.WaitWorkflow {
		t.Error("expected wait_workflow true by default")
	}
}

func TestLoadConfig_CustomValues(t *testing.T) {
	os.Setenv("INPUT_OWNER", "test-owner")
	os.Setenv("INPUT_REPO", "test-repo")
	os.Setenv("INPUT_GITHUB_TOKEN", "test-token")
	os.Setenv("INPUT_WORKFLOW_FILE_NAME", "test.yml")
	os.Setenv("INPUT_REF", "develop")
	os.Setenv("INPUT_WAIT_INTERVAL", "30")
	os.Setenv("INPUT_TRIGGER_TIMEOUT", "300")
	os.Setenv("INPUT_PROPAGATE_FAILURE", "false")
	os.Setenv("INPUT_CLIENT_PAYLOAD", `{"key": "value", "num": 123}`)
	defer func() {
		os.Unsetenv("INPUT_OWNER")
		os.Unsetenv("INPUT_REPO")
		os.Unsetenv("INPUT_GITHUB_TOKEN")
		os.Unsetenv("INPUT_WORKFLOW_FILE_NAME")
		os.Unsetenv("INPUT_REF")
		os.Unsetenv("INPUT_WAIT_INTERVAL")
		os.Unsetenv("INPUT_TRIGGER_TIMEOUT")
		os.Unsetenv("INPUT_PROPAGATE_FAILURE")
		os.Unsetenv("INPUT_CLIENT_PAYLOAD")
	}()

	config, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig failed: %v", err)
	}

	if config.Ref != "develop" {
		t.Errorf("expected ref 'develop', got '%s'", config.Ref)
	}
	if config.WaitInterval != 30*time.Second {
		t.Errorf("expected wait_interval 30s, got %v", config.WaitInterval)
	}
	if config.TriggerTimeout != 300*time.Second {
		t.Errorf("expected trigger_timeout 300s, got %v", config.TriggerTimeout)
	}
	if config.PropagateFailure {
		t.Error("expected propagate_failure false")
	}
	if config.ClientPayload["key"] != "value" {
		t.Errorf("expected client_payload key='value', got %v", config.ClientPayload["key"])
	}
}

func TestLoadConfig_MissingRequired(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		errMsg  string
	}{
		{
			name:    "missing owner",
			envVars: map[string]string{"INPUT_REPO": "repo", "INPUT_GITHUB_TOKEN": "token", "INPUT_WORKFLOW_FILE_NAME": "test.yml"},
			errMsg:  "owner",
		},
		{
			name:    "missing repo",
			envVars: map[string]string{"INPUT_OWNER": "owner", "INPUT_GITHUB_TOKEN": "token", "INPUT_WORKFLOW_FILE_NAME": "test.yml"},
			errMsg:  "repo",
		},
		{
			name:    "missing token",
			envVars: map[string]string{"INPUT_OWNER": "owner", "INPUT_REPO": "repo", "INPUT_WORKFLOW_FILE_NAME": "test.yml"},
			errMsg:  "github_token",
		},
		{
			name:    "missing workflow",
			envVars: map[string]string{"INPUT_OWNER": "owner", "INPUT_REPO": "repo", "INPUT_GITHUB_TOKEN": "token"},
			errMsg:  "workflow_file_name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all env vars
			os.Unsetenv("INPUT_OWNER")
			os.Unsetenv("INPUT_REPO")
			os.Unsetenv("INPUT_GITHUB_TOKEN")
			os.Unsetenv("INPUT_WORKFLOW_FILE_NAME")

			// Set test env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer func() {
				for k := range tt.envVars {
					os.Unsetenv(k)
				}
			}()

			_, err := loadConfig()
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !contains(err.Error(), tt.errMsg) {
				t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
			}
		})
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	os.Setenv("INPUT_OWNER", "test-owner")
	os.Setenv("INPUT_REPO", "test-repo")
	os.Setenv("INPUT_GITHUB_TOKEN", "test-token")
	os.Setenv("INPUT_WORKFLOW_FILE_NAME", "test.yml")
	os.Setenv("INPUT_CLIENT_PAYLOAD", "invalid json")
	defer func() {
		os.Unsetenv("INPUT_OWNER")
		os.Unsetenv("INPUT_REPO")
		os.Unsetenv("INPUT_GITHUB_TOKEN")
		os.Unsetenv("INPUT_WORKFLOW_FILE_NAME")
		os.Unsetenv("INPUT_CLIENT_PAYLOAD")
	}()

	_, err := loadConfig()
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		value    string
		defVal   bool
		expected bool
	}{
		{"", true, true},
		{"", false, false},
		{"true", false, true},
		{"TRUE", false, true},
		{"True", false, true},
		{"false", true, false},
		{"FALSE", true, false},
		{"other", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			os.Setenv("TEST_BOOL", tt.value)
			defer os.Unsetenv("TEST_BOOL")

			result := getEnvBool("TEST_BOOL", tt.defVal)
			if result != tt.expected {
				t.Errorf("getEnvBool(%q, %v) = %v, want %v", tt.value, tt.defVal, result, tt.expected)
			}
		})
	}
}

func TestAPIRequest_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("missing or incorrect Authorization header")
		}
		if r.Header.Get("Accept") != "application/vnd.github.v3+json" {
			t.Error("missing or incorrect Accept header")
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	config := &Config{
		Owner:        "owner",
		Repo:         "repo",
		GitHubToken:  "test-token",
		GitHubAPIURL: server.URL,
	}

	resp, err := apiRequest(config, "GET", "test/path", nil)
	if err != nil {
		t.Fatalf("apiRequest failed: %v", err)
	}

	var result map[string]string
	json.Unmarshal(resp, &result)
	if result["status"] != "ok" {
		t.Errorf("unexpected response: %s", string(resp))
	}
}

func TestAPIRequest_Failure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{"message": "Validation failed"}`))
	}))
	defer server.Close()

	config := &Config{
		Owner:        "owner",
		Repo:         "repo",
		GitHubToken:  "test-token",
		GitHubAPIURL: server.URL,
	}

	_, err := apiRequest(config, "POST", "test/path", []byte(`{}`))
	if err == nil {
		t.Fatal("expected error for 422 response, got nil")
	}
	if !contains(err.Error(), "422") {
		t.Errorf("expected error containing '422', got '%s'", err.Error())
	}
}

func TestFindWorkflowRun(t *testing.T) {
	startTime := time.Now()
	distinctID := "test-distinct-123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Runs endpoint - return matching display_title
		response := WorkflowRunsResponse{
			WorkflowRuns: []WorkflowRun{
				{
					ID:           12345,
					CreatedAt:    startTime.Add(1 * time.Second).Format(time.RFC3339),
					DisplayTitle: "Self-test [" + distinctID + "]",
				},
				{ID: 12344, CreatedAt: startTime.Add(-1 * time.Hour).Format(time.RFC3339)},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &Config{
		Owner:            "owner",
		Repo:             "repo",
		GitHubToken:      "test-token",
		GitHubAPIURL:     server.URL,
		WorkflowFileName: "test.yml",
		Ref:              "main",
		DistinctID:       distinctID,
		DistinctIDName:   "distinct_id",
	}

	runID, err := findWorkflowRun(config, startTime)
	if err != nil {
		t.Fatalf("findWorkflowRun failed: %v", err)
	}
	if runID != 12345 {
		t.Errorf("expected runID 12345, got %d", runID)
	}
}

func TestFindWorkflowRun_NoMatch(t *testing.T) {
	startTime := time.Now()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := WorkflowRunsResponse{
			WorkflowRuns: []WorkflowRun{
				{ID: 12345, CreatedAt: startTime.Add(-1 * time.Hour).Format(time.RFC3339)},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &Config{
		Owner:            "owner",
		Repo:             "repo",
		GitHubToken:      "test-token",
		GitHubAPIURL:     server.URL,
		WorkflowFileName: "test.yml",
		Ref:              "main",
	}

	runID, err := findWorkflowRun(config, startTime)
	if err != nil {
		t.Fatalf("findWorkflowRun failed: %v", err)
	}
	if runID != 0 {
		t.Errorf("expected runID 0 (no match), got %d", runID)
	}
}

func TestFindWorkflowRun_DistinctIDFallbackToTimeBased(t *testing.T) {
	startTime := time.Now()
	distinctID := "test-distinct-123"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate target workflow that doesn't include distinct_id in run-name
		response := WorkflowRunsResponse{
			WorkflowRuns: []WorkflowRun{
				{
					ID:           12346,
					CreatedAt:    startTime.Add(2 * time.Second).Format(time.RFC3339),
					DisplayTitle: "Deploy", // No distinct_id in title
				},
				{
					ID:           12345,
					CreatedAt:    startTime.Add(1 * time.Second).Format(time.RFC3339),
					DisplayTitle: "Deploy", // No distinct_id in title
				},
				{
					ID:           12344,
					CreatedAt:    startTime.Add(-1 * time.Hour).Format(time.RFC3339),
					DisplayTitle: "Deploy",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &Config{
		Owner:            "owner",
		Repo:             "repo",
		GitHubToken:      "test-token",
		GitHubAPIURL:     server.URL,
		WorkflowFileName: "test.yml",
		Ref:              "main",
		DistinctID:       distinctID,
		DistinctIDName:   "distinct_id",
	}

	runID, err := findWorkflowRun(config, startTime)
	if err != nil {
		t.Fatalf("findWorkflowRun failed: %v", err)
	}

	// When distinct_id is enabled but not found in display_title,
	// it should NOT fall back to time-based matching and return 0
	if runID != 0 {
		t.Errorf("expected runID 0 (no distinct_id match), got %d", runID)
	}
}

func TestFindWorkflowRun_TimeBasedWithMultipleRuns(t *testing.T) {
	startTime := time.Now()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Multiple runs after startTime, should match the first one chronologically
		response := WorkflowRunsResponse{
			WorkflowRuns: []WorkflowRun{
				{
					ID:           12346,
					CreatedAt:    startTime.Add(2 * time.Second).Format(time.RFC3339),
					DisplayTitle: "Deploy production",
				},
				{
					ID:           12345,
					CreatedAt:    startTime.Add(1 * time.Second).Format(time.RFC3339),
					DisplayTitle: "Deploy staging",
				},
				{
					ID:           12344,
					CreatedAt:    startTime.Add(-1 * time.Hour).Format(time.RFC3339),
					DisplayTitle: "Deploy dev",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &Config{
		Owner:            "owner",
		Repo:             "repo",
		GitHubToken:      "test-token",
		GitHubAPIURL:     server.URL,
		WorkflowFileName: "test.yml",
		Ref:              "main",
		// No DistinctID - pure time-based matching
	}

	runID, err := findWorkflowRun(config, startTime)
	if err != nil {
		t.Fatalf("findWorkflowRun failed: %v", err)
	}

	// Should match the first run created after startTime (12346, not 12345)
	if runID != 12346 {
		t.Errorf("expected runID 12346 (first after startTime), got %d", runID)
	}
}

func TestSetOutput(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "github_output")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	os.Setenv("GITHUB_OUTPUT", tmpFile.Name())
	defer os.Unsetenv("GITHUB_OUTPUT")

	setOutput("test_key", "test_value")
	setOutput("workflow_id", "12345")

	content, _ := os.ReadFile(tmpFile.Name())
	if !contains(string(content), "test_key=test_value") {
		t.Errorf("expected output to contain 'test_key=test_value', got: %s", string(content))
	}
	if !contains(string(content), "workflow_id=12345") {
		t.Errorf("expected output to contain 'workflow_id=12345', got: %s", string(content))
	}
}

func TestTriggerWorkflow_Success(t *testing.T) {
	triggerCalled := false
	startTime := time.Now()
	distinctID := "test-trigger-id"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if contains(r.URL.Path, "dispatches") {
			triggerCalled = true
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if contains(r.URL.Path, "runs") {
			response := WorkflowRunsResponse{
				WorkflowRuns: []WorkflowRun{
					{
						ID:           99999,
						CreatedAt:    startTime.Add(1 * time.Second).Format(time.RFC3339),
						DisplayTitle: "Test [" + distinctID + "]",
					},
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
	}))
	defer server.Close()

	config := &Config{
		Owner:            "owner",
		Repo:             "repo",
		GitHubToken:      "test-token",
		GitHubAPIURL:     server.URL,
		WorkflowFileName: "test.yml",
		Ref:              "main",
		ClientPayload:    map[string]interface{}{},
		WaitInterval:     100 * time.Millisecond,
		TriggerTimeout:   5 * time.Second,
		DistinctID:       distinctID,
		DistinctIDName:   "distinct_id",
	}

	runID, err := triggerWorkflow(config)
	if err != nil {
		t.Fatalf("triggerWorkflow failed: %v", err)
	}
	if !triggerCalled {
		t.Error("dispatch endpoint was not called")
	}
	if runID != 99999 {
		t.Errorf("expected runID 99999, got %d", runID)
	}
}

func TestTriggerWorkflow_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if contains(r.URL.Path, "dispatches") {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if contains(r.URL.Path, "runs") {
			// Return empty runs - run never appears
			json.NewEncoder(w).Encode(WorkflowRunsResponse{WorkflowRuns: []WorkflowRun{}})
			return
		}
	}))
	defer server.Close()

	config := &Config{
		Owner:            "owner",
		Repo:             "repo",
		GitHubToken:      "test-token",
		GitHubAPIURL:     server.URL,
		WorkflowFileName: "test.yml",
		Ref:              "main",
		ClientPayload:    map[string]interface{}{},
		WaitInterval:     50 * time.Millisecond,
		TriggerTimeout:   200 * time.Millisecond,
		DistinctID:       "timeout-test",
		DistinctIDName:   "distinct_id",
	}

	_, err := triggerWorkflow(config)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !contains(err.Error(), "timeout") {
		t.Errorf("expected error containing 'timeout', got '%s'", err.Error())
	}
}

func TestWaitForWorkflow_Success(t *testing.T) {
	pollCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pollCount++
		run := WorkflowRun{
			ID:         12345,
			Status:     "completed",
			Conclusion: "success",
		}
		json.NewEncoder(w).Encode(run)
	}))
	defer server.Close()

	tmpFile, _ := os.CreateTemp("", "github_output")
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()
	os.Setenv("GITHUB_OUTPUT", tmpFile.Name())
	defer os.Unsetenv("GITHUB_OUTPUT")

	config := &Config{
		Owner:            "owner",
		Repo:             "repo",
		GitHubToken:      "test-token",
		GitHubAPIURL:     server.URL,
		GitHubServerURL:  "https://github.com",
		WaitInterval:     50 * time.Millisecond,
		PropagateFailure: true,
	}

	err := waitForWorkflow(config, 12345)
	if err != nil {
		t.Fatalf("waitForWorkflow failed: %v", err)
	}

	content, _ := os.ReadFile(tmpFile.Name())
	if !contains(string(content), "workflow_id=12345") {
		t.Error("expected workflow_id output")
	}
	if !contains(string(content), "conclusion=success") {
		t.Error("expected conclusion=success output")
	}
}

func TestWaitForWorkflow_FailurePropagated(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		run := WorkflowRun{
			ID:         12345,
			Status:     "completed",
			Conclusion: "failure",
		}
		json.NewEncoder(w).Encode(run)
	}))
	defer server.Close()

	tmpFile, _ := os.CreateTemp("", "github_output")
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()
	os.Setenv("GITHUB_OUTPUT", tmpFile.Name())
	defer os.Unsetenv("GITHUB_OUTPUT")

	config := &Config{
		Owner:            "owner",
		Repo:             "repo",
		GitHubToken:      "test-token",
		GitHubAPIURL:     server.URL,
		GitHubServerURL:  "https://github.com",
		WaitInterval:     50 * time.Millisecond,
		PropagateFailure: true,
	}

	err := waitForWorkflow(config, 12345)
	if err == nil {
		t.Fatal("expected error for failed workflow, got nil")
	}
	if !contains(err.Error(), "failure") {
		t.Errorf("expected error containing 'failure', got '%s'", err.Error())
	}
}

func TestWaitForWorkflow_FailureNotPropagated(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		run := WorkflowRun{
			ID:         12345,
			Status:     "completed",
			Conclusion: "failure",
		}
		json.NewEncoder(w).Encode(run)
	}))
	defer server.Close()

	tmpFile, _ := os.CreateTemp("", "github_output")
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()
	os.Setenv("GITHUB_OUTPUT", tmpFile.Name())
	defer os.Unsetenv("GITHUB_OUTPUT")

	config := &Config{
		Owner:            "owner",
		Repo:             "repo",
		GitHubToken:      "test-token",
		GitHubAPIURL:     server.URL,
		GitHubServerURL:  "https://github.com",
		WaitInterval:     50 * time.Millisecond,
		PropagateFailure: false,
	}

	err := waitForWorkflow(config, 12345)
	if err != nil {
		t.Fatalf("expected no error when propagate_failure=false, got: %v", err)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestRemoveEmptyValues(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name: "removes empty strings",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": "",
				"key3": "value3",
			},
			expected: map[string]interface{}{
				"key1": "value1",
				"key3": "value3",
			},
		},
		{
			name: "removes nil values",
			input: map[string]interface{}{
				"key1": "value1",
				"key2": nil,
				"key3": "value3",
			},
			expected: map[string]interface{}{
				"key1": "value1",
				"key3": "value3",
			},
		},
		{
			name: "keeps non-empty values",
			input: map[string]interface{}{
				"string":  "value",
				"number":  123,
				"boolean": true,
				"zero":    0,
			},
			expected: map[string]interface{}{
				"string":  "value",
				"number":  123,
				"boolean": true,
				"zero":    0,
			},
		},
		{
			name: "handles nested maps",
			input: map[string]interface{}{
				"outer": map[string]interface{}{
					"inner1": "value",
					"inner2": "",
					"inner3": nil,
				},
				"key": "value",
			},
			expected: map[string]interface{}{
				"outer": map[string]interface{}{
					"inner1": "value",
				},
				"key": "value",
			},
		},
		{
			name: "removes empty nested maps",
			input: map[string]interface{}{
				"outer": map[string]interface{}{
					"inner": "",
				},
				"key": "value",
			},
			expected: map[string]interface{}{
				"key": "value",
			},
		},
		{
			name: "handles all empty values",
			input: map[string]interface{}{
				"key1": "",
				"key2": nil,
			},
			expected: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeEmptyValues(tt.input)

			// Check length
			if len(result) != len(tt.expected) {
				t.Errorf("expected length %d, got %d", len(tt.expected), len(result))
			}

			// Check each key-value
			for key, expectedVal := range tt.expected {
				actualVal, exists := result[key]
				if !exists {
					t.Errorf("expected key %q to exist", key)
					continue
				}

				// Handle nested maps
				if expectedMap, ok := expectedVal.(map[string]interface{}); ok {
					actualMap, ok := actualVal.(map[string]interface{})
					if !ok {
						t.Errorf("expected %q to be a map", key)
						continue
					}
					if len(actualMap) != len(expectedMap) {
						t.Errorf("nested map %q: expected length %d, got %d", key, len(expectedMap), len(actualMap))
					}
					for nestedKey, nestedExpected := range expectedMap {
						if actualMap[nestedKey] != nestedExpected {
							t.Errorf("nested map %q[%q]: expected %v, got %v", key, nestedKey, nestedExpected, actualMap[nestedKey])
						}
					}
				} else if actualVal != expectedVal {
					t.Errorf("key %q: expected %v, got %v", key, expectedVal, actualVal)
				}
			}
		})
	}
}

func TestLoadConfig_ClientPayloadWithEmptyValues(t *testing.T) {
	os.Setenv("INPUT_OWNER", "test-owner")
	os.Setenv("INPUT_REPO", "test-repo")
	os.Setenv("INPUT_GITHUB_TOKEN", "test-token")
	os.Setenv("INPUT_WORKFLOW_FILE_NAME", "test.yml")
	os.Setenv("INPUT_CLIENT_PAYLOAD", `{"env": "prod", "version": "", "debug": false, "optional": null}`)
	defer func() {
		os.Unsetenv("INPUT_OWNER")
		os.Unsetenv("INPUT_REPO")
		os.Unsetenv("INPUT_GITHUB_TOKEN")
		os.Unsetenv("INPUT_WORKFLOW_FILE_NAME")
		os.Unsetenv("INPUT_CLIENT_PAYLOAD")
	}()

	config, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig failed: %v", err)
	}

	// Check that empty string was removed
	if _, exists := config.ClientPayload["version"]; exists {
		t.Error("expected 'version' (empty string) to be removed")
	}

	// Check that null was removed
	if _, exists := config.ClientPayload["optional"]; exists {
		t.Error("expected 'optional' (null) to be removed")
	}

	// Check that non-empty values remain
	if config.ClientPayload["env"] != "prod" {
		t.Errorf("expected env='prod', got %v", config.ClientPayload["env"])
	}

	if config.ClientPayload["debug"] != false {
		t.Errorf("expected debug=false, got %v", config.ClientPayload["debug"])
	}
}

// Benchmark tests
func BenchmarkLoadConfig(b *testing.B) {
	os.Setenv("INPUT_OWNER", "test-owner")
	os.Setenv("INPUT_REPO", "test-repo")
	os.Setenv("INPUT_GITHUB_TOKEN", "test-token")
	os.Setenv("INPUT_WORKFLOW_FILE_NAME", "test.yml")
	os.Setenv("INPUT_CLIENT_PAYLOAD", `{"key": "value"}`)
	defer func() {
		os.Unsetenv("INPUT_OWNER")
		os.Unsetenv("INPUT_REPO")
		os.Unsetenv("INPUT_GITHUB_TOKEN")
		os.Unsetenv("INPUT_WORKFLOW_FILE_NAME")
		os.Unsetenv("INPUT_CLIENT_PAYLOAD")
	}()

	for i := 0; i < b.N; i++ {
		loadConfig()
	}
}
