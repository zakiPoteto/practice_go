package main

import (
	"fmt"
	"net/http"

	// "practice_go/handler"
	handler "todo-api/handler"
	model "todo-api/model"
	repo "todo-api/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite" // Sqlite driver based on CGO
	"gorm.io/gorm"
)

func main() {
	r := gin.Default()
	r.GET("/hello", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello, World!")
	})

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		fmt.Println("failed to connect database:", err)
		return
	}
	db.AutoMigrate(&model.Task{}, &model.User{})

	taskRepo := repo.NewTaskRepository(db)
	userRepo := repo.NewUserRepository(db)
	authHandler := handler.NewAuthHandler(userRepo)
	h := handler.NewHandler(taskRepo)

	r.POST("/auth/register", authHandler.Register)
	r.POST("/auth/login", authHandler.Login)

	tasks := r.Group("/tasks", handler.AthMiddleware())
	{
		tasks.POST("", h.CreateTask)
		tasks.GET("", h.GetAllTasks)
		tasks.GET("/:id", h.GetTasksById)
		tasks.DELETE("", h.DeleteAllTasks)
		tasks.DELETE("/:id", h.DeleteTaskById)
	}

	r.Run(":8080")

}
