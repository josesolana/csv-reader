package database

import "database/sql"

// DB represents available database operations
type DB interface {
	Begin() error
	Commit() error
	Read() (*sql.Rows, error)
	SetAsProcessed(id int) error
	IncreaseRetry(id int) error
	Close() []error
}
