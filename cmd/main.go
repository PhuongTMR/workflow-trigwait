package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Owner            string
	Repo             string
	GitHubToken      string
	WorkflowFileName string
	Ref              string
	ClientPayload    map[string]interface{}
	WaitInterval     time.Duration
	TriggerTimeout   time.Duration
	PropagateFailure bool
	TriggerWorkflow  bool
	WaitWorkflow     bool
	GitHubAPIURL     string
	GitHubServerURL  string
	DistinctID       string
	DistinctIDName   string
}

type WorkflowRun struct {
	ID           int64  `json:"id"`
	Status       string `json:"status"`
	Conclusion   string `json:"conclusion"`
	CreatedAt    string `json:"created_at"`
	DisplayTitle string `json:"display_title"`
}

type WorkflowRunsResponse struct {
	WorkflowRuns []WorkflowRun `json:"workflow_runs"`
}

func main() {
	config, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v", err)
		os.Exit(1)
	}

	// Print header
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	var runID int64
	if config.TriggerWorkflow {
		runID, err = triggerWorkflow(config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: %v", err)
			os.Exit(1)
		}
	} else {
		fmt.Println("⏭ Skipping workflow trigger")
	}

	if config.WaitWorkflow && runID > 0 {
		err = waitForWorkflow(config, runID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: %v", err)
			os.Exit(1)
		}
	} else if runID > 0 {
		// Set outputs even when not waiting
		workflowURL := fmt.Sprintf("%s/%s/%s/actions/runs/%d", config.GitHubServerURL, config.Owner, config.Repo, runID)
		setOutput("workflow_id", strconv.FormatInt(runID, 10))
		setOutput("workflow_url", workflowURL)
		fmt.Printf("⏭ Skipping wait (workflow started)")
		fmt.Printf("   URL: %s", workflowURL)
	}
}

func loadConfig() (*Config, error) {
	config := &Config{
		Owner:            os.Getenv("INPUT_OWNER"),
		Repo:             os.Getenv("INPUT_REPO"),
		GitHubToken:      os.Getenv("INPUT_GITHUB_TOKEN"),
		WorkflowFileName: os.Getenv("INPUT_WORKFLOW_FILE_NAME"),
		Ref:              getEnvOrDefault("INPUT_REF", "main"),
		PropagateFailure: getEnvBool("INPUT_PROPAGATE_FAILURE", true),
		TriggerWorkflow:  getEnvBool("INPUT_TRIGGER_WORKFLOW", true),
		WaitWorkflow:     getEnvBool("INPUT_WAIT_WORKFLOW", true),
		GitHubAPIURL:     getEnvOrDefault("GITHUB_API_URL", "https://api.github.com"),
		GitHubServerURL:  getEnvOrDefault("GITHUB_SERVER_URL", "https://github.com"),
		DistinctIDName:   os.Getenv("INPUT_DISTINCT_ID_NAME"),
	}

	// Parse durations
	waitInterval, _ := strconv.Atoi(getEnvOrDefault("INPUT_WAIT_INTERVAL", "10"))
	config.WaitInterval = time.Duration(waitInterval) * time.Second

	triggerTimeout, _ := strconv.Atoi(getEnvOrDefault("INPUT_TRIGGER_TIMEOUT", "120"))
	config.TriggerTimeout = time.Duration(triggerTimeout) * time.Second

	// Parse client payload
	payloadStr := os.Getenv("INPUT_CLIENT_PAYLOAD")
	if payloadStr != "" {
		if err := json.Unmarshal([]byte(payloadStr), &config.ClientPayload); err != nil {
			return nil, fmt.Errorf("invalid client_payload JSON: %w", err)
		}
	} else {
		config.ClientPayload = make(map[string]interface{})
	}

	// Remove empty values from client_payload
	config.ClientPayload = removeEmptyValues(config.ClientPayload)

	// Generate distinct_id for correlating the triggered workflow run (only if enabled)
	if config.DistinctIDName != "" {
		config.DistinctID = generateDistinctID()
		config.ClientPayload[config.DistinctIDName] = config.DistinctID
	}

	// Validate required fields
	if config.Owner == "" {
		return nil, fmt.Errorf("owner is a required argument")
	}
	if config.Repo == "" {
		return nil, fmt.Errorf("repo is a required argument")
	}
	if config.GitHubToken == "" {
		return nil, fmt.Errorf("github_token is required")
	}
	if config.WorkflowFileName == "" {
		return nil, fmt.Errorf("workflow_file_name is required")
	}

	return config, nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return strings.ToLower(value) == "true"
}

func removeEmptyValues(payload map[string]interface{}) map[string]interface{} {
	cleaned := make(map[string]interface{})
	for key, value := range payload {
		// Skip nil values
		if value == nil {
			continue
		}

		// Check for empty string
		if strVal, ok := value.(string); ok && strVal == "" {
			continue
		}

		// Recursively clean nested maps
		if mapVal, ok := value.(map[string]interface{}); ok {
			cleanedNested := removeEmptyValues(mapVal)
			if len(cleanedNested) > 0 {
				cleaned[key] = cleanedNested
			}
			continue
		}

		// Keep non-empty values
		cleaned[key] = value
	}
	return cleaned
}

func generateDistinctID() string {
	// Generate 6 random bytes
	b := make([]byte, 6)
	rand.Read(b)

	// Use base32 encoding without padding for shorter, more readable IDs
	// Results in ~10 characters (e.g., "K7M4N2P9QR")
	encoded := strings.ToUpper(hex.EncodeToString(b))

	// Take first 8 characters for even shorter ID
	if len(encoded) > 8 {
		return encoded[:8]
	}
	return encoded
}

func triggerWorkflow(config *Config) (int64, error) {
	startTime := time.Now()
	deadline := startTime.Add(config.TriggerTimeout)

	// Prepare dispatch payload
	payload := map[string]interface{}{
		"ref":    config.Ref,
		"inputs": config.ClientPayload,
	}
	payloadBytes, _ := json.Marshal(payload)

	// Print compact header
	fmt.Printf("🚀 Triggering %s/%s → %s @ %s", config.Owner, config.Repo, config.WorkflowFileName, config.Ref)
	if config.DistinctID != "" {
		fmt.Printf(" [%s]", config.DistinctID)
		setOutput("distinct_id", config.DistinctID)
	}
	if len(config.ClientPayload) > 0 {
		inputsJSON, _ := json.Marshal(config.ClientPayload)
		fmt.Printf("   Inputs: %s", string(inputsJSON))
	}
	fmt.Println()

	// Trigger the workflow
	path := fmt.Sprintf("workflows/%s/dispatches", config.WorkflowFileName)
	_, err := apiRequest(config, "POST", path, payloadBytes)
	if err != nil {
		return 0, fmt.Errorf("failed to trigger workflow: %w", err)
	}

	// Wait for the run to appear
	retryInterval := config.WaitInterval
	lastPrintTime := time.Now()
	for {
		if time.Now().After(deadline) {
			return 0, fmt.Errorf("timeout: workflow run did not appear within %v", config.TriggerTimeout)
		}

		time.Sleep(retryInterval)

		runID, err := findWorkflowRun(config, startTime)
		if err != nil {
			// Only print errors occasionally to avoid spam
			if time.Since(lastPrintTime) > 10*time.Second {
				fmt.Fprintf(os.Stderr, "\r⚠ Error checking runs (retrying...)")
				lastPrintTime = time.Now()
			}
		}
		if runID > 0 {
			fmt.Printf("   ✓ Triggered run #%d", runID)
			return runID, nil
		}

		// Show progress dot every 10 seconds
		if time.Since(lastPrintTime) > 10*time.Second {
			elapsed := time.Since(startTime).Round(time.Second)
			fmt.Printf("\r   Finding run... %v", elapsed)
			lastPrintTime = time.Now()
		}

		// Exponential backoff up to 60 seconds
		retryInterval *= 2
		if retryInterval > 60*time.Second {
			retryInterval = 60 * time.Second
		}
	}
}

func findWorkflowRun(config *Config, startTime time.Time) (int64, error) {
	// Build query with filters
	query := fmt.Sprintf("event=workflow_dispatch&branch=%s&per_page=10", config.Ref)

	path := fmt.Sprintf("workflows/%s/runs?%s", config.WorkflowFileName, query)
	respBody, err := apiRequest(config, "GET", path, nil)
	if err != nil {
		return 0, err
	}

	var response WorkflowRunsResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		return 0, fmt.Errorf("failed to parse response: %w", err)
	}

	// Find a run created after startTime
	for _, run := range response.WorkflowRuns {
		createdAt, err := time.Parse(time.RFC3339, run.CreatedAt)
		if err != nil {
			continue
		}
		if createdAt.Unix() >= startTime.Unix() {
			// If distinct_id is enabled, verify by checking display_title (run-name)
			if config.DistinctID != "" {
				if strings.Contains(run.DisplayTitle, config.DistinctID) {
					return run.ID, nil
				}
			} else {
				// Fall back to time-based matching
				return run.ID, nil
			}
		}
	}

	return 0, nil
}

func waitForWorkflow(config *Config, runID int64) error {
	workflowURL := fmt.Sprintf("%s/%s/%s/actions/runs/%d", config.GitHubServerURL, config.Owner, config.Repo, runID)

	fmt.Printf("⏳ Waiting for workflow completion...")
	fmt.Printf("   URL: %s", workflowURL)

	setOutput("workflow_id", strconv.FormatInt(runID, 10))
	setOutput("workflow_url", workflowURL)

	startTime := time.Now()
	lastStatus := ""
	pollInterval := config.WaitInterval
	lastPrintTime := time.Now()

	// Poll for completion with adaptive intervals
	for {
		time.Sleep(pollInterval)

		run, err := getWorkflowRun(config, runID)
		if err != nil {
			// Only show errors occasionally
			if time.Since(lastPrintTime) > 10*time.Second {
				fmt.Fprintf(os.Stderr, "\r⚠ Error fetching status (retrying...)")
				lastPrintTime = time.Now()
			}
			continue
		}

		elapsed := time.Since(startTime).Round(time.Second)
		setOutput("conclusion", run.Conclusion)

		if run.Status == "completed" {
			fmt.Printf("\r")
			if run.Conclusion == "success" {
				fmt.Printf("   ✅ Completed successfully in %v", elapsed)
			} else {
				fmt.Printf("   ❌ Failed with status: %s (duration: %v)", run.Conclusion, elapsed)
			}

			if run.Conclusion != "success" && config.PropagateFailure {
				return fmt.Errorf("workflow failed with conclusion: %s", run.Conclusion)
			}
			return nil
		}

		// Only print status changes to reduce log noise
		if run.Status != lastStatus {
			statusIcon := "⏳"
			statusText := run.Status
			switch run.Status {
			case "queued", "waiting", "pending":
				statusIcon = "🔄"
				statusText = "queued"
			case "in_progress":
				statusIcon = "▶️"
				statusText = "running"
			}
			fmt.Printf("\r   %s Status: %s (elapsed: %v)", statusIcon, statusText, elapsed)
			lastStatus = run.Status
			lastPrintTime = time.Now()
		} else if time.Since(lastPrintTime) > 30*time.Second {
			// Update elapsed time every 30 seconds even if status hasn't changed
			statusText := "running"
			if run.Status == "queued" || run.Status == "waiting" || run.Status == "pending" {
				statusText = "queued"
			}
			fmt.Printf("\r   %s Status: %s (elapsed: %v)", "⏳", statusText, elapsed)
			lastPrintTime = time.Now()
		}

		// Adaptive polling: slower when queued, faster when in_progress
		switch run.Status {
		case "queued", "waiting", "pending":
			pollInterval = maxDuration(config.WaitInterval, 30*time.Second)
		case "in_progress":
			pollInterval = config.WaitInterval
		}
	}
}

func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}

func getWorkflowRun(config *Config, runID int64) (*WorkflowRun, error) {
	path := fmt.Sprintf("runs/%d", runID)
	respBody, err := apiRequest(config, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var run WorkflowRun
	if err := json.Unmarshal(respBody, &run); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &run, nil
}

func apiRequest(config *Config, method, path string, body []byte) ([]byte, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/actions/%s", config.GitHubAPIURL, config.Owner, config.Repo, path)

	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+config.GitHubToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 204 No Content is success for dispatch
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return respBody, nil
	}

	return nil, fmt.Errorf("API request failed: %sResponse: %s", resp.Status, string(respBody))
}

func setOutput(name, value string) {
	outputFile := os.Getenv("GITHUB_OUTPUT")
	if outputFile == "" {
		return
	}

	f, err := os.OpenFile(outputFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open GITHUB_OUTPUT: %v", err)
		return
	}
	defer f.Close()

	fmt.Fprintf(f, "%s=%s", name, value)
}
