package model

import "gorm.io/gorm"

type Task struct {
	gorm.Model
	UserID uint   `json:"user_id"`
	Title  string `json:"title" binding:"required"`
	Status string `json:"status" binding:"required"`
}
