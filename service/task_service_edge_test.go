package service

import (
	"errors"
	"testing"
	model "todo-api/model"
)

// GetByID / Delete 等で任意のエラーを返す fake。
type errTaskRepo struct {
	*fakeTaskRepo
	getErr error
}

func (e *errTaskRepo) GetByID(id int) (model.Task, error) {
	return model.Task{}, e.getErr
}

func TestGetAllTasks_OnlyOwn(t *testing.T) {
	repo := newFakeTaskRepo()
	svc := NewTaskService(repo)
	seed(t, repo, 1)
	seed(t, repo, 1)
	seed(t, repo, 2)

	got, err := svc.GetAllTasks(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("自分のタスク2件のはず: %d", len(got))
	}
	for _, task := range got {
		if task.UserID != 1 {
			t.Fatalf("他ユーザーのタスクが混ざっている: %+v", task)
		}
	}
}

func TestGetAllTasks_NoTasks(t *testing.T) {
	svc := NewTaskService(newFakeTaskRepo())

	got, err := svc.GetAllTasks(42)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("タスクなしのユーザーは0件のはず: %d", len(got))
	}
}

func TestDeleteTask_NotFound(t *testing.T) {
	svc := NewTaskService(newFakeTaskRepo())

	if err := svc.DeleteTask(999, 1); !errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("存在しないIDは ErrTaskNotFound のはず: %v", err)
	}
}

func TestDeleteTask_KeepsOtherTasks(t *testing.T) {
	repo := newFakeTaskRepo()
	svc := NewTaskService(repo)
	mine := seed(t, repo, 1)
	other := seed(t, repo, 2)

	if err := svc.DeleteTask(mine, 1); err != nil {
		t.Fatal(err)
	}
	if _, ok := repo.tasks[other]; !ok {
		t.Fatalf("他ユーザーのタスクが消えている")
	}
}

func TestUpdateTask_ForbiddenKeepsStatus(t *testing.T) {
	repo := newFakeTaskRepo()
	svc := NewTaskService(repo)
	id := seed(t, repo, 1)

	if _, err := svc.UpdateTask(id, 2, "hack", "done"); !errors.Is(err, ErrForbidden) {
		t.Fatalf("ErrForbidden のはず: %v", err)
	}
	if repo.tasks[id].Status != "todo" || repo.tasks[id].Title != "t" {
		t.Fatalf("Forbidden なのに変更されている: %+v", repo.tasks[id])
	}
}

func TestCheckOwnership_PropagatesOtherErrors(t *testing.T) {
	boom := errors.New("db down")
	svc := NewTaskService(&errTaskRepo{fakeTaskRepo: newFakeTaskRepo(), getErr: boom})

	if _, err := svc.GetTaskByID(1, 1); !errors.Is(err, boom) || errors.Is(err, ErrTaskNotFound) {
		t.Fatalf("NotFound 以外のエラーはそのまま返るはず: %v", err)
	}
	if _, err := svc.UpdateTask(1, 1, "a", "b"); !errors.Is(err, boom) {
		t.Fatalf("更新時もエラーがそのまま返るはず: %v", err)
	}
	if err := svc.DeleteTask(1, 1); !errors.Is(err, boom) {
		t.Fatalf("削除時もエラーがそのまま返るはず: %v", err)
	}
}
