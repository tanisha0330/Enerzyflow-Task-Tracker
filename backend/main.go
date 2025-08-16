package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers" // 1. Import the handlers package
	"github.com/gorilla/mux"
	_ "github.com/lib/pq" // The underscore means we need this package for its side effects (registering the driver)
)

func main() {
	// 1. LOAD ENVIRONMENT VARIABLES
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// 2. CONNECT TO DATABASE
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Check if the connection is successful
	err = db.Ping()
	if err != nil {
		log.Fatal("Database connection test failed:", err)
	}
	fmt.Println("Successfully connected to the database!")

	// 3. SET UP ROUTER
	r := mux.NewRouter()

	// Public routes (no authentication needed)
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("API is running"))
	}).Methods("GET")
	r.HandleFunc("/auth/register", RegisterHandler(db)).Methods("POST")
	r.HandleFunc("/auth/login", LoginHandler(db)).Methods("POST")

	// Protected routes (require JWT)
	protectedRoutes := r.PathPrefix("/tasks").Subrouter()
	protectedRoutes.Use(JWTMiddleware)

	// Wire up the task handlers
	protectedRoutes.HandleFunc("", GetTasksHandler(db)).Methods("GET")
	protectedRoutes.HandleFunc("", CreateTaskHandler(db)).Methods("POST")
	protectedRoutes.HandleFunc("/{id}", UpdateTaskHandler(db)).Methods("PUT")
	protectedRoutes.HandleFunc("/{id}", DeleteTaskHandler(db)).Methods("DELETE")

	// 4. CONFIGURE CORS
	// This tells the backend to allow requests from your frontend server
	allowedOrigins := handlers.AllowedOrigins([]string{"http://localhost:3000"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})
	allowedHeaders := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})

	// 5. START THE SERVER WITH CORS MIDDLEWARE
	fmt.Println("Starting server on port 8080...")
	// We wrap our main router 'r' with the CORS handler
	log.Fatal(http.ListenAndServe(":8080", handlers.CORS(allowedOrigins, allowedMethods, allowedHeaders)(r)))
}
