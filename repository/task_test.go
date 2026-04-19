package task

import (
	"errors"
	"fmt"
	"testing"

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

	created := NewTask("buy milk", "todo")
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

	if tasks[0].Title != "buy milk" {
		t.Fatalf("titleが不正です: 期待=buy milk 実際=%s", tasks[0].Title)
	}

	if tasks[0].Status != "todo" {
		t.Fatalf("statusが不正です: 期待=todo 実際=%s", tasks[0].Status)
	}
}

func TestTaskRepositoryGetTasksById(t *testing.T) {
	repo := setupTestRepo(t)

	created := NewTask("read book", "done")
	if err := repo.Create(created); err != nil {
		t.Fatalf("タスク作成に失敗しました: %v", err)
	}

	got, err := repo.GetTasksById(created.ID)
	if err != nil {
		t.Fatalf("ID指定のタスク取得に失敗しました: %v", err)
	}

	if got.Title != "read book" {
		t.Fatalf("titleが不正です: 期待=read book 実際=%s", got.Title)
	}

	if got.Status != "done" {
		t.Fatalf("statusが不正です: 期待=done 実際=%s", got.Status)
	}
}

func TestTaskRepositoryGetTasksById_NotFound(t *testing.T) {
	repo := setupTestRepo(t)

	_, err := repo.GetTasksById(999)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("想定外のエラーです: 期待=ErrRecordNotFound 実際=%v", err)
	}
}
