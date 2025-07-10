package main

import (
	"fmt"
	"strings"
	"redis-analyzer-api/parser"
	"redis-analyzer-api/semantic"
)

func main() {
	fmt.Println("=== Analizador Semántico de Redis ===\n")
	
	analyzer := semantic.New()
	
	// Ejemplos de comandos Redis válidos e inválidos
	commands := []string{
		// Comandos válidos
		"GET mykey",
		`SET user:123 "John Doe" EX 3600`,
		"ZADD leaderboard 100 player1 95 player2",
		"ZRANGE scores 0 -1 WITHSCORES",
		"SCAN 0 MATCH user:* COUNT 10",
		"HSET user:456 name Alice age 30",
		
		// Comandos con errores semánticos
		"GET",                                    // Muy pocos argumentos
		"GET key1 key2",                         // Demasiados argumentos
		"UNKNOWN command",                       // Comando desconocido
		`SET key "value" EX -60`,               // Valor de expiración negativo
		`SET key "value" NX XX`,                // Opciones conflictivas
		`SET key "value" EX`,                   // Opción sin valor
		`ZADD myset "invalid" member1`,         // Score inválido
		`SCAN "invalid"`,                       // Cursor inválido
	}
	
	for i, cmd := range commands {
		fmt.Printf("Comando %d: %s\n", i+1, cmd)
		fmt.Println(strings.Repeat("-", 60))
		
		// Parsear el comando
		parsedCmd, parseErrors := parser.ParseCommand(cmd)
		
		if len(parseErrors) > 0 {
			fmt.Printf("❌ Errores de parsing: %v\n\n", parseErrors)
			continue
		}
		
		// Validar semánticamente
		result := analyzer.ValidateCommand(parsedCmd)
		
		if result.Valid {
			fmt.Printf("✅ Comando válido\n")
		} else {
			fmt.Printf("❌ Comando inválido\n")
		}
		
		// Mostrar información del comando
		if info := result.CommandInfo; len(info) > 0 {
			fmt.Printf("Información del comando:\n")
			if name, ok := info["name"]; ok {
				fmt.Printf("  - Nombre: %v\n", name)
			}
			if desc, ok := info["description"]; ok {
				fmt.Printf("  - Descripción: %v\n", desc)
			}
			if hasKey, ok := info["has_key"]; ok {
				fmt.Printf("  - Tiene clave: %v\n", hasKey)
			}
			if key, ok := info["key"]; ok {
				fmt.Printf("  - Clave: %v\n", key)
			}
		}
		
		// Mostrar errores semánticos
		if len(result.Errors) > 0 {
			fmt.Printf("Errores semánticos:\n")
			for _, err := range result.Errors {
				fmt.Printf("  - [%s] %s\n", err.Type, err.Message)
			}
		}
		
		// Mostrar advertencias
		if len(result.Warnings) > 0 {
			fmt.Printf("Advertencias:\n")
			for _, warning := range result.Warnings {
				fmt.Printf("  - %s\n", warning)
			}
		}
		
		fmt.Println()
	}
	
	// Ejemplo con múltiples comandos
	fmt.Println("=== Validación de Programa Completo ===")
	multiCmd := `SET key1 value1
GET key1
INVALID command
SET key2 value2 EX -30
DEL key1 key2`
	
	fmt.Printf("Programa:\n%s\n", multiCmd)
	fmt.Println(strings.Repeat("-", 60))
	
	program, parseErrors := parser.ParseCommands(multiCmd)
	if len(parseErrors) > 0 {
		fmt.Printf("❌ Errores de parsing: %v\n", parseErrors)
	} else {
		results := analyzer.ValidateProgram(program)
		
		fmt.Printf("Resultados de validación:\n")
		validCount := 0
		for i, result := range results {
			if result.Valid {
				fmt.Printf("  Comando %d: ✅ Válido\n", i+1)
				validCount++
			} else {
				fmt.Printf("  Comando %d: ❌ Inválido\n", i+1)
				for _, err := range result.Errors {
					fmt.Printf("    - %s\n", err.Message)
				}
			}
		}
		
		fmt.Printf("\nResumen: %d/%d comandos válidos\n", validCount, len(results))
	}
	
	// Mostrar especificaciones de comandos disponibles
	fmt.Println("\n=== Comandos Soportados ===")
	specs := analyzer.GetCommandSpecs()
	
	for name, spec := range specs {
		fmt.Printf("- %s: %s\n", name, spec.Description)
		fmt.Printf("  Argumentos: %d", spec.MinArgs)
		if spec.MaxArgs == -1 {
			fmt.Printf("+\n")
		} else {
			fmt.Printf("-%d\n", spec.MaxArgs)
		}
		
		if len(spec.Options) > 0 {
			fmt.Printf("  Opciones: ")
			optionNames := make([]string, 0, len(spec.Options))
			for optName := range spec.Options {
				optionNames = append(optionNames, optName)
			}
			fmt.Printf("%s\n", strings.Join(optionNames, ", "))
		}
	}
}

