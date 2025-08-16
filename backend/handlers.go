package main

import (
	"database/sql"
	"encoding/json"
	"log" // Import the log package
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
	"github.com/lib/pq" // Import the postgres driver
)

// We need a global variable for the JWT secret key.
// IMPORTANT: In a real application, load this from an environment variable!
var jwtKey = []byte("my_super_secret_key")

// Claims struct for our JWT
type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

// --- AUTHENTICATION HANDLERS ---

// RegisterHandler handles new user registration.
func RegisterHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds Credentials
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		hashedPassword, err := hashPassword(creds.Password)
		if err != nil {
			log.Printf("Error hashing password: %v", err) // Added logging
			respondWithError(w, http.StatusInternalServerError, "Could not process request")
			return
		}

		var userID int
		err = db.QueryRow("INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id",
			creds.Email, hashedPassword).Scan(&userID)

		// --- IMPROVED ERROR HANDLING ---
		if err != nil {
			// Check if the error is a unique constraint violation (duplicate email)
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
				respondWithError(w, http.StatusConflict, "Email address already in use")
			} else {
				// Log the actual, more general error for debugging
				log.Printf("Error creating user: %v", err)
				respondWithError(w, http.StatusInternalServerError, "Could not create user")
			}
			return
		}
		// --- END OF IMPROVEMENT ---

		respondWithJSON(w, http.StatusCreated, map[string]interface{}{"message": "User created successfully", "userID": userID})
	}
}

// LoginHandler handles user authentication and returns a JWT.
func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds Credentials
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		var user User
		var storedHash string

		err := db.QueryRow("SELECT id, email, password_hash FROM users WHERE email = $1", creds.Email).Scan(&user.ID, &user.Email, &storedHash)
		if err != nil {
			if err == sql.ErrNoRows {
				respondWithError(w, http.StatusUnauthorized, "Invalid email or password")
			} else {
				log.Printf("Database error during login: %v", err) // Added logging
				respondWithError(w, http.StatusInternalServerError, "Database error")
			}
			return
		}

		if !checkPasswordHash(creds.Password, storedHash) {
			respondWithError(w, http.StatusUnauthorized, "Invalid email or password")
			return
		}

		expirationTime := time.Now().Add(24 * time.Hour)
		claims := &Claims{
			UserID: user.ID,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(expirationTime),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(jwtKey)
		if err != nil {
			log.Printf("Error creating token: %v", err) // Added logging
			respondWithError(w, http.StatusInternalServerError, "Could not create token")
			return
		}

		respondWithJSON(w, http.StatusOK, map[string]string{"token": tokenString})
	}
}

// --- TASK HANDLERS ---

// CreateTaskHandler handles the creation of a new task.
func CreateTaskHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("userID").(int)

		var task Task
		if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		err := db.QueryRow(
			"INSERT INTO tasks (user_id, title, description, status) VALUES ($1, $2, $3, $4) RETURNING id, created_at",
			userID, task.Title, task.Description, "Pending",
		).Scan(&task.ID, &task.CreatedAt)

		if err != nil {
			log.Printf("Error creating task: %v", err) // Added logging
			respondWithError(w, http.StatusInternalServerError, "Could not create task")
			return
		}

		task.UserID = userID
		task.Status = "Pending"
		respondWithJSON(w, http.StatusCreated, task)
	}
}

// GetTasksHandler retrieves all tasks for the logged-in user.
func GetTasksHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("userID").(int)

		rows, err := db.Query("SELECT id, title, description, status, created_at FROM tasks WHERE user_id = $1 ORDER BY created_at DESC", userID)
		if err != nil {
			log.Printf("Database error getting tasks: %v", err) // Added logging
			respondWithError(w, http.StatusInternalServerError, "Database error")
			return
		}
		defer rows.Close()

		tasks := []Task{}
		for rows.Next() {
			var task Task
			if err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.CreatedAt); err != nil {
				log.Printf("Error scanning tasks: %v", err) // Added logging
				respondWithError(w, http.StatusInternalServerError, "Error scanning tasks")
				return
			}
			task.UserID = userID
			tasks = append(tasks, task)
		}

		respondWithJSON(w, http.StatusOK, tasks)
	}
}

// UpdateTaskHandler updates a task's status.
func UpdateTaskHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("userID").(int)
		vars := mux.Vars(r)
		taskID, err := strconv.Atoi(vars["id"])
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid task ID")
			return
		}

		var payload struct {
			Status string `json:"status"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		if payload.Status != "Pending" && payload.Status != "Done" {
			respondWithError(w, http.StatusBadRequest, "Invalid status value")
			return
		}

		result, err := db.Exec("UPDATE tasks SET status = $1 WHERE id = $2 AND user_id = $3", payload.Status, taskID, userID)
		if err != nil {
			log.Printf("Error updating task: %v", err) // Added logging
			respondWithError(w, http.StatusInternalServerError, "Could not update task")
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			respondWithError(w, http.StatusNotFound, "Task not found or you do not have permission to update it")
			return
		}

		respondWithJSON(w, http.StatusOK, map[string]string{"message": "Task updated successfully"})
	}
}

// DeleteTaskHandler deletes a task.
func DeleteTaskHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("userID").(int)
		vars := mux.Vars(r)
		taskID, err := strconv.Atoi(vars["id"])
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid task ID")
			return
		}

		result, err := db.Exec("DELETE FROM tasks WHERE id = $1 AND user_id = $2", taskID, userID)
		if err != nil {
			log.Printf("Error deleting task: %v", err) // Added logging
			respondWithError(w, http.StatusInternalServerError, "Could not delete task")
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			respondWithError(w, http.StatusNotFound, "Task not found or you do not have permission to delete it")
			return
		}

		respondWithJSON(w, http.StatusOK, map[string]string{"message": "Task deleted successfully"})
	}
}
