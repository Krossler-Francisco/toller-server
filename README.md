# Toller Server

API REST de autenticación construida con Go, PostgreSQL y JWT.

## Requisitos

- Go 1.21 o superior
- PostgreSQL 12 o superior

## Instalación

1. Clona el repositorio:
```bash
git clone https://github.com/TU_USUARIO/toller-server.git
cd toller-server
```

2. Instala las dependencias:
```bash
go mod download
```

3. Configura las variables de entorno:

Crea un archivo `.env` en la raíz del proyecto:

```env
JWT_SECRET=tu_secreto_jwt_super_seguro
DB_URL=postgres://postgres:1234@localhost:5432/toller?sslmode=disable
PORT=8080
```

4. Crea la base de datos y las tablas:

```sql
CREATE DATABASE toller;

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## 🏃 Ejecución

### Desarrollo
```bash
go run cmd/server/main.go
```

### Producción
```bash
go build -o bin/server cmd/server/main.go
./bin/server
```

El servidor se ejecutará en `http://localhost:8080`

## 📡 Endpoints

### Registro de usuario
```http
POST /register
Content-Type: application/json

{
  "username": "usuario",
  "email": "usuario@example.com",
  "password": "contraseña123"
}
```

**Respuesta exitosa (201):**
```json
{
  "message": "Usuario registrado exitosamente",
  "user": {
    "id": 1,
    "username": "usuario",
    "email": "usuario@example.com"
  }
}
```

### Login
```http
POST /login
Content-Type: application/json

{
  "email": "usuario@example.com",
  "password": "contraseña123"
}
```

**Respuesta exitosa (200):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "message": "Login exitoso"
}
```

## 🛠️ Tecnologías

- [Go](https://golang.org/) - Lenguaje de programación
- [Gorilla Mux](https://github.com/gorilla/mux) - Router HTTP
- [PostgreSQL](https://www.postgresql.org/) - Base de datos
- [JWT](https://github.com/golang-jwt/jwt) - Autenticación
- [bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt) - Hash de contraseñas

## 📁 Estructura del proyecto

```
toller-server/
├── cmd/
│   └── server/
│       └── main.go          # Punto de entrada
├── internal/
│   └── auth/
│       ├── handler.go       # Controladores HTTP
│       ├── service.go       # Lógica de negocio
│       ├── repository.go    # Acceso a datos
│       └── models.go        # Modelos de datos
├── .env                     # Variables de entorno (no subir a git)
├── go.mod
├── go.sum
└── README.md
```

## 🌐 Deploy en Render

1. Crea una base de datos PostgreSQL en Render
2. Crea un Web Service conectado a tu repo de GitHub
3. Configura las variables de entorno en Render:
   - `JWT_SECRET`
   - `DB_URL` (obtenida de la base de datos de Render)
   - `PORT` (opcional, Render lo asigna automáticamente)
4. Build Command: `go build -o bin/server cmd/server/main.go`
5. Start Command: `./bin/server`

## 📝 Notas

- Las contraseñas se hashean con bcrypt antes de almacenarse
- Los tokens JWT expiran después de 72 horas
- La base de datos gratuita de Render se elimina después de 90 días

## 📄 Licencia

MIT