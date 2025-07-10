package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Estructuras para las requests/responses de la API
type AnalyzeRequest struct {
	Command string `json:"command"`
}

type ExecuteRequest struct {
	Command string `json:"command"`
}

func main() {
	fmt.Println("=== Demostración de la API REST Redis Analyzer ===\n")
	
	// URL base de la API
	baseURL := "http://localhost:8080/api/v1"
	
	// Esperar a que el servidor esté listo
	fmt.Println("Esperando a que el servidor esté listo...")
	if !waitForServer(baseURL + "/health") {
		fmt.Println("❌ El servidor no está disponible. Asegúrate de ejecutar:")
		fmt.Println("   go run main.go")
		return
	}
	fmt.Println("✅ Servidor listo\n")
	
	// Test 1: Health Check
	fmt.Println("=== 1. Health Check ===")
	testHealthCheck(baseURL)
	
	// Test 2: Obtener especificaciones de comandos
	fmt.Println("\n=== 2. Especificaciones de Comandos ===")
	testCommandSpecs(baseURL)
	
	// Test 3: Analizar comandos
	fmt.Println("\n=== 3. Análisis de Comandos ===")
	testAnalyzeCommands(baseURL)
	
	// Test 4: Ejecutar comandos
	fmt.Println("\n=== 4. Ejecución de Comandos ===")
	testExecuteCommands(baseURL)
	
	// Test 5: Información de base de datos
	fmt.Println("\n=== 5. Información de Base de Datos ===")
	testDatabaseInfo(baseURL)
	
	// Test 6: Gestión de claves
	fmt.Println("\n=== 6. Gestión de Claves ===")
	testKeyManagement(baseURL)
	
	fmt.Println("\n🎉 Demostración completada exitosamente!")
}

func waitForServer(healthURL string) bool {
	for i := 0; i < 30; i++ { // Esperar hasta 30 segundos
		resp, err := http.Get(healthURL)
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return true
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}
	return false
}

func testHealthCheck(baseURL string) {
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode == 200 {
		fmt.Printf("✅ Status: %d\n", resp.StatusCode)
		
		var result map[string]interface{}
		json.Unmarshal(body, &result)
		fmt.Printf("   Estado: %v\n", result["status"])
		fmt.Printf("   Redis: %v\n", result["redis"])
		fmt.Printf("   Versión: %v\n", result["version"])
	} else {
		fmt.Printf("❌ Status: %d\n", resp.StatusCode)
	}
}

func testCommandSpecs(baseURL string) {
	resp, err := http.Get(baseURL + "/commands")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode == 200 {
		var result map[string]interface{}
		json.Unmarshal(body, &result)
		
		commands := result["commands"].(map[string]interface{})
		fmt.Printf("✅ Comandos disponibles: %d\n", len(commands))
		
		// Mostrar algunos comandos
		count := 0
		for name, spec := range commands {
			if count >= 5 { // Mostrar solo los primeros 5
				break
			}
			specMap := spec.(map[string]interface{})
			fmt.Printf("   - %s: %v\n", name, specMap["description"])
			count++
		}
		if len(commands) > 5 {
			fmt.Printf("   ... y %d más\n", len(commands)-5)
		}
	} else {
		fmt.Printf("❌ Status: %d\n", resp.StatusCode)
	}
}

func testAnalyzeCommands(baseURL string) {
	commands := []string{
		"GET mykey",
		`SET user:123 "John Doe" EX 3600`,
		"ZADD scores 100 player1 95 player2",
		"GET", // Comando inválido
		"UNKNOWN command", // Comando desconocido
	}
	
	for i, cmd := range commands {
		fmt.Printf("Comando %d: %s\n", i+1, cmd)
		
		request := AnalyzeRequest{Command: cmd}
		jsonData, _ := json.Marshal(request)
		
		resp, err := http.Post(baseURL+"/analyze", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("   ❌ Error: %v\n", err)
			continue
		}
		defer resp.Body.Close()
		
		body, _ := io.ReadAll(resp.Body)
		
		if resp.StatusCode == 200 {
			var result map[string]interface{}
			json.Unmarshal(body, &result)
			
			valid := result["valid"].(bool)
			if valid {
				fmt.Printf("   ✅ Válido\n")
				if commandInfo, ok := result["command_info"].(map[string]interface{}); ok {
					if name, ok := commandInfo["name"]; ok {
						fmt.Printf("      Comando: %v\n", name)
					}
				}
			} else {
				fmt.Printf("   ❌ Inválido\n")
				if validation, ok := result["validation"].(map[string]interface{}); ok {
					if errors, ok := validation["errors"].([]interface{}); ok && len(errors) > 0 {
						for _, err := range errors {
							if errMap, ok := err.(map[string]interface{}); ok {
								fmt.Printf("      Error: %v\n", errMap["message"])
							}
						}
					}
				}
			}
		} else {
			fmt.Printf("   ❌ Status: %d\n", resp.StatusCode)
		}
		fmt.Println()
	}
}

func testExecuteCommands(baseURL string) {
	commands := []string{
		`SET demo:key "Hello Redis!"`,
		"GET demo:key",
		`SET demo:temp "temporary" EX 60`,
		"ZADD demo:scores 100 alice 95 bob 87 charlie",
		"ZRANGE demo:scores 0 -1 WITHSCORES",
	}
	
	for i, cmd := range commands {
		fmt.Printf("Ejecutando %d: %s\n", i+1, cmd)
		
		request := ExecuteRequest{Command: cmd}
		jsonData, _ := json.Marshal(request)
		
		resp, err := http.Post(baseURL+"/execute", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("   ❌ Error: %v\n", err)
			continue
		}
		defer resp.Body.Close()
		
		body, _ := io.ReadAll(resp.Body)
		
		if resp.StatusCode == 200 {
			var result map[string]interface{}
			json.Unmarshal(body, &result)
			
			success := result["success"].(bool)
			if success {
				fmt.Printf("   ✅ Éxito (tiempo: %v)\n", result["execution_time"])
				if result["result"] != nil {
					fmt.Printf("      Resultado: %v\n", result["result"])
				}
			} else {
				fmt.Printf("   ❌ Error: %v\n", result["error"])
			}
		} else {
			fmt.Printf("   ❌ Status: %d\n", resp.StatusCode)
		}
		fmt.Println()
	}
}

func testDatabaseInfo(baseURL string) {
	resp, err := http.Get(baseURL + "/database/info")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode == 200 {
		var result map[string]interface{}
		json.Unmarshal(body, &result)
		
		fmt.Printf("✅ Información de la base de datos:\n")
		fmt.Printf("   Versión: %v\n", result["version"])
		fmt.Printf("   Número de claves: %v\n", result["key_count"])
		
		if memory, ok := result["memory"].(map[string]interface{}); ok {
			for key, value := range memory {
				fmt.Printf("   %s: %v\n", key, value)
				break // Solo mostrar uno para no saturar
			}
		}
	} else {
		fmt.Printf("❌ Status: %d\n", resp.StatusCode)
	}
}

func testKeyManagement(baseURL string) {
	// Listar claves
	fmt.Println("Listando claves que empiezan con 'demo:'")
	resp, err := http.Get(baseURL + "/keys?pattern=demo:*&limit=10")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := io.ReadAll(resp.Body)
	
	if resp.StatusCode == 200 {
		var result map[string]interface{}
		json.Unmarshal(body, &result)
		
		keys := result["keys"].([]interface{})
		fmt.Printf("✅ Encontradas %d claves:\n", len(keys))
		for _, key := range keys {
			fmt.Printf("   - %v\n", key)
		}
		
		// Obtener información de la primera clave
		if len(keys) > 0 {
			firstKey := keys[0].(string)
			fmt.Printf("\nInformación de la clave '%s':\n", firstKey)
			
			resp2, err := http.Get(baseURL + "/keys/" + firstKey)
			if err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				return
			}
			defer resp2.Body.Close()
			
			body2, _ := io.ReadAll(resp2.Body)
			
			if resp2.StatusCode == 200 {
				var keyResult map[string]interface{}
				json.Unmarshal(body2, &keyResult)
				
				if keyResult["exists"].(bool) {
					fmt.Printf("✅ La clave existe\n")
					if info, ok := keyResult["info"].(map[string]interface{}); ok {
						for k, v := range info {
							fmt.Printf("   %s: %v\n", k, v)
						}
					}
				} else {
					fmt.Printf("❌ La clave no existe\n")
				}
			}
		}
	} else {
		fmt.Printf("❌ Status: %d\n", resp.StatusCode)
	}
	
	// Limpiar claves de demostración
	fmt.Println("\nLimpiando claves de demostración...")
	cleanupKeys := []string{"demo:key", "demo:temp", "demo:scores"}
	for _, key := range cleanupKeys {
		req, _ := http.NewRequest("DELETE", baseURL+"/keys/"+key, nil)
		client := &http.Client{}
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == 200 {
			fmt.Printf("✅ Eliminada: %s\n", key)
			resp.Body.Close()
		}
	}
}

