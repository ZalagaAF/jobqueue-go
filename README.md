## Estado del proyecto

✅ Semana 1 completa — núcleo de la cola (sin concurrencia).
🚧 Semana 2 en progreso — worker pool + API REST.

## Decisiones de diseño

- **Job como interfaz de un solo método** (Strategy pattern, GoF): el motor
  de la cola nunca conoce EmailJob ni ImageResizeJob, solo el contrato
  `Execute(ctx) error`.
- **Payload dentro del struct del Job, no como parámetro genérico**: evita
  type assertions y mantiene type-safety (convención similar a
  `http.Handler.ServeHTTP`).
- **Status como string tipado, no iota**: prioriza legibilidad en logs/JSON
  sobre unos bytes de eficiencia, dado que el proyecto se expone via API REST.
- **Backoff exponencial con loop y techo (30s)**, en vez de fórmula cerrada
  con `math.Pow`: evita overflow silencioso en `time.Duration` (int64) ante
  inputs inesperados.
- **IDs de Task via crypto/rand (8 bytes, hex)**: sin dependencias externas,
  thread-safe de entrada (relevante para la concurrencia de Semana 2).
- **DeadLetterQueue usa map[string]*Task, Queue usa []*Task**: la estructura
  de datos se elige según el patrón de acceso predominante — DLQ necesita
  búsqueda/remoción por ID (O(1)), Queue necesita orden FIFO estricto.
- **ProcessTask no duerme el delay de backoff real**: separa "decidir qué
  hacer" de "ejecutar la espera", para no acoplar la lógica de negocio a
  time.Sleep antes de tener concurrencia real (Semana 2).