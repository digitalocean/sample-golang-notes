package notes

import (
	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// Note represents a single note.
type Note struct {
	gorm.Model
	Uuid string
	Body string
}
