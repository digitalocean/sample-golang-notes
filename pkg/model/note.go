package model

import "github.com/jinzhu/gorm"

// Note represents a single note.
type Note struct {
	gorm.Model
	UUID string
	Body string
}
