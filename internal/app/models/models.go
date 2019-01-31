package models

type Process struct {
	ID        int       `db:"id"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
}
