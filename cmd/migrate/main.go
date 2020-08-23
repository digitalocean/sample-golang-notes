package main

import (
	"log"

	"github.com/digitalocean/sample-golang-notes/pkg/notes"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func main() {
	db, err := notes.GetDatabaseConnection()
	if err != nil {
		log.Fatalf("getting database connection: %s", err)
	}
	defer db.Close()

	db.AutoMigrate(&notes.Note{})
}
