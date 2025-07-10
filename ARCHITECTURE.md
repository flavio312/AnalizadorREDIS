# Arquitectura Técnica - Redis Analyzer

## Visión General del Sistema

Redis Analyzer es un sistema distribuido que combina análisis estático de comandos Redis con ejecución dinámica, proporcionando una interfaz web moderna para interactuar con bases de datos Redis.

### Componentes Principales

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Backend API   │    │   Redis DB      │
│   (React)       │◄──►│   (Go/Gin)      │◄──►│   (Redis)       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │
         │              ┌─────────────────┐
         │              │   Analyzers     │
         │              │   (Lex/Parse/   │
         └──────────────┤   Semantic)     │
                        └─────────────────┘
```

## Arquitectura del Backend

### 1. Analizador Léxico (Lexer)

**Ubicación**: `backend/lexer/`

**Responsabilidades**:
- Tokenización de comandos Redis
- Reconocimiento de patrones léxicos
- Manejo de cadenas, números y símbolos

**Componentes**:

```go
// Token representa un elemento léxico
type Token struct {
    Type     TokenType
    Literal  string
    Position int
}

// Lexer principal
type Lexer struct {
    input        string
    position     int
    readPosition int
    ch           byte
}
```

**Tipos de Tokens Soportados**:
- `IDENT`: Identificadores y comandos
- `STRING`: Cadenas con comillas
- `INT`: Números enteros
- `FLOAT`: Números decimales
- Símbolos especiales: `*`, `:`, `[`, `]`, etc.
- Palabras clave: `EX`, `PX`, `NX`, `XX`, `MATCH`, `COUNT`, etc.

**Algoritmo de Tokenización**:
1. Leer carácter actual
2. Determinar tipo de token basado en el carácter
3. Consumir caracteres hasta completar el token
4. Avanzar posición y repetir

### 2. Analizador Sintáctico (Parser)

**Ubicación**: `backend/parser/`

**Responsabilidades**:
- Construcción del AST (Abstract Syntax Tree)
- Validación sintáctica de comandos
- Manejo de gramática Redis

**Estructura del AST**:

```go
// Nodo base del AST
type Node interface {
    String() string
}

// Programa completo
type Program struct {
    Statements []Statement
}

// Comando Redis
type RedisCommand struct {
    Name       *Identifier
    Arguments  []Expression
    Options    map[string]Expression
}
```

**Gramática Soportada**:
```
Program     := Statement*
Statement   := RedisCommand
RedisCommand := IDENT Expression* Option*
Expression  := IDENT | STRING | INT | FLOAT | Pattern
Pattern     := IDENT (":" | "*")*
Option      := KEYWORD Expression?
```

**Algoritmo de Parsing**:
- Parser descendente recursivo
- Precedencia de operadores implícita
- Manejo de errores con recuperación

### 3. Analizador Semántico

**Ubicación**: `backend/semantic/`

**Responsabilidades**:
- Validación semántica de comandos
- Verificación de tipos y rangos
- Detección de conflictos entre opciones

**Especificaciones de Comandos**:

```go
type CommandSpec struct {
    Name        string
    Description string
    MinArgs     int
    MaxArgs     int
    Options     map[string]OptionSpec
}

type OptionSpec struct {
    RequiresValue bool
    ValueType     string
    Conflicts     []string
}
```

**Reglas de Validación**:
1. **Número de argumentos**: Verificar min/max args
2. **Tipos de datos**: Validar tipos de argumentos
3. **Opciones válidas**: Verificar opciones soportadas
4. **Conflictos**: Detectar opciones mutuamente excluyentes
5. **Rangos**: Validar rangos de valores (ej: TTL > 0)

### 4. Cliente Redis

**Ubicación**: `backend/redis/`

**Responsabilidades**:
- Conexión y comunicación con Redis
- Ejecución de comandos validados
- Manejo de errores y timeouts

**Arquitectura del Cliente**:

```go
type RedisClient struct {
    rdb      *redis.Client
    ctx      context.Context
    analyzer *semantic.Analyzer
}
```

**Pool de Conexiones**:
- Configuración automática de pool
- Reutilización de conexiones
- Manejo de reconexión automática

**Comandos Implementados**:
- **Strings**: GET, SET, DEL
- **Hashes**: HGET, HSET
- **Sorted Sets**: ZADD, ZRANGE
- **Utilidades**: SCAN, INFO

### 5. API REST

**Ubicación**: `backend/api/`

**Responsabilidades**:
- Exposición de endpoints HTTP
- Serialización JSON
- Manejo de CORS
- Middleware de logging

**Endpoints Principales**:

```
POST /api/v1/analyze     - Análisis de comandos
POST /api/v1/execute     - Ejecución de comandos
GET  /api/v1/database/info - Información de DB
GET  /api/v1/keys        - Listado de claves
GET  /api/v1/commands    - Especificaciones
GET  /api/v1/health      - Health check
```

**Middleware Stack**:
1. **CORS**: Permitir requests cross-origin
2. **Logging**: Log de requests/responses
3. **Recovery**: Manejo de panics
4. **Rate Limiting**: Control de tasa (futuro)

## Arquitectura del Frontend

### Estructura de Componentes

```
App.jsx
├── Header (Estado del servidor)
├── Tabs (Navegación)
│   ├── AnalyzeTab
│   ├── ExecuteTab
│   ├── KeysTab
│   ├── DatabaseTab
│   └── CommandsTab
└── CommandInput (Compartido)
```

### Estado de la Aplicación

**Estado Global**:
```javascript
const [activeTab, setActiveTab] = useState('analyze')
const [command, setCommand] = useState('')
const [serverStatus, setServerStatus] = useState(null)
```

**Estado por Componente**:
- `analyzeResult`: Resultado del análisis
- `executeResult`: Resultado de ejecución
- `dbInfo`: Información de base de datos
- `keys`: Lista de claves
- `commands`: Especificaciones de comandos

### Comunicación con API

**Cliente HTTP**:
```javascript
const API_BASE = 'http://localhost:8080/api/v1'

const analyzeCommand = async () => {
  const response = await fetch(`${API_BASE}/analyze`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ command })
  })
  return response.json()
}
```

**Manejo de Errores**:
- Try-catch en todas las llamadas
- Estados de loading
- Mensajes de error user-friendly

## Flujo de Datos

### Análisis de Comandos

```
1. Usuario ingresa comando
   ↓
2. Frontend envía POST /analyze
   ↓
3. Backend tokeniza (Lexer)
   ↓
4. Backend parsea (Parser)
   ↓
5. Backend valida (Semantic)
   ↓
6. Backend responde con resultado
   ↓
7. Frontend muestra análisis
```

### Ejecución de Comandos

```
1. Usuario ejecuta comando
   ↓
2. Frontend envía POST /execute
   ↓
3. Backend analiza comando
   ↓
4. Si válido: ejecuta en Redis
   ↓
5. Backend mide tiempo de ejecución
   ↓
6. Backend responde con resultado
   ↓
7. Frontend muestra resultado
```

## Patrones de Diseño Utilizados

### Backend (Go)

1. **Factory Pattern**: Creación de analizadores
2. **Strategy Pattern**: Diferentes tipos de validación
3. **Builder Pattern**: Construcción del AST
4. **Singleton Pattern**: Cliente Redis global
5. **Observer Pattern**: Logging y métricas

### Frontend (React)

1. **Component Pattern**: Componentes reutilizables
2. **Hook Pattern**: Custom hooks para API
3. **Provider Pattern**: Context para estado global
4. **Render Props**: Componentes de alto orden

## Optimizaciones de Rendimiento

### Backend

1. **Pool de Conexiones Redis**:
```go
rdb := redis.NewClient(&redis.Options{
    Addr:         "localhost:6379",
    PoolSize:     10,
    MinIdleConns: 5,
})
```

2. **Cache de Especificaciones**:
```go
var commandSpecs = map[string]*CommandSpec{
    "GET": {Name: "GET", MinArgs: 1, MaxArgs: 1},
    // ... cacheado en memoria
}
```

3. **Parsing Eficiente**:
- Parser descendente recursivo O(n)
- Reutilización de tokens
- Minimal backtracking

### Frontend

1. **Lazy Loading**: Componentes cargados bajo demanda
2. **Memoization**: React.memo para componentes pesados
3. **Debouncing**: Análisis automático con delay
4. **Virtual Scrolling**: Para listas grandes de claves

## Seguridad

### Validación de Entrada

1. **Sanitización**: Escape de caracteres especiales
2. **Validación de Longitud**: Límites en comandos
3. **Whitelist de Comandos**: Solo comandos permitidos
4. **Rate Limiting**: Control de frecuencia de requests

### Autenticación (Futuro)

```go
// JWT middleware
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if !validateJWT(token) {
            c.JSON(401, gin.H{"error": "Unauthorized"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

## Escalabilidad

### Horizontal Scaling

1. **Stateless API**: Sin estado en el servidor
2. **Load Balancer**: Nginx/HAProxy
3. **Redis Cluster**: Múltiples instancias Redis
4. **CDN**: Archivos estáticos del frontend

### Vertical Scaling

1. **Goroutines**: Concurrencia nativa de Go
2. **Connection Pooling**: Reutilización de conexiones
3. **Memory Management**: GC optimizado de Go
4. **CPU Optimization**: Algoritmos eficientes

## Monitoreo y Observabilidad

### Métricas

```go
type Metrics struct {
    RequestCount    int64
    ResponseTime    time.Duration
    ErrorRate       float64
    RedisConnections int
}
```

### Logging

```go
log.WithFields(log.Fields{
    "command": cmd,
    "duration": duration,
    "success": success,
}).Info("Command executed")
```

### Health Checks

```go
func healthCheck() gin.HandlerFunc {
    return func(c *gin.Context) {
        // Verificar Redis
        pong := rdb.Ping(ctx)
        
        c.JSON(200, gin.H{
            "status": "ok",
            "redis": pong.Val(),
            "timestamp": time.Now().Unix(),
        })
    }
}
```

## Testing Strategy

### Niveles de Testing

1. **Unit Tests**: Componentes individuales
2. **Integration Tests**: Comunicación entre componentes
3. **End-to-End Tests**: Flujo completo usuario
4. **Performance Tests**: Carga y estrés

### Coverage Goals

- **Backend**: >90% cobertura de código
- **Frontend**: >80% cobertura de componentes
- **API**: 100% cobertura de endpoints
- **Critical Paths**: 100% cobertura

## Deployment Architecture

### Development

```
Developer Machine
├── Go Backend (localhost:8080)
├── React Frontend (localhost:5173)
└── Redis (localhost:6379)
```

### Production

```
Load Balancer (nginx)
├── Frontend Servers (Static Files)
├── API Servers (Go Binaries)
│   ├── Instance 1
│   ├── Instance 2
│   └── Instance N
└── Redis Cluster
    ├── Master
    └── Replicas
```

## Futuras Mejoras

### Funcionalidades

1. **Más Comandos Redis**: Streams, Pub/Sub, Modules
2. **Query Builder**: Constructor visual de comandos
3. **Batch Operations**: Ejecución de múltiples comandos
4. **Export/Import**: Backup y restore de datos

### Arquitectura

1. **Microservicios**: Separar analizadores
2. **Event Sourcing**: Historial de comandos
3. **CQRS**: Separar lectura y escritura
4. **GraphQL**: API más flexible

### DevOps

1. **CI/CD Pipeline**: Automatización completa
2. **Container Orchestration**: Kubernetes
3. **Service Mesh**: Istio para comunicación
4. **Observability**: Prometheus + Grafana

---

Esta arquitectura proporciona una base sólida y escalable para el análisis y gestión de comandos Redis, con separación clara de responsabilidades y patrones de diseño probados.

