package database

import (
		"log"
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
	
}


func Create(pin *message.Pin) int{
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


func GetLastCreatedTime(userIP string) (time.Time, error){
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

