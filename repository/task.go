package repository

import (
	model "todo-api/model"

	"gorm.io/gorm"
)

type TaskRepository struct {
	db *gorm.DB
}

func NewTask(title string, status string) *model.Task {
	return &model.Task{

		Title:  title,
		Status: status,
	}
}

// 挿入
func (r *TaskRepository) Create(task *model.Task) error {
	return r.db.Create(task).Error
}
func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{
		db: db,
	}
}
func (r *TaskRepository) GetAll() ([]model.Task, error) {
	var tasks []model.Task
	err := r.db.Find(&tasks).Error
	return tasks, err
}
func (r *TaskRepository) GetTasksById(id int) (model.Task, error) {
	var tasks model.Task
	err := r.db.First(&tasks, "id = ?", id).Error
	return tasks, err
}
func (r *TaskRepository) DeleteAll() error {
	return r.db.Exec("DELETE FROM tasks").Error
}
func (r *TaskRepository) DeleteById(id int) error {
	result := r.db.Delete(&model.Task{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
