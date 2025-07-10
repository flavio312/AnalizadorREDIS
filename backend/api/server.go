package api

import (
	"net/http"
	"strconv"
	"time"
	
	"github.com/gin-gonic/gin"
	"redis-analyzer-api/redis"
	"redis-analyzer-api/parser"
	"redis-analyzer-api/semantic"
)

// Server representa el servidor API
type Server struct {
	router      *gin.Engine
	redisClient *redis.Client
	analyzer    *semantic.Analyzer
}

// AnalyzeRequest representa una solicitud de análisis
type AnalyzeRequest struct {
	Command string `json:"command" binding:"required"`
}

// AnalyzeResponse representa la respuesta del análisis
type AnalyzeResponse struct {
	Valid        bool                           `json:"valid"`
	ParsedAST    string                        `json:"parsed_ast"`
	Validation   *semantic.ValidationResult    `json:"validation"`
	CommandInfo  map[string]interface{}        `json:"command_info"`
	ParseErrors  []string                      `json:"parse_errors,omitempty"`
}

// ExecuteRequest representa una solicitud de ejecución
type ExecuteRequest struct {
	Command string `json:"command" binding:"required"`
}

// ExecuteResponse representa la respuesta de ejecución
type ExecuteResponse struct {
	Success       bool                        `json:"success"`
	Result        interface{}                 `json:"result,omitempty"`
	Error         string                      `json:"error,omitempty"`
	ExecutionTime string                      `json:"execution_time"`
	Validation    *semantic.ValidationResult `json:"validation"`
}

// DatabaseInfoResponse representa información de la base de datos
type DatabaseInfoResponse struct {
	Version      string            `json:"version"`
	KeyCount     int64             `json:"key_count"`
	Memory       map[string]string `json:"memory"`
	Clients      map[string]string `json:"clients"`
	Stats        map[string]string `json:"stats"`
}

// KeysResponse representa la respuesta de listado de claves
type KeysResponse struct {
	Keys    []string `json:"keys"`
	Count   int      `json:"count"`
	Pattern string   `json:"pattern"`
}

// KeyInfoResponse representa información de una clave
type KeyInfoResponse struct {
	Key    string                 `json:"key"`
	Exists bool                   `json:"exists"`
	Info   map[string]interface{} `json:"info,omitempty"`
	Error  string                 `json:"error,omitempty"`
}

// CommandSpecsResponse representa las especificaciones de comandos
type CommandSpecsResponse struct {
	Commands map[string]CommandSpecInfo `json:"commands"`
}

// CommandSpecInfo representa información de especificación de comando
type CommandSpecInfo struct {
	Name        string                        `json:"name"`
	MinArgs     int                           `json:"min_args"`
	MaxArgs     int                           `json:"max_args"`
	Description string                        `json:"description"`
	Options     map[string]OptionSpecInfo     `json:"options,omitempty"`
}

// OptionSpecInfo representa información de especificación de opción
type OptionSpecInfo struct {
	HasValue    bool     `json:"has_value"`
	ValueType   string   `json:"value_type,omitempty"`
	Description string   `json:"description"`
	Conflicts   []string `json:"conflicts,omitempty"`
}

// NewServer crea un nuevo servidor API
func NewServer(redisConfig redis.Config) *Server {
	// Configurar Gin en modo release para producción
	gin.SetMode(gin.ReleaseMode)
	
	router := gin.Default()
	
	// Configurar CORS
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})
	
	// Crear cliente Redis
	redisClient := redis.NewClient(redisConfig)
	analyzer := semantic.New()
	
	server := &Server{
		router:      router,
		redisClient: redisClient,
		analyzer:    analyzer,
	}
	
	// Configurar rutas
	server.setupRoutes()
	
	return server
}

// setupRoutes configura las rutas de la API
func (s *Server) setupRoutes() {
	api := s.router.Group("/api/v1")
	
	// Rutas de análisis
	api.POST("/analyze", s.analyzeCommand)
	api.GET("/commands", s.getCommandSpecs)
	
	// Rutas de ejecución
	api.POST("/execute", s.executeCommand)
	
	// Rutas de base de datos
	api.GET("/database/info", s.getDatabaseInfo)
	api.DELETE("/database/flush", s.flushDatabase)
	
	// Rutas de claves
	api.GET("/keys", s.listKeys)
	api.GET("/keys/:key", s.getKeyInfo)
	api.DELETE("/keys/:key", s.deleteKey)
	
	// Ruta de salud
	api.GET("/health", s.healthCheck)
	
	// Servir archivos estáticos (para el frontend)
	s.router.Static("/static", "./web/static")
	s.router.StaticFile("/", "./web/index.html")
}

// analyzeCommand analiza un comando Redis sin ejecutarlo
func (s *Server) analyzeCommand(c *gin.Context) {
	var req AnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Parsear el comando
	cmd, parseErrors := parser.ParseCommand(req.Command)
	
	response := AnalyzeResponse{
		ParseErrors: parseErrors,
	}
	
	if len(parseErrors) > 0 {
		response.Valid = false
		c.JSON(http.StatusOK, response)
		return
	}
	
	// Obtener AST como string
	response.ParsedAST = cmd.String()
	
	// Validar semánticamente
	validation := s.analyzer.ValidateCommand(cmd)
	response.Validation = &validation
	response.Valid = validation.Valid
	
	// Obtener información del comando
	response.CommandInfo = parser.GetCommandInfo(cmd)
	
	c.JSON(http.StatusOK, response)
}

// executeCommand ejecuta un comando Redis
func (s *Server) executeCommand(c *gin.Context) {
	var req ExecuteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Ejecutar comando
	result := s.redisClient.ExecuteCommand(req.Command)
	
	response := ExecuteResponse{
		Success:       result.Success,
		Result:        result.Result,
		Error:         result.Error,
		ExecutionTime: result.ExecutionTime.String(),
		Validation:    result.Validation,
	}
	
	c.JSON(http.StatusOK, response)
}

// getDatabaseInfo obtiene información de la base de datos
func (s *Server) getDatabaseInfo(c *gin.Context) {
	info, err := s.redisClient.GetDatabaseInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	response := DatabaseInfoResponse{
		Version:  info.Version,
		KeyCount: info.KeyCount,
		Memory:   info.Memory,
		Clients:  info.Clients,
		Stats:    info.Stats,
	}
	
	c.JSON(http.StatusOK, response)
}

// listKeys lista las claves que coinciden con un patrón
func (s *Server) listKeys(c *gin.Context) {
	pattern := c.DefaultQuery("pattern", "*")
	limitStr := c.DefaultQuery("limit", "100")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 100
	}
	
	keys, err := s.redisClient.ListKeys(pattern, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	response := KeysResponse{
		Keys:    keys,
		Count:   len(keys),
		Pattern: pattern,
	}
	
	c.JSON(http.StatusOK, response)
}

// getKeyInfo obtiene información de una clave específica
func (s *Server) getKeyInfo(c *gin.Context) {
	key := c.Param("key")
	
	info, err := s.redisClient.GetKeyInfo(key)
	
	response := KeyInfoResponse{
		Key:    key,
		Exists: err == nil,
	}
	
	if err != nil {
		response.Error = err.Error()
	} else {
		response.Info = info
	}
	
	c.JSON(http.StatusOK, response)
}

// deleteKey elimina una clave
func (s *Server) deleteKey(c *gin.Context) {
	key := c.Param("key")
	
	result := s.redisClient.ExecuteCommand("DEL " + key)
	
	if result.Success {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Key deleted successfully",
		})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   result.Error,
		})
	}
}

// flushDatabase limpia la base de datos
func (s *Server) flushDatabase(c *gin.Context) {
	err := s.redisClient.FlushDatabase()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Database flushed successfully",
	})
}

// getCommandSpecs obtiene las especificaciones de comandos
func (s *Server) getCommandSpecs(c *gin.Context) {
	specs := s.analyzer.GetCommandSpecs()
	
	commands := make(map[string]CommandSpecInfo)
	for name, spec := range specs {
		options := make(map[string]OptionSpecInfo)
		for optName, optSpec := range spec.Options {
			options[optName] = OptionSpecInfo{
				HasValue:    optSpec.HasValue,
				ValueType:   optSpec.ValueType,
				Description: optSpec.Description,
				Conflicts:   optSpec.Conflicts,
			}
		}
		
		commands[name] = CommandSpecInfo{
			Name:        spec.Name,
			MinArgs:     spec.MinArgs,
			MaxArgs:     spec.MaxArgs,
			Description: spec.Description,
			Options:     options,
		}
	}
	
	response := CommandSpecsResponse{
		Commands: commands,
	}
	
	c.JSON(http.StatusOK, response)
}

// healthCheck verifica el estado del servidor
func (s *Server) healthCheck(c *gin.Context) {
	// Verificar conexión a Redis
	err := s.redisClient.Connect()
	redisStatus := "ok"
	if err != nil {
		redisStatus = "error: " + err.Error()
	}
	
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
		"redis":     redisStatus,
		"version":   "1.0.0",
	})
}

// Start inicia el servidor
func (s *Server) Start(port string) error {
	// Conectar a Redis
	if err := s.redisClient.Connect(); err != nil {
		return err
	}
	
	// Iniciar servidor
	return s.router.Run("0.0.0.0:" + port)
}

// Stop detiene el servidor
func (s *Server) Stop() error {
	return s.redisClient.Close()
}

