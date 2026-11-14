package service

import (
	"errors"
	"testing"
	"time"
	model "todo-api/model"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Create が常に失敗する fake。
type failingUserRepo struct {
	*fakeUserRepo
	createErr error
}

func (f *failingUserRepo) Create(user *model.User) error {
	return f.createErr
}

func TestRegister_PasswordMatchesHash(t *testing.T) {
	repo := newFakeUserRepo()
	svc := NewAuthService(repo)
	registerUser(t, svc)

	stored := repo.users["taro@example.com"]
	if err := bcrypt.CompareHashAndPassword([]byte(stored.Password), []byte("secret")); err != nil {
		t.Fatalf("ハッシュが元のパスワードと一致しない: %v", err)
	}
}

func TestRegister_PropagatesRepoError(t *testing.T) {
	dup := errors.New("UNIQUE constraint failed: users.email")
	svc := NewAuthService(&failingUserRepo{fakeUserRepo: newFakeUserRepo(), createErr: dup})

	err := svc.Register(&model.User{Name: "taro", Email: "taro@example.com", Password: "secret"})
	if !errors.Is(err, dup) {
		t.Fatalf("repo のエラーがそのまま返るはず: %v", err)
	}
}

func TestLogin_EmptyCredentials(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	svc := NewAuthService(newFakeUserRepo())
	registerUser(t, svc)

	cases := map[string][2]string{
		"empty password": {"taro@example.com", ""},
		"empty email":    {"", "secret"},
		"both empty":     {"", ""},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := svc.Login(c[0], c[1]); !errors.Is(err, ErrInvalidCredentials) {
				t.Fatalf("ErrInvalidCredentials のはず: %v", err)
			}
		})
	}
}

func TestLogin_FailureReturnsNoToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	svc := NewAuthService(newFakeUserRepo())
	registerUser(t, svc)

	token, err := svc.Login("taro@example.com", "wrong")
	if err == nil || token != "" {
		t.Fatalf("失敗時はトークンが空のはず: token=%q err=%v", token, err)
	}
}

func TestLogin_TokenExpiresIn24Hours(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	svc := NewAuthService(newFakeUserRepo())
	registerUser(t, svc)

	tokenString, err := svc.Login("taro@example.com", "secret")
	if err != nil {
		t.Fatal(err)
	}
	token, err := jwt.Parse(tokenString, func(*jwt.Token) (any, error) { return []byte("test-secret"), nil })
	if err != nil {
		t.Fatal(err)
	}
	exp, err := token.Claims.GetExpirationTime()
	if err != nil || exp == nil {
		t.Fatalf("exp が取得できない: %v", err)
	}
	remain := time.Until(exp.Time)
	if remain < 23*time.Hour || remain > 24*time.Hour+time.Minute {
		t.Fatalf("有効期限が約24時間でない: %v", remain)
	}
}

func TestLogin_TokenRejectedWithWrongSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret")
	svc := NewAuthService(newFakeUserRepo())
	registerUser(t, svc)

	tokenString, err := svc.Login("taro@example.com", "secret")
	if err != nil {
		t.Fatal(err)
	}
	_, err = jwt.Parse(tokenString, func(*jwt.Token) (any, error) { return []byte("other-secret"), nil })
	if err == nil {
		t.Fatalf("別の秘密鍵では検証に失敗するはず")
	}
}
