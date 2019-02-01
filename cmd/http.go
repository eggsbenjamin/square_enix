package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/eggsbenjamin/square_enix/internal/app/httphandlers"
	"github.com/eggsbenjamin/square_enix/internal/app/processor"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func startHTTPListeners(proc processor.Processor, port int) {
	startHandler := httphandlers.NewStartHandler(proc)
	pauseHandler := httphandlers.NewPauseHandler(proc)
	statHandler := httphandlers.NewStatHandler(proc)

	mux := chi.NewRouter()

	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)
	mux.Use(middleware.Timeout(30 * time.Second))

	mux.Route("/process", func(r chi.Router) {
		r.Put("/start", startHandler.Handle)
		r.Put("/pause", pauseHandler.Handle)
		r.Get("/stat", statHandler.Handle)
	})

	log.Printf("listening on port %d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), mux))
}
