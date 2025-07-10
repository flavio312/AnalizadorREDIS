package redis

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
	
	"github.com/redis/go-redis/v9"
	"redis-analyzer-api/parser"
	"redis-analyzer-api/semantic"
)

// Client representa el cliente Redis con capacidades de análisis
type Client struct {
	rdb      *redis.Client
	analyzer *semantic.Analyzer
	ctx      context.Context
}

// Config contiene la configuración para conectar a Redis
type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// ExecutionResult contiene el resultado de ejecutar un comando
type ExecutionResult struct {
	Success      bool
	Result       interface{}
	Error        string
	ExecutionTime time.Duration
	Command      string
	Validation   *semantic.ValidationResult
}

// DatabaseInfo contiene información sobre la base de datos Redis
type DatabaseInfo struct {
	Version      string
	Memory       map[string]string
	Clients      map[string]string
	Stats        map[string]string
	KeyCount     int64
	DatabaseSize int64
}

// NewClient crea un nuevo cliente Redis
func NewClient(config Config) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})
	
	return &Client{
		rdb:      rdb,
		analyzer: semantic.New(),
		ctx:      context.Background(),
	}
}

// Connect establece la conexión con Redis
func (c *Client) Connect() error {
	_, err := c.rdb.Ping(c.ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}
	return nil
}

// Close cierra la conexión con Redis
func (c *Client) Close() error {
	return c.rdb.Close()
}

// ExecuteCommand ejecuta un comando Redis después de analizarlo
func (c *Client) ExecuteCommand(commandStr string) ExecutionResult {
	start := time.Now()
	
	result := ExecutionResult{
		Command:       commandStr,
		ExecutionTime: 0,
		Success:       false,
	}
	
	// Parsear el comando
	cmd, parseErrors := parser.ParseCommand(commandStr)
	if len(parseErrors) > 0 {
		result.Error = fmt.Sprintf("Parse errors: %v", parseErrors)
		result.ExecutionTime = time.Since(start)
		return result
	}
	
	// Validar semánticamente
	validation := c.analyzer.ValidateCommand(cmd)
	result.Validation = &validation
	
	if !validation.Valid {
		result.Error = fmt.Sprintf("Semantic errors: %v", validation.Errors)
		result.ExecutionTime = time.Since(start)
		return result
	}
	
	// Ejecutar el comando
	res, err := c.executeRedisCommand(cmd)
	if err != nil {
		result.Error = err.Error()
	} else {
		result.Success = true
		result.Result = res
	}
	
	result.ExecutionTime = time.Since(start)
	return result
}

// executeRedisCommand ejecuta el comando Redis parseado
func (c *Client) executeRedisCommand(cmd *parser.RedisCommand) (interface{}, error) {
	commandName := strings.ToUpper(cmd.Command.Value)
	
	switch commandName {
	case "GET":
		if len(cmd.Arguments) != 1 {
			return nil, fmt.Errorf("GET requires exactly 1 argument")
		}
		key := c.extractStringValue(cmd.Arguments[0])
		return c.rdb.Get(c.ctx, key).Result()
		
	case "SET":
		if len(cmd.Arguments) < 2 {
			return nil, fmt.Errorf("SET requires at least 2 arguments")
		}
		key := c.extractStringValue(cmd.Arguments[0])
		value := c.extractStringValue(cmd.Arguments[1])
		
		// Manejar opciones
		var expiration time.Duration
		var nx, xx bool
		
		for i := 2; i < len(cmd.Arguments); i++ {
			arg := cmd.Arguments[i]
			if arg.Type() == "KeywordExpression" {
				option := strings.ToUpper(arg.String())
				switch option {
				case "EX":
					if i+1 < len(cmd.Arguments) {
						if seconds, err := c.extractIntValue(cmd.Arguments[i+1]); err == nil {
							expiration = time.Duration(seconds) * time.Second
							i++
						}
					}
				case "PX":
					if i+1 < len(cmd.Arguments) {
						if millis, err := c.extractIntValue(cmd.Arguments[i+1]); err == nil {
							expiration = time.Duration(millis) * time.Millisecond
							i++
						}
					}
				case "NX":
					nx = true
				case "XX":
					xx = true
				}
			}
		}
		
		// Ejecutar SET con opciones
		if nx {
			return c.rdb.SetNX(c.ctx, key, value, expiration).Result()
		} else if xx {
			return c.rdb.SetXX(c.ctx, key, value, expiration).Result()
		} else {
			return c.rdb.Set(c.ctx, key, value, expiration).Result()
		}
		
	case "DEL":
		if len(cmd.Arguments) == 0 {
			return nil, fmt.Errorf("DEL requires at least 1 argument")
		}
		keys := make([]string, len(cmd.Arguments))
		for i, arg := range cmd.Arguments {
			keys[i] = c.extractStringValue(arg)
		}
		return c.rdb.Del(c.ctx, keys...).Result()
		
	case "HGET":
		if len(cmd.Arguments) != 2 {
			return nil, fmt.Errorf("HGET requires exactly 2 arguments")
		}
		key := c.extractStringValue(cmd.Arguments[0])
		field := c.extractStringValue(cmd.Arguments[1])
		return c.rdb.HGet(c.ctx, key, field).Result()
		
	case "HSET":
		if len(cmd.Arguments) < 3 {
			return nil, fmt.Errorf("HSET requires at least 3 arguments")
		}
		key := c.extractStringValue(cmd.Arguments[0])
		
		// Construir pares field-value
		fields := make([]interface{}, 0, len(cmd.Arguments)-1)
		for i := 1; i < len(cmd.Arguments); i++ {
			fields = append(fields, c.extractStringValue(cmd.Arguments[i]))
		}
		
		return c.rdb.HSet(c.ctx, key, fields...).Result()
		
	case "ZADD":
		if len(cmd.Arguments) < 3 {
			return nil, fmt.Errorf("ZADD requires at least 3 arguments")
		}
		key := c.extractStringValue(cmd.Arguments[0])
		
		// Construir miembros con scores
		members := make([]redis.Z, 0)
		for i := 1; i < len(cmd.Arguments); i += 2 {
			if i+1 >= len(cmd.Arguments) {
				break
			}
			score, err := c.extractFloatValue(cmd.Arguments[i])
			if err != nil {
				return nil, fmt.Errorf("invalid score: %v", err)
			}
			member := c.extractStringValue(cmd.Arguments[i+1])
			members = append(members, redis.Z{Score: score, Member: member})
		}
		
		return c.rdb.ZAdd(c.ctx, key, members...).Result()
		
	case "ZRANGE":
		if len(cmd.Arguments) < 3 {
			return nil, fmt.Errorf("ZRANGE requires at least 3 arguments")
		}
		key := c.extractStringValue(cmd.Arguments[0])
		start, err := c.extractIntValue(cmd.Arguments[1])
		if err != nil {
			return nil, fmt.Errorf("invalid start index: %v", err)
		}
		stop, err := c.extractIntValue(cmd.Arguments[2])
		if err != nil {
			return nil, fmt.Errorf("invalid stop index: %v", err)
		}
		
		// Verificar si hay WITHSCORES
		withScores := false
		for i := 3; i < len(cmd.Arguments); i++ {
			if strings.ToUpper(cmd.Arguments[i].String()) == "WITHSCORES" {
				withScores = true
				break
			}
		}
		
		if withScores {
			return c.rdb.ZRangeWithScores(c.ctx, key, start, stop).Result()
		} else {
			return c.rdb.ZRange(c.ctx, key, start, stop).Result()
		}
		
	case "SCAN":
		if len(cmd.Arguments) < 1 {
			return nil, fmt.Errorf("SCAN requires at least 1 argument")
		}
		cursor, err := c.extractIntValue(cmd.Arguments[0])
		if err != nil {
			return nil, fmt.Errorf("invalid cursor: %v", err)
		}
		
		// Opciones por defecto
		match := ""
		count := int64(10)
		
		// Parsear opciones
		for i := 1; i < len(cmd.Arguments); i++ {
			if cmd.Arguments[i].Type() == "KeywordExpression" {
				option := strings.ToUpper(cmd.Arguments[i].String())
				switch option {
				case "MATCH":
					if i+1 < len(cmd.Arguments) {
						match = c.extractStringValue(cmd.Arguments[i+1])
						i++
					}
				case "COUNT":
					if i+1 < len(cmd.Arguments) {
						if c, err := c.extractIntValue(cmd.Arguments[i+1]); err == nil {
							count = c
							i++
						}
					}
				}
			}
		}
		
		keys, nextCursor, err := c.rdb.Scan(c.ctx, uint64(cursor), match, count).Result()
		if err != nil {
			return nil, err
		}
		
		// Devolver un mapa con las claves y el cursor
		return map[string]interface{}{
			"keys":   keys,
			"cursor": nextCursor,
		}, nil
		
	default:
		return nil, fmt.Errorf("command %s not implemented", commandName)
	}
}

// extractStringValue extrae el valor string de una expresión
func (c *Client) extractStringValue(expr parser.Expression) string {
	switch e := expr.(type) {
	case *parser.Identifier:
		return e.Value
	case *parser.StringLiteral:
		return e.Value
	case *parser.PatternExpression:
		return e.Value
	case *parser.IntegerLiteral:
		return strconv.FormatInt(e.Value, 10)
	case *parser.FloatLiteral:
		return strconv.FormatFloat(e.Value, 'f', -1, 64)
	default:
		return expr.String()
	}
}

// extractIntValue extrae el valor entero de una expresión
func (c *Client) extractIntValue(expr parser.Expression) (int64, error) {
	switch e := expr.(type) {
	case *parser.IntegerLiteral:
		return e.Value, nil
	case *parser.Identifier:
		return strconv.ParseInt(e.Value, 10, 64)
	case *parser.StringLiteral:
		return strconv.ParseInt(e.Value, 10, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to integer", expr)
	}
}

// extractFloatValue extrae el valor flotante de una expresión
func (c *Client) extractFloatValue(expr parser.Expression) (float64, error) {
	switch e := expr.(type) {
	case *parser.FloatLiteral:
		return e.Value, nil
	case *parser.IntegerLiteral:
		return float64(e.Value), nil
	case *parser.Identifier:
		return strconv.ParseFloat(e.Value, 64)
	case *parser.StringLiteral:
		return strconv.ParseFloat(e.Value, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float", expr)
	}
}

// GetDatabaseInfo obtiene información sobre la base de datos Redis
func (c *Client) GetDatabaseInfo() (DatabaseInfo, error) {
	info := DatabaseInfo{
		Memory:  make(map[string]string),
		Clients: make(map[string]string),
		Stats:   make(map[string]string),
	}
	
	// Obtener información del servidor
	infoResult, err := c.rdb.Info(c.ctx).Result()
	if err != nil {
		return info, err
	}
	
	// Parsear la información
	lines := strings.Split(infoResult, "\r\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				
				switch {
				case strings.HasPrefix(key, "redis_version"):
					info.Version = value
				case strings.HasPrefix(key, "used_memory"):
					info.Memory[key] = value
				case strings.HasPrefix(key, "connected_clients"):
					info.Clients[key] = value
				case strings.HasPrefix(key, "total_commands_processed"):
					info.Stats[key] = value
				}
			}
		}
	}
	
	// Obtener número de claves
	dbSize, err := c.rdb.DBSize(c.ctx).Result()
	if err == nil {
		info.KeyCount = dbSize
	}
	
	return info, nil
}

// ListKeys lista las claves que coinciden con un patrón
func (c *Client) ListKeys(pattern string, limit int) ([]string, error) {
	if pattern == "" {
		pattern = "*"
	}
	
	keys := make([]string, 0)
	iter := c.rdb.Scan(c.ctx, 0, pattern, int64(limit)).Iterator()
	
	for iter.Next(c.ctx) {
		keys = append(keys, iter.Val())
		if len(keys) >= limit {
			break
		}
	}
	
	return keys, iter.Err()
}

// GetKeyInfo obtiene información sobre una clave específica
func (c *Client) GetKeyInfo(key string) (map[string]interface{}, error) {
	info := make(map[string]interface{})
	
	// Tipo de la clave
	keyType, err := c.rdb.Type(c.ctx, key).Result()
	if err != nil {
		return nil, err
	}
	info["type"] = keyType
	
	// TTL
	ttl, err := c.rdb.TTL(c.ctx, key).Result()
	if err == nil {
		info["ttl"] = ttl.Seconds()
	}
	
	// Tamaño (aproximado)
	switch keyType {
	case "string":
		length, err := c.rdb.StrLen(c.ctx, key).Result()
		if err == nil {
			info["length"] = length
		}
	case "list":
		length, err := c.rdb.LLen(c.ctx, key).Result()
		if err == nil {
			info["length"] = length
		}
	case "set":
		length, err := c.rdb.SCard(c.ctx, key).Result()
		if err == nil {
			info["length"] = length
		}
	case "hash":
		length, err := c.rdb.HLen(c.ctx, key).Result()
		if err == nil {
			info["length"] = length
		}
	case "zset":
		length, err := c.rdb.ZCard(c.ctx, key).Result()
		if err == nil {
			info["length"] = length
		}
	}
	
	return info, nil
}

// FlushDatabase limpia la base de datos actual
func (c *Client) FlushDatabase() error {
	return c.rdb.FlushDB(c.ctx).Err()
}

