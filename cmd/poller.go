package main

import (
	"log"
	"time"

	"github.com/eggsbenjamin/square_enix/internal/app/processor"
)

func pollProcess(proc processor.Processor, batchSize, pollInterval int) {
	for {
		log.Println("Querying processes")
		runningProcess, err := proc.RunningProcessExists()
		if err != nil {
			log.Fatalf("error getting process info: %q\n", err)
		}

		if runningProcess {
			log.Println("Running process found. Processing batch...")

			if err := proc.ProcessBatch(batchSize); err != nil {
				log.Fatalf("error processing batch: %q\n", err)
			}
		}

		time.Sleep(time.Duration(pollInterval) * time.Second)
	}
}
