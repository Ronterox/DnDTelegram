# Proceso de Migración y Decisiones de Diseño

## Decisiones de Diseño

### 1. Transición a Microservicios
Originalmente, el sistema era monolítico o dependía de procesos locales. Se decidió usar **Docker Compose** para:
- Aislar dependencias (Bun, Go, Redis, RabbitMQ).
- Garantizar que el entorno de desarrollo sea idéntico al de producción.
- Facilitar el escalado de componentes específicos (p.ej. levantar más instancias de la D&D API para procesar la cola).

### 2. Implementación de RabbitMQ
Se eligió RabbitMQ sobre HTTP directo para la comunicación DM por:
- **Resiliencia**: Si el motor de IA está saturado, los mensajes permanecen en la cola en lugar de dar timeout HTTP.
- **Priorización**: Permite que las acciones críticas de los jugadores se procesen antes que el chat ambiental.
- **Observabilidad**: La interfaz de RabbitMQ permite ver cuellos de botella en tiempo real.

### 3. Persistencia Híbrida
- **Redis**: Para el estado rápido y volátil de la partida activa.
- **SQLite (users.db)**: Para datos de autenticación y persistencia a largo plazo dentro de contenedores con volúmenes dedicados.

## Proceso de Migración
1.  **Contenedorización**: Creación de Dockerfiles específicos optimizados para Alpine Linux.
2.  **Externalización de Configuración**: Refactorización de código en Go y TS para usar variables de entorno en lugar de `localhost` hardcoded.
3.  **Puente RPC**: Implementación de `QueueManager` en Go y el consumidor en TS para migrar el flujo crítico de chat a mensajería asíncrona.
4.  **Volúmenes Compartidos**: Implementación de un volumen compartido (`shared_data`) para que el bot en Go pueda exportar estados JSON que el frontend React consume instantáneamente.
5.  **Resolución de Conflictos**: Mapeo de puertos alternativos (`6380`, `3002`, `3003`) para permitir la ejecución simultánea de servicios locales y contenedores.
