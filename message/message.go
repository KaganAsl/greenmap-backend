package message


import (	    
	"gorm.io/gorm"
	"time"


)

type Pin struct {
		gorm.Model
		LocationID int 
    Location Loc `json:"location"`
		UserIP   string `json:"user_ip"` 
    Title    string `json:"title"`
    Text     string `json:"text"`
    PhotoID  string `json:"photo_id"`
    CreatedAt time.Time  `json:"-"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
		UpdatedAt time.Time `json:"-"`
		ID uint `json:"-"`
}

type Loc struct {
	ID uint `gorm:"primary_key"`
	Lat string `json:"lat"` 
	Long string `json:"long"` 

}

type PinWithoutUserIP struct {
    Pin
    UserIP string `json:"-"`
}


