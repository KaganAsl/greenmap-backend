package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"pawmap/database"
	"pawmap/server"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func main() {
	// Initialize the database
	database.InitDB()

	// Init log folder
	if _, err := os.Stat("./logs"); os.IsNotExist(err) {
		os.Mkdir("./logs", 0755)
	}

	// Init Uploads Folder
	if _, err := os.Stat("./uploads"); os.IsNotExist(err) {
		os.Mkdir("./uploads", 0755)
	}

	// Init Logger
	f, err := os.OpenFile("./logs/api.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	wrt := io.MultiWriter(os.Stdout, f)
	log.SetOutput(wrt)

	// Initialize the router
	r := mux.NewRouter()
	v1 := r.PathPrefix("/api/v1").Subrouter()

	// Pins
	v1.HandleFunc("/pin/submitPin", server.SubmitPinHandler).Methods("POST")
	v1.HandleFunc("/pin/getAllPins", server.GetAllPinsHandler).Methods("GET")
	v1.HandleFunc("/pin/getPinsByLocation", server.GetPinsByLocationHandler).Methods("GET")

	// Users
	v1.HandleFunc("/user/createUser", server.CreateUserHandler).Methods("POST")
	v1.HandleFunc("/user/getUserByUsername", server.GetUserByUsernameHandler).Methods("GET")
	v1.HandleFunc("/user/getUserByEmail", server.GetUserByMailHandler).Methods("GET")
	v1.HandleFunc("/user/updateUser", server.UpdateUserHandler).Methods("POST")
	v1.HandleFunc("/user/deleteUser", server.DeleteUserHandler).Methods("DELETE")

	// Sessions
	v1.HandleFunc("/session/createSession", server.CreateSessionHandler).Methods("POST")
	v1.HandleFunc("/session/getSession", server.GetSessionHandler).Methods("GET")
	v1.HandleFunc("/session/checkSession", server.CheckSessionHandler).Methods("GET")
	v1.HandleFunc("/session/deleteSession", server.DeleteSessionHandler).Methods("DELETE")

	// Login/Logout
	v1.HandleFunc("/login", server.LoginHandler).Methods("POST")
	v1.HandleFunc("/logout", server.LogoutHandler).Methods("POST")

	// Categories
	v1.HandleFunc("/category/getAllCategories", server.GetAllCategoriesHandler).Methods("GET")

	// Images
	v1.HandleFunc("/images/uploadFile", server.UploadFileHandler).Methods("POST") // Very dangerous functions TODO: Make it only for images
	v1.HandleFunc("/images/getFile", server.GetFileByIDHandler).Methods("GET")    // Very dangerous functions TODO: Make it only for images

	//allowedOrigin := os.Getenv("ALLOWED_ORIGIN")

	// CORS middleware
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := corsMiddleware.Handler(r)

	// Start the server
	fmt.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe("localhost:8080", handler))
}
