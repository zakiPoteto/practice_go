package service

import (
	"errors"
	"testing"
	model "todo-api/model"

	"gorm.io/gorm"
)

// DB を使わない fake。TaskRepository interface を満たす。
type fakeTaskRepo struct {
	tasks  map[int]model.Task
	nextID int
}

func newFakeTaskRepo() *fakeTaskRepo {
	return &fakeTaskRepo{tasks: map[int]model.Task{}, nextID: 1}
}

func (f *fakeTaskRepo) Create(task *model.Task) error {
	task.ID = uint(f.nextID)
	f.tasks[f.nextID] = *task
	f.nextID++
	return nil
}

func (f *fakeTaskRepo) GetAllByUser(userID uint) ([]model.Task, error) {
	var out []model.Task
	for _, t := range f.tasks {
		if t.UserID == userID {
			out = append(out, t)
		}
	}
	return out, nil
}

func (f *fakeTaskRepo) GetByID(id int) (model.Task, error) {
	t, ok := f.tasks[id]
	if !ok {
		return model.Task{}, gorm.ErrRecordNotFound
	}
	return t, nil
}

func (f *fakeTaskRepo) Update(task *model.Task) error {
	f.tasks[int(task.ID)] = *task
	return nil
}

func (f *fakeTaskRepo) DeleteAllByUser(userID uint) error {
	for id, t := range f.tasks {
		if t.UserID == userID {
			delete(f.tasks, id)
		}
	}
	return nil
}

func (f *fakeTaskRepo) DeleteByID(id int) error {
	delete(f.tasks, id)
	return nil
}

func seed(t *testing.T, repo *fakeTaskRepo, userID uint) int {
	t.Helper()
	task := &model.Task{UserID: userID, Title: "t", Status: "todo"}
	if err := repo.Create(task); err != nil {
		t.Fatal(err)
	}
	return int(task.ID)
}

func TestGetTaskByID(t *testing.T) {
	repo := newFakeTaskRepo()
	svc := NewTaskService(repo)
	id := seed(t, repo, 1)

	if _, err := svc.GetTaskByID(id, 1); err != nil {
		t.Fatalf("自分のタスクは取得できるはず: %v", err)
	}
	if _, err := svc.GetTaskByID(id, 2); !errors.Is(err, ErrForbidden) {
		t.Fatalf("他人のタスクは ErrForbidden のはず: %v", err)
	}
	if _, err := svc.GetTaskByID(999, 1); !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("存在しないIDは ErrTaskNotFound のはず: %v", err)
	}
}

func TestUpdateTask(t *testing.T) {
	repo := newFakeTaskRepo()
	svc := NewTaskService(repo)
	id := seed(t, repo, 1)

	got, err := svc.UpdateTask(id, 1, "new", "done")
	if err != nil {
		t.Fatalf("更新に失敗: %v", err)
	}
	if got.Title != "new" || got.Status != "done" {
		t.Fatalf("戻り値が不正: %+v", got)
	}
	if repo.tasks[id].Title != "new" {
		t.Fatalf("repoに反映されていない")
	}

	if _, err := svc.UpdateTask(id, 2, "hack", "done"); !errors.Is(err, ErrForbidden) {
		t.Fatalf("他人のタスク更新は ErrForbidden のはず: %v", err)
	}
	if repo.tasks[id].Title != "new" {
		t.Fatalf("Forbidden なのに更新されている")
	}
	if _, err := svc.UpdateTask(999, 1, "a", "b"); !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("存在しないIDは ErrTaskNotFound のはず: %v", err)
	}
}

func TestDeleteTask(t *testing.T) {
	repo := newFakeTaskRepo()
	svc := NewTaskService(repo)
	id := seed(t, repo, 1)

	if err := svc.DeleteTask(id, 2); !errors.Is(err, ErrForbidden) {
		t.Fatalf("他人のタスク削除は ErrForbidden のはず: %v", err)
	}
	if _, ok := repo.tasks[id]; !ok {
		t.Fatalf("Forbidden なのに削除されている")
	}
	if err := svc.DeleteTask(id, 1); err != nil {
		t.Fatalf("自分のタスクは削除できるはず: %v", err)
	}
	if err := svc.DeleteTask(id, 1); !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("削除済みは ErrTaskNotFound のはず: %v", err)
	}
}

func TestDeleteAllTasks_OnlyOwn(t *testing.T) {
	repo := newFakeTaskRepo()
	svc := NewTaskService(repo)
	seed(t, repo, 1)
	other := seed(t, repo, 2)

	if err := svc.DeleteAllTasks(1); err != nil {
		t.Fatal(err)
	}
	if len(repo.tasks) != 1 {
		t.Fatalf("残件数が不正: %d", len(repo.tasks))
	}
	if _, ok := repo.tasks[other]; !ok {
		t.Fatalf("他ユーザーのタスクが消えている")
	}
}
