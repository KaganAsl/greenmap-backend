package database

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"strconv"
	"time"

	"pawmap/message"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// initDB initializes the SQLite database
func InitDB() {

	var err error
	DB, err = gorm.Open(sqlite.Open("pins.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate the schema
	DB.AutoMigrate(&message.Pin{})
	DB.AutoMigrate(&message.User{})
	DB.AutoMigrate(&message.Session{})
	DB.AutoMigrate(&message.Category{})

	// Init Category
	InitCategory()
}

func InitCategory() {
	jsonFile, err := os.Open("./database/category.json")
	if err != nil {
		log.Fatal("Failed to open categories.json:", err)
	}
	defer jsonFile.Close()

	byteValue, _ := os.ReadFile("./database/category.json")

	type CategoryList struct {
		Categories []string `json:"Categories"`
	}

	var categories CategoryList
	json.Unmarshal(byteValue, &categories)

	for _, category := range categories.Categories {
		var cat message.Category
		result := DB.Where("type = ?", category).First(&cat)
		if result.Error == nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				cat.Type = category
				log.Println("Category not found, Adding", category)
				DB.Create(&cat)
			} else {
				log.Println("Category Init, Passing", category)
			}
		}
	}

}

// Pin functions

func CreatePin(pin *message.Pin) int {
	// Save the pin to the database
	result := DB.Create(pin)
	if result.Error != nil {
		return 0
	}

	return 1

}

func GetAllPins() ([]message.Pin, error) {

	var pins []message.Pin

	result := DB.Find(&pins)

	if result.Error != nil {
		log.Println("Could not get all pins", result.Error)
		return nil, result.Error
	}

	for i := 0; i < len(pins); i++ {
		var location message.Loc
		locResult := DB.Where("id = ?", pins[i].LocationID).First(&location)
		if locResult.Error != nil {
			log.Println("Could not get Location of pin", result.Error, pins[i])
			return nil, result.Error

		}
		pins[i].Location = location
	}

	return pins, nil
}

func GetLastCreatedTime(userIP string) (time.Time, error) {
	// Save the pin to the database
	var latestPin message.Pin
	err := DB.Where("user_ip = ?", userIP).
		Order("created_at desc").
		First(&latestPin).Error

	if err != nil {
		return time.Now(), err
	}

	return latestPin.CreatedAt, nil

}

func GetPinsByLocation(lat, long string, radius int) ([]message.Pin, error) {
	var latMin, latMax, longMin, longMax float64

	latFloat, err := strconv.ParseFloat(lat, 64)
	if err != nil {
		return nil, err
	}
	longFloat, err := strconv.ParseFloat(long, 64)
	if err != nil {
		return nil, err
	}

	latMin = latFloat - float64(radius)
	latMax = latFloat + float64(radius)
	longMin = longFloat - float64(radius)
	longMax = longFloat + float64(radius)

	var pins []message.Pin

	result := DB.Where("lat >= ? AND lat <= ? AND long >= ? AND long <= ?", latMin, latMax, longMin, longMax).Find(&pins)
	if result.Error != nil {
		log.Println("Could not get pins by location", result.Error)
		return nil, result.Error
	}
	return pins, nil
}

// User functions

func CreateUser(user *message.User) int {
	// Save the user to the database
	isUserAlreadyCreated, err := GetUserByUsername(&user.Username)
	if err == nil {
		if isUserAlreadyCreated.Username != "" {
			return 0
		}
	}

	result := DB.Create(user)
	if result.Error != nil {
		log.Println("Could not add user", result.Error)
		return 0
	}
	return 1
}

func GetUserByID(userID uint) (message.User, error) {
	var user message.User
	result := DB.Where("user_id = ?", userID).First(&user)
	if result.Error != nil {
		log.Println("Could not get user", result.Error)
		return message.User{}, result.Error
	}
	return user, nil
}

func GetUserByUsername(username *string) (message.User, error) {
	var user message.User
	result := DB.Where("username = ?", username).First(&user)
	if result.Error != nil {
		log.Println("Could not get user", result.Error)
		return message.User{}, result.Error
	}
	return user, nil
}

func GetUserByMail(mail *string) (message.User, error) {
	var user message.User
	result := DB.Where("email = ?", mail).First(&user)
	if result.Error != nil {
		log.Println("Could not get user", result.Error)
		return message.User{}, result.Error
	}
	return user, nil
}

func GetUsers() ([]message.User, error) {
	var users []message.User
	result := DB.Find(&users)
	if result.Error != nil {
		log.Println("Could not get all users", result.Error)
		return nil, result.Error
	}
	return users, nil
}

func CreateSession(session *message.Session) int {
	currentSession, error := GetSessionByUserID(session.UserID)
	println(currentSession.ID, currentSession.UserID)
	if error == nil {
		if currentSession.ID != 0 {
			return 0
		}
	}
	result := DB.Create(session)
	if result.Error != nil {
		log.Println("Could not add session", result.Error)
		return 0
	}
	return 1
}

func DeleteSession(userID uint) int {
	result := DB.Where("user_id = ?", userID).Delete(&message.Session{})
	if result.Error != nil || result.RowsAffected == 0 {
		log.Println("Could not delete session", result.Error)
		return 0
	}
	return 1
}

func GetSessionBySessionID(sessionID uint) (message.Session, error) {
	var session message.Session
	result := DB.Where("id = ?", sessionID).First(&session)
	if result.Error != nil {
		log.Println("Could not get session", result.Error)
		return message.Session{}, result.Error
	}
	return session, nil
}

func GetSessionByUserID(userID uint) (message.Session, error) {
	var session message.Session
	result := DB.Where("user_id = ?", userID).First(&session)
	if result.Error != nil {
		log.Println("Could not get session", result.Error)
		return message.Session{}, result.Error
	}
	return session, nil
}

func GetCategories() ([]message.Category, error) {
	var categories []message.Category
	result := DB.Find(&categories)
	if result.Error != nil {
		log.Println("Could not get categories", result.Error)
		return nil, result.Error
	}
	return categories, nil
}
