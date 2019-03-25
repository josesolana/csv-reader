package database

// DB represents available database operations
type DB interface {
	Insert(row ...string) error
	Close() error
}
