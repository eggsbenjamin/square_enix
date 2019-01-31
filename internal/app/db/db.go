package db

type DB interface {
	Select(dest interface{}, query string, args ...interface{}) error
}
