// backend/models.go
package main

import "time"

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Don't send password hash in JSON responses
	CreatedAt time.Time `json:"created_at"`
}

type Task struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// Credentials is used for decoding login/register requests
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
