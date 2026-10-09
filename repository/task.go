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
func (r *TaskRepository) GetAllByUser(userID uint) ([]model.Task, error) {
	var tasks []model.Task
	err := r.db.Where("user_id=?", userID).Find(&tasks).Error
	return tasks, err
}
func (r *TaskRepository) GetByID(id int) (model.Task, error) {
	var task model.Task
	err := r.db.First(&task, "id = ?", id).Error
	return task, err
}
func (r *TaskRepository) Update(task *model.Task) error {
	return r.db.Save(task).Error
}
func (r *TaskRepository) DeleteAllByUser(userID uint) error {
	return r.db.Where("user_id=?", userID).Delete(&model.Task{}).Error
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{
		db: db,
	}
}
func (r *TaskRepository) DeleteByID(id int) error {
	result := r.db.Delete(&model.Task{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
