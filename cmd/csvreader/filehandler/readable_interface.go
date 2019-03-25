package filehandler

// Readable Interface to wraps CSV Reader
type Readable interface {
	Read() ([]string, error)
	Close() error
}
