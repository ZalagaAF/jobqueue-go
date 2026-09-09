# Hoja de ruta — Job Queue con Dead-Letter Queue (Go + TDD)

**Objetivo del proyecto:** sistema de colas de tareas en background, con reintentos y Dead-Letter Queue, expuesto como API REST reutilizable por terceros. Proyecto de portafolio con foco principal en **aprender** — cada pieza se implementa entendiendo qué hace, por qué se diseña así, y de dónde viene la práctica.

**Restricciones:**
- 3-4 horas por día disponibles
- Máximo 3 semanas (objetivo interno: núcleo terminado en 2, semana 3 de colchón + pulido)
- TDD estricto: rojo → verde → refactor, en cada pieza de lógica
- Sin frontend de escritorio — solo API REST + dashboard mínimo de demostración
- Diseño extensible vía interfaces de Go (patrón Strategy + Registry), no herencia clásica

**Tareas demo elegidas:** envío de email (simulado, con probabilidad de falla configurable) y procesamiento de imágenes (real: resize/thumbnail).

---

## Semana 1 — Núcleo de la cola (sin concurrencia todavía)

**Meta de la semana:** que la lógica de negocio (cola, estados, reintentos, DLQ) esté completa y testeada, corriendo en memoria y de forma secuencial. Nada de HTTP, nada de goroutines todavía — eso es ruido para entender la lógica central.

- [x] Definir la interfaz `Job` (contrato mínimo: `Execute(ctx, payload) error`)
- [x] Definir el ciclo de vida de una tarea (pendiente → en proceso → completada / fallida → reintentando → muerta)
- [x] Cola en memoria (FIFO) con tests TDD: encolar, desencolar, orden, cola vacía
- [x] Lógica de reintentos con backoff exponencial — tests de casos límite (falla siempre, falla las primeras N veces, éxito al primer intento)
- [x] Dead-Letter Queue: mover tarea tras agotar reintentos, guardar payload + motivo del último error
- [x] Checkpoint: toda la lógica anterior corre con `go test ./...` en verde, sin ningún servidor levantado

**Horas estimadas:** ~24-32h (encaja en 6-8 días a 3-4h/día)

---

## Semana 2 — Concurrencia + API REST

**Meta de la semana:** que múltiples workers puedan tomar tareas de la cola sin pisarse, y que todo sea accesible por HTTP.

- [x] Worker pool con goroutines y channels
- [x] Tests de condiciones de carrera con `go test -race`
- [x] Definir cuántos workers corren en paralelo (configurable) y qué pasa si todos están ocupados
- [ ] Endpoints REST: `POST /jobs` (encolar), `GET /jobs/:id` (estado), `GET /dlq` (listar muertas), `POST /dlq/:id/retry` (reencolar manual)
- [ ] Tests de integración de los endpoints (no solo unitarios)
- [ ] Checkpoint: se puede levantar el servidor y encolar/consultar tareas con `curl`

**Horas estimadas:** ~24-30h

---

## Semana 3 — Tareas demo, dashboard y pulido (semana de colchón)

**Meta de la semana:** conectar las tareas reales/simuladas vía el patrón Registry, armar la demo visual, y dejar el repo presentable.

- [ ] Registry pattern: mapa `nombre_de_tarea → constructor de Job`
- [ ] `EmailJob` (simulado, con falla configurable) implementando la interfaz `Job`
- [ ] `ImageResizeJob` (real) implementando la interfaz `Job`
- [ ] Dashboard mínimo (HTML + JS liviano o templates de Go) que consuma la API: ver cola, ver DLQ, reintentar manualmente
- [ ] README con: diagrama de arquitectura, cómo correr el proyecto, decisiones de diseño explicadas (por qué interfaces, por qué DLQ, por qué backoff exponencial)
- [ ] Badge de cobertura de tests
- [ ] Colchón para lo que se haya atrasado de semanas 1-2 (probablemente concurrencia)

**Horas estimadas:** ~20-26h

---

## Principios a mantener durante todo el proyecto

- Cada feature nueva arranca con un test que falla (rojo) antes de escribir la implementación
- El motor de la cola nunca debe conocer detalles de `EmailJob` ni `ImageResizeJob` — solo la interfaz `Job`
- Preguntar siempre "¿de dónde viene esta práctica?" — buscar el patrón de diseño con nombre o la librería real que lo usa
- Commit por ciclo rojo-verde-refactor, no por sesión de trabajo completa
