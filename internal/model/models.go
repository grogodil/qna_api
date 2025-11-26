package model

import (
	"time"
)

type Question struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Text      string    `gorm:"not null" json:"text"`
	CreatedAt time.Time `json:"created_at"`
	Answers   []Answer  `gorm:"constraint:OnDelete:CASCADE;" json:"answers,omitempty"`
}

type Answer struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	QuestionID uint      `gorm:"index;not null" json:"question_id"`
	UserID     string    `gorm:"type:uuid;not null" json:"user_id"`
	Text       string    `gorm:"not null" json:"text"`
	CreatedAt  time.Time `json:"created_at"`
}
