package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/xo/dburl"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

const (
	defaultPort        = "80"
	defaultDatabaseURL = "postgresql://postgres:postgres@127.0.0.1:5432/notes/?sslmode=disable"
)

// Note represents a single note.
type Note struct {
	gorm.Model
	Uuid string
	Body string
}

func initialMigration(db *gorm.DB) {
	db.AutoMigrate(&Note{})
}

func main() {
	// Parse connection config.
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = defaultDatabaseURL
	}
	dbURL, err := dburl.Parse(databaseURL)
	requireNoError(err, "parsing DATABASE_URL")

	// Open a DB connection.
	dbPassword, _ := dbURL.User.Password()
	dbName := strings.Trim(dbURL.Path, "/")
	connectionString := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s", dbURL.Hostname(), dbURL.Port(), dbURL.User.Username(), dbName, dbPassword, dbURL.Query().Get("sslmode"))
	db, err := gorm.Open("postgres", connectionString)
	requireNoError(err, "connecting to database")
	defer db.Close()
	initialMigration(db)

	// Initialize the listening port.
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Initialize and start the HTTP server.
	r := mux.NewRouter()
	r.HandleFunc("/", notesHandler(db)).Methods("GET", "POST")
	r.HandleFunc("/{note_id}", noteHandler(db)).Methods("GET")
	http.ListenAndServe(fmt.Sprintf(":%s", port), r)
}

func notesHandler(db *gorm.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var body string

			if r.Body != nil {
				buf := new(strings.Builder)
				_, err := io.Copy(buf, r.Body)
				if !requireNoErrorInHandler(w, err, "reading request body") {
					return
				}
				body = buf.String()
			}

			if body == "" {
				w.WriteHeader(http.StatusUnprocessableEntity)
				fmt.Fprintf(w, "invalid empty note")
				return
			}

			noteUUID, err := uuid.NewV4()
			if !requireNoErrorInHandler(w, err, "creating note uuid") {
				return
			}

			err = db.Create(&Note{
				Uuid: noteUUID.String(),
				Body: body,
			}).Error
			if !requireNoErrorInHandler(w, err, "creating note in db") {
				return
			}
			log.Printf("POST %s %s\n", noteUUID.String(), body)

			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, noteUUID.String())
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "POST a note")
	}
}

func noteHandler(db *gorm.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		noteID := vars["note_id"]

		if noteID == "" {
			w.WriteHeader(http.StatusUnprocessableEntity)
			fmt.Fprintf(w, "note_id required")
			return
		}

		var note Note
		err := db.Where("uuid = ?", noteID).Take(&note).Error
		if gorm.IsRecordNotFoundError(err) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if !requireNoErrorInHandler(w, err, "getting note from db") {
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, note.Body)
	}
}

func requireNoErrorInHandler(w http.ResponseWriter, err error, msg string) bool {
	if err != nil {
		log.Printf(errors.Wrap(err, msg).Error())
		w.WriteHeader(http.StatusInternalServerError)
		return false
	}
	return true
}

func requireNoError(err error, msg string) {
	if err != nil {
		panic(errors.Wrap(err, msg))
	}
}
