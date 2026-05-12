package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required" gorm:"uniqueIndex"`
	Password string `json:"-" binding:"required"`
}
