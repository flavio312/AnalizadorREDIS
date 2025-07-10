package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	
	"redis-analyzer-api/api"
	"redis-analyzer-api/redis"
)

func main() {
	// Configurar flags de línea de comandos
	var (
		port         = flag.String("port", "8080", "Puerto del servidor")
		redisHost    = flag.String("redis-host", "localhost", "Host de Redis")
		redisPort    = flag.Int("redis-port", 6379, "Puerto de Redis")
		redisDB      = flag.Int("redis-db", 0, "Base de datos de Redis")
		redisPass    = flag.String("redis-password", "", "Contraseña de Redis")
		help         = flag.Bool("help", false, "Mostrar ayuda")
	)
	
	flag.Parse()
	
	if *help {
		fmt.Println("Redis Analyzer API Server")
		fmt.Println("========================")
		fmt.Println()
		fmt.Println("Una API REST para analizar y ejecutar comandos Redis con validación")
		fmt.Println("léxica, sintáctica y semántica integrada.")
		fmt.Println()
		fmt.Println("Uso:")
		flag.PrintDefaults()
		fmt.Println()
		fmt.Println("Variables de entorno:")
		fmt.Println("  PORT              Puerto del servidor (default: 8080)")
		fmt.Println("  REDIS_HOST        Host de Redis (default: localhost)")
		fmt.Println("  REDIS_PORT        Puerto de Redis (default: 6379)")
		fmt.Println("  REDIS_DB          Base de datos de Redis (default: 0)")
		fmt.Println("  REDIS_PASSWORD    Contraseña de Redis")
		fmt.Println()
		fmt.Println("Endpoints principales:")
		fmt.Println("  POST /api/v1/analyze     - Analizar comando sin ejecutar")
		fmt.Println("  POST /api/v1/execute     - Ejecutar comando Redis")
		fmt.Println("  GET  /api/v1/database/info - Información de la base de datos")
		fmt.Println("  GET  /api/v1/keys        - Listar claves")
		fmt.Println("  GET  /api/v1/commands    - Especificaciones de comandos")
		fmt.Println("  GET  /api/v1/health      - Estado del servidor")
		return
	}
	
	// Leer configuración de variables de entorno si están disponibles
	if envPort := os.Getenv("PORT"); envPort != "" {
		*port = envPort
	}
	if envHost := os.Getenv("REDIS_HOST"); envHost != "" {
		*redisHost = envHost
	}
	if envPort := os.Getenv("REDIS_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			*redisPort = p
		}
	}
	if envDB := os.Getenv("REDIS_DB"); envDB != "" {
		if db, err := strconv.Atoi(envDB); err == nil {
			*redisDB = db
		}
	}
	if envPass := os.Getenv("REDIS_PASSWORD"); envPass != "" {
		*redisPass = envPass
	}
	
	// Configurar Redis
	redisConfig := redis.Config{
		Host:     *redisHost,
		Port:     *redisPort,
		Password: *redisPass,
		DB:       *redisDB,
	}
	
	// Crear servidor
	server := api.NewServer(redisConfig)
	
	// Mostrar información de inicio
	fmt.Println("🚀 Iniciando Redis Analyzer API Server")
	fmt.Printf("   Puerto: %s\n", *port)
	fmt.Printf("   Redis: %s:%d (DB: %d)\n", *redisHost, *redisPort, *redisDB)
	fmt.Println()
	fmt.Println("📚 Documentación de la API:")
	fmt.Printf("   Health Check: http://localhost:%s/api/v1/health\n", *port)
	fmt.Printf("   Comandos:     http://localhost:%s/api/v1/commands\n", *port)
	fmt.Printf("   Interfaz Web: http://localhost:%s/\n", *port)
	fmt.Println()
	
	// Iniciar servidor
	log.Printf("Servidor iniciado en puerto %s", *port)
	if err := server.Start(*port); err != nil {
		log.Fatalf("Error iniciando servidor: %v", err)
	}
}

