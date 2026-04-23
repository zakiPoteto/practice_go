package task

import (
	"errors"
	"fmt"
	"testing"
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

	if err := db.AutoMigrate(&Task{}); err != nil {
		t.Fatalf("テストDBのマイグレーションに失敗しました: %v", err)
	}

	return NewTaskRepository(db)
}

func TestTaskRepositoryCreateAndGetAll(t *testing.T) {
	repo := setupTestRepo(t)

	input := testdata.DefaultTaskInputs[0]
	created := NewTask(input.Title, input.Status)
	if err := repo.Create(created); err != nil {
		t.Fatalf("タスク作成に失敗しました: %v", err)
	}

	tasks, err := repo.GetAll()
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

func TestTaskRepositoryGetTasksById(t *testing.T) {
	repo := setupTestRepo(t)

	input := testdata.DefaultTaskInputs[4]
	created := NewTask(input.Title, input.Status)
	if err := repo.Create(created); err != nil {
		t.Fatalf("タスク作成に失敗しました: %v", err)
	}

	got, err := repo.GetTasksById(created.ID)
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

func TestTaskRepositoryGetTasksById_NotFound(t *testing.T) {
	repo := setupTestRepo(t)

	_, err := repo.GetTasksById(999)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("想定外のエラーです: 期待=ErrRecordNotFound 実際=%v", err)
	}
}

func TestTaskRepositoryDeleteById_Success(t *testing.T) {
	repo := setupTestRepo(t)

	input := testdata.DefaultTaskInputs[2]
	created := NewTask(input.Title, input.Status)
	if err := repo.Create(created); err != nil {
		t.Fatalf("タスク作成に失敗しました: %v", err)
	}

	if err := repo.DeleteById(created.ID); err != nil {
		t.Fatalf("タスク削除に失敗しました: %v", err)
	}

	_, err := repo.GetTasksById(created.ID)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("削除後にタスクが残っています: 実際のエラー=%v", err)
	}
}

func TestTaskRepositoryDeleteById_NotFound(t *testing.T) {
	repo := setupTestRepo(t)

	err := repo.DeleteById(999)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("想定外のエラーです: 期待=ErrRecordNotFound 実際=%v", err)
	}
}
