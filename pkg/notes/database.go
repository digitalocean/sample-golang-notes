package notes

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/xo/dburl"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

const (
	defaultDatabaseURL = "postgresql://postgres:postgres@127.0.0.1:5432/notes/?sslmode=disable"
)

// GetDatabaseConnection returns a database connection.
// The user must db.Close() the returned instance.
func GetDatabaseConnection() (*gorm.DB, error) {
	// Parse connection config.
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = defaultDatabaseURL
	}
	dbURL, err := dburl.Parse(databaseURL)
	if err != nil {
		return nil, errors.Wrap(err, "parsing DATABASE_URL")
	}

	// Open a DB connection.
	dbPassword, _ := dbURL.User.Password()
	dbName := strings.Trim(dbURL.Path, "/")
	connectionString := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s", dbURL.Hostname(), dbURL.Port(), dbURL.User.Username(), dbName, dbPassword, dbURL.Query().Get("sslmode"))
	db, err := gorm.Open("postgres", connectionString)
	if err != nil {
		log.Fatalf("connecting to database: %s", err)
		return nil, errors.Wrap(err, "connection to database")
	}

	return db, nil
}
