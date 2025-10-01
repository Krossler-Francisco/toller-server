
# Contexto del Módulo de Equipos (`teams`)

Este documento describe la funcionalidad, arquitectura y estructura de la base de datos para el módulo de equipos (`teams`) del servidor Toller-Chat. Su propósito es servir como un contexto claro para futuras implementaciones y desarrollos asistidos por IA.

## Visión General del Módulo de Equipos

El módulo `teams` es fundamental para la organización de usuarios y canales. Permite a los usuarios crear equipos, unirse a ellos, gestionarlos y ver quién más es miembro. La pertenencia a un equipo es un requisito previo para poder acceder a sus canales. Todas las rutas de este módulo están protegidas y requieren autenticación mediante JWT.

El flujo general de operaciones es el siguiente:

1.  **Autenticación**: Todas las solicitudes a los endpoints de `teams` deben estar autenticadas. El `JWTMiddleware` valida el token y extrae el `user_id` del solicitante.

2.  **Creación de un Equipo**: Un usuario autenticado puede crear un nuevo equipo. Al hacerlo, se convierte automáticamente en el `admin` de ese equipo.

3.  **Gestión de Miembros**:
    *   Un `admin` del equipo puede añadir a otros usuarios registrados en la plataforma al equipo.
    *   Un `admin` puede eliminar a otros miembros del equipo (excepto a sí mismo).
    *   Cualquier miembro puede optar por abandonar un equipo.

4.  **Autorización Basada en Roles**: Las acciones críticas como añadir/eliminar miembros o actualizar la información del equipo están restringidas a los usuarios con el rol de `admin` dentro de ese equipo. El `TeamService` se encarga de verificar estos permisos antes de ejecutar la acción.

5.  **Acceso a la Información**:
    *   Un usuario puede listar todos los equipos a los que pertenece.
    *   Un miembro de un equipo puede ver los detalles de ese equipo y la lista de todos sus miembros.

## Estructura del Proyecto y Componentes

### `main.go`

El archivo `main.go` es el punto de entrada que configura la aplicación:

*   Inicializa las dependencias del módulo `teams`: `TeamRepository`, `TeamService` y `TeamHandler`.
*   Registra las rutas del módulo `teams` en el enrutador principal (`mux.Router`).
*   Aplica el `auth.JWTMiddleware` a todas las rutas de `teams` para garantizar que solo los usuarios autenticados puedan acceder.

### Módulo `teams`

El módulo está estructurado de la siguiente manera:

*   **`handler.go`**: Contiene el `TeamHandler`, que maneja las solicitudes HTTP. Es responsable de decodificar los cuerpos de las solicitudes, extraer parámetros de la URL (como el `team_id`), obtener el `user_id` del contexto de la solicitud y llamar a los métodos apropiados en el `TeamService`.

*   **`service.go`**: Alberga la lógica de negocio. El `TeamService` orquesta las operaciones, como la creación de equipos y la gestión de miembros. Realiza todas las validaciones de permisos, asegurando que solo los administradores puedan realizar acciones administrativas.

*   **`repository.go`**: La capa de acceso a datos. El `TeamRepository` contiene todas las consultas SQL necesarias para interactuar con las tablas `teams` y `user_teams`. Abstrae la lógica de la base de datos del resto de la aplicación.

*   **`models.go`**: Define las estructuras de datos (`structs`) utilizadas en todo el módulo, como `Team`, `UserTeam`, `TeamWithRole` (que incluye el rol del usuario en el equipo) y `TeamMember` (que combina información del usuario y su rol en el equipo).

*   **`routes.go` (implícito en `handler.go`)**: La función `RegisterRoutes` en `handler.go` se encarga de definir todas las rutas de la API para el módulo `teams`, asociando cada endpoint (por ejemplo, `POST /teams`, `GET /teams/{id}/members`) a su función de `handler` correspondiente.

## Esquema de la Base de Datos

Las tablas clave para este módulo, definidas en `pkg/config/001_init.sql`, son `teams` y `user_teams`.

### Tabla `teams`

Almacena la información básica de cada equipo.

```sql
CREATE TABLE teams (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

*   **`id`**: Identificador único del equipo.
*   **`name`**: Nombre del equipo.
*   **`description`**: Descripción opcional del equipo.
*   **`created_at`**: Fecha y hora de creación.

### Tabla `user_teams`

Esta tabla de unión gestiona la relación muchos a muchos entre usuarios y equipos, definiendo el rol de cada usuario dentro de un equipo.

```sql
CREATE TABLE user_teams (
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    team_id INT REFERENCES teams(id) ON DELETE CASCADE,
    role VARCHAR(20) DEFAULT 'member', -- admin / member
    PRIMARY KEY (user_id, team_id)
);
```

*   **`user_id`**: Clave foránea que referencia al usuario.
*   **`team_id`**: Clave foránea que referencia al equipo.
*   **`role`**: El rol del usuario en el equipo, que puede ser `admin` o `member` (por defecto).
*   **Clave Primaria**: La combinación de `user_id` y `team_id` es única, lo que garantiza que un usuario solo pueda tener un rol en un equipo.

Este contexto proporciona una base completa para que una IA entienda cómo funciona la gestión de equipos y pueda realizar modificaciones o añadir nuevas funcionalidades de forma coherente.
