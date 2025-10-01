
# Contexto del Módulo de Autenticación (`auth`)

Este documento proporciona una visión detallada del flujo de autenticación y la estructura de la base de datos para el servidor de Toller-Chat. Está diseñado para ser utilizado como contexto para futuras implementaciones y desarrollos asistidos por IA.

## Visión General del Flujo de Autenticación

El flujo de autenticación se encarga de registrar nuevos usuarios y de permitir que los usuarios existentes inicien sesión. Una vez que un usuario inicia sesión, se genera un token JWT que se utiliza para autenticar las solicitudes a las rutas protegidas de la API.

El flujo es el siguiente:

1.  **Registro (`/register`)**:
    *   Un usuario envía su `username`, `email` y `password` a través de una solicitud POST.
    *   El `AuthHandler` recibe la solicitud y llama al `AuthService`.
    *   El `AuthService` hashea la contraseña y utiliza el `UserRepository` para crear un nuevo usuario en la base de datos.
    *   Se devuelve el usuario recién creado (sin la contraseña).

2.  **Inicio de Sesión (`/login`)**:
    *   Un usuario envía su `email` y `password` a través de una solicitud POST.
    *   El `AuthHandler` recibe la solicitud y llama al `AuthService`.
    *   El `AuthService` utiliza el `UserRepository` para buscar al usuario por su email.
    *   Si el usuario existe, se compara la contraseña enviada con el hash almacenado en la base de datos.
    *   Si la contraseña es correcta, se genera un token JWT que contiene el `user_id` y una fecha de expiración.
    *   Se devuelve el token JWT al cliente.

3.  **Autenticación de Rutas Protegidas**:
    *   Para acceder a las rutas protegidas, el cliente debe incluir el token JWT en el encabezado `Authorization` de la solicitud, con el formato `Bearer <token>`.
    *   El `JWTMiddleware` intercepta la solicitud, verifica la validez del token y extrae el `user_id`.
    *   Si el token es válido, el `user_id` se añade al contexto de la solicitud para que pueda ser utilizado por los `handlers` de las rutas protegidas.
    *   Si el token no es válido, se devuelve un error de "no autorizado".

## Estructura del Proyecto y Componentes

### `main.go`

El archivo `main.go` es el punto de entrada de la aplicación. Realiza las siguientes tareas:

*   Carga las variables de entorno desde un archivo `.env`.
*   Establece la conexión con la base de datos PostgreSQL.
*   Inicializa las dependencias de cada módulo:
    *   `UserRepository`
    *   `AuthService`
    *   `AuthHandler`
*   Configura el enrutador (`mux.Router`) y registra las rutas para cada módulo.
*   Aplica el `JWTMiddleware` a las rutas que requieren autenticación.
*   Inicia el servidor HTTP.

### Módulo `auth`

El módulo `auth` está organizado en varios archivos, cada uno con una responsabilidad específica:

*   **`handler.go`**: Define los `http.Handler` que manejan las solicitudes HTTP para `/register` y `/login`. Se encarga de decodificar las solicitudes, llamar al servicio correspondiente y enviar la respuesta.

*   **`service.go`**: Contiene la lógica de negocio para el registro y el inicio de sesión. Se comunica con el `UserRepository` para interactuar con la base de datos y es responsable de hashear contraseñas y generar tokens JWT.

*   **`repository.go`**: Proporciona una capa de abstracción sobre la base de datos. Contiene las consultas SQL para crear y buscar usuarios.

*   **`middleware.go`**: Implementa el `JWTMiddleware` que protege las rutas de la API. Verifica la firma y la validez de los tokens JWT.

*   **`model.go`**: Define la estructura de datos `User`, que representa a un usuario en el sistema.

## Esquema de la Base de Datos

El archivo `pkg/config/001_init.sql` define el esquema de la base de datos. A continuación se muestran las tablas relevantes para el flujo de autenticación y la gestión de usuarios.

### Tabla `users`

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

*   **`id`**: Identificador único para cada usuario.
*   **`username`**: Nombre de usuario, debe ser único.
*   **`email`**: Correo electrónico del usuario, debe ser único.
*   **`password`**: Contraseña hasheada del usuario.
*   **`created_at`**: Fecha y hora de creación del usuario.

### Otras Tablas Relevantes

Aunque no forman parte directa del módulo `auth`, estas tablas están relacionadas con los usuarios:

*   **`teams`**: Almacena los equipos creados.
*   **`user_teams`**: Tabla de unión para la relación muchos a muchos entre usuarios y equipos.
*   **`channels`**: Almacena los canales de comunicación, que pertenecen a un equipo.
*   **`channel_users`**: Tabla de unión para la relación muchos a muchos entre usuarios y canales.
*   **`messages`**: Almacena los mensajes enviados en los canales.
*   **`friends`**: Almacena las relaciones de amistad entre usuarios.

Este contexto debería ser suficiente para que una IA pueda entender el funcionamiento del sistema de autenticación y realizar nuevas implementaciones o modificaciones de forma coherente con la arquitectura existente.
