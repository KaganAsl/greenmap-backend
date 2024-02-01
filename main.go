package main

import (
    "fmt"
    "log"
    "net/http"
		"io"
		"os"

		"pawmap/server"
		"pawmap/database"	

    "github.com/gorilla/mux"
		"github.com/rs/cors"
)


func main() {
		// Initialize the database
		database.InitDB()

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

    // Define the API endpoint for submitting pins
    r.HandleFunc("/submitPin", server.SubmitPinHandler).Methods("POST")
	r.HandleFunc("/getAllPins", server.GetAllPinsHandler).Methods("GET")

		//allowedOrigin := os.Getenv("ALLOWED_ORIGIN")

		// CORS middleware
		corsMiddleware := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		})

		handler := corsMiddleware.Handler(r)

    // Start the server
    fmt.Println("Server listening on :8080")
    log.Fatal(http.ListenAndServe(":8080", handler))
}


