package storer

import (
	"fmt"
	"strings"

	"github.com/digitalocean-apps/sample-with-database/pkg/model"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/xo/dburl"

	_ "github.com/jinzhu/gorm/dialects/mysql"
)

// MySQL is the MySQL storer implementation
type MySQL struct {
	DB *gorm.DB
}

// NewMySQLClient creates a mysql client
func NewMySQLClient(connection string) (*MySQL, error) {
	dbURL, err := dburl.Parse(connection)
	if err != nil {
		return nil, errors.Wrap(err, "parsing DATABASE_URL")
	}

	// Open a DB connection.
	dbPassword, _ := dbURL.User.Password()
	dbName := strings.Trim(dbURL.Path, "/")
	connectionString := fmt.Sprintf("%s:%s@(%s:%s)/%s?charset=utf8&parseTime=true", dbURL.User.Username(), dbPassword, dbURL.Hostname(), dbURL.Port(), dbName)
	db, err := gorm.Open("mysql", connectionString)
	if err != nil {
		return nil, errors.Wrap(err, "connecting to database")
	}

	db.AutoMigrate(&model.Note{})

	return &MySQL{DB: db}, nil
}

// Get gets a Note from the DB
func (p *MySQL) Get(id string) (*model.Note, error) {
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
func (p *MySQL) Create(note *model.Note) error {
	err := p.DB.Create(note).Error
	if err != nil {
		return errors.Wrap(err, "creating note in db")
	}

	return nil
}

// Close the DB connection
func (p *MySQL) Close() error {
	return p.DB.Close()
}
