package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	
	"redis-analyzer-api/redis"
)

func TestAnalyzeEndpoint(t *testing.T) {
	// Crear servidor de prueba
	config := redis.Config{
		Host: "localhost",
		Port: 6379,
		DB:   1, // Usar DB diferente para pruebas
	}
	
	server := NewServer(config)
	
	tests := []struct {
		name           string
		request        AnalyzeRequest
		expectedStatus int
		expectValid    bool
	}{
		{
			name:           "Valid GET command",
			request:        AnalyzeRequest{Command: "GET mykey"},
			expectedStatus: http.StatusOK,
			expectValid:    true,
		},
		{
			name:           "Valid SET command with options",
			request:        AnalyzeRequest{Command: `SET key "value" EX 60`},
			expectedStatus: http.StatusOK,
			expectValid:    true,
		},
		{
			name:           "Invalid command - too few args",
			request:        AnalyzeRequest{Command: "GET"},
			expectedStatus: http.StatusOK,
			expectValid:    false,
		},
		{
			name:           "Unknown command",
			request:        AnalyzeRequest{Command: "UNKNOWN key"},
			expectedStatus: http.StatusOK,
			expectValid:    false,
		},
		{
			name:           "Empty command",
			request:        AnalyzeRequest{Command: ""},
			expectedStatus: http.StatusBadRequest,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Preparar request
			jsonData, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest("POST", "/api/v1/analyze", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			
			// Crear response recorder
			w := httptest.NewRecorder()
			
			// Ejecutar request
			server.router.ServeHTTP(w, req)
			
			// Verificar status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			
			// Si esperamos 200, verificar el contenido
			if tt.expectedStatus == http.StatusOK {
				var response AnalyzeResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Error parsing response: %v", err)
				}
				
				if response.Valid != tt.expectValid {
					t.Errorf("Expected valid=%v, got valid=%v", tt.expectValid, response.Valid)
				}
				
				if tt.expectValid {
					if response.ParsedAST == "" {
						t.Error("Expected non-empty parsed AST")
					}
					if response.Validation == nil {
						t.Error("Expected validation result")
					}
				}
			}
		})
	}
}

func TestExecuteEndpoint(t *testing.T) {
	// Crear servidor de prueba
	config := redis.Config{
		Host: "localhost",
		Port: 6379,
		DB:   1, // Usar DB diferente para pruebas
	}
	
	server := NewServer(config)
	
	// Verificar que Redis esté disponible
	if err := server.redisClient.Connect(); err != nil {
		t.Skipf("Redis not available, skipping integration tests: %v", err)
		return
	}
	defer server.redisClient.Close()
	
	tests := []struct {
		name           string
		request        ExecuteRequest
		expectedStatus int
		expectSuccess  bool
	}{
		{
			name:           "Valid SET command",
			request:        ExecuteRequest{Command: `SET testkey "testvalue"`},
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "Valid GET command",
			request:        ExecuteRequest{Command: "GET testkey"},
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
		},
		{
			name:           "Invalid command syntax",
			request:        ExecuteRequest{Command: "GET"},
			expectedStatus: http.StatusOK,
			expectSuccess:  false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Preparar request
			jsonData, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest("POST", "/api/v1/execute", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			
			// Crear response recorder
			w := httptest.NewRecorder()
			
			// Ejecutar request
			server.router.ServeHTTP(w, req)
			
			// Verificar status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
			
			// Verificar contenido de respuesta
			if tt.expectedStatus == http.StatusOK {
				var response ExecuteResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Error parsing response: %v", err)
				}
				
				if response.Success != tt.expectSuccess {
					t.Errorf("Expected success=%v, got success=%v. Error: %s", 
						tt.expectSuccess, response.Success, response.Error)
				}
				
				if response.Validation == nil {
					t.Error("Expected validation result")
				}
				
				if response.ExecutionTime == "" {
					t.Error("Expected non-empty execution time")
				}
			}
		})
	}
	
	// Limpiar datos de prueba
	server.redisClient.ExecuteCommand("DEL testkey")
}

func TestDatabaseInfoEndpoint(t *testing.T) {
	config := redis.Config{
		Host: "localhost",
		Port: 6379,
		DB:   1,
	}
	
	server := NewServer(config)
	
	if err := server.redisClient.Connect(); err != nil {
		t.Skipf("Redis not available, skipping test: %v", err)
		return
	}
	defer server.redisClient.Close()
	
	req, _ := http.NewRequest("GET", "/api/v1/database/info", nil)
	w := httptest.NewRecorder()
	
	server.router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response DatabaseInfoResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Error parsing response: %v", err)
	}
	
	if response.Version == "" {
		t.Error("Expected non-empty version")
	}
}

func TestKeysEndpoint(t *testing.T) {
	config := redis.Config{
		Host: "localhost",
		Port: 6379,
		DB:   1,
	}
	
	server := NewServer(config)
	
	if err := server.redisClient.Connect(); err != nil {
		t.Skipf("Redis not available, skipping test: %v", err)
		return
	}
	defer server.redisClient.Close()
	
	// Insertar algunas claves de prueba
	server.redisClient.ExecuteCommand(`SET testkey1 "value1"`)
	server.redisClient.ExecuteCommand(`SET testkey2 "value2"`)
	
	// Test listar claves
	req, _ := http.NewRequest("GET", "/api/v1/keys?pattern=testkey*", nil)
	w := httptest.NewRecorder()
	
	server.router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response KeysResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Error parsing response: %v", err)
	}
	
	if len(response.Keys) < 2 {
		t.Errorf("Expected at least 2 keys, got %d", len(response.Keys))
	}
	
	// Test información de clave
	req, _ = http.NewRequest("GET", "/api/v1/keys/testkey1", nil)
	w = httptest.NewRecorder()
	
	server.router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var keyResponse KeyInfoResponse
	err = json.Unmarshal(w.Body.Bytes(), &keyResponse)
	if err != nil {
		t.Errorf("Error parsing response: %v", err)
	}
	
	if !keyResponse.Exists {
		t.Error("Expected key to exist")
	}
	
	// Limpiar
	server.redisClient.ExecuteCommand("DEL testkey1 testkey2")
}

func TestHealthEndpoint(t *testing.T) {
	config := redis.Config{
		Host: "localhost",
		Port: 6379,
		DB:   1,
	}
	
	server := NewServer(config)
	
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()
	
	server.router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Error parsing response: %v", err)
	}
	
	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got %v", response["status"])
	}
	
	if response["version"] == nil {
		t.Error("Expected version in response")
	}
}

func TestCommandSpecsEndpoint(t *testing.T) {
	config := redis.Config{
		Host: "localhost",
		Port: 6379,
		DB:   1,
	}
	
	server := NewServer(config)
	
	req, _ := http.NewRequest("GET", "/api/v1/commands", nil)
	w := httptest.NewRecorder()
	
	server.router.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response CommandSpecsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Error parsing response: %v", err)
	}
	
	if len(response.Commands) == 0 {
		t.Error("Expected non-empty commands list")
	}
	
	// Verificar que algunos comandos básicos estén presentes
	expectedCommands := []string{"GET", "SET", "DEL"}
	for _, cmd := range expectedCommands {
		if _, exists := response.Commands[cmd]; !exists {
			t.Errorf("Expected command %s to be present", cmd)
		}
	}
}

func TestCORS(t *testing.T) {
	config := redis.Config{
		Host: "localhost",
		Port: 6379,
		DB:   1,
	}
	
	server := NewServer(config)
	
	// Test OPTIONS request
	req, _ := http.NewRequest("OPTIONS", "/api/v1/health", nil)
	w := httptest.NewRecorder()
	
	server.router.ServeHTTP(w, req)
	
	if w.Code != 204 {
		t.Errorf("Expected status 204 for OPTIONS, got %d", w.Code)
	}
	
	// Verificar headers CORS
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("Expected CORS origin header")
	}
	
	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("Expected CORS methods header")
	}
}

