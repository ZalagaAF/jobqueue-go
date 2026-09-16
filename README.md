cat > README.md << 'EOF'
# jobqueue-go

Cola de trabajos en memoria con Dead-Letter Queue (DLQ), escrita en Go
puro sin dependencias externas como proyecto de portafolio backend,
construido con TDD estricto (rojo-verde-refactor) de punta a punta.
El fin de este proyecto era el aprendizaje e implementación de técnicas de diseño estudiadas. 

## Qué hace

- Recibe trabajos por HTTP (`POST /jobs`), los procesa de forma
  concurrente con un pool de workers, y reintenta automáticamente los
  que fallan con backoff exponencial.
- Si un trabajo agota sus reintentos, cae a una Dead-Letter Queue
  consultable y reintentable manualmente.
- Dos tipos de trabajo implementados: envío de email (simulado) y
  redimensión de imágenes (real, con `image/jpeg` y `image/png` de la
  stdlib). Agregar un tipo nuevo no requiere tocar el motor de la cola
  — ver "Arquitectura".
- Dashboard HTML mínimo para ver el procesamiento concurrente en vivo
  y reintentar tareas de la DLQ con un clic.

## Requisitos

- Go 1.22 o superior (usa `net/http.ServeMux` con path variables,
  disponible desde 1.22)
- Sin dependencias externas — todo el proyecto usa únicamente la
  librería estándar de Go

## Uso rápido

```bash
go run ./cmd/server
```

Flags disponibles:

| Flag                | Default | Descripción                                          |
|---------------------|---------|-------------------------------------------------------|
| `-addr`              | `:8080` | Dirección donde escucha el servidor                    |
| `-workers`           | `4`     | Cantidad de workers concurrentes                        |
| `-channel-capacity`  | `16`    | Capacidad del channel bounded entre dispatcher y workers |

Con el servidor corriendo, abrí `http://localhost:8080/dashboard` en
el navegador, o probá la API directo con curl:

```bash
curl -X POST http://localhost:8080/jobs \
  -H "Content-Type: application/json" \
  -d '{"type":"email","payload":{"to":"a@b.com","subject":"hola"}}'
```

## Endpoints

| Método | Path                | Descripción                                          |
|--------|----------------------|-------------------------------------------------------|
| POST   | `/jobs`              | Crea un trabajo. Body: `{"type": "...", "payload": {...}}` |
| GET    | `/jobs/{id}`         | Consulta el estado de un trabajo por ID                |
| GET    | `/dlq`               | Lista los trabajos que agotaron sus reintentos          |
| POST   | `/dlq/{id}/retry`    | Reintenta manualmente un trabajo de la DLQ              |
| GET    | `/dashboard`         | Dashboard HTML de demostración                          |

### Tipos de trabajo (`type`)

**`email`** — simulado, no envía nada de verdad.
```json
{"type": "email", "payload": {"to": "x@y.com", "subject": "..."}}
```

**`image_resize`** — real. Redimensiona con nearest neighbor y guarda
el resultado en el mismo directorio que el original, con sufijo
`_resized` antes de la extensión (`foto.jpg` → `foto_resized.jpg`,
sobrescribiendo si ya existía).
```json
{"type": "image_resize", "payload": {"source_path": "/ruta/foto.jpg", "width": 200, "height": 200}}
```

## Arquitectura

- **Dominio** (`internal/queue`): `Job` es una interfaz (Strategy) —
  cualquier struct con `Execute(ctx) error` sirve. `JobRegistry`
  mapea el string `type` del JSON a la factory que construye el `Job`
  concreto (Registry). Agregar un tipo nuevo es escribir el struct +
  una llamada a `Register` en `main.go`; nada del motor de la cola,
  ni los handlers HTTP, necesitan cambiar.
- **Motor de concurrencia**: `Submit()` encola en un inbox ilimitado
  (nunca bloquea al llamador HTTP), un dispatcher lo mueve a un
  channel bounded, y N workers lo consumen. El backpressure real vive
  en ese channel bounded, no en el inbox.
- **Reintentos**: backoff exponencial con techo de 30s, hasta
  `MaxAttempts = 5` intentos. Al agotarlos, la tarea va a la DLQ.
- **HTTP**: `net/http.ServeMux` nativo de Go 1.22, sin librerías de
  routing externas.

## Tests

```bash
go test -race ./...
```

Cada ciclo de desarrollo se hizo con `-race` activado sin excepciones.

## Limitaciones conocidas

- **Estado en memoria**: todo (tareas pendientes, DLQ, historial) vive
  en RAM. Reiniciar el proceso borra todo sin dejar rastro — no hay
  persistencia en disco ni base de datos.
- **Resize con nearest neighbor**: prioriza simplicidad y
  testeabilidad sobre calidad visual. En escalados grandes (por
  ejemplo, una imagen 4K reducida a 200x200) el resultado se ve
  notoriamente pixelado. Un algoritmo bilinear daría mejor calidad a
  costa de más complejidad.
- **Sin jitter en el backoff**: todos los workers reintentando el
  mismo tipo de fallo en simultáneo esperan exactamente el mismo
  delay, lo que puede generar picos de carga sincronizados
  ("thundering herd") en un escenario con volumen alto.
- **DLQ sin orden garantizado**: `DeadLetterQueue.List()` itera un
  `map`, así que el orden de la lista no es determinístico ni refleja
  cuándo murió cada tarea.
- **Backoff no configurable en runtime**: `backoffBase` y
  `backoffFactor` son constantes fijas en el código, no flags ni
  variables de entorno.

## Posibles mejoras futuras (no implementadas)

- Jitter en el backoff para evitar thundering herd
- `backoffBase` / `backoffFactor` configurables
- Persistencia en disco o base de datos
- Ordenar la DLQ por un campo `DeadAt time.Time`
- Ring buffer o lista enlazada para el inbox si el slicing se vuelve
  cuello de botella con volumen alto

## Estructura del repo