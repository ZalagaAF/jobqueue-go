package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/ZalagaAF/jobqueue-go/internal/api"
	"github.com/ZalagaAF/jobqueue-go/internal/queue"
)

func main() {
	addr := flag.String("addr", ":8080", "direccion donde escucha el servidor")
	numWorkers := flag.Int("workers", 4, "cantidad de workers concurrentes")
	channelCapacity := flag.Int("channel-capacity", 16, "capacidad del channel bounded entre el dispatcher y los workers")
	flag.Parse()

	// ctx se cancela solo cuando llega SIGINT (Ctrl+C) o SIGTERM (lo
	// que manda, por ejemplo, `docker stop` o un orquestador). Es el
	// mismo ctx que ya usan wp.Start, el dispatcher y cada worker —
	// un solo mecanismo de apagado para todo el sistema, consistente
	// con la decisión de Semana 2.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	dlq := queue.NewDeadLetterQueue()
	wp := queue.NewWorkerPool(*numWorkers, *channelCapacity, dlq)

	// Acá es donde se registran los tipos de Job concretos 
	jobRegistry := queue.NewJobRegistry()
	jobRegistry.Register("email", func(payload json.RawMessage) (queue.Job, error) {
		var j queue.EmailJob
		if err := json.Unmarshal(payload, &j); err != nil {
			return nil, err
		}
		return j, nil
	})

	jobRegistry.Register("image_resize", func(payload json.RawMessage) (queue.Job, error) {
	var j queue.ImageResizeJob
	if err := json.Unmarshal(payload, &j); err != nil {
		return nil, err
	}
	return j, nil
	})

	wp.Start(ctx)

	srv := api.NewServer(wp, jobRegistry, dlq)

	httpServer := &http.Server{
		Addr:    *addr,
		Handler: srv.Routes(),
	}

	go func() {
		log.Printf("jobqueue-go escuchando en %s (workers=%d, channel-capacity=%d)", *addr, *numWorkers, *channelCapacity)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("error del servidor HTTP: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("señal de apagado recibida, cerrando...")

	// ctx ya está cancelado (por eso llegamos acá) — Shutdown necesita
	// su PROPIO contexto, con un margen fresco, para saber cuánto
	// esperar a que terminen las requests en vuelo antes de forzar
	// el cierre.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("error cerrando el servidor HTTP: %v", err)
	}
}