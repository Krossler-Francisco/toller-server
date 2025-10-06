# Toller Server — Chat y colaboración en tiempo real

Este repo es el backend de una aplicación de chat y colaboración. Está escrito en Go, usa PostgreSQL, y expone HTTP + WebSockets. La idea fue construir una arquitectura limpia y modular, fácil de probar y de hacer crecer.

## Mis objetivos

- Autenticación JWT simple y segura
- Módulos independientes (auth, users, teams, channels, friends, dms, chatWs)
- Servicios con lógica de negocio, handlers y repositorios para acceso a datos
- WebSockets para chat en tiempo real (broadcast por canal, historial, typing)
- Pruebas de integración vía `tests/api.http` y tests en `tests/*.go`

## Arquitectura y estructura del código

Separé el código por módulos y dentro de cada módulo seguí un patrón muy claro:

- Handler (HTTP/WebSocket): adapta la entrada/salida HTTP y WS
- Service: encapsula la lógica de negocio
- Repository: habla con la base de datos (SQL)
- Routes: registra las rutas del módulo en el router principal

El router global vive en `main.go` y centraliza las “salidas” HTTP. Además, ahí se inicializa el Hub de WebSockets, la conexión a la DB y el middleware de autenticación.

```
main.go
  ├─ modules/
  │   ├─ auth/
  │   ├─ users/
  │   ├─ teams/
  │   ├─ channels/
  │   ├─ friends/
  │   ├─ dms/
  │   └─ chat/ (handler WS, hub, client, repository)
  ├─ pkg/config/001_init.sql (schema SQL)
  ├─ tests/ (HTTP, tests de integración y tests end to end)
  └─ static/ (tentativa de documentación HTML y clientes de prueba simples)
```

Intente que este enfoque haga que cada módulo sea autónomo y reemplazable, y permite testear servicios y repositorios en aislamiento. Los handlers son “adapters” que traducen HTTP ↔ objetos de dominio.

## Autenticación (JWT)

El flujo de auth vive en `modules/auth` y funciona así:

- Registro: `POST /api/v1/auth/register` crea un usuario (username, email, password hasheado con bcrypt)
- Login: `POST /api/v1/auth/login` devuelve `{ token, user }` con la identidad del usuario
- Middleware: `JWTMiddleware` valida el token y coloca `user_id` en el contexto de la request para rutas protegidas

El token viaja en `Authorization: Bearer <token>` o en el query param `token` para WebSockets.

## WebSockets (Chat en tiempo real)

En `modules/chat` hay un `Hub` que orquesta salas por `channel_id`, `Client` que maneja la conexión WebSocket y los pumps de lectura/escritura, y un `Repository` para persistencia de mensajes.

- Conexión: `ws://localhost:8080/ws/channel/{channel_id}?token=<JWT>`
- Mensajes entrantes (desde el cliente):
  - `{ "type": "message", "content": "Hola a todos" }`
  - `{ "type": "typing" }`
- Persistencia: cada mensaje `type: "message"` se guarda en `messages (channel_id, user_id, content)` y se rebotea a los clientes del canal.
- Broadcast: el Hub entrega a todos los clientes conectados un `OutgoingMessage` con `{ type, content, user_id, channel_id, message_id, created_at }`.

Esto permite historial, notificaciones y extensiones como “typing” sin bloquear.

Entiendo que el envio de JWT en la URL no es lo ideal, pero es un compromiso común para WebSockets donde los headers son más difíciles de manejar desde clientes web. En node existen librerías que permiten enviar headers personalizados en la conexión WS, pero en Go no pude encontrar una solución simple y no quise adentrarme en ese tema.

## Módulos y endpoints principales

- Auth: `POST /auth/register`, `POST /auth/login`
- Users: `GET /users`, `GET /users/{id}`, `GET /users/search?query=...`
- Teams: `POST /teams`, `GET /teams`, `GET /teams/{id}`, `GET /teams/{id}/members`, `PUT /teams/{id}` (update), `POST /teams/{team_id}/members`, `DELETE /teams/{team_id}/members/{user_id}`
- Channels: `POST /teams/{team_id}/channels`, `GET /teams/{team_id}/channels`, `GET /channels/{channel_id}`, `PUT /channels/{channel_id}`, `DELETE /channels/{channel_id}`, `GET /channels/{channel_id}/members`, `POST /channels/{channel_id}/members`, `DELETE /channels/{channel_id}/members/{user_id}`
- Friends: `POST /friends/requests`, `PUT /friends/requests/{friendID}`, `GET /friends`, `GET /friends/requests/pending`
- DMs: `POST /dms`, `GET /dms`, `GET /dms/{channelID}/messages`, `POST /dms/{channelID}/read`
- WebSocket: `GET /ws/channel/{channel_id}` (upgrade WS)

Hay documentación viva en `tests/api.http` con ejemplos de request y respuestas esperadas.

## Base de datos (PostgreSQL)

El esquema está en `pkg/config/001_init.sql` e incluye:

- `users`: identidad y credenciales
- `teams` y `user_teams`: equipos y membresía (roles)
- `channels` y `channel_users`: canales (públicos por team o DMs) y membresía
- `messages`: mensajes persistidos (por canal y user)
- `friends`: solicitudes y relaciones de amistad (`pending`, `accepted`, `blocked`)
- `last_read`: para marcadores de lectura por canal

Todas las claves foráneas usan `ON DELETE CASCADE` para mantener integridad.

## Para probar

Variables necesarias:
- `DB_URL`: cadena de conexión a Postgres
- `JWT_SECRET`: secreto para firmar JWT
- `PORT` (opcional): puerto HTTP (por defecto 8080)

Pasos:

```bash
# 1) Instalar dependencias
go mod download

# 2) Crear base y correr el schema
psql "$DB_URL" -f pkg/config/001_init.sql

# 3) Levantar el servidor
go run .
```

El servidor expone CORS abierto (solo dev) y sirve archivos estáticos en `/static`.

## Pruebas y calidad

- `tests/api.http`: colección de requests para cubrir auth, teams, channels, users, friends, DMs y errores comunes.
- `tests/*.go`: pruebas de flujos e2e de usuarios, permisos, chat, etc.

## Decisiones de diseño y aprendizajes

- Modularidad: facilita mantener y evolucionar cada dominio sin romper el resto.
- Separación Handler/Service/Repository: claridad entre entrada/salida, negocio y persistencia.
- WebSockets con Hub: escalable y aislado por canal, fácil de instrumentar y testear.
- JWT en middleware: simple, efectivo, y reutilizable por todos los módulos.
- Documentación viva: `api.http` funciona como contrato de API y base para tests manuales.

## Próximos pasos (ideas)

- Roles por canal más avanzados (moderación)
- Paginación y búsqueda avanzada en mensajes y usuarios
- Métricas y trazas (Prometheus/OpenTelemetry)
- Integración de notificaciones push
- Endpoints de administración y reportes globales
- Agregar documentacion OpenAPI/Swagger, que se documente automáticamente
- Tests unitarios para servicios y repositorios

---
