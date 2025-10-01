
# Contexto del Módulo de Chat (`chat`)

Este documento detalla la arquitectura y el flujo de datos del sistema de chat en tiempo real de Toller-Chat, que se basa en WebSockets. Está diseñado para servir como contexto para futuras implementaciones y desarrollos asistidos por IA.

## Visión General del Módulo de Chat

El módulo `chat` permite la comunicación en tiempo real dentro de un canal (`channel`). Utiliza WebSockets para una comunicación bidireccional eficiente entre el cliente y el servidor. La funcionalidad principal incluye el envío y recepción de mensajes, la persistencia de los mismos en la base de datos y la notificación de eventos como "usuario escribiendo".

El flujo de comunicación es el siguiente:

1.  **Conexión WebSocket**: Un cliente autenticado establece una conexión WebSocket a la ruta `/ws/channel/{channel_id}`. La autenticación se realiza mediante un token JWT que puede ser enviado como un parámetro de consulta (`?token=...`) o en el encabezado `Authorization`.

2.  **Autenticación y Mejora de Conexión**: El `ChatHandler` intercepta la solicitud, valida el token JWT para obtener el `user_id` y "mejora" la conexión HTTP a una conexión WebSocket persistente.

3.  **Registro del Cliente**: Una vez establecida la conexión, se crea una instancia de `Client` que representa a ese usuario en ese canal. Este cliente se registra en el `Hub`, que lo asocia al `channel_id` correspondiente.

4.  **Carga del Historial**: Inmediatamente después del registro, el servidor carga los últimos 50 mensajes del canal desde la base de datos y se los envía al cliente recién conectado.

5.  **Comunicación en Tiempo Real**:
    *   **Mensajes Entrantes**: Cuando un cliente envía un mensaje (`IncomingMessage`), el método `readPump` del cliente lo recibe. El mensaje se guarda en la base de datos a través del `Repository`.
    *   **Difusión (Broadcast)**: Después de guardar el mensaje, el `Hub` lo difunde (`Broadcast`) a todos los demás clientes conectados en el mismo canal. El mensaje ahora es un `OutgoingMessage`, que incluye el `message_id` y `created_at` de la base de datos.
    *   **Mensajes Salientes**: Cada cliente tiene un `writePump` que escucha en un canal (`send`) y escribe los mensajes que recibe en su propia conexión WebSocket.

6.  **Desconexión**: Si un cliente se desconecta, el `readPump` termina, se llama a `hub.Unregister` para eliminar al cliente del `Hub`, y la conexión WebSocket se cierra.

## Estructura del Proyecto y Componentes

### `main.go`

El punto de entrada de la aplicación, `main.go`, se encarga de:

*   Inicializar el `Hub` de chat.
*   Crear una instancia del `ChatHandler`, inyectándole la conexión a la base de datos, el secreto del JWT y el `Hub`.
*   Registrar la ruta WebSocket `/ws/channel/{channel_id}` en el enrutador principal, asociándola al método `ServeWS` del `ChatHandler`.

### Módulo `chat`

Este módulo es el núcleo de la funcionalidad de chat en tiempo real:

*   **`handler.go`**: Define el `ChatHandler`, que gestiona las nuevas conexiones WebSocket. Es responsable de la autenticación del usuario a través del token JWT y de la creación de la estructura `Client` para cada conexión exitosa.

*   **`hub.go`**: Actúa como un concentrador central para todas las conexiones de chat. Mantiene un mapa de `rooms` (canales) y los clientes dentro de cada uno. Sus responsabilidades son:
    *   `Register`: Registrar un nuevo cliente en un canal.
    *   `Unregister`: Eliminar un cliente de un canal.
    *   `Broadcast`: Enviar un mensaje a todos los clientes de un canal, excepto al remitente.

*   **`client.go`**: Representa a un cliente (usuario) conectado a un canal a través de una única conexión WebSocket. Cada `Client` tiene dos bucles principales (goroutines):
    *   `readPump`: Lee los mensajes JSON que llegan desde el cliente (navegador).
    *   `writePump`: Escribe los mensajes JSON que se envían desde el servidor hacia el cliente.
    Define también las estructuras `IncomingMessage` y `OutgoingMessage`.

*   **`repository.go`**: Es la capa de acceso a datos para el chat. Se encarga de:
    *   `SaveMessage`: Guardar un nuevo mensaje en la tabla `messages`.
    *   `LoadLastMessages`: Cargar los mensajes más recientes de un canal para enviarlos como historial.

## Esquema de la Base de Datos

La tabla principal para este módulo es `messages`.

### Tabla `messages`

```sql
CREATE TABLE messages (
    id SERIAL PRIMARY KEY,
    channel_id INT REFERENCES channels(id) ON DELETE CASCADE,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

*   **`id`**: Identificador único del mensaje.
*   **`channel_id`**: Clave foránea que vincula el mensaje al canal donde fue enviado. Si el canal se elimina, los mensajes se eliminan en cascada.
*   **`user_id`**: Clave foránea que vincula el mensaje al usuario que lo envió.
*   **`content`**: El contenido de texto del mensaje.
*   **`created_at`**: La fecha y hora en que se guardó el mensaje.

Este contexto proporciona una visión completa del sistema de chat, crucial para que una IA pueda depurar, modificar o extender su funcionalidad.
