package storer

import (
	"errors"
	"fmt"
	"strings"

	"github.com/digitalocean-apps/sample-with-database/pkg/model"
)

var (
	// ErrNotFound is a standard not found err
	ErrNotFound = errors.New("NotFound")
)

// Storer is an interface for persistence mediusm
type Storer interface {
	// Returns a Note
	Get(id string) (*model.Note, error)
	// Creates a Note
	Create(note *model.Note) error
	// Close closes the DB connection
	Close() error
}

// NewStorer creates a storer client
func NewStorer(connection string, ca string) (Storer, error) {
	if strings.HasPrefix(connection, "postgres") {
		return NewPostgresClient(connection)
	}

	if strings.HasPrefix(connection, "mongodb+srv://") {
		return NewMongoClient(connection, ca)
	}

	if strings.HasPrefix(connection, "mysql") {
		return NewMySQLClient(connection)
	}

	return nil, (fmt.Errorf("Improper connection string format: %v", connection))
}
