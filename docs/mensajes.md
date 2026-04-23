# Esquema de Mensajes y Colas

## Patrón RPC (Request/Reply)
El sistema utiliza RabbitMQ para realizar llamadas síncronas simuladas entre el Bot (Cliente) y la API D&D (Servidor).

### Colas Utilizadas
-   **`rpc_queue`**: Cola principal donde el Bot envía solicitudes. Soporta prioridades (0-10).
-   **`amq.gen-*` (Direct)**: Colas de respuesta exclusivas y temporales creadas por el Bot para recibir resultados.
-   **`dlx_queue`**: Cola de "Dead Letter" vinculada a `dlx_exchange` para capturar mensajes expirados o fallidos.

## Estructura de Mensajes (JSON)

### Solicitud (RPCRequest)
```json
{
  "method": "chat | init",
  "sessionId": "string",
  "payload": {
    "message": "string",
    "format": "object (opcional)"
  }
}
```

### Respuesta (RPCResponse)
```json
{
  "success": true,
  "sessionId": "string",
  "response": {
    "narrative": "string"
  },
  "type": "structured | fallback",
  "error": "string (opcional)"
}
```

## Políticas de Mensajería
1.  **Prioridad**:
    *   `8`: Acciones críticas (tiradas de dados, combate).
    *   `5`: Conversación normal y comandos de información.
2.  **TTL (Time-To-Live)**:
    *   Los mensajes de chat expiran en **120 segundos** para evitar respuestas de IA obsoletas tras una caída del servicio.
3.  **DLX (Dead Letter Exchange)**:
    *   Cualquier mensaje que no sea procesado dentro del TTL o que sea rechazado por el consumidor se redirige a la `dlx_queue` para auditoría.
