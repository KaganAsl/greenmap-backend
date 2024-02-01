package server

 import (
    "encoding/json"
    "net/http"
		"log"
	

		"pawmap/message"
		"pawmap/database"
		"pawmap/utils"
	)			



func SubmitPinHandler(w http.ResponseWriter, r *http.Request) {
    // Decode JSON from request body
    var pin message.Pin

    err := json.NewDecoder(r.Body).Decode(&pin)
    if err != nil {
				log.Println("Invalid Json Format", err)
        http.Error(w, "Invalid JSON format", http.StatusConflict)
        return
    }

		ipAdr := utils.GetIPAddress(r)
		pin.UserIP = ipAdr
		
		if utils.CheckMessageData(&pin) == false {
				log.Println("Invalid Data", pin)
			  http.Error(w, "Fill all fields", http.StatusConflict)
        return
		}
		
		var locations []message.Loc
		result := database.DB.Where("lat = ? AND long = ?", pin.Location.Lat,
	pin.Location.Long).Find(&locations)

		if result.RowsAffected > int64(0) {
				log.Println("Location Must Be Unique", pin, result)
			  http.Error(w, "There is a diffrent pin at that location", http.StatusBadRequest)
        return
		}

		
		if pin.UserIP == "" || utils.IsRateLimited(pin.UserIP) {
			log.Println("Rate Limit || No User IP", pin)
			http.Error(w, "You need to wait before submitting new Pin", http.StatusInternalServerError)
			return
		}
		
		w.WriteHeader(http.StatusOK)

		if database.Create(&pin) == 1 {
			log.Println("Pin submitted successfully")
		} else {
			log.Println("Database could not save")
			http.Error(w, "Error saving pin to database", http.StatusInternalServerError)
			return
	}
}

func GetAllPinsHandler(w http.ResponseWriter, r *http.Request) {

	pins, err := database.GetAllPins()
	
	if err != nil {	
		http.Error(w, "Error Getting Values", http.StatusInternalServerError)
		return
	}
	
	/* Alternatively Can Be Done In This Way More easy
		
	func getPinJSON(pin Pin) ([]byte, error) {
    type PinWithoutUserIP struct {
        Pin
        UserIP string `json:"-"`
    }

    pinWithoutUserIP := PinWithoutUserIP{Pin: pin}
    return json.Marshal(pinWithoutUserIP)
	}
	
	*/
	

	for i := 0; i < len(pins); i++ {
		utils.GetPinModifier(&pins[i])
	}

	w.Header().Set("Content-Type", "application/json")
			
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(pins)
	if err != nil {
		log.Println("Error encoding pins to JSON", err)	
		http.Error(w, "Error encoding pins to JSON", http.StatusInternalServerError)
		return	
	}	
}




