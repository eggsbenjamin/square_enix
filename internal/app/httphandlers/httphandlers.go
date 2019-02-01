package httphandlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/eggsbenjamin/square_enix/internal/app/processor"
)

type StartHandler struct {
	proc processor.Processor
}

func NewStartHandler(proc processor.Processor) *StartHandler {
	return &StartHandler{
		proc: proc,
	}
}

func (s *StartHandler) Handle(w http.ResponseWriter, req *http.Request) {
	log.Print("starting process")

	w.Header().Set("Content-Type", "application/json")

	if err := s.proc.Start(); err != nil {
		if err == processor.ErrRunningProcessExists {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"message":"running process exists"}`))
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"internal server error"}`))
		return
	}

	log.Print("process started")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"message":"process started"}`))
}

type StatHandler struct {
	proc processor.Processor
}

func NewStatHandler(proc processor.Processor) *StatHandler {
	return &StatHandler{
		proc: proc,
	}
}

func (s *StatHandler) Handle(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	stat, err := s.proc.GetLatestsStat()
	if err != nil {
		if err == processor.ErrNoProcessExists {
			w.WriteHeader(http.StatusPreconditionFailed)
			w.Write([]byte(`{"message":"no processes exist"}`))
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"internal server error"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"stat": %d}`, stat)))
	return
}

type PauseHandler struct {
	proc processor.Processor
}

func NewPauseHandler(proc processor.Processor) *PauseHandler {
	return &PauseHandler{
		proc: proc,
	}
}

func (s *PauseHandler) Handle(w http.ResponseWriter, req *http.Request) {
	log.Print("pausing process")

	w.Header().Set("Content-Type", "application/json")

	if err := s.proc.Pause(); err != nil {
		if err == processor.ErrNoProcessExists ||
			err == processor.ErrNoRunningProcessExists {
			w.WriteHeader(http.StatusPreconditionFailed)
			w.Write([]byte(`{"message":"no applicable processes exist"}`))
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"internal server error"}`))
		return
	}

	log.Print("process paused")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"message":"process paused"}`))
}
