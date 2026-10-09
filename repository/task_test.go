package repository

import (
	"errors"
	"fmt"
	"testing"
	model "todo-api/model"
	testdata "todo-api/test"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestRepo(t *testing.T) *TaskRepository {
	t.Helper()

	// テスト名をDSNに入れて、他テストとDB状態を分離する。
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("テストDBの接続に失敗しました: %v", err)
	}

	if err := db.AutoMigrate(&model.Task{}); err != nil {
		t.Fatalf("テストDBのマイグレーションに失敗しました: %v", err)
	}

	return NewTaskRepository(db)
}

func TestTaskRepositoryCreateAndGetAllByUser(t *testing.T) {
	repo := setupTestRepo(t)

	input := testdata.DefaultTaskInputs[0]
	created := NewTask(input.Title, input.Status)
	created.UserID = 1
	if err := repo.Create(created); err != nil {
		t.Fatalf("タスク作成に失敗しました: %v", err)
	}
	// 他ユーザーのタスクは含まれないこと
	other := NewTask("other", "todo")
	other.UserID = 2
	if err := repo.Create(other); err != nil {
		t.Fatalf("タスク作成に失敗しました: %v", err)
	}

	tasks, err := repo.GetAllByUser(1)
	if err != nil {
		t.Fatalf("タスク一覧取得に失敗しました: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("件数が不正です: 期待=1 実際=%d", len(tasks))
	}
	if tasks[0].Title != input.Title {
		t.Fatalf("titleが不正です: 期待=%s 実際=%s", input.Title, tasks[0].Title)
	}
	if tasks[0].Status != input.Status {
		t.Fatalf("statusが不正です: 期待=%s 実際=%s", input.Status, tasks[0].Status)
	}
}

func TestTaskRepositoryGetByID(t *testing.T) {
	repo := setupTestRepo(t)

	input := testdata.DefaultTaskInputs[4]
	created := NewTask(input.Title, input.Status)
	if err := repo.Create(created); err != nil {
		t.Fatalf("タスク作成に失敗しました: %v", err)
	}

	got, err := repo.GetByID(int(created.ID))
	if err != nil {
		t.Fatalf("ID指定のタスク取得に失敗しました: %v", err)
	}

	if got.Title != input.Title {
		t.Fatalf("titleが不正です: 期待=%s 実際=%s", input.Title, got.Title)
	}
	if got.Status != input.Status {
		t.Fatalf("statusが不正です: 期待=%s 実際=%s", input.Status, got.Status)
	}
}

func TestTaskRepositoryGetByID_NotFound(t *testing.T) {
	repo := setupTestRepo(t)

	_, err := repo.GetByID(999)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("想定外のエラーです: 期待=ErrRecordNotFound 実際=%v", err)
	}
}

func TestTaskRepositoryUpdate(t *testing.T) {
	repo := setupTestRepo(t)

	created := NewTask("before", "todo")
	if err := repo.Create(created); err != nil {
		t.Fatalf("タスク作成に失敗しました: %v", err)
	}
	created.Title = "after"
	created.Status = "done"
	if err := repo.Update(created); err != nil {
		t.Fatalf("タスク更新に失敗しました: %v", err)
	}

	got, err := repo.GetByID(int(created.ID))
	if err != nil {
		t.Fatalf("タスク取得に失敗しました: %v", err)
	}
	if got.Title != "after" || got.Status != "done" {
		t.Fatalf("更新が反映されていません: %+v", got)
	}
}

func TestTaskRepositoryDeleteAllByUser(t *testing.T) {
	repo := setupTestRepo(t)

	for _, uid := range []uint{1, 1, 2} {
		task := NewTask("t", "todo")
		task.UserID = uid
		if err := repo.Create(task); err != nil {
			t.Fatalf("タスク作成に失敗しました: %v", err)
		}
	}
	if err := repo.DeleteAllByUser(1); err != nil {
		t.Fatalf("全削除に失敗しました: %v", err)
	}

	mine, _ := repo.GetAllByUser(1)
	if len(mine) != 0 {
		t.Fatalf("自分のタスクが残っています: %d件", len(mine))
	}
	others, _ := repo.GetAllByUser(2)
	if len(others) != 1 {
		t.Fatalf("他ユーザーのタスクが消えています: %d件", len(others))
	}
}

func TestTaskRepositoryDeleteByID_Success(t *testing.T) {
	repo := setupTestRepo(t)

	input := testdata.DefaultTaskInputs[2]
	created := NewTask(input.Title, input.Status)
	if err := repo.Create(created); err != nil {
		t.Fatalf("タスク作成に失敗しました: %v", err)
	}

	if err := repo.DeleteByID(int(created.ID)); err != nil {
		t.Fatalf("タスク削除に失敗しました: %v", err)
	}

	_, err := repo.GetByID(int(created.ID))
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("削除後にタスクが残っています: 実際のエラー=%v", err)
	}
}

func TestTaskRepositoryDeleteByID_NotFound(t *testing.T) {
	repo := setupTestRepo(t)

	err := repo.DeleteByID(999)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("想定外のエラーです: 期待=ErrRecordNotFound 実際=%v", err)
	}
}
