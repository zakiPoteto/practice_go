package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	model "todo-api/model"
	task "todo-api/repository"
	"todo-api/service"
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

	if err := db.AutoMigrate(&model.Task{}); err != nil {
		t.Fatalf("テストDBのマイグレーションに失敗しました: %v", err)
	}

	return db
}

// 認証ミドルウェアの代わりに、固定の userID をコンテキストへ入れる
func fakeAuth(userID uint) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("userID", userID)
		c.Next()
	}
}

// 指定ユーザーとしてログイン済みのルーターを組み立てる
func setupRouter(repo *task.TaskRepository, userID uint) *gin.Engine {
	gin.SetMode(gin.TestMode)

	h := NewHandler(service.NewTaskService(repo))

	r := gin.Default()
	g := r.Group("/tasks", fakeAuth(userID))
	g.POST("", h.CreateTask)
	g.GET("", h.GetAllTasks)
	g.GET("/:id", h.GetTasksById)
	g.PUT("/:id", h.UpdateTask)
	g.DELETE("", h.DeleteAllTasks)
	g.DELETE("/:id", h.DeleteTaskById)

	return r
}

func setupTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	return setupRouter(task.NewTaskRepository(setupTestDB(t)), 1)
}

func seedTask(t *testing.T, repo *task.TaskRepository, userID uint, title, status string) *model.Task {
	t.Helper()
	seed := task.NewTask(title, status)
	seed.UserID = userID
	if err := repo.Create(seed); err != nil {
		t.Fatalf("テストデータ投入に失敗しました: %v", err)
	}
	return seed
}

func doRequest(r *gin.Engine, method, url string, body []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestCreateTask_Success(t *testing.T) {
	r := setupTestRouter(t)

	input := testdata.TaskInput{Title: "write tests", Status: "todo"}
	body, err := testdata.TaskJSONBody(input)
	if err != nil {
		t.Fatalf("リクエストJSONの生成に失敗しました: %v", err)
	}

	w := doRequest(r, http.MethodPost, "/tasks", body)
	if w.Code != http.StatusCreated {
		t.Fatalf("ステータスコードが不正です: 期待=201 実際=%d", w.Code)
	}

	var got map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("レスポンスJSONの解析に失敗しました: %v", err)
	}
	if got["title"] != input.Title {
		t.Fatalf("titleが不正です: 期待=%s 実際=%v", input.Title, got["title"])
	}
	if got["status"] != input.Status {
		t.Fatalf("statusが不正です: 期待=%s 実際=%v", input.Status, got["status"])
	}
}

func TestCreateTask_IgnoresUserIDInBody(t *testing.T) {
	repo := task.NewTaskRepository(setupTestDB(t))
	r := setupRouter(repo, 1)

	w := doRequest(r, http.MethodPost, "/tasks", []byte(`{"title":"a","status":"todo","user_id":99}`))
	if w.Code != http.StatusCreated {
		t.Fatalf("ステータスコードが不正です: 期待=201 実際=%d", w.Code)
	}

	mine, _ := repo.GetAllByUser(1)
	if len(mine) != 1 {
		t.Fatalf("トークン由来のユーザーで作成されていません: %d件", len(mine))
	}
}

func TestCreateTask_BadRequest(t *testing.T) {
	r := setupTestRouter(t)

	body, err := testdata.TaskJSONBody(testdata.TaskInput{Title: "missing status"})
	if err != nil {
		t.Fatalf("リクエストJSONの生成に失敗しました: %v", err)
	}

	w := doRequest(r, http.MethodPost, "/tasks", body)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("ステータスコードが不正です: 期待=400 実際=%d", w.Code)
	}
}

func TestGetAllTasks_OnlyOwn(t *testing.T) {
	repo := task.NewTaskRepository(setupTestDB(t))
	for _, input := range testdata.DefaultTaskInputs[:2] {
		seedTask(t, repo, 1, input.Title, input.Status)
	}
	seedTask(t, repo, 2, "someone else", "todo")

	r := setupRouter(repo, 1)
	w := doRequest(r, http.MethodGet, "/tasks", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("ステータスコードが不正です: 期待=200 実際=%d", w.Code)
	}

	var got []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("レスポンスJSONの解析に失敗しました: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("取得件数が不正です: 期待=2 実際=%d", len(got))
	}
	for i, want := range testdata.DefaultTaskInputs[:2] {
		if got[i]["title"] != want.Title {
			t.Fatalf("%d件目titleが不正です: 期待=%s 実際=%v", i+1, want.Title, got[i]["title"])
		}
	}
}

func TestGetTasksById_Success(t *testing.T) {
	repo := task.NewTaskRepository(setupTestDB(t))
	seed := seedTask(t, repo, 1, "single", "todo")

	r := setupRouter(repo, 1)
	w := doRequest(r, http.MethodGet, "/tasks/"+strconv.Itoa(int(seed.ID)), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("ステータスコードが不正です: 期待=200 実際=%d", w.Code)
	}

	var got map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("レスポンスJSONの解析に失敗しました: %v", err)
	}
	if got["title"] != "single" {
		t.Fatalf("titleが不正です: 期待=single 実際=%v", got["title"])
	}
}

func TestGetTasksById_InvalidID(t *testing.T) {
	r := setupTestRouter(t)

	w := doRequest(r, http.MethodGet, "/tasks/not-number", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("ステータスコードが不正です: 期待=400 実際=%d", w.Code)
	}
}

func TestGetTasksById_NotFound(t *testing.T) {
	r := setupTestRouter(t)

	w := doRequest(r, http.MethodGet, "/tasks/999", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("ステータスコードが不正です: 期待=404 実際=%d", w.Code)
	}
}

func TestGetTasksById_Forbidden(t *testing.T) {
	repo := task.NewTaskRepository(setupTestDB(t))
	seed := seedTask(t, repo, 2, "not mine", "todo")

	r := setupRouter(repo, 1)
	w := doRequest(r, http.MethodGet, "/tasks/"+strconv.Itoa(int(seed.ID)), nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("ステータスコードが不正です: 期待=403 実際=%d", w.Code)
	}
}

func TestUpdateTask_Success(t *testing.T) {
	repo := task.NewTaskRepository(setupTestDB(t))
	seed := seedTask(t, repo, 1, "before", "todo")

	r := setupRouter(repo, 1)
	w := doRequest(r, http.MethodPut, "/tasks/"+strconv.Itoa(int(seed.ID)),
		[]byte(`{"title":"after","status":"done"}`))
	if w.Code != http.StatusOK {
		t.Fatalf("ステータスコードが不正です: 期待=200 実際=%d", w.Code)
	}

	got, err := repo.GetByID(int(seed.ID))
	if err != nil {
		t.Fatalf("取得に失敗しました: %v", err)
	}
	if got.Title != "after" || got.Status != "done" {
		t.Fatalf("更新が反映されていません: %+v", got)
	}
}

func TestUpdateTask_BadRequest(t *testing.T) {
	repo := task.NewTaskRepository(setupTestDB(t))
	seed := seedTask(t, repo, 1, "before", "todo")

	r := setupRouter(repo, 1)
	w := doRequest(r, http.MethodPut, "/tasks/"+strconv.Itoa(int(seed.ID)), []byte(`{"title":"only title"}`))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("ステータスコードが不正です: 期待=400 実際=%d", w.Code)
	}
}

func TestUpdateTask_NotFound(t *testing.T) {
	r := setupTestRouter(t)

	w := doRequest(r, http.MethodPut, "/tasks/999", []byte(`{"title":"a","status":"done"}`))
	if w.Code != http.StatusNotFound {
		t.Fatalf("ステータスコードが不正です: 期待=404 実際=%d", w.Code)
	}
}

func TestUpdateTask_Forbidden(t *testing.T) {
	repo := task.NewTaskRepository(setupTestDB(t))
	seed := seedTask(t, repo, 2, "not mine", "todo")

	r := setupRouter(repo, 1)
	w := doRequest(r, http.MethodPut, "/tasks/"+strconv.Itoa(int(seed.ID)),
		[]byte(`{"title":"hack","status":"done"}`))
	if w.Code != http.StatusForbidden {
		t.Fatalf("ステータスコードが不正です: 期待=403 実際=%d", w.Code)
	}
}

func TestDeleteAllTasks_OnlyOwn(t *testing.T) {
	repo := task.NewTaskRepository(setupTestDB(t))
	seedTask(t, repo, 1, "mine", "todo")
	seedTask(t, repo, 2, "theirs", "todo")

	r := setupRouter(repo, 1)
	w := doRequest(r, http.MethodDelete, "/tasks", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("ステータスコードが不正です: 期待=200 実際=%d", w.Code)
	}

	if mine, _ := repo.GetAllByUser(1); len(mine) != 0 {
		t.Fatalf("自分のタスクが残っています: %d件", len(mine))
	}
	if others, _ := repo.GetAllByUser(2); len(others) != 1 {
		t.Fatalf("他ユーザーのタスクが消えています: %d件", len(others))
	}
}

func TestDeleteTaskById_Success(t *testing.T) {
	repo := task.NewTaskRepository(setupTestDB(t))
	seed := seedTask(t, repo, 1, "to be deleted", "todo")

	r := setupRouter(repo, 1)
	w := doRequest(r, http.MethodDelete, "/tasks/"+strconv.Itoa(int(seed.ID)), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("ステータスコードが不正です: 期待=200 実際=%d", w.Code)
	}

	if _, err := repo.GetByID(int(seed.ID)); err == nil {
		t.Fatalf("削除後にタスクが残っています")
	}
}

func TestDeleteTaskById_InvalidID(t *testing.T) {
	r := setupTestRouter(t)

	w := doRequest(r, http.MethodDelete, "/tasks/not-number", nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("ステータスコードが不正です: 期待=400 実際=%d", w.Code)
	}
}

func TestDeleteTaskById_NotFound(t *testing.T) {
	r := setupTestRouter(t)

	w := doRequest(r, http.MethodDelete, "/tasks/999", nil)
	if w.Code != http.StatusNotFound {
		t.Fatalf("ステータスコードが不正です: 期待=404 実際=%d", w.Code)
	}
}

func TestDeleteTaskById_Forbidden(t *testing.T) {
	repo := task.NewTaskRepository(setupTestDB(t))
	seed := seedTask(t, repo, 2, "not mine", "todo")

	r := setupRouter(repo, 1)
	w := doRequest(r, http.MethodDelete, "/tasks/"+strconv.Itoa(int(seed.ID)), nil)
	if w.Code != http.StatusForbidden {
		t.Fatalf("ステータスコードが不正です: 期待=403 実際=%d", w.Code)
	}
	if _, err := repo.GetByID(int(seed.ID)); err != nil {
		t.Fatalf("他人のタスクが削除されています: %v", err)
	}
}
