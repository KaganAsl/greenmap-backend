package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"pawmap/database"
	"pawmap/message"
	"pawmap/utils"
)

func SubmitPinHandler(w http.ResponseWriter, r *http.Request) {

	authToken := r.Header.Get("Authorization")
	log.Println(authToken)
	if authToken == "" || authToken == "undefined" {
		log.Println("No Token")
		http.Error(w, "No Token OR User Is not Authenticated", http.StatusBadRequest)
		return
	}

	// Decode JSON from request body
	// In 'Data' form value, we have the JSON string
	data := r.FormValue("Data")

	var pin message.Pin
	//err = json.NewDecoder(r.Body).Decode(&pin)
	err := json.Unmarshal([]byte(data), &pin)
	if err != nil {
		log.Println("Invalid Json Format", err)
		http.Error(w, "Invalid JSON format", http.StatusConflict)
		return
	}

	ipAdr := utils.GetIPAddress(r)
	pin.UserIP = ipAdr

	category, err := database.GetCategoryByID(pin.CategoryID)
	if err != nil {
		log.Println("Invalid Category ID", pin.CategoryID, err)
		http.Error(w, "Error Getting Values", http.StatusInternalServerError)
		return
	}

	pin.CategoryID = category.ID

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

	_, _, err = r.FormFile("File")

	var image message.File
	if err == nil {
		filename, err := utils.UploadFile(r)
		if err != nil {
			log.Println("Error uploading the file", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		status := database.CreateFile(filename)
		if status != 1 {
			http.Error(w, "Error saving file to database", http.StatusInternalServerError)
			return
		}
		image, err = database.GetFileByName(filename)
		if err != nil {
			http.Error(w, "Error Getting Values", http.StatusInternalServerError)
			return
		}
	}

	pin.PhotoID = image.ID
	//pin.Photo = image

	if database.CreatePin(&pin) == 1 {
		log.Println("Pin submitted successfully")
		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusOK)
		// Return the pin as JSON

		err = json.NewEncoder(w).Encode(&message.PinOutput{
			ID:        pin.ID,
			Location:  message.LocOutput{Lat: pin.Location.Lat, Long: pin.Location.Long},
			Category:  message.CategoryOutput{ID: category.ID, Type: category.Type},
			Title:     pin.Title,
			Text:      pin.Text,
			Photo:     message.FileOutput{ID: image.ID, Name: image.Name, Link: image.Link},
			CreatedAt: pin.CreatedAt,
		})
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

	/*
		for i := 0; i < len(pins); i++ {
			utils.GetPinModifier(&pins[i])
		}
	*/

	pinsOutput := make([]message.PinOutput, len(pins))
	for i, pin := range pins {
		category, err := database.GetCategoryByID(pin.CategoryID)
		if err != nil {
			log.Println("Invalid Category ID", pin.CategoryID, err)
			http.Error(w, "Error Getting Category Values ", http.StatusInternalServerError)
			return
		}

		image, err := database.GetFileByID(pin.PhotoID)
		if err != nil && pin.PhotoID != 0 {
			log.Println("Invalid Image ID", pin.PhotoID, err)
			http.Error(w, "Error Getting File Values", http.StatusInternalServerError)
			return
		}

		messagePinOutput := message.PinOutput{
			ID:       pin.ID,
			Location: message.LocOutput{Lat: pin.Location.Lat, Long: pin.Location.Long},
			Category: message.CategoryOutput{ID: category.ID, Type: category.Type},
			Title:    pin.Title,
			Text:     pin.Text,
			//Photo:     message.FileOutput{ID: image.ID, Name: image.Name, Link: image.Link},
			CreatedAt: pin.CreatedAt,
		}

		if pin.PhotoID != 0 {
			messagePinOutput.Photo = message.FileOutput{ID: image.ID, Name: image.Name, Link: image.Link}
		}

		pinsOutput[i] = messagePinOutput
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(pinsOutput)
	if err != nil {
		log.Println("Error encoding pins to JSON", err)
		http.Error(w, "Error encoding pins to JSON", http.StatusInternalServerError)
		return
	}
}

// Filtering

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

func GetPinsByCategoryHandler(w http.ResponseWriter, r *http.Request) {
}

// Users

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

	if database.CreateUser(&user) == 1 {
		log.Println("User Created successfully")
		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusOK)
		// Return the username as JSON
		returnUser := message.User{Username: user.Username}
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

func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
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

	if database.UpdateUser(&user) == 1 {
		log.Println("User Updated successfully")
		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusOK)
		// Return the username as JSON
		returnUser := message.User{Username: user.Username}
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

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	type User struct {
		Username string `json:"username"`
	}

	var userInput User

	err := json.NewDecoder(r.Body).Decode(&userInput)

	if err != nil {
		log.Println("Invalid Json Format", err)
		http.Error(w, "Invalid JSON format", http.StatusConflict)
		return
	}

	user, err := database.GetUserByUsername(&userInput.Username)

	if err != nil {
		http.Error(w, "Error Getting Values", http.StatusNotFound)
		return
	}

	if database.DeleteUser(user.ID) == 1 {
		log.Println("User Deleted successfully")
		w.WriteHeader(http.StatusOK)
	} else {
		log.Println("Database could not delete user, User does not exist")
		http.Error(w, "Error deleting user from database", http.StatusNotFound)
		return
	}
}

func GetUserByIdHandler(w http.ResponseWriter, r http.Request) {
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

	userOutput := message.UserOutput{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}

	err = json.NewEncoder(w).Encode(&userOutput)
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

// Login

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var user message.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		log.Println("Invalid Json Format", err)
		http.Error(w, "Invalid JSON format", http.StatusConflict)
		return
	}

	dbUser, err := database.GetUserByUsername(&user.Username)
	if err != nil {
		log.Println("Could not get user, username: ", user.Username, err)
		http.Error(w, "Error getting user", http.StatusNotFound)
		return
	}

	if user.Password != dbUser.Password {
		log.Println("Invalid Password", user.Username)
		http.Error(w, "Invalid Credientals", http.StatusUnauthorized)
		return
	}
	token := utils.CreateSession(dbUser.ID)
	if token == "" {
		log.Println("Token could not be created, Session already active")
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		return
	}

	// TODO: Improve this, It can be impelemented in a better way

	sessionRes, err := database.GetSessionByUserID(dbUser.ID)
	if err != nil {
		http.Error(w, "Cannot Get Session", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")

	ck := http.Cookie{
		Name:  "GreenMap_AUTH",
		Value: token,
		Path:  "/",
		//HttpOnly: true,
		// Secure:   true, // Make sure to use this only if you have HTTPS enabled
		Expires: sessionRes.ExpiresAt,
		MaxAge:  14400,
	}

	http.SetCookie(w, &ck)

	w.WriteHeader(http.StatusOK)

	log.Println("Session Created successfully")

	type UserResponse struct {
		Username string `json:"username"`
		UserID   uint   `json:"user_id"`
	}

	err = json.NewEncoder(w).Encode(&UserResponse{Username: dbUser.Username, UserID: dbUser.ID})
	if err != nil {
		log.Println("Error encoding user to JSON", err)
		http.Error(w, "Error encoding user to JSON", http.StatusInternalServerError)
		return
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
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
	userID := user.ID

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

// Session

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

	token := utils.CreateSession(session.UserID)
	if token == "" {
		log.Println("Token could not be created")
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		return
	}

	// TODO: Improve this, It can be impelemented in a better way

	sessionRes, err := database.GetSessionByUserID(session.UserID)
	if err != nil {
		http.Error(w, "Cannot Get Session", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")

	ck := http.Cookie{
		Name:  "GreenMap_AUTH",
		Value: token,
		Path:  "/",
		//HttpOnly: true,
		// Secure:   true, // Make sure to use this only if you have HTTPS enabled
		Expires: sessionRes.ExpiresAt,
		MaxAge:  14400,
	}

	http.SetCookie(w, &ck)

	w.WriteHeader(http.StatusOK)

	log.Println("Session Created successfully")
	err = json.NewEncoder(w).Encode(struct{ Token string }{Token: token})
	if err != nil {
		log.Println("Error encoding session to JSON", err)
		http.Error(w, "Error encoding session to JSON", http.StatusInternalServerError)
		return
	}

	session.StartedAt = time.Now()
	session.ExpiresAt = time.Now().Add(time.Hour * 2)

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

	session, error := database.GetSessionByUserID(user.ID)
	if error != nil {
		log.Println("Could not get session, userID: ", session.UserID, error)
		http.Error(w, "Error getting session", http.StatusNotFound)
		return
	}

	parsedEndTime, error := strconv.ParseInt(endTime, 10, 64)
	if error != nil {
		log.Println("Could not parse end time", endTime, error)
		http.Error(w, "Error parsing end time", http.StatusInternalServerError)
		return
	}

	if time.Now().After(time.Unix(parsedEndTime, 0)) {
		log.Println("Token Expired", username)
		http.Error(w, "Token Expired", http.StatusUnauthorized)
		/*
			userID, err := strconv.Atoi(username)
			if err != nil {
				log.Println("Invalid User ID, Converting Failed!", username, err)
				http.Error(w, "Error Getting Values, Session expired", http.StatusNotFound)
				return
			}
		*/
		database.DeleteSession(uint(user.ID))
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

	startedAt := strconv.FormatInt(session.StartedAt.Unix(), 10)
	expiresAt := strconv.FormatInt(session.ExpiresAt.Unix(), 10)
	token := utils.Base64EncodeString(utils.CreateToken(user.Username, startedAt, expiresAt))
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
	userID := user.ID

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

// Categories

func GetAllCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	categories, err := database.GetCategories()

	if err != nil {
		http.Error(w, "Error Getting Category Values", http.StatusInternalServerError)
		//log.Println("Error Getting Category Values", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)

	type CategoryResponse struct {
		ID   uint   `json:"id"`
		Type string `json:"type"`
	}

	var categoriesResponse []CategoryResponse
	for _, category := range categories {
		categoriesResponse = append(categoriesResponse, CategoryResponse{ID: category.ID, Type: category.Type})
	}

	response := map[string][]CategoryResponse{
		"Categories": categoriesResponse,
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println("Error encoding categories to JSON", err)
		http.Error(w, "Error encoding categories to JSON", http.StatusInternalServerError)
		return
	}

}

// File Stuff

func UploadFileHandler(w http.ResponseWriter, r *http.Request) {
	filename, err := utils.UploadFile(r)
	if err != nil {
		log.Println("Error uploading the file", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	println(filename)
	status := database.CreateFile(filename)
	if status != 1 {
		http.Error(w, "Error saving file to database", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	type FileResponse struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}

	result, err := database.GetFileByName(filename)
	if err != nil {
		http.Error(w, "Error Getting Values", http.StatusInternalServerError)
		return
	}
	response := FileResponse{ID: result.ID, Name: result.Name}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Println("Error encoding file to JSON", err)
		http.Error(w, "Error encoding file to JSON", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func GetFileByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("ID")
	fileID, err := strconv.Atoi(id)
	if err != nil {
		log.Println("Invalid File ID, Converting Failed!", id, err)
		http.Error(w, "Error Getting Values", http.StatusInternalServerError)
		return
	}

	file, err := database.GetFileByID(uint(fileID))
	if err != nil {
		http.Error(w, "Error Getting Values", http.StatusNotFound)
		return
	}

	// Assuming that the filename is the file ID with a .jpg extension
	// and the files are stored in the uploads directory
	filename := fmt.Sprintf("./uploads/%s", file.Name)

	// Send the image file
	http.ServeFile(w, r, filename)
}
