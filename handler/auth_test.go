package handler

import (
	"encoding/json"
	"net/http"
	"testing"
	model "todo-api/model"
	repo "todo-api/repository"
	"todo-api/service"

	"github.com/gin-gonic/gin"
)

func setupAuthRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db := setupTestDB(t)
	if err := db.AutoMigrate(&model.User{}); err != nil {
		t.Fatalf("マイグレーションに失敗しました: %v", err)
	}
	h := NewAuthHandler(service.NewAuthService(repo.NewUserRepository(db)))

	r := gin.Default()
	r.POST("/auth/register", h.Register)
	r.POST("/auth/login", h.Login)
	return r
}

const registerBody = `{"name":"taro","email":"taro@example.com","password":"secret123"}`

func TestRegister_Success(t *testing.T) {
	r := setupAuthRouter(t)

	w := doRequest(r, http.MethodPost, "/auth/register", []byte(registerBody))
	if w.Code != http.StatusCreated {
		t.Fatalf("ステータスコードが不正です: 期待=201 実際=%d body=%s", w.Code, w.Body.String())
	}

	var got map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("レスポンスJSONの解析に失敗しました: %v", err)
	}
	if got["email"] != "taro@example.com" {
		t.Fatalf("emailが不正です: %v", got["email"])
	}
	if _, ok := got["password"]; ok {
		t.Fatalf("レスポンスにパスワードが含まれています")
	}
}

func TestRegister_BadRequest(t *testing.T) {
	r := setupAuthRouter(t)

	w := doRequest(r, http.MethodPost, "/auth/register", []byte(`{"name":"taro","email":"taro@example.com"}`))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("ステータスコードが不正です: 期待=400 実際=%d", w.Code)
	}
}

func TestLogin_AfterRegister(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	r := setupAuthRouter(t)
	doRequest(r, http.MethodPost, "/auth/register", []byte(registerBody))

	w := doRequest(r, http.MethodPost, "/auth/login", []byte(`{"email":"taro@example.com","password":"secret123"}`))
	if w.Code != http.StatusOK {
		t.Fatalf("ステータスコードが不正です: 期待=200 実際=%d body=%s", w.Code, w.Body.String())
	}
	var got map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || got["token"] == "" {
		t.Fatalf("tokenが返っていません: %s", w.Body.String())
	}

	w = doRequest(r, http.MethodPost, "/auth/login", []byte(`{"email":"taro@example.com","password":"wrong"}`))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("誤ったパスワードで 401 になりません: 実際=%d", w.Code)
	}
}
