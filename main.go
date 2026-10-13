package main

import (
	"errors"
	"log"
	"net/http"
	"os"

	// "practice_go/handler"
	handler "todo-api/handler"
	model "todo-api/model"
	repo "todo-api/repository"
	"todo-api/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// 必須の環境変数を取得する。未設定なら起動を中止する。
func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("環境変数 %s が設定されていません", key)
	}
	return v
}

func main() {
	// ローカル開発用。.env が無いのは許容する（本番は環境変数を直接渡す）。
	// ただし存在するのに読み込めない（書式エラーなど）場合は、原因を隠さず終了する。
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatalf(".env の読み込みに失敗しました: %v", err)
	}

	dsn := mustEnv("DATABASE_URL")
	mustEnv("JWT_SECRET") // 空のまま署名・検証されるのを防ぐため、起動時に存在確認する
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := gin.Default()
	r.GET("/hello", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello, World!")
	})

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	if err := db.AutoMigrate(&model.Task{}, &model.User{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	taskRepo := repo.NewTaskRepository(db)
	userRepo := repo.NewUserRepository(db)
	taskService := service.NewTaskService(taskRepo)
	authService := service.NewAuthService(userRepo)
	authHandler := handler.NewAuthHandler(authService)
	h := handler.NewHandler(taskService)

	r.POST("/auth/register", authHandler.Register)
	r.POST("/auth/login", authHandler.Login)

	tasks := r.Group("/tasks", handler.AthMiddleware())
	{
		tasks.POST("", h.CreateTask)
		tasks.GET("", h.GetAllTasks)
		tasks.GET("/:id", h.GetTasksById)
		tasks.PUT("/:id", h.UpdateTask)
		tasks.DELETE("", h.DeleteAllTasks)
		tasks.DELETE("/:id", h.DeleteTaskById)
	}

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
