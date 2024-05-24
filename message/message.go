package message

import (
	"time"

	"gorm.io/gorm"
)

type Pin struct {
	gorm.Model
	LocationID int
	Location   Loc            `json:"location"`
	UserIP     string         `json:"user_ip"`
	Title      string         `json:"title"`
	Text       string         `json:"text"`
	PhotoID    string         `json:"photo_id"`
	CreatedAt  time.Time      `json:"-"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	UpdatedAt  time.Time      `json:"-"`
	ID         uint           `json:"-"`
}

type Loc struct {
	ID        uint   `gorm:"primary_key"`
	Lat       string `json:"lat"`
	Long      string `json:"lng"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Look Server GetAllPinsHandler
type PinWithoutUserIP struct {
	Pin
	UserIP string `json:"-"`
}

type User struct {
	UserID    uint   `gorm:"primary_key"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Email     string `json:"email"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Session struct {
	gorm.Model
	UserID    uint      `json:"user_id"`
	StartedAt time.Time `json:"started_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Category struct {
	gorm.Model
	Type string `json:"type"`
}
