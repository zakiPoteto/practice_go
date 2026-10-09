package handler

import (
	"errors"
	"net/http"
	"strconv"
	model "todo-api/model"
	"todo-api/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	taskService *service.TaskService
}

func NewHandler(s *service.TaskService) *Handler {
	return &Handler{
		taskService: s,
	}
}

// Service が返すエラーを HTTP ステータスに変換する
func respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrTaskNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
	case errors.Is(err, service.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func taskResponse(t model.Task) gin.H {
	return gin.H{
		"id":     t.ID,
		"title":  t.Title,
		"status": t.Status,
	}
}

func (h *Handler) CreateTask(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	task := new(model.Task)
	if err := c.ShouldBindJSON(task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// リクエストボディの user_id は信用せず、トークン由来の値で上書きする
	task.UserID = userID

	if err := h.taskService.CreateTask(task); err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, taskResponse(*task))
}

func (h *Handler) GetAllTasks(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	tasks, err := h.taskService.GetAllTasks(userID)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	response := make([]gin.H, 0, len(tasks))
	for _, t := range tasks {
		response = append(response, taskResponse(t))
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) GetTasksById(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	task, err := h.taskService.GetTaskByID(id, userID)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, taskResponse(task))
}

func (h *Handler) UpdateTask(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	var input struct {
		Title  string `json:"title" binding:"required"`
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	task, err := h.taskService.UpdateTask(id, userID, input.Title, input.Status)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, taskResponse(task))
}

func (h *Handler) DeleteAllTasks(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	if err := h.taskService.DeleteAllTasks(userID); err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "all tasks deleted"})
}

func (h *Handler) DeleteTaskById(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id format"})
		return
	}
	if err := h.taskService.DeleteTask(id, userID); err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "task deleted"})
}
