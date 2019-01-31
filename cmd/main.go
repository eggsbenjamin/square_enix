package main

import (
	"fmt"
	"log"

	"github.com/eggsbenjamin/square_enix/internal/app/db"
	"github.com/eggsbenjamin/square_enix/internal/app/processor"
	"github.com/eggsbenjamin/square_enix/internal/app/repository"
	"github.com/eggsbenjamin/square_enix/pkg/env"
	"github.com/jmoiron/sqlx"
)

func main() {
	dsn := fmt.Sprintf(
		"%s@tcp(%s:3306)/%s?parseTime=true",
		env.MustGetEnv("MYSQL_USER"),
		env.MustGetEnv("MYSQL_HOST"),
		env.MustGetEnv("MYSQL_DB"),
	)
	conn, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		log.Fatalf("error opening db: %q", err)
	}

	db := db.NewDB(conn)
	proc := processor.NewProcessor(
		db,
		repository.NewProcessRepositoryFactory(),
		repository.NewElementRepositoryFactory(),
	)

	go pollProcess(
		proc,
		env.MustGetIntEnv("BATCH_SIZE"),
		env.MustGetIntEnv("POLL_INTERVAL"),
	)

	startHTTPListeners(
		proc,
		env.MustGetIntEnv("PORT"),
	)
}
