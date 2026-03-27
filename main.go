package main

import (
	"fmt"
	"net/http"

	// "practice_go/handler"
	"practice_go/task"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite" // Sqlite driver based on CGO
	"gorm.io/gorm"
)
func main() {
	r := gin.Default()
	r.GET("/hello", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello, World!")
	})
	
	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		fmt.Println("failed to connect database:", err)
		return
	}
	db.AutoMigrate(&task.Task{})

	r.Run(":8080")

}
