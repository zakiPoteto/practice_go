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

type TaskRepository struct {
	db *gorm.DB
}

func (r *TaskRepository) Create(task *Task) error {
	return r.db.Create(task).Error
}
