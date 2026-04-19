package main

import (
	"fmt"
	"net/http"

	// "practice_go/handler"
	handler "todo-api/handler"
	task "todo-api/repository"

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
	db.AutoMigrate(&task.Task{})

	taskRepo := task.NewTaskRepository(db)
	h := handler.NewHandler(taskRepo)

	r.POST("/tasks", h.CreateTask)
	r.GET("/tasks", h.GetAllTasks)
	r.GET("/tasks/:id", h.GetTasksById)

	r.Run(":8080")

}
