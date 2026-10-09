package service

import (
	"errors"
	"testing"
	model "todo-api/model"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type fakeUserRepo struct {
	users map[string]*model.User
}

func newFakeUserRepo() *fakeUserRepo {
	return &fakeUserRepo{users: map[string]*model.User{}}
}

func (f *fakeUserRepo) Create(user *model.User) error {
	user.ID = uint(len(f.users) + 1)
	f.users[user.Email] = user
	return nil
}

func (f *fakeUserRepo) FindByEmail(email string) (*model.User, error) {
	u, ok := f.users[email]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return u, nil
}

func registerUser(t *testing.T, svc *AuthService) {
	t.Helper()
	if err := svc.Register(&model.User{Name: "taro", Email: "taro@example.com", Password: "secret"}); err != nil {
		t.Fatalf("登録に失敗: %v", err)
	}
}

func TestRegister_HashesPassword(t *testing.T) {
	repo := newFakeUserRepo()
	svc := NewAuthService(repo)
	registerUser(t, svc)

	stored := repo.users["taro@example.com"]
	if stored.Password == "secret" || stored.Password == "" {
		t.Fatalf("パスワードがハッシュ化されていない: %q", stored.Password)
	}
}

func TestLogin_Success(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	svc := NewAuthService(newFakeUserRepo())
	registerUser(t, svc)

	tokenString, err := svc.Login("taro@example.com", "secret")
	if err != nil {
		t.Fatalf("ログインに失敗: %v", err)
	}
	token, err := jwt.Parse(tokenString, func(*jwt.Token) (any, error) { return []byte("test-secret"), nil })
	if err != nil || !token.Valid {
		t.Fatalf("発行されたトークンが不正: %v", err)
	}
	claims := token.Claims.(jwt.MapClaims)
	if claims["user_id"].(float64) != 1 {
		t.Fatalf("user_id が不正: %v", claims["user_id"])
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	svc := NewAuthService(newFakeUserRepo())
	registerUser(t, svc)

	cases := map[string][2]string{
		"wrong password": {"taro@example.com", "wrong"},
		"unknown email":  {"nobody@example.com", "secret"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := svc.Login(c[0], c[1]); !errors.Is(err, ErrInvalidCredentials) {
				t.Fatalf("ErrInvalidCredentials のはず: %v", err)
			}
		})
	}
}
