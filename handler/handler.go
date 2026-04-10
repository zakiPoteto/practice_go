package handler

import (
	"net/http"
	task "todo-api/repository"

	"github.com/gin-gonic/gin"
)

type tmp struct {
	ID     int    `json:"-"`
	Title  string `json:"title" binding:"required"`
	Status string `json:"status" binding:"required"`
}
type Handler struct {
	taskRepo *task.TaskRepository
}

func NewHandler(repo *task.TaskRepository) *Handler {

	return &Handler{
		taskRepo: repo,
	}
}
func (h *Handler) CreateTask(c *gin.Context) {
	// Implementation for creating a task
	task := new(task.Task)
	if err := c.Bind(task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.taskRepo.Create(task); err != nil { // データベースにユーザーを作成
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}) // 作成エラー時に500を返す
		return
	}

	c.JSON(http.StatusCreated, task) // 成功時に201と作成されたユーザーを返す
}
func (h *Handler) GetAllTasks(c *gin.Context) {
	tasks, err := h.taskRepo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tasks)
}
