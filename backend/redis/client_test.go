package redis

import (
	"testing"
	"time"
)

// MockRedisClient para pruebas sin conexión real a Redis
type MockRedisClient struct {
	data map[string]interface{}
}

func NewMockClient() *MockRedisClient {
	return &MockRedisClient{
		data: make(map[string]interface{}),
	}
}

func TestExecuteCommand(t *testing.T) {
	// Nota: Estas pruebas requieren una instancia de Redis ejecutándose
	// Para pruebas unitarias reales, se debería usar un mock
	
	tests := []struct {
		name        string
		command     string
		expectError bool
	}{
		{
			name:        "Valid SET command",
			command:     `SET testkey "testvalue"`,
			expectError: false,
		},
		{
			name:        "Valid GET command",
			command:     "GET testkey",
			expectError: false,
		},
		{
			name:        "Invalid command syntax",
			command:     "INVALID",
			expectError: true,
		},
		{
			name:        "SET with expiration",
			command:     `SET expkey "value" EX 60`,
			expectError: false,
		},
		{
			name:        "GET non-existent key",
			command:     "GET nonexistent",
			expectError: false, // Redis devuelve nil, no error
		},
	}
	
	// Crear cliente (esto requiere Redis ejecutándose)
	config := Config{
		Host: "localhost",
		Port: 6379,
		DB:   0,
	}
	
	client := NewClient(config)
	
	// Intentar conectar (si falla, saltar las pruebas)
	if err := client.Connect(); err != nil {
		t.Skipf("Redis not available, skipping integration tests: %v", err)
		return
	}
	defer client.Close()
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.ExecuteCommand(tt.command)
			
			if tt.expectError && result.Success {
				t.Errorf("Expected error but command succeeded")
			}
			
			if !tt.expectError && !result.Success && result.Error != "" {
				t.Errorf("Expected success but got error: %s", result.Error)
			}
			
			// Verificar que la validación semántica se ejecutó
			if result.Validation == nil {
				t.Errorf("Expected validation result")
			}
			
			// Verificar que se midió el tiempo de ejecución
			if result.ExecutionTime == 0 {
				t.Errorf("Expected non-zero execution time")
			}
		})
	}
}

func TestExtractValues(t *testing.T) {
	client := NewClient(Config{})
	
	// Crear expresiones de prueba manualmente
	// Nota: En un test real, usaríamos el parser para crear estas expresiones
	
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		// Estas pruebas requerirían crear expresiones del parser
		// Por simplicidad, las omitimos aquí
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Implementar pruebas de extracción de valores
			// cuando tengamos las expresiones del parser
		})
	}
}

func TestCommandValidation(t *testing.T) {
	client := NewClient(Config{})
	
	tests := []struct {
		name          string
		command       string
		expectValid   bool
		expectError   string
	}{
		{
			name:        "Valid SET command",
			command:     `SET key "value"`,
			expectValid: true,
		},
		{
			name:        "Invalid GET command - too many args",
			command:     "GET key1 key2",
			expectValid: false,
			expectError: "Too many arguments",
		},
		{
			name:        "Invalid SET command - conflicting options",
			command:     `SET key "value" NX XX`,
			expectValid: false,
			expectError: "conflicts with",
		},
		{
			name:        "Unknown command",
			command:     "UNKNOWN key",
			expectValid: false,
			expectError: "Unknown command",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.ExecuteCommand(tt.command)
			
			if result.Validation == nil {
				t.Fatal("Expected validation result")
			}
			
			if result.Validation.Valid != tt.expectValid {
				t.Errorf("Expected valid=%v, got valid=%v", tt.expectValid, result.Validation.Valid)
			}
			
			if !tt.expectValid && tt.expectError != "" {
				found := false
				for _, err := range result.Validation.Errors {
					if contains(err.Message, tt.expectError) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error containing '%s', got errors: %v", tt.expectError, result.Validation.Errors)
				}
			}
		})
	}
}

func TestDatabaseOperations(t *testing.T) {
	config := Config{
		Host: "localhost",
		Port: 6379,
		DB:   0,
	}
	
	client := NewClient(config)
	
	if err := client.Connect(); err != nil {
		t.Skipf("Redis not available, skipping integration tests: %v", err)
		return
	}
	defer client.Close()
	
	// Test GetDatabaseInfo
	t.Run("GetDatabaseInfo", func(t *testing.T) {
		info, err := client.GetDatabaseInfo()
		if err != nil {
			t.Errorf("GetDatabaseInfo failed: %v", err)
		}
		
		if info.Version == "" {
			t.Error("Expected non-empty version")
		}
	})
	
	// Test ListKeys
	t.Run("ListKeys", func(t *testing.T) {
		// Primero insertar algunas claves de prueba
		client.ExecuteCommand(`SET testkey1 "value1"`)
		client.ExecuteCommand(`SET testkey2 "value2"`)
		
		keys, err := client.ListKeys("testkey*", 10)
		if err != nil {
			t.Errorf("ListKeys failed: %v", err)
		}
		
		if len(keys) < 2 {
			t.Errorf("Expected at least 2 keys, got %d", len(keys))
		}
		
		// Limpiar
		client.ExecuteCommand("DEL testkey1 testkey2")
	})
	
	// Test GetKeyInfo
	t.Run("GetKeyInfo", func(t *testing.T) {
		// Insertar una clave de prueba
		client.ExecuteCommand(`SET infokey "value"`)
		
		info, err := client.GetKeyInfo("infokey")
		if err != nil {
			t.Errorf("GetKeyInfo failed: %v", err)
		}
		
		if info["type"] != "string" {
			t.Errorf("Expected type 'string', got %v", info["type"])
		}
		
		// Limpiar
		client.ExecuteCommand("DEL infokey")
	})
}

func TestPerformance(t *testing.T) {
	config := Config{
		Host: "localhost",
		Port: 6379,
		DB:   0,
	}
	
	client := NewClient(config)
	
	if err := client.Connect(); err != nil {
		t.Skipf("Redis not available, skipping performance tests: %v", err)
		return
	}
	defer client.Close()
	
	// Test de rendimiento básico
	start := time.Now()
	iterations := 100
	
	for i := 0; i < iterations; i++ {
		result := client.ExecuteCommand(`SET perfkey "value"`)
		if !result.Success {
			t.Errorf("Command failed at iteration %d: %s", i, result.Error)
		}
	}
	
	duration := time.Since(start)
	avgTime := duration / time.Duration(iterations)
	
	t.Logf("Average execution time: %v", avgTime)
	
	// Verificar que el tiempo promedio sea razonable (menos de 10ms)
	if avgTime > 10*time.Millisecond {
		t.Logf("Warning: Average execution time is high: %v", avgTime)
	}
	
	// Limpiar
	client.ExecuteCommand("DEL perfkey")
}

// Helper function para verificar si una cadena contiene una subcadena
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
			(s[:len(substr)] == substr || 
			 s[len(s)-len(substr):] == substr || 
			 containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

