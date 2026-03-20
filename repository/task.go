package task

import (
	"gorm.io/gorm"
)

type Task struct {
	gorm.Model
	Id     int
	Title  string
	Status string
}
