
# Contexto del Módulo de Canales (`channels`)

Este documento proporciona una visión detallada del funcionamiento del módulo de canales, su estructura y su interacción con la base de datos en el servidor de Toller-Chat. Está diseñado para ser utilizado como contexto para futuras implementaciones y desarrollos asistidos por IA.

## Visión General del Módulo de Canales

El módulo `channels` gestiona todo lo relacionado con los canales de comunicación dentro de un equipo (`team`). Permite a los usuarios crear, listar, ver, actualizar y eliminar canales, así como administrar los miembros de cada canal. Todas las rutas de este módulo requieren autenticación a través de un token JWT.

El flujo general es el siguiente:

1.  **Autenticación**: Un usuario autenticado (con un `user_id` válido en el contexto de la solicitud, validado por el `JWTMiddleware`) puede interactuar con los endpoints de este módulo.

2.  **Autorización a Nivel de Equipo**: Antes de realizar la mayoría de las acciones sobre canales (crear, listar), el sistema verifica que el usuario pertenezca al equipo (`team`) correspondiente.

3.  **Autorización a Nivel de Canal**: Para acciones específicas sobre un canal (ver detalles, añadir/eliminar miembros, actualizar, eliminar), el sistema verifica que el usuario sea miembro de dicho canal. Ciertas acciones, como añadir/eliminar miembros o eliminar el canal, requieren además que el usuario tenga un rol de `admin` en ese canal.

4.  **Operaciones CRUD y Gestión de Miembros**:
    *   **Crear un canal**: Un miembro de un equipo puede crear un nuevo canal dentro de ese equipo. El creador se convierte automáticamente en `admin` del canal.
    *   **Listar canales**: Un miembro de un equipo puede ver la lista de canales a los que pertenece dentro de ese equipo.
    *   **Ver detalles y miembros de un canal**: Un miembro de un canal puede ver sus detalles y la lista de otros miembros.
    *   **Añadir/Eliminar miembros**: Un `admin` del canal puede añadir o eliminar a otros usuarios del equipo al canal.
    *   **Actualizar un canal**: Un `admin` puede cambiar el nombre del canal.
    *   **Eliminar un canal**: Un `admin` puede eliminar el canal.

## Estructura del Proyecto y Componentes

### `main.go`

El punto de entrada de la aplicación, `main.go`, se encarga de:

*   Inicializar las dependencias del módulo `channels`: `ChannelRepository`, `ChannelService` y `ChannelHandler`.
*   Registrar las rutas del módulo `channels` en el enrutador principal (`mux.Router`).
*   Aplicar el `auth.JWTMiddleware` a todas las rutas de canales para asegurar que solo los usuarios autenticados puedan acceder a ellas.

### Módulo `channels`

El módulo está organizado en varios archivos con responsabilidades claras:

*   **`handler.go`**: Define los `http.Handler` que procesan las solicitudes HTTP para las operaciones de canales. Extrae los parámetros de la ruta (como `team_id`, `channel_id`), decodifica el cuerpo de la solicitud, obtiene el `user_id` del contexto y llama a los métodos del `ChannelService`.

*   **`service.go`**: Contiene la lógica de negocio. Orquesta las operaciones, realiza las validaciones de permisos (¿el usuario pertenece al equipo? ¿es admin del canal?) y se comunica con el `ChannelRepository` para acceder a la base de datos.

*   **`repository.go`**: Es la capa de abstracción de la base de datos. Contiene todas las consultas SQL para crear, leer, actualizar y eliminar canales y sus miembros. También define los modelos de datos como `Channel`, `ChannelMember` y `ChannelWithRole`.

*   **`routes.go`**: Define y registra todas las rutas de la API para el módulo `channels` en un sub-enrutador (`/api/v1`). Asocia cada ruta y método HTTP con su `handler` correspondiente.

## Esquema de la Base de Datos

Las siguientes tablas del archivo `pkg/config/001_init.sql` son fundamentales para el módulo `channels`.

### Tabla `channels`

```sql
CREATE TABLE channels (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    team_id INT REFERENCES teams(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

*   **`id`**: Identificador único del canal.
*   **`name`**: Nombre del canal.
*   **`team_id`**: Clave foránea que vincula el canal a un equipo (`teams`). Si el equipo se elimina, sus canales también se eliminan en cascada.
*   **`created_at`**: Fecha y hora de creación del canal.

### Tabla `channel_users`

Esta tabla de unión gestiona la relación muchos a muchos entre usuarios y canales, asignando un rol a cada usuario dentro de un canal.

```sql
CREATE TABLE channel_users (
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    channel_id INT REFERENCES channels(id) ON DELETE CASCADE,
    role VARCHAR(20) DEFAULT 'user', -- admin / user
    PRIMARY KEY (user_id, channel_id)
);
```

*   **`user_id`**: Clave foránea que referencia al usuario.
*   **`channel_id`**: Clave foránea que referencia al canal.
*   **`role`**: Rol del usuario en el canal. Puede ser `admin` o `user` (por defecto).
*   **Clave Primaria**: La combinación de `user_id` y `channel_id` es única, asegurando que un usuario solo pueda estar una vez en cada canal.

### Tablas Relacionadas

*   **`users`**: La tabla de usuarios, de donde se obtiene la información del usuario.
*   **`teams`**: La tabla de equipos, a la que cada canal debe pertenecer.
*   **`user_teams`**: Esencial para verificar si un usuario es miembro de un equipo antes de permitirle crear o unirse a canales dentro de ese equipo.

Este contexto proporciona una base sólida para que una IA pueda comprender y extender la funcionalidad del módulo de canales de manera segura y coherente.
