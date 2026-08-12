package service

import (
	"errors"
	model "todo-api/model"

	"gorm.io/gorm"
)

type TaskRepository interface {
	Create(task *model.Task) error
	GetAllByUser(userID uint) ([]model.Task, error)
	GetByID(id int) (model.Task, error)
	Update(task *model.Task) error
	DeleteAllByUser(userID uint) error
	DeleteByID(id int) error
}
type TaskService struct {
	repo TaskRepository
}

func NewTaskService(repo TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}
func (s *TaskService) CreateTask(task *model.Task) error {
	return s.repo.Create(task)
}
func (s *TaskService) GetAllTasks(userID uint) ([]model.Task, error) {
	return s.repo.GetAllByUser(userID)
}
func (s *TaskService) GetTaskByID(id int, userID uint) (model.Task, error) {
	return s.checkOwnership(id, userID)
}
func (s *TaskService) UpdateTask(id int, userID uint, title string, status string) (model.Task, error) {
	task, err := s.checkOwnership(id, userID)
	if err != nil {
		return model.Task{}, err
	}
	task.Title = title
	task.Status = status
	if err := s.repo.Update(&task); err != nil {
		return model.Task{}, err
	}
	return task, nil
}
func (s *TaskService) DeleteTask(id int, userID uint) error {
	if _, err := s.checkOwnership(id, userID); err != nil {
		return err
	}
	return s.repo.DeleteByID(id)
}
func (s *TaskService) DeleteAllTasks(userID uint) error {
	return s.repo.DeleteAllByUser(userID)
}

func (s *TaskService) checkOwnership(id int, userID uint) (model.Task, error) {
	task, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Task{}, ErrTaskNotFound
		}
		return model.Task{}, err
	}
	if task.UserID != userID {
		return model.Task{}, ErrForbidden
	}
	return task, nil
}
