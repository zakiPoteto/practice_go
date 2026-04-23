package task

import (
	"gorm.io/gorm"
)

type Task struct {
	ID     int    `json:"-"`
	Title  string `json:"title" binding:"required"`
	Status string `json:"status" binding:"required"`
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

// 挿入
func (r *TaskRepository) Create(task *Task) error {
	return r.db.Create(task).Error
}
func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{
		db: db,
	}
}
func (r *TaskRepository) GetAll() ([]Task, error) {
	var tasks []Task
	err := r.db.Find(&tasks).Error
	return tasks, err
}
func (r *TaskRepository) GetTasksById(id int) (Task, error) {
	var tasks Task
	err := r.db.First(&tasks, "id = ?", id).Error
	return tasks, err
}
func (r *TaskRepository) DeleteAll() error {
	return r.db.Exec("DELETE FROM tasks").Error
}
