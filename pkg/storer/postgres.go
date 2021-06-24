package storer

import (
	"fmt"
	"strings"

	"github.com/digitalocean-apps/sample-with-database/pkg/model"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/xo/dburl"
)

// PG is the postgres storer implementation
type PG struct {
	DB *gorm.DB
}

// NewPostgresClient creates a postgres client
func NewPostgresClient(connection string) (*PG, error) {
	dbURL, err := dburl.Parse(connection)
	if err != nil {
		return nil, errors.Wrap(err, "parsing DATABASE_URL")
	}

	// Open a DB connection.
	dbPassword, _ := dbURL.User.Password()
	dbName := strings.Trim(dbURL.Path, "/")
	connectionString := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s", dbURL.Hostname(), dbURL.Port(), dbURL.User.Username(), dbName, dbPassword, dbURL.Query().Get("sslmode"))
	db, err := gorm.Open("postgres", connectionString)
	if err != nil {
		return nil, errors.Wrap(err, "connecting to database")
	}

	initialMigration(db)

	return &PG{DB: db}, nil
}

func initialMigration(db *gorm.DB) {
	db.AutoMigrate(&model.Note{})
}

// Get gets a Note from the DB
func (p *PG) Get(id string) (*model.Note, error) {
	var note model.Note
	err := p.DB.Where("uuid = ?", id).Take(&note).Error
	if gorm.IsRecordNotFoundError(err) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "getting note from db")
	}

	return &note, nil
}

// Create creates a note
func (p *PG) Create(note *model.Note) error {
	err := p.DB.Create(note).Error
	if err != nil {
		return errors.Wrap(err, "creating note in db")
	}

	return nil
}

// Close the DB connection
func (p *PG) Close() error {
	return p.DB.Close()
}
