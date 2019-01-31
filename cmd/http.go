package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/eggsbenjamin/square_enix/internal/app/httphandlers"
	"github.com/eggsbenjamin/square_enix/internal/app/processor"
	"github.com/go-chi/chi"
)

func startHTTPListeners(proc processor.Processor, port int) {
	startHandler := httphandlers.NewStartHandler(proc)
	pauseHandler := httphandlers.NewPauseHandler(proc)
	statHandler := httphandlers.NewStatHandler(proc)

	mux := chi.NewRouter()
	mux.Put("/start", startHandler.Handle)
	mux.Put("/pause", pauseHandler.Handle)
	mux.Get("/stat", statHandler.Handle)

	log.Printf("listening on port %d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), mux))
}
