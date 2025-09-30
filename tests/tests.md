# 🧪 Guía de Testing - Toller API

## 📋 Requisitos

- Tener el servidor corriendo: `go run cmd/server/main.go`
- Extensión REST Client para VS Code (recomendado) o usar curl

## 🚀 Cómo usar los tests

### Opción 1: VS Code REST Client (Recomendado)

1. Instala la extensión **REST Client** en VS Code
2. Abre el archivo `tests/api.http`
3. Verás un botón **"Send Request"** sobre cada request
4. Click para ejecutar cada test

### Opción 2: Postman

Importa las requests manualmente o usa curl.

### Opción 3: curl

Puedes copiar los requests del archivo `.http` y adaptarlos a curl.

## 📝 Flujo de Testing Recomendado

### 1️⃣ Setup Inicial

```http
# Registrar usuarios de prueba
POST /register  (Fran)
POST /register  (Maria)
POST /register  (Juan)
```

### 2️⃣ Autenticación

```http
# Login con Fran
POST /login

# ⚠️ IMPORTANTE: Copia el token del response y pégalo en la variable @token
```

### 3️⃣ Gestión de Teams

```http
# Crear teams
POST /teams  (Team Alpha)
POST /teams  (Team Beta)

# Ver mis teams
GET /teams

# Ver detalles de un team
GET /teams/1
```

### 4️⃣ Gestión de Miembros

```http
# Ver miembros actuales
GET /teams/1/members

# Agregar miembros
POST /teams/1/members  (Maria - user_id: 2)
POST /teams/1/members  (Juan - user_id: 3)

# Ver miembros actualizados
GET /teams/1/members

# Remover miembro
DELETE /teams/1/members/3
```

## 🎯 Tests por Categoría

### ✅ Tests de Autenticación

- ✓ Registro exitoso
- ✓ Registro con email duplicado (debe fallar)
- ✓ Login exitoso
- ✓ Login con credenciales incorrectas (debe fallar)

### ✅ Tests de Teams

- ✓ Crear team exitosamente
- ✓ Crear team sin nombre (debe fallar)
- ✓ Crear team sin autenticación (debe fallar)
- ✓ Obtener lista de teams del usuario
- ✓ Obtener team específico
- ✓ Actualizar team (solo admin)
- ✓ Obtener team que no existe (debe fallar)

### ✅ Tests de Miembros

- ✓ Ver miembros de un team
- ✓ Agregar miembro al team (solo admin)
- ✓ Agregar miembro duplicado (debe fallar)
- ✓ Remover miembro del team (solo admin)
- ✓ Remover al admin (debe fallar)
- ✓ Salir del team (leave)

### ✅ Tests de Seguridad

- ✓ Acceso sin token (debe retornar 401)
- ✓ Acceso con token inválido (debe retornar 401)
- ✓ Acceso con token mal formado (debe retornar 401)
- ✓ Acceso con token expirado (debe retornar 401)

### ✅ Tests de Edge Cases

- ✓ ID no numérico en URL
- ✓ ID negativo
- ✓ JSON malformado
- ✓ Campos faltantes en request
- ✓ Team/User que no existe

## 🔧 Configuración de Variables

En `api.http`, actualiza estas variables según tu setup:

```http
@baseUrl = http://localhost:8080  # Cambia si usas otro puerto
@token = tu_token_jwt_aqui        # Actualiza después de hacer login
```

## 📊 Respuestas Esperadas

### ✅ Registro Exitoso (201)
```json
{
  "message": "Usuario registrado exitosamente",
  "user": {
    "id": 1,
    "username": "fran",
    "email": "fran@example.com"
  }
}
```

### ✅ Login Exitoso (200)
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "message": "Login exitoso"
}
```

### ✅ Team Creado (201)
```json
{
  "message": "Equipo creado exitosamente",
  "team": {
    "id": 1,
    "name": "Team Alpha",
    "description": "Equipo de desarrollo",
    "created_at": "2025-09-30T..."
  }
}
```

### ✅ Lista de Teams (200)
```json
[
  {
    "id": 1,
    "name": "Team Alpha",
    "description": "Equipo de desarrollo",
    "created_at": "2025-09-30T...",
    "user_role": "admin"
  }
]
```

### ❌ Error de Autenticación (401)
```json
{
  "error": "unauthorized",
  "message": "Token no proporcionado"
}
```

### ❌ Error de Permisos (403)
```json
{
  "error": "access_denied",
  "message": "Solo los admins pueden agregar miembros"
}
```

## 🐛 Troubleshooting

### El token no funciona
- Verifica que copiaste el token completo
- Asegúrate de usar el formato: `Bearer tu_token`
- El token expira en 72 horas, haz login nuevamente

### Error de conexión
- Verifica que el servidor esté corriendo
- Verifica el puerto correcto (8080 por defecto)
- Revisa los logs del servidor

### Error 500
- Revisa los logs del servidor
- Verifica que la base de datos esté corriendo
- Verifica que las tablas existan

## 📚 Próximos Módulos

Próximamente agregaremos tests para:
- 📢 Channels
- 💬 Messages
- 👫 Friends (DMs)

## 💡 Tips

1. **Orden importa**: Ejecuta primero los tests de registro y login
2. **Guarda el token**: Actualiza la variable `@token` después del login
3. **IDs dinámicos**: Los IDs en los ejemplos (1, 2, 3) pueden variar
4. **Limpia la DB**: Para empezar de cero, trunca las tablas
5. **Logs útiles**: Revisa los logs del servidor para debugging

## 🔄 Reset de Base de Datos

Si necesitas empezar de cero:

```sql
TRUNCATE users, teams, user_teams CASCADE;
```

⚠️ Esto borrará todos los datos de prueba.