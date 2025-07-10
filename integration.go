package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// Colores para output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
)

type TestResult struct {
	Name     string
	Success  bool
	Duration time.Duration
	Error    string
}

type IntegrationTester struct {
	apiURL      string
	frontendURL string
	results     []TestResult
}

func NewIntegrationTester() *IntegrationTester {
	return &IntegrationTester{
		apiURL:      "http://localhost:8080/api/v1",
		frontendURL: "http://localhost:5173",
		results:     make([]TestResult, 0),
	}
}

func (t *IntegrationTester) runTest(name string, testFunc func() error) {
	fmt.Printf("%s[TEST]%s %s... ", ColorBlue, ColorReset, name)
	
	start := time.Now()
	err := testFunc()
	duration := time.Since(start)
	
	result := TestResult{
		Name:     name,
		Success:  err == nil,
		Duration: duration,
	}
	
	if err != nil {
		result.Error = err.Error()
		fmt.Printf("%s✗ FAIL%s (%v) - %s\n", ColorRed, ColorReset, duration, err.Error())
	} else {
		fmt.Printf("%s✓ PASS%s (%v)\n", ColorGreen, ColorReset, duration)
	}
	
	t.results = append(t.results, result)
}

func (t *IntegrationTester) testAPIHealth() error {
	resp, err := http.Get(t.apiURL + "/health")
	if err != nil {
		return fmt.Errorf("failed to connect to API: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return fmt.Errorf("API health check failed with status %d", resp.StatusCode)
	}
	
	var health map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return fmt.Errorf("failed to decode health response: %v", err)
	}
	
	if health["status"] != "ok" {
		return fmt.Errorf("API status is not ok: %v", health["status"])
	}
	
	return nil
}

func (t *IntegrationTester) testAnalyzeEndpoint() error {
	testCases := []struct {
		command     string
		expectValid bool
	}{
		{"GET mykey", true},
		{"SET key value", true},
		{"GET", false},
		{"UNKNOWN command", false},
	}
	
	for _, tc := range testCases {
		payload := map[string]string{"command": tc.command}
		jsonData, _ := json.Marshal(payload)
		
		resp, err := http.Post(t.apiURL+"/analyze", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return fmt.Errorf("failed to analyze command '%s': %v", tc.command, err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != 200 {
			return fmt.Errorf("analyze endpoint returned status %d for command '%s'", resp.StatusCode, tc.command)
		}
		
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return fmt.Errorf("failed to decode analyze response: %v", err)
		}
		
		valid, ok := result["valid"].(bool)
		if !ok {
			return fmt.Errorf("analyze response missing 'valid' field")
		}
		
		if valid != tc.expectValid {
			return fmt.Errorf("command '%s' expected valid=%v but got valid=%v", tc.command, tc.expectValid, valid)
		}
	}
	
	return nil
}

func (t *IntegrationTester) testExecuteEndpoint() error {
	// Test SET command
	payload := map[string]string{"command": "SET test:integration:key \"test value\""}
	jsonData, _ := json.Marshal(payload)
	
	resp, err := http.Post(t.apiURL+"/execute", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to execute SET command: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return fmt.Errorf("execute endpoint returned status %d", resp.StatusCode)
	}
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode execute response: %v", err)
	}
	
	success, ok := result["success"].(bool)
	if !ok || !success {
		return fmt.Errorf("SET command execution failed: %v", result["error"])
	}
	
	// Test GET command
	payload = map[string]string{"command": "GET test:integration:key"}
	jsonData, _ = json.Marshal(payload)
	
	resp, err = http.Post(t.apiURL+"/execute", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to execute GET command: %v", err)
	}
	defer resp.Body.Close()
	
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode GET response: %v", err)
	}
	
	success, ok = result["success"].(bool)
	if !ok || !success {
		return fmt.Errorf("GET command execution failed: %v", result["error"])
	}
	
	if result["result"] != "test value" {
		return fmt.Errorf("GET command returned unexpected value: %v", result["result"])
	}
	
	// Cleanup
	payload = map[string]string{"command": "DEL test:integration:key"}
	jsonData, _ = json.Marshal(payload)
	http.Post(t.apiURL+"/execute", "application/json", bytes.NewBuffer(jsonData))
	
	return nil
}

func (t *IntegrationTester) testDatabaseInfo() error {
	resp, err := http.Get(t.apiURL + "/database/info")
	if err != nil {
		return fmt.Errorf("failed to get database info: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return fmt.Errorf("database info endpoint returned status %d", resp.StatusCode)
	}
	
	var info map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return fmt.Errorf("failed to decode database info response: %v", err)
	}
	
	if info["version"] == nil {
		return fmt.Errorf("database info missing version field")
	}
	
	return nil
}

func (t *IntegrationTester) testKeysEndpoint() error {
	// First, create some test keys
	testKeys := []string{
		"SET test:keys:1 value1",
		"SET test:keys:2 value2",
		"SET test:keys:3 value3",
	}
	
	for _, cmd := range testKeys {
		payload := map[string]string{"command": cmd}
		jsonData, _ := json.Marshal(payload)
		http.Post(t.apiURL+"/execute", "application/json", bytes.NewBuffer(jsonData))
	}
	
	// Test listing keys
	resp, err := http.Get(t.apiURL + "/keys?pattern=test:keys:*&limit=10")
	if err != nil {
		return fmt.Errorf("failed to list keys: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return fmt.Errorf("keys endpoint returned status %d", resp.StatusCode)
	}
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode keys response: %v", err)
	}
	
	keys, ok := result["keys"].([]interface{})
	if !ok {
		return fmt.Errorf("keys response missing 'keys' field")
	}
	
	if len(keys) < 3 {
		return fmt.Errorf("expected at least 3 keys, got %d", len(keys))
	}
	
	// Test deleting a key
	req, _ := http.NewRequest("DELETE", t.apiURL+"/keys/test:keys:1", nil)
	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete key: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return fmt.Errorf("delete key endpoint returned status %d", resp.StatusCode)
	}
	
	// Cleanup remaining keys
	for _, key := range []string{"test:keys:2", "test:keys:3"} {
		req, _ := http.NewRequest("DELETE", t.apiURL+"/keys/"+key, nil)
		client.Do(req)
	}
	
	return nil
}

func (t *IntegrationTester) testCommandSpecs() error {
	resp, err := http.Get(t.apiURL + "/commands")
	if err != nil {
		return fmt.Errorf("failed to get command specs: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return fmt.Errorf("commands endpoint returned status %d", resp.StatusCode)
	}
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode commands response: %v", err)
	}
	
	commands, ok := result["commands"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("commands response missing 'commands' field")
	}
	
	expectedCommands := []string{"GET", "SET", "DEL", "HGET", "HSET", "ZADD", "ZRANGE", "SCAN"}
	for _, cmd := range expectedCommands {
		if _, exists := commands[cmd]; !exists {
			return fmt.Errorf("expected command '%s' not found in specs", cmd)
		}
	}
	
	return nil
}

func (t *IntegrationTester) testFrontendAccess() error {
	resp, err := http.Get(t.frontendURL)
	if err != nil {
		return fmt.Errorf("failed to access frontend: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 200 {
		return fmt.Errorf("frontend returned status %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read frontend response: %v", err)
	}
	
	bodyStr := string(body)
	if !strings.Contains(bodyStr, "Redis Analyzer") {
		return fmt.Errorf("frontend does not contain expected title")
	}
	
	return nil
}

func (t *IntegrationTester) testCORS() error {
	req, _ := http.NewRequest("OPTIONS", t.apiURL+"/health", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to test CORS: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != 204 {
		return fmt.Errorf("CORS preflight returned status %d", resp.StatusCode)
	}
	
	allowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
	if allowOrigin != "*" {
		return fmt.Errorf("CORS Allow-Origin header incorrect: %s", allowOrigin)
	}
	
	return nil
}

func (t *IntegrationTester) runAllTests() {
	fmt.Printf("%s=== Redis Analyzer Integration Tests ===%s\n\n", ColorCyan, ColorReset)
	
	// API Tests
	fmt.Printf("%s[API TESTS]%s\n", ColorPurple, ColorReset)
	t.runTest("API Health Check", t.testAPIHealth)
	t.runTest("Analyze Endpoint", t.testAnalyzeEndpoint)
	t.runTest("Execute Endpoint", t.testExecuteEndpoint)
	t.runTest("Database Info", t.testDatabaseInfo)
	t.runTest("Keys Management", t.testKeysEndpoint)
	t.runTest("Command Specifications", t.testCommandSpecs)
	t.runTest("CORS Configuration", t.testCORS)
	
	// Frontend Tests
	fmt.Printf("\n%s[FRONTEND TESTS]%s\n", ColorPurple, ColorReset)
	t.runTest("Frontend Access", t.testFrontendAccess)
	
	// Summary
	t.printSummary()
}

func (t *IntegrationTester) printSummary() {
	fmt.Printf("\n%s=== TEST SUMMARY ===%s\n", ColorCyan, ColorReset)
	
	passed := 0
	failed := 0
	totalDuration := time.Duration(0)
	
	for _, result := range t.results {
		totalDuration += result.Duration
		if result.Success {
			passed++
		} else {
			failed++
		}
	}
	
	fmt.Printf("Total Tests: %d\n", len(t.results))
	fmt.Printf("%sPassed: %d%s\n", ColorGreen, passed, ColorReset)
	if failed > 0 {
		fmt.Printf("%sFailed: %d%s\n", ColorRed, failed, ColorReset)
	} else {
		fmt.Printf("Failed: %d\n", failed)
	}
	fmt.Printf("Total Duration: %v\n", totalDuration)
	
	if failed > 0 {
		fmt.Printf("\n%s[FAILED TESTS]%s\n", ColorRed, ColorReset)
		for _, result := range t.results {
			if !result.Success {
				fmt.Printf("- %s: %s\n", result.Name, result.Error)
			}
		}
	}
	
	if failed == 0 {
		fmt.Printf("\n%s🎉 All tests passed! The Redis Analyzer is working correctly.%s\n", ColorGreen, ColorReset)
	} else {
		fmt.Printf("\n%s❌ Some tests failed. Please check the errors above.%s\n", ColorRed, ColorReset)
	}
}

func waitForServer(url string, timeout time.Duration) bool {
	fmt.Printf("Waiting for server at %s...\n", url)
	
	start := time.Now()
	for time.Since(start) < timeout {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			fmt.Printf("✓ Server is ready\n")
			return true
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}
	
	fmt.Printf("✗ Server not ready after %v\n", timeout)
	return false
}

func checkRedis() bool {
	fmt.Printf("Checking Redis connection...\n")
	
	cmd := exec.Command("redis-cli", "ping")
	output, err := cmd.Output()
	if err != nil {
		fmt.Printf("✗ Redis not available: %v\n", err)
		return false
	}
	
	if strings.TrimSpace(string(output)) == "PONG" {
		fmt.Printf("✓ Redis is running\n")
		return true
	}
	
	fmt.Printf("✗ Redis ping failed: %s\n", output)
	return false
}

func startServers() (*exec.Cmd, *exec.Cmd, error) {
	fmt.Printf("%s=== Starting Servers ===%s\n", ColorCyan, ColorReset)
	
	// Start API server
	fmt.Printf("Starting API server...\n")
	apiCmd := exec.Command("go", "run", "main.go")
	apiCmd.Dir = "/home/ubuntu/redis-analyzer-api/backend"
	apiCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	
	if err := apiCmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("failed to start API server: %v", err)
	}
	
	// Wait for API server
	if !waitForServer("http://localhost:8080/api/v1/health", 30*time.Second) {
		apiCmd.Process.Kill()
		return nil, nil, fmt.Errorf("API server failed to start")
	}
	
	// Start frontend server
	fmt.Printf("Starting frontend server...\n")
	frontendCmd := exec.Command("pnpm", "run", "dev", "--host")
	frontendCmd.Dir = "/home/ubuntu/redis-analyzer-api/frontend"
	frontendCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	
	if err := frontendCmd.Start(); err != nil {
		apiCmd.Process.Kill()
		return nil, nil, fmt.Errorf("failed to start frontend server: %v", err)
	}
	
	// Wait for frontend server
	if !waitForServer("http://localhost:5173", 30*time.Second) {
		apiCmd.Process.Kill()
		frontendCmd.Process.Kill()
		return nil, nil, fmt.Errorf("frontend server failed to start")
	}
	
	fmt.Printf("✓ Both servers are running\n\n")
	return apiCmd, frontendCmd, nil
}

func stopServers(apiCmd, frontendCmd *exec.Cmd) {
	fmt.Printf("\n%s=== Stopping Servers ===%s\n", ColorCyan, ColorReset)
	
	if apiCmd != nil && apiCmd.Process != nil {
		syscall.Kill(-apiCmd.Process.Pid, syscall.SIGTERM)
		apiCmd.Wait()
		fmt.Printf("✓ API server stopped\n")
	}
	
	if frontendCmd != nil && frontendCmd.Process != nil {
		syscall.Kill(-frontendCmd.Process.Pid, syscall.SIGTERM)
		frontendCmd.Wait()
		fmt.Printf("✓ Frontend server stopped\n")
	}
}

func main() {
	// Check prerequisites
	if !checkRedis() {
		fmt.Printf("%sPlease start Redis server first:%s\n", ColorRed, ColorReset)
		fmt.Printf("sudo systemctl start redis-server\n")
		os.Exit(1)
	}
	
	// Start servers
	apiCmd, frontendCmd, err := startServers()
	if err != nil {
		fmt.Printf("%sError starting servers: %v%s\n", ColorRed, err, ColorReset)
		os.Exit(1)
	}
	
	// Ensure servers are stopped on exit
	defer stopServers(apiCmd, frontendCmd)
	
	// Run tests
	tester := NewIntegrationTester()
	tester.runAllTests()
	
	// Exit with appropriate code
	failed := 0
	for _, result := range tester.results {
		if !result.Success {
			failed++
		}
	}
	
	if failed > 0 {
		os.Exit(1)
	}
}

