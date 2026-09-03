# Job Queue con Dead-Letter Queue

Sistema de colas de tareas en background, con reintentos y Dead-Letter Queue (DLQ), expuesto como API REST. Proyecto de portafolio construido con TDD estricto (rojo-verde-refactor).

## Objetivo

Implementar desde cero, entendiendo cada decisión de diseño, un sistema de:
- Cola de tareas en memoria (FIFO)
- Reintentos con backoff exponencial
- Dead-Letter Queue para tareas que agotaron reintentos
- Worker pool concurrente
- API REST para encolar/consultar/reintentar tareas
- Extensibilidad vía interfaces de Go (patrón Strategy + Registry)

## Stack

- Go (sin frameworks externos para la lógica core)
- TDD estricto: cada pieza de lógica arranca con un test que falla

## Cómo correr los tests

\`\`\`bash
go test ./...
go test -v ./...       # verbose
go test -race ./...    # detección de race conditions
\`\`\`

## Estado del proyecto

🚧 En desarrollo — Semana 1: núcleo de la cola.

## Decisiones de diseño

_(se va completando a medida que el proyecto avanza)_

## Roadmap

Ver [roadmap-job-queue-dlq.md](./roadmap-job-queue-dlq.md).
