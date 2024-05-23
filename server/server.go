package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"pawmap/database"
	"pawmap/message"
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

	if !utils.CheckMessageData(&pin) {
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

	/*
		if pin.UserIP == "" || utils.IsRateLimited(pin.UserIP) {
			log.Println("Rate Limit || No User IP", pin)
			http.Error(w, "You need to wait before submitting new Pin", http.StatusInternalServerError)
			return
		}
	*/

	w.WriteHeader(http.StatusOK)

	if database.CreatePin(&pin) == 1 {
		log.Println("Pin submitted successfully")
		w.Header().Set("Content-Type", "application/json")
		// Return the pin as JSON
		err = json.NewEncoder(w).Encode(pin)
		if err != nil {
			log.Println("Error encoding pin to JSON", err)
			http.Error(w, "Error encoding pin to JSON", http.StatusInternalServerError)
			return
		}
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

func GetPinsByLocationHandler(w http.ResponseWriter, r *http.Request) {
	lat := r.URL.Query().Get("lat")
	long := r.URL.Query().Get("long")
	radius := r.URL.Query().Get("radius")
	radius_int, err := strconv.Atoi(radius)

	if err != nil {
		log.Println("Invalid Radius Pin Radius", radius, err)
		http.Error(w, "Error Getting Values", http.StatusInternalServerError)
		return
	}

	pins, err := database.GetPinsByLocation(lat, long, radius_int)

	if err != nil {
		http.Error(w, "Error Getting Values", http.StatusInternalServerError)
		return
	}

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

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var user message.User

	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		log.Println("Invalid Json Format", err)
		http.Error(w, "Invalid JSON format", http.StatusConflict)
		return
	}

	if !utils.CheckUserData(&user) {
		log.Println("Invalid Data", user)
		http.Error(w, "Fill all fields", http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusOK)

	if database.CreateUser(&user) == 1 {
		log.Println("User Created successfully")
		w.Header().Set("Content-Type", "application/json")
		// Return the username as JSON
		returnUser := message.User{Username: user.Username, UserID: user.UserID}
		err = json.NewEncoder(w).Encode(returnUser)

		if err != nil {
			log.Println("Error encoding user to JSON", err)
			http.Error(w, "Error encoding user to JSON", http.StatusInternalServerError)
			return
		}
	} else {
		log.Println("Database could not save, User already exists")
		http.Error(w, "Error saving User to database, User Already exists", http.StatusBadRequest)
		return
	}
}

func GetUserByIdHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	userID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Invalid User ID, Converting Failed!", id, err)
		http.Error(w, "Error Getting Values", http.StatusInternalServerError)
		return
	}
	user, err := database.GetUserByID(uint(userID))
	if err != nil {
		http.Error(w, "Error Getting Values", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		log.Println("Error encoding user to JSON", err)
		http.Error(w, "Error encoding user to JSON", http.StatusInternalServerError)
		return
	}
}

func GetUserByUsernameHandler(w http.ResponseWriter, r *http.Request) {

	username := r.URL.Query().Get("username")
	user, err := database.GetUserByUsername(&username)
	if err != nil {
		http.Error(w, "Error Getting Values", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		log.Println("Error encoding user to JSON", err)
		http.Error(w, "Error encoding user to JSON", http.StatusInternalServerError)
		return
	}
}

func GetUserByMailHandler(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")

	user, err := database.GetUserByMail(&email)

	if err != nil {
		http.Error(w, "Error Getting Values", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		log.Println("Error encoding user to JSON", err)
		http.Error(w, "Error encoding user to JSON", http.StatusInternalServerError)
		return
	}
}

func GetUsersHandler(w http.ResponseWriter, r *http.Request) {

	users, err := database.GetUsers()

	if err != nil {
		http.Error(w, "Error Getting Values", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(users)
	if err != nil {
		log.Println("Error encoding users to JSON", err)
		http.Error(w, "Error encoding users to JSON", http.StatusInternalServerError)
		return
	}
}

func CreateSessionHandler(w http.ResponseWriter, r *http.Request) {
	var session message.Session
	err := json.NewDecoder(r.Body).Decode(&session)
	if err != nil {
		log.Println("Invalid Json Format", err)
		http.Error(w, "Invalid JSON format", http.StatusConflict)
		return
	}

	if err != nil {
		log.Println("Invalid Json Format", err)
		http.Error(w, "Invalid JSON format", http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusOK)

	// Set Session Time

	session.StartedAt = time.Now()
	session.ExpiresAt = time.Now().Add(time.Hour * 2)

	if database.CreateSession(&session) == 1 {
		// Create Token For Session
		user, error := database.GetUserByID(session.UserID)
		if error != nil {
			log.Println("Could not get user, userID", error)
			http.Error(w, "Error getting user", http.StatusNotFound)
			return
		}
		token := utils.Base64EncodeString(utils.CreateToken(user.Username, session.StartedAt.String(), session.ExpiresAt.String()))
		log.Println("Session Created successfully")
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(struct{ Token string }{Token: token})
		if err != nil {
			log.Println("Error encoding session to JSON", err)
			http.Error(w, "Error encoding session to JSON", http.StatusInternalServerError)
			return
		}
	} else {
		log.Println("Database could not save")
		http.Error(w, "Error saving session to database, There is a active session", http.StatusInternalServerError)
		return
	}
}

func CheckSessionHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		log.Println("No Token")
		http.Error(w, "No Token", http.StatusBadRequest)
		return
	}

	decodedToken, err := utils.Base64DecodeString(token)
	if err != nil {
		log.Println("Invalid Token")
		http.Error(w, "Invalid Token", http.StatusUnauthorized)
		return
	}

	username, _, endTime := utils.ValidateToken(decodedToken)

	user, error := database.GetUserByUsername(&username)
	if error != nil {
		log.Println("Could not get user, username: ", username, error)
		http.Error(w, "Error getting user", http.StatusNotFound)
		return
	}

	session, error := database.GetSessionByUserID(user.UserID)
	if error != nil {
		log.Println("Could not get session, userID: ", session.UserID, error)
		http.Error(w, "Error getting session", http.StatusNotFound)
		return
	}

	if time.Now().String() > endTime {
		log.Println("Token Expired", username)
		http.Error(w, "Token Expired", http.StatusUnauthorized)
		userID, err := strconv.Atoi(username)
		if err != nil {
			log.Println("Invalid User ID, Converting Failed!", username, err)
			http.Error(w, "Error Getting Values, Session expired", http.StatusNotFound)
			return
		}
		database.DeleteSession(uint(userID))
		return
	}

	log.Println("Token Valid, Username: ", username)
	w.WriteHeader(http.StatusOK)
}

func GetSessionHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID uint `json:"user_id"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Println("Invalid Request Body")
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	session, err := database.GetSessionByUserID(req.UserID)
	if err != nil {
		log.Println("Invalid Session ID, Cannot Get Session", session.UserID, err)
		http.Error(w, "Error Getting Values", http.StatusNotFound)
		json.NewEncoder(w).Encode(struct{ Token string }{Token: ""})
		return
	}

	user, err := database.GetUserByID(req.UserID)
	if err != nil {
		log.Println("Invalid User ID, Cannot Get Session", user.Username, err)
		http.Error(w, "Error Getting Values", http.StatusNotFound)
		json.NewEncoder(w).Encode(struct{ Token string }{Token: ""})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	token := utils.Base64EncodeString(utils.CreateToken(user.Username, session.StartedAt.String(), session.ExpiresAt.String()))
	log.Println("Session Created successfully")
	err = json.NewEncoder(w).Encode(struct{ Token string }{Token: token})
	if err != nil {
		log.Println("Error encoding session to JSON", err)
		http.Error(w, "Error encoding session to JSON", http.StatusInternalServerError)
		return
	}
}

func DeleteSessionHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		log.Println("No Token")
		http.Error(w, "No Token", http.StatusBadRequest)
		return
	}

	decodedToken, err := utils.Base64DecodeString(token)
	if err != nil {
		log.Println("Invalid Token")
		http.Error(w, "Invalid Token", http.StatusUnauthorized)
		return
	}

	userName, _, _ := utils.ValidateToken(decodedToken)

	user, err := database.GetUserByUsername(&userName)
	userID := user.UserID

	if err != nil {
		log.Println("Invalid User ID, Converting Failed!", userID, err)
		http.Error(w, "Error Getting Values", http.StatusBadRequest)
		return
	}

	if database.DeleteSession(userID) == 1 {
		log.Println("Session Deleted successfully")
		w.WriteHeader(http.StatusOK)
	} else {
		log.Println("Database could not delete session, There is no session for that user")
		http.Error(w, "Error deleting session from database", http.StatusNotFound)
		return
	}
}
