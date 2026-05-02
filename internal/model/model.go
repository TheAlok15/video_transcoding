package model

import "time"

type Job struct {
	ID           string `gorm:"primaryKey"`
	RetryCount int `gorm:"default:0"`
	InputURL     string `gorm:"size:500"`
	Output360URL string `gorm:"size:500"`
	Output480URL string `gorm:"size:500"`
	Output720URL string `gorm:"size:500"`
	Status       string `gorm:"size:20"`
	ErrorMessage string `gorm:"size:1000"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
