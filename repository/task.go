package task

import (
	"gorm.io/gorm"
)

type Task struct {
	gorm.Model
	Title  string
	Status string
}
type TaskRepository struct {
	db *gorm.DB
}


func NewTask(title string, status string) *Task {
	return &Task{
		Title:  title,
		Status: status,
	}
}
//挿入
func (r *TaskRepository) Create(task *Task) error {
	return r.db.Create(task).Error
}
func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{
		db: db,
	}
}