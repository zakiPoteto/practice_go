package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	task "todo-api/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// テスト名をDSNに含め、テスト間でDBデータが混ざらないようにする。
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("テストDBの接続に失敗しました: %v", err)
	}

	if err := db.AutoMigrate(&task.Task{}); err != nil {
		t.Fatalf("テストDBのマイグレーションに失敗しました: %v", err)
	}

	return db
}

func setupTestRouter(t *testing.T) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)

	db := setupTestDB(t)
	repo := task.NewTaskRepository(db)
	h := NewHandler(repo)

	r := gin.Default()
	r.POST("/tasks", h.CreateTask)
	r.GET("/tasks", h.GetAllTasks)
	r.GET("/tasks/:id", h.GetTasksById)

	return r
}

func TestCreateTask_Success(t *testing.T) {
	r := setupTestRouter(t)

	// Arrange: 正常なリクエストボディ
	body := `{"title":"write tests","status":"todo"}`
	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("ステータスコードが不正です: 期待=201 実際=%d", w.Code)
	}

	var got map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("レスポンスJSONの解析に失敗しました: %v", err)
	}

	if got["title"] != "write tests" {
		t.Fatalf("titleが不正です: 期待=write tests 実際=%s", got["title"])
	}

	if got["status"] != "todo" {
		t.Fatalf("statusが不正です: 期待=todo 実際=%s", got["status"])
	}
}

func TestCreateTask_BadRequest(t *testing.T) {
	r := setupTestRouter(t)

	body := `{"title":"missing status"}`
	req := httptest.NewRequest(http.MethodPost, "/tasks", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("ステータスコードが不正です: 期待=400 実際=%d", w.Code)
	}
}

func TestGetAllTasks_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	repo := task.NewTaskRepository(db)
	h := NewHandler(repo)

	// 取得APIの検証用に事前データを投入する。
	if err := repo.Create(task.NewTask("a", "todo")); err != nil {
		t.Fatalf("テストデータ投入に失敗しました: %v", err)
	}
	if err := repo.Create(task.NewTask("b", "done")); err != nil {
		t.Fatalf("テストデータ投入に失敗しました: %v", err)
	}

	r := gin.Default()
	r.GET("/tasks", h.GetAllTasks)

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("ステータスコードが不正です: 期待=200 実際=%d", w.Code)
	}

	var got []map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("レスポンスJSONの解析に失敗しました: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("取得件数が不正です: 期待=2 実際=%d", len(got))
	}
}

func TestGetTasksById_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	repo := task.NewTaskRepository(db)
	h := NewHandler(repo)

	seed := task.NewTask("single", "todo")
	if err := repo.Create(seed); err != nil {
		t.Fatalf("テストデータ投入に失敗しました: %v", err)
	}

	r := gin.Default()
	r.GET("/tasks/:id", h.GetTasksById)

	url := "/tasks/" + strconv.Itoa(seed.ID)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("ステータスコードが不正です: 期待=200 実際=%d", w.Code)
	}

	var got map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("レスポンスJSONの解析に失敗しました: %v", err)
	}

	if got["title"] != "single" {
		t.Fatalf("titleが不正です: 期待=single 実際=%s", got["title"])
	}
}

func TestGetTasksById_InvalidID(t *testing.T) {
	r := setupTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/tasks/not-number", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("ステータスコードが不正です: 期待=400 実際=%d", w.Code)
	}
}

func TestGetTasksById_NotFound(t *testing.T) {
	r := setupTestRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/tasks/999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("ステータスコードが不正です: 期待=404 実際=%d", w.Code)
	}
}
