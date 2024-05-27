package message

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Pin struct {
	BaseModel
	LocationID uint
	Location   Loc    `gorm:"foreignKey:LocationID" json:"location"`
	CategoryID uint   `json:"category_id"`
	UserIP     string `json:"user_ip"`
	Title      string `json:"title"`
	Text       string `json:"text"`
	PhotoID    uint   `json:"photo_id"`
}

type Loc struct {
	BaseModel
	Lat  string `json:"lat"`
	Long string `json:"lng"`
}

type PinWithoutUserIP struct {
	Pin
	UserIP string `json:"-"`
}

type User struct {
	BaseModel
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}
type Session struct {
	BaseModel
	UserID    uint      `json:"user_id"`
	StartedAt time.Time `json:"started_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Category struct {
	BaseModel
	Type string `json:"type"`
}

type File struct {
	BaseModel
	Name string `gorm:"unique" json:"name"`
	Link string `json:"link"`
}

// Outputs

type PinOutput struct {
	ID        uint           `json:"id"`
	Location  LocOutput      `json:"location"`
	Category  CategoryOutput `json:"category"`
	Title     string         `json:"title"`
	Text      string         `json:"text"`
	Photo     FileOutput     `json:"photo"`
	CreatedAt time.Time      `json:"created_at"`
}

type LocOutput struct {
	Lat  string `json:"lat"`
	Long string `json:"lng"`
}

type UserOutput struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type SessionOutput struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	StartedAt time.Time `json:"started_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

type CategoryOutput struct {
	ID   uint   `json:"id"`
	Type string `json:"type"`
}

type FileOutput struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Link string `json:"link"`
}
