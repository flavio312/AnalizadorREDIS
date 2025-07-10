package semantic

import (
	"testing"
	"redis-analyzer-api/parser"
)

func TestValidateSimpleCommands(t *testing.T) {
	analyzer := New()
	
	tests := []struct {
		name        string
		input       string
		expectValid bool
		expectError string
	}{
		{
			name:        "Valid GET command",
			input:       "GET mykey",
			expectValid: true,
		},
		{
			name:        "Valid SET command",
			input:       `SET mykey "value"`,
			expectValid: true,
		},
		{
			name:        "GET with too many arguments",
			input:       "GET key1 key2",
			expectValid: false,
			expectError: "Too many arguments",
		},
		{
			name:        "GET with too few arguments",
			input:       "GET",
			expectValid: false,
			expectError: "Too few arguments",
		},
		{
			name:        "Unknown command",
			input:       "UNKNOWN key",
			expectValid: false,
			expectError: "Unknown command",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, parseErrors := parser.ParseCommand(tt.input)
			if len(parseErrors) > 0 {
				t.Fatalf("Parse error: %v", parseErrors)
			}
			
			result := analyzer.ValidateCommand(cmd)
			
			if result.Valid != tt.expectValid {
				t.Errorf("Expected valid=%v, got valid=%v", tt.expectValid, result.Valid)
			}
			
			if !tt.expectValid && tt.expectError != "" {
				found := false
				for _, err := range result.Errors {
					if contains(err.Message, tt.expectError) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error containing '%s', got errors: %v", tt.expectError, result.Errors)
				}
			}
		})
	}
}

func TestValidateSetCommandWithOptions(t *testing.T) {
	analyzer := New()
	
	tests := []struct {
		name        string
		input       string
		expectValid bool
		expectError string
	}{
		{
			name:        "SET with EX option",
			input:       `SET key "value" EX 60`,
			expectValid: true,
		},
		{
			name:        "SET with PX option",
			input:       `SET key "value" PX 60000`,
			expectValid: true,
		},
		{
			name:        "SET with NX option",
			input:       `SET key "value" NX`,
			expectValid: true,
		},
		{
			name:        "SET with conflicting options NX and XX",
			input:       `SET key "value" NX XX`,
			expectValid: false,
			expectError: "conflicts with",
		},
		{
			name:        "SET with conflicting options EX and PX",
			input:       `SET key "value" EX 60 PX 60000`,
			expectValid: false,
			expectError: "conflicts with",
		},
		{
			name:        "SET with EX but no value",
			input:       `SET key "value" EX`,
			expectValid: false,
			expectError: "requires a value",
		},
		{
			name:        "SET with negative expiration",
			input:       `SET key "value" EX -60`,
			expectValid: false,
			expectError: "must be positive",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, parseErrors := parser.ParseCommand(tt.input)
			if len(parseErrors) > 0 {
				t.Fatalf("Parse error: %v", parseErrors)
			}
			
			result := analyzer.ValidateCommand(cmd)
			
			if result.Valid != tt.expectValid {
				t.Errorf("Expected valid=%v, got valid=%v. Errors: %v", tt.expectValid, result.Valid, result.Errors)
			}
			
			if !tt.expectValid && tt.expectError != "" {
				found := false
				for _, err := range result.Errors {
					if contains(err.Message, tt.expectError) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error containing '%s', got errors: %v", tt.expectError, result.Errors)
				}
			}
		})
	}
}

func TestValidateZAddCommand(t *testing.T) {
	analyzer := New()
	
	tests := []struct {
		name        string
		input       string
		expectValid bool
		expectError string
	}{
		{
			name:        "Valid ZADD command",
			input:       "ZADD myset 1.5 member1",
			expectValid: true,
		},
		{
			name:        "ZADD with multiple members",
			input:       "ZADD myset 1.5 member1 2.0 member2",
			expectValid: true,
		},
		{
			name:        "ZADD with invalid score type",
			input:       `ZADD myset "invalid" member1`,
			expectValid: false,
			expectError: "should be a numeric score",
		},
		{
			name:        "ZADD with too few arguments",
			input:       "ZADD myset 1.5",
			expectValid: false,
			expectError: "Too few arguments",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, parseErrors := parser.ParseCommand(tt.input)
			if len(parseErrors) > 0 {
				t.Fatalf("Parse error: %v", parseErrors)
			}
			
			result := analyzer.ValidateCommand(cmd)
			
			if result.Valid != tt.expectValid {
				t.Errorf("Expected valid=%v, got valid=%v. Errors: %v", tt.expectValid, result.Valid, result.Errors)
			}
			
			if !tt.expectValid && tt.expectError != "" {
				found := false
				for _, err := range result.Errors {
					if contains(err.Message, tt.expectError) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error containing '%s', got errors: %v", tt.expectError, result.Errors)
				}
			}
		})
	}
}

func TestValidateScanCommand(t *testing.T) {
	analyzer := New()
	
	tests := []struct {
		name        string
		input       string
		expectValid bool
		expectError string
	}{
		{
			name:        "Valid SCAN command",
			input:       "SCAN 0",
			expectValid: true,
		},
		{
			name:        "SCAN with MATCH option",
			input:       "SCAN 0 MATCH user:*",
			expectValid: true,
		},
		{
			name:        "SCAN with COUNT option",
			input:       "SCAN 0 COUNT 10",
			expectValid: true,
		},
		{
			name:        "SCAN with both MATCH and COUNT",
			input:       "SCAN 0 MATCH user:* COUNT 10",
			expectValid: true,
		},
		{
			name:        "SCAN with invalid cursor type",
			input:       `SCAN "invalid"`,
			expectValid: false,
			expectError: "should be a cursor",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, parseErrors := parser.ParseCommand(tt.input)
			if len(parseErrors) > 0 {
				t.Fatalf("Parse error: %v", parseErrors)
			}
			
			result := analyzer.ValidateCommand(cmd)
			
			if result.Valid != tt.expectValid {
				t.Errorf("Expected valid=%v, got valid=%v. Errors: %v", tt.expectValid, result.Valid, result.Errors)
			}
			
			if !tt.expectValid && tt.expectError != "" {
				found := false
				for _, err := range result.Errors {
					if contains(err.Message, tt.expectError) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error containing '%s', got errors: %v", tt.expectError, result.Errors)
				}
			}
		})
	}
}

func TestValidateProgram(t *testing.T) {
	analyzer := New()
	
	input := `SET key1 value1
GET key1
INVALID command
DEL key1`
	
	program, parseErrors := parser.ParseCommands(input)
	if len(parseErrors) > 0 {
		t.Fatalf("Parse errors: %v", parseErrors)
	}
	
	results := analyzer.ValidateProgram(program)
	
	if len(results) != 4 {
		t.Errorf("Expected 4 validation results, got %d", len(results))
	}
	
	// Verificar que los comandos válidos pasen
	if !results[0].Valid {
		t.Errorf("SET command should be valid, errors: %v", results[0].Errors)
	}
	
	if !results[1].Valid {
		t.Errorf("GET command should be valid, errors: %v", results[1].Errors)
	}
	
	// Verificar que el comando inválido falle
	if results[2].Valid {
		t.Errorf("INVALID command should not be valid")
	}
	
	if !results[3].Valid {
		t.Errorf("DEL command should be valid, errors: %v", results[3].Errors)
	}
}

func TestCommandInfo(t *testing.T) {
	analyzer := New()
	
	cmd, parseErrors := parser.ParseCommand(`SET mykey "value" EX 60`)
	if len(parseErrors) > 0 {
		t.Fatalf("Parse error: %v", parseErrors)
	}
	
	result := analyzer.ValidateCommand(cmd)
	
	if !result.Valid {
		t.Fatalf("Command should be valid, errors: %v", result.Errors)
	}
	
	info := result.CommandInfo
	
	if info["name"] != "SET" {
		t.Errorf("Expected command name 'SET', got %v", info["name"])
	}
	
	if info["has_key"] != true {
		t.Errorf("Expected has_key to be true")
	}
	
	if info["key"] != "mykey" {
		t.Errorf("Expected key 'mykey', got %v", info["key"])
	}
	
	if info["description"] == "" {
		t.Errorf("Expected non-empty description")
	}
}

func TestGetCommandSpecs(t *testing.T) {
	analyzer := New()
	
	specs := analyzer.GetCommandSpecs()
	
	if len(specs) == 0 {
		t.Error("Expected non-empty command specs")
	}
	
	// Verificar que algunos comandos básicos estén presentes
	expectedCommands := []string{"GET", "SET", "DEL", "HGET", "HSET", "ZADD", "ZRANGE", "SCAN"}
	
	for _, cmd := range expectedCommands {
		if _, exists := specs[cmd]; !exists {
			t.Errorf("Expected command spec for %s", cmd)
		}
	}
}

func TestAddCommandSpec(t *testing.T) {
	analyzer := New()
	
	// Añadir un comando personalizado
	customSpec := CommandSpec{
		Name:        "CUSTOM",
		MinArgs:     1,
		MaxArgs:     2,
		KeyPosition: 0,
		ValueTypes:  []string{"key", "value"},
		Description: "Custom command for testing",
	}
	
	analyzer.AddCommandSpec(customSpec)
	
	// Verificar que se añadió correctamente
	specs := analyzer.GetCommandSpecs()
	if _, exists := specs["CUSTOM"]; !exists {
		t.Error("Custom command spec was not added")
	}
	
	// Probar validación del comando personalizado
	cmd, parseErrors := parser.ParseCommand("CUSTOM mykey myvalue")
	if len(parseErrors) > 0 {
		t.Fatalf("Parse error: %v", parseErrors)
	}
	
	result := analyzer.ValidateCommand(cmd)
	if !result.Valid {
		t.Errorf("Custom command should be valid, errors: %v", result.Errors)
	}
}

// Helper function to check if a string contains a substring
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

