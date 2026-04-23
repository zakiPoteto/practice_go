package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	task "todo-api/repository"
	testdata "todo-api/test"

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
	r.DELETE("/tasks", h.DeleteAllTasks)
	r.DELETE("/tasks/:id", h.DeleteTaskById)

	return r
}

func TestCreateTask_Success(t *testing.T) {
	r := setupTestRouter(t)

	input := testdata.TaskInput{Title: "write tests", Status: "todo"}
	body, err := testdata.TaskJSONBody(input)
	if err != nil {
		t.Fatalf("リクエストJSONの生成に失敗しました: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(body))
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

	if got["title"] != input.Title {
		t.Fatalf("titleが不正です: 期待=%s 実際=%s", input.Title, got["title"])
	}

	if got["status"] != input.Status {
		t.Fatalf("statusが不正です: 期待=%s 実際=%s", input.Status, got["status"])
	}
}

func TestCreateTask_BadRequest(t *testing.T) {
	r := setupTestRouter(t)

	input := testdata.TaskInput{Title: "missing status"}
	body, err := testdata.TaskJSONBody(input)
	if err != nil {
		t.Fatalf("リクエストJSONの生成に失敗しました: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(body))
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

	// 取得APIの検証用に共通モックデータを投入する。
	for _, input := range testdata.DefaultTaskInputs[:2] {
		if err := repo.Create(task.NewTask(input.Title, input.Status)); err != nil {
			t.Fatalf("テストデータ投入に失敗しました: %v", err)
		}
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

	if got[0]["title"] != testdata.DefaultTaskInputs[0].Title {
		t.Fatalf("1件目titleが不正です: 期待=%s 実際=%s", testdata.DefaultTaskInputs[0].Title, got[0]["title"])
	}

	if got[1]["title"] != testdata.DefaultTaskInputs[1].Title {
		t.Fatalf("2件目titleが不正です: 期待=%s 実際=%s", testdata.DefaultTaskInputs[1].Title, got[1]["title"])
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

func TestDeleteTaskById_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)
	repo := task.NewTaskRepository(db)
	h := NewHandler(repo)

	seed := task.NewTask("to be deleted", "todo")
	if err := repo.Create(seed); err != nil {
		t.Fatalf("テストデータ投入に失敗しました: %v", err)
	}

	r := gin.Default()
	r.DELETE("/tasks/:id", h.DeleteTaskById)

	url := "/tasks/" + strconv.Itoa(seed.ID)
	req := httptest.NewRequest(http.MethodDelete, url, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("ステータスコードが不正です: 期待=200 実際=%d", w.Code)
	}

	_, err := repo.GetTasksById(seed.ID)
	if err == nil {
		t.Fatalf("削除後にタスクが残っています")
	}
}

func TestDeleteTaskById_InvalidID(t *testing.T) {
	r := setupTestRouter(t)

	req := httptest.NewRequest(http.MethodDelete, "/tasks/not-number", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("ステータスコードが不正です: 期待=400 実際=%d", w.Code)
	}
}

func TestDeleteTaskById_NotFound(t *testing.T) {
	r := setupTestRouter(t)

	req := httptest.NewRequest(http.MethodDelete, "/tasks/999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("ステータスコードが不正です: 期待=404 実際=%d", w.Code)
	}
}
