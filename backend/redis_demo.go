package main

import (
	"fmt"
	"strings"
	"time"
	"redis-analyzer-api/redis"
)

func main() {
	fmt.Println("=== Cliente Redis con Analizador Integrado ===\n")
	
	// Configurar cliente Redis
	config := redis.Config{
		Host: "localhost",
		Port: 6379,
		DB:   0,
	}
	
	client := redis.NewClient(config)
	
	// Conectar a Redis
	fmt.Println("Conectando a Redis...")
	if err := client.Connect(); err != nil {
		fmt.Printf("❌ Error conectando a Redis: %v\n", err)
		fmt.Println("Asegúrate de que Redis esté ejecutándose en localhost:6379")
		return
	}
	defer client.Close()
	fmt.Println("✅ Conectado a Redis exitosamente\n")
	
	// Obtener información de la base de datos
	fmt.Println("=== Información de la Base de Datos ===")
	dbInfo, err := client.GetDatabaseInfo()
	if err != nil {
		fmt.Printf("Error obteniendo información: %v\n", err)
	} else {
		fmt.Printf("Versión de Redis: %s\n", dbInfo.Version)
		fmt.Printf("Número de claves: %d\n", dbInfo.KeyCount)
		if memory, ok := dbInfo.Memory["used_memory_human"]; ok {
			fmt.Printf("Memoria utilizada: %s\n", memory)
		}
	}
	fmt.Println()
	
	// Ejemplos de comandos Redis
	fmt.Println("=== Ejecutando Comandos Redis ===")
	
	commands := []string{
		// Comandos básicos de strings
		`SET usuario:123 "Juan Pérez"`,
		`SET usuario:123 "Juan Pérez" EX 3600`,
		`GET usuario:123`,
		
		// Comandos de hash
		`HSET perfil:123 nombre "Ana García" edad "28" ciudad "Madrid"`,
		`HGET perfil:123 nombre`,
		
		// Comandos de sorted sets
		`ZADD puntuaciones 100 jugador1 95 jugador2 87 jugador3`,
		`ZRANGE puntuaciones 0 -1 WITHSCORES`,
		
		// Comandos de utilidad
		`SCAN 0 MATCH usuario:* COUNT 10`,
		
		// Comandos con errores para demostrar validación
		`GET`,                           // Muy pocos argumentos
		`SET clave`,                     // Muy pocos argumentos
		`SET clave "valor" EX -60`,      // Valor inválido
		`SET clave "valor" NX XX`,       // Opciones conflictivas
		`COMANDO_INEXISTENTE clave`,     // Comando desconocido
	}
	
	for i, cmd := range commands {
		fmt.Printf("Comando %d: %s\n", i+1, cmd)
		fmt.Println(strings.Repeat("-", 50))
		
		// Ejecutar comando
		result := client.ExecuteCommand(cmd)
		
		// Mostrar resultado
		if result.Success {
			fmt.Printf("✅ Éxito (tiempo: %v)\n", result.ExecutionTime)
			if result.Result != nil {
				fmt.Printf("Resultado: %v\n", result.Result)
			}
		} else {
			fmt.Printf("❌ Error: %s\n", result.Error)
		}
		
		// Mostrar información de validación
		if result.Validation != nil {
			if result.Validation.Valid {
				fmt.Printf("✅ Validación semántica: Comando válido\n")
			} else {
				fmt.Printf("❌ Validación semántica: Comando inválido\n")
				for _, err := range result.Validation.Errors {
					fmt.Printf("  - [%s] %s\n", err.Type, err.Message)
				}
			}
			
			if len(result.Validation.Warnings) > 0 {
				fmt.Printf("⚠️  Advertencias:\n")
				for _, warning := range result.Validation.Warnings {
					fmt.Printf("  - %s\n", warning)
				}
			}
		}
		
		fmt.Println()
	}
	
	// Demostrar operaciones de gestión de claves
	fmt.Println("=== Gestión de Claves ===")
	
	// Listar claves
	fmt.Println("Listando claves que empiezan con 'usuario:'")
	keys, err := client.ListKeys("usuario:*", 10)
	if err != nil {
		fmt.Printf("Error listando claves: %v\n", err)
	} else {
		fmt.Printf("Claves encontradas: %v\n", keys)
	}
	
	// Información de una clave específica
	if len(keys) > 0 {
		key := keys[0]
		fmt.Printf("\nInformación de la clave '%s':\n", key)
		keyInfo, err := client.GetKeyInfo(key)
		if err != nil {
			fmt.Printf("Error obteniendo información: %v\n", err)
		} else {
			for k, v := range keyInfo {
				fmt.Printf("  %s: %v\n", k, v)
			}
		}
	}
	
	// Demostrar rendimiento
	fmt.Println("\n=== Prueba de Rendimiento ===")
	iterations := 100
	start := time.Now()
	
	successCount := 0
	for i := 0; i < iterations; i++ {
		result := client.ExecuteCommand(fmt.Sprintf(`SET perf:key%d "value%d"`, i, i))
		if result.Success {
			successCount++
		}
	}
	
	duration := time.Since(start)
	avgTime := duration / time.Duration(iterations)
	
	fmt.Printf("Ejecutadas %d operaciones SET en %v\n", iterations, duration)
	fmt.Printf("Tiempo promedio por operación: %v\n", avgTime)
	fmt.Printf("Operaciones exitosas: %d/%d\n", successCount, iterations)
	fmt.Printf("Operaciones por segundo: %.2f\n", float64(iterations)/duration.Seconds())
	
	// Limpiar claves de prueba
	fmt.Println("\n=== Limpieza ===")
	cleanupCommands := []string{
		`DEL usuario:123`,
		`DEL perfil:123`,
		`DEL puntuaciones`,
	}
	
	for _, cmd := range cleanupCommands {
		result := client.ExecuteCommand(cmd)
		if result.Success {
			fmt.Printf("✅ %s\n", cmd)
		} else {
			fmt.Printf("❌ %s: %s\n", cmd, result.Error)
		}
	}
	
	// Limpiar claves de rendimiento
	for i := 0; i < iterations; i++ {
		client.ExecuteCommand(fmt.Sprintf(`DEL perf:key%d`, i))
	}
	
	fmt.Println("\n🎉 Demostración completada exitosamente!")
}

