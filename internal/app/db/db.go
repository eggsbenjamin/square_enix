//go:generate mockgen -package db -source=db.go -destination ./mocks/db.go

package db

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Querier interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	NamedExec(query string, arg interface{}) (sql.Result, error)
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
}

type DB interface {
	Querier
	Beginx() (*sqlx.Tx, error)
}

type querier struct {
	*sqlx.DB
}

func NewQuerier(conn *sqlx.DB) Querier {
	return &querier{
		conn,
	}
}

type db struct {
	*sqlx.DB
}

func NewDB(conn *sqlx.DB) DB {
	return &db{
		conn,
	}
}
