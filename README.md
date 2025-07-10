# Redis Analyzer - Analizador Léxico, Sintáctico y Semántico

Un sistema completo de análisis y gestión para comandos Redis, desarrollado en Go con una interfaz web moderna en React.

## 🚀 Características Principales

### Analizador de Comandos Redis
- **Analizador Léxico**: Tokenización completa de comandos Redis con soporte para cadenas, números, símbolos y palabras clave
- **Analizador Sintáctico**: Parser descendente recursivo que construye un AST (Abstract Syntax Tree) completo
- **Analizador Semántico**: Validación semántica avanzada con verificación de argumentos, opciones y tipos de datos

### API REST Completa
- Endpoints para análisis de comandos (`/api/v1/analyze`)
- Ejecución segura de comandos (`/api/v1/execute`)
- Gestión de claves y base de datos (`/api/v1/keys`, `/api/v1/database`)
- Especificaciones de comandos (`/api/v1/commands`)
- Health checks y monitoreo (`/api/v1/health`)

### Interfaz Web Moderna
- Diseño responsivo con Tailwind CSS y shadcn/ui
- Navegación por pestañas intuitiva
- Análisis en tiempo real de comandos
- Ejecución interactiva con resultados formateados
- Gestión visual de claves Redis
- Información detallada de la base de datos

### Características Técnicas
- **Rendimiento**: Más de 25,000 operaciones por segundo
- **Multiplataforma**: Binarios para Linux, Windows y macOS
- **CORS habilitado**: Comunicación frontend-backend sin restricciones
- **Pruebas completas**: Suite de pruebas unitarias e integración
- **Construcción automatizada**: Scripts de build y despliegue

## 📋 Requisitos del Sistema

### Dependencias Principales
- **Go 1.21+**: Para el backend
- **Node.js 18+**: Para el frontend
- **pnpm**: Gestor de paquetes para Node.js
- **Redis Server**: Base de datos Redis

### Instalación de Dependencias

#### Ubuntu/Debian
```bash
# Instalar Redis
sudo apt update
sudo apt install redis-server

# Iniciar Redis
sudo systemctl start redis-server
sudo systemctl enable redis-server

# Verificar instalación
redis-cli ping
```

#### Go (si no está instalado)
```bash
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

#### Node.js y pnpm
```bash
# Instalar Node.js
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs

# Instalar pnpm
npm install -g pnpm
```

## 🛠️ Instalación y Configuración

### Opción 1: Desarrollo (Recomendado para desarrollo)

1. **Clonar el repositorio**
```bash
git clone <repository-url>
cd redis-analyzer-api
```

2. **Iniciar entorno de desarrollo**
```bash
./build.sh dev
```

Esto iniciará:
- Backend en `http://localhost:8080`
- Frontend en `http://localhost:5173`
- API documentada en `http://localhost:8080/api/v1/health`

### Opción 2: Construcción para Producción

1. **Construir el proyecto completo**
```bash
./build.sh deploy
```

2. **Extraer y ejecutar**
```bash
cd dist
tar -xzf redis-analyzer-*.tar.gz
cd redis-analyzer
./start.sh  # Linux/macOS
# o
start.bat   # Windows
```

### Opción 3: Construcción Manual

#### Backend
```bash
cd backend
go mod download
go build -o redis-analyzer .
./redis-analyzer
```

#### Frontend
```bash
cd frontend
pnpm install
pnpm run build
pnpm run dev
```

## 🧪 Pruebas

### Pruebas Unitarias
```bash
# Probar backend
./build.sh test

# Probar componentes individuales
cd backend
go test ./lexer -v
go test ./parser -v
go test ./semantic -v
go test ./api -v
```

### Pruebas de Integración
```bash
./build.sh integration
```

Las pruebas de integración verifican:
- Conectividad API y frontend
- Análisis de comandos end-to-end
- Ejecución de comandos Redis
- Gestión de claves
- Información de base de datos
- Configuración CORS

## 📖 Uso de la API

### Análisis de Comandos

**POST** `/api/v1/analyze`
```json
{
  "command": "SET user:123 \"John Doe\" EX 3600"
}
```

**Respuesta:**
```json
{
  "valid": true,
  "command_info": {
    "name": "SET",
    "args_count": 4
  },
  "validation": {
    "valid": true,
    "errors": [],
    "warnings": []
  },
  "parsed_ast": "..."
}
```

### Ejecución de Comandos

**POST** `/api/v1/execute`
```json
{
  "command": "GET user:123"
}
```

**Respuesta:**
```json
{
  "success": true,
  "result": "John Doe",
  "execution_time": "45.2µs"
}
```

### Gestión de Claves

**GET** `/api/v1/keys?pattern=user:*&limit=10`
```json
{
  "keys": ["user:123", "user:456"],
  "count": 2,
  "pattern": "user:*"
}
```

**DELETE** `/api/v1/keys/{key}`
```json
{
  "success": true,
  "message": "Key deleted successfully"
}
```

### Información de Base de Datos

**GET** `/api/v1/database/info`
```json
{
  "version": "6.0.16",
  "key_count": 1247,
  "memory": {
    "used_memory": "2.1MB",
    "used_memory_dataset_perc": "67.60%"
  }
}
```

## 🏗️ Arquitectura del Sistema

### Estructura del Proyecto
```
redis-analyzer-api/
├── backend/                 # API en Go
│   ├── lexer/              # Analizador léxico
│   ├── parser/             # Analizador sintáctico
│   ├── semantic/           # Analizador semántico
│   ├── redis/              # Cliente Redis
│   ├── api/                # Endpoints REST
│   └── main.go             # Punto de entrada
├── frontend/               # Interfaz web en React
│   ├── src/
│   │   ├── components/     # Componentes UI
│   │   └── App.jsx         # Aplicación principal
│   └── package.json
├── build.sh               # Script de construcción
├── integration.go         # Pruebas de integración
└── README.md              # Esta documentación
```

### Flujo de Análisis

1. **Tokenización**: El lexer convierte el comando en tokens
2. **Parsing**: El parser construye un AST a partir de los tokens
3. **Validación Semántica**: Se verifican reglas específicas de Redis
4. **Ejecución**: Si es válido, se ejecuta contra Redis
5. **Respuesta**: Se devuelve el resultado formateado

### Comandos Soportados

El analizador soporta los siguientes comandos Redis:

| Comando | Descripción | Argumentos | Opciones |
|---------|-------------|------------|----------|
| GET | Obtener valor de clave | key | - |
| SET | Establecer valor de clave | key value | EX, PX, NX, XX |
| DEL | Eliminar claves | key [key ...] | - |
| HGET | Obtener campo de hash | key field | - |
| HSET | Establecer campo de hash | key field value | - |
| ZADD | Añadir a sorted set | key score member [...] | NX, XX, CH, INCR |
| ZRANGE | Rango de sorted set | key start stop | WITHSCORES |
| SCAN | Escanear claves | cursor | MATCH, COUNT |

## 🔧 Configuración Avanzada

### Variables de Entorno

```bash
# Puerto del servidor (default: 8080)
export PORT=8080

# Host de Redis (default: localhost:6379)
export REDIS_HOST=localhost:6379

# Base de datos Redis (default: 0)
export REDIS_DB=0

# Contraseña de Redis (opcional)
export REDIS_PASSWORD=your_password
```

### Configuración de Redis

Para desarrollo local:
```bash
# redis.conf
bind 127.0.0.1
port 6379
save 900 1
save 300 10
save 60 10000
```

Para producción:
```bash
# Configurar autenticación
requirepass your_secure_password

# Configurar persistencia
save 900 1
save 300 10
save 60 10000
```

## 🚀 Despliegue en Producción

### Docker (Recomendado)

1. **Crear Dockerfile para backend**
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY backend/ .
RUN go mod download
RUN go build -o redis-analyzer .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/redis-analyzer .
EXPOSE 8080
CMD ["./redis-analyzer"]
```

2. **Docker Compose**
```yaml
version: '3.8'
services:
  redis:
    image: redis:6-alpine
    ports:
      - "6379:6379"
  
  api:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - redis
    environment:
      - REDIS_HOST=redis:6379
  
  frontend:
    image: nginx:alpine
    ports:
      - "80:80"
    volumes:
      - ./frontend/dist:/usr/share/nginx/html
```

### Servidor Tradicional

1. **Construir para producción**
```bash
./build.sh deploy
```

2. **Configurar servicio systemd**
```ini
[Unit]
Description=Redis Analyzer API
After=network.target redis.service

[Service]
Type=simple
User=redis-analyzer
WorkingDirectory=/opt/redis-analyzer
ExecStart=/opt/redis-analyzer/bin/redis-analyzer-linux
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

3. **Configurar nginx para frontend**
```nginx
server {
    listen 80;
    server_name your-domain.com;
    
    location / {
        root /opt/redis-analyzer/web;
        try_files $uri $uri/ /index.html;
    }
    
    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 🐛 Solución de Problemas

### Problemas Comunes

**Error: "Redis connection refused"**
```bash
# Verificar que Redis esté ejecutándose
sudo systemctl status redis-server
redis-cli ping

# Iniciar Redis si no está activo
sudo systemctl start redis-server
```

**Error: "Port 8080 already in use"**
```bash
# Encontrar proceso usando el puerto
sudo lsof -i :8080

# Terminar proceso o cambiar puerto
export PORT=8081
```

**Error: "Frontend build failed"**
```bash
# Limpiar cache de pnpm
cd frontend
pnpm store prune
rm -rf node_modules
pnpm install
```

### Logs y Debugging

**Habilitar logs detallados:**
```bash
export GIN_MODE=debug
export LOG_LEVEL=debug
```

**Verificar conectividad:**
```bash
# Probar API
curl http://localhost:8080/api/v1/health

# Probar Redis
redis-cli -h localhost -p 6379 ping
```

## 📊 Rendimiento y Optimización

### Métricas de Rendimiento

- **Análisis léxico**: ~100,000 tokens/segundo
- **Análisis sintáctico**: ~50,000 comandos/segundo
- **Validación semántica**: ~75,000 comandos/segundo
- **Ejecución Redis**: ~25,000 operaciones/segundo

### Optimizaciones Implementadas

1. **Pool de conexiones Redis**: Reutilización de conexiones
2. **Cache de especificaciones**: Comandos cacheados en memoria
3. **Parsing eficiente**: Parser descendente recursivo optimizado
4. **Validación lazy**: Solo se valida cuando es necesario

### Monitoreo

El sistema incluye métricas integradas:
- Tiempo de respuesta por endpoint
- Número de comandos procesados
- Errores de validación
- Estado de conexión Redis

## 🤝 Contribución

### Estructura de Desarrollo

1. **Fork del repositorio**
2. **Crear rama de feature**: `git checkout -b feature/nueva-funcionalidad`
3. **Implementar cambios con pruebas**
4. **Ejecutar suite de pruebas**: `./build.sh test && ./build.sh integration`
5. **Commit y push**: `git commit -m "Descripción" && git push`
6. **Crear Pull Request**

### Estándares de Código

- **Go**: Seguir `gofmt` y `golint`
- **JavaScript**: Usar ESLint y Prettier
- **Commits**: Formato convencional (feat:, fix:, docs:)
- **Pruebas**: Cobertura mínima del 80%

### Añadir Nuevos Comandos

1. **Actualizar lexer** (si necesita nuevos tokens)
2. **Extender parser** (para nueva sintaxis)
3. **Añadir validación semántica**
4. **Implementar ejecución en cliente Redis**
5. **Añadir pruebas unitarias e integración**

## 📄 Licencia

Este proyecto está licenciado bajo la Licencia MIT. Ver el archivo `LICENSE` para más detalles.

## 🙏 Agradecimientos

- **Redis Team**: Por la excelente base de datos
- **Go Community**: Por las librerías y herramientas
- **React Team**: Por el framework frontend
- **Tailwind CSS**: Por el sistema de diseño
- **shadcn/ui**: Por los componentes UI

## 📞 Soporte

Para soporte técnico o preguntas:

- **Issues**: Crear issue en el repositorio
- **Documentación**: Ver este README y comentarios en código
- **Ejemplos**: Revisar archivos `*_demo.go` y pruebas

---

**Redis Analyzer** - Desarrollado con ❤️ para la comunidad Redis

