package main

import (
	"fmt"
	"strings"
	"redis-analyzer-api/parser"
)

func main() {
	fmt.Println("=== Analizador Sintáctico de Redis ===\n")
	
	// Ejemplos de comandos Redis
	commands := []string{
		"GET mykey",
		`SET user:123 "John Doe" EX 3600`,
		"ZADD leaderboard 100 player1 95 player2",
		"ZRANGE scores 0 -1 WITHSCORES",
		"SCAN 0 MATCH user:* COUNT 10",
		"HSET user:456 name Alice age 30",
	}
	
	for i, cmd := range commands {
		fmt.Printf("Comando %d: %s\n", i+1, cmd)
		fmt.Println(strings.Repeat("-", 50))
		
		// Parsear el comando
		parsedCmd, errors := parser.ParseCommand(cmd)
		
		if len(errors) > 0 {
			fmt.Printf("Errores: %v\n", errors)
			continue
		}
		
		// Mostrar información del comando
		fmt.Printf("Comando: %s\n", parsedCmd.Command.Value)
		fmt.Printf("Número de argumentos: %d\n", len(parsedCmd.Arguments))
		
		// Mostrar argumentos
		fmt.Println("Argumentos:")
		for j, arg := range parsedCmd.Arguments {
			fmt.Printf("  [%d] %s: %s\n", j, arg.Type(), arg.String())
		}
		
		// Obtener información adicional
		info := parser.GetCommandInfo(parsedCmd)
		fmt.Printf("Tiene clave: %v\n", info["has_key"])
		fmt.Printf("Tiene valor: %v\n", info["has_value"])
		if options, ok := info["options"].([]string); ok && len(options) > 0 {
			fmt.Printf("Opciones: %v\n", options)
		}
		
		// Mostrar AST como string
		fmt.Printf("AST: %s\n", parsedCmd.String())
		
		fmt.Println()
	}
	
	// Ejemplo con múltiples comandos
	fmt.Println("=== Múltiples Comandos ===")
	multiCmd := `SET key1 value1
GET key1
DEL key1`
	
	fmt.Printf("Comandos:\n%s\n", multiCmd)
	fmt.Println(strings.Repeat("-", 50))
	
	program, errors := parser.ParseCommands(multiCmd)
	if len(errors) > 0 {
		fmt.Printf("Errores: %v\n", errors)
	} else {
		fmt.Printf("Número de comandos: %d\n", len(program.Statements))
		for i, stmt := range program.Statements {
			if cmd, ok := stmt.(*parser.RedisCommand); ok {
				fmt.Printf("  [%d] %s\n", i+1, cmd.String())
			}
		}
		fmt.Printf("\nPrograma completo:\n%s\n", program.String())
	}
}

