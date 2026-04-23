# Documentación de la API (Gateway)

Aunque el sistema prioriza RabbitMQ, los microservicios mantienen interfaces HTTP para compatibilidad y fallback.

## 1. D&D API (Puerto 3002)
Utilizada para interactuar con el Dungeon Master de IA.

### `POST /api/init`
Inicializa una sesión de OpenCode.
- **Body**: `{ "sessionId": "opcional" }`
- **Response**: `{ "success": true, "sessionId": "id" }`

### `POST /api/chat`
Envía un mensaje narrativo al DM.
- **Body**: `{ "message": "string", "sessionId": "string" }`
- **Response**: `{ "response": { "narrative": "..." }, "sessionId": "..." }`

## 2. Game API (Puerto 3003)
Utilizada para la persistencia del estado del juego en Redis.

### `POST /api/games`
Guarda el estado completo de una partida.
- **Body**: Objeto `Game` serializado.

### `GET /api/games/:sessionId`
Recupera el estado de una partida por su ID de chat de Telegram.

### `GET /api/games?field=X&value=Y`
Buscador de partidas indexadas (p.ej. por nombre de jugador).

## 3. RabbitMQ Management (Puerto 15672)
Interfaz visual para administración.
- **User/Pass**: `guest` / `guest`

## 4. Web Viewer (Puerto 5173)
Interfaz de usuario para jugadores.
- **URL**: `http://localhost:5173/?id={chatId}`
