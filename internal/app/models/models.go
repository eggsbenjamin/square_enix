package models

import (
	"time"
)

const (
	PROCESS_STATUS_RUNNING  = "RUNNING"
	PROCESS_STATUS_COMPLETE = "COMPLETE"
	PROCESS_STATUS_PAUSED   = "PAUSED"
)

type Process struct {
	ID        int       `db:"id"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
}

type Element struct {
	ID        int       `db:"id"`
	Data      string    `db:"data"`
	CreatedAt time.Time `db:"created_at"`
}
