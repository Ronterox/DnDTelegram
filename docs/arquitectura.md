# Arquitectura del Sistema

## Vista General
El sistema utiliza una arquitectura de microservicios desacoplados mediante un broker de mensajería (RabbitMQ) y contenedores Docker.

## Diagrama de Componentes
```
┌──────────────────┐      ┌──────────────────┐      ┌──────────────────┐
│  Telegram Bot    │◄────►│   RabbitMQ       │◄────►│    D&D API       │
│      (Go)        │      │ (Broker/Colas)   │      │  (Bun/Express)   │
└────────┬─────────┘      └──────────────────┘      └────────┬─────────┘
         │                                                   │
         │                ┌──────────────────┐      ┌────────▼─────────┐
         ├───────────────►│    Game API      │      │   OpenCode AI    │
         │                │ (Redis Storage)  │      │   (DM Engine)    │
         │                └──────────────────┘      └──────────────────┘
         │
         │                ┌──────────────────┐
         └───────────────►│  SixSevenStory   │
                          │  (Web Viewer)    │
                          └──────────────────┘
```

## Descripción de Servicios
1.  **Bot (Go)**: Orquestador principal. Gestiona comandos de Telegram, lógica de turnos y exportación de estado para la web.
2.  **RabbitMQ**: Middleware de comunicación. Implementa patrones RPC, priorización de mensajes y manejo de errores mediante DLX.
3.  **D&D API (Node.js)**: Servicio de lógica narrativa. Procesa peticiones de chat e inicialización de sesiones.
4.  **OpenCode AI**: Motor de IA local que genera las respuestas del Dungeon Master.
5.  **Game API (Node.js)**: Microservicio CRUD que persiste el estado de las partidas en Redis.
6.  **SixSevenStory (React)**: Dashboard visual que consume archivos JSON generados por el Bot.

## Red y Comunicación
-   **Interna (Docker)**: Los servicios se comunican usando nombres de servicio (p.ej., `http://rabbitmq:5672`).
-   **Externa (Host)**:
    -   Web: Port 5173
    -   API D&D: Port 3002
    -   API Game: Port 3003
    -   RabbitMQ Admin: Port 15672
