package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"
    "fmt"
    "math/rand"
	"github.com/dgrijalva/jwt-go"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
    "gorm.io/gorm"
)

var db *sql.DB

// User struct

// User struct
type User struct {
	ID               uint   `json:"id"`
	Email            string `json:"email"`
	Password         string `json:"password"`
	Role             string `json:"role"`
	OTP              string `json:"otp"` // New field for OTP
	NewPassword      string `json:"new_password"` // New field for new password
	ConfirmPassword  string `json:"confirm_password"` // New field for confirming new password
}

// Error struct
type Error struct {
	Message string `json:"message"`
}

// Claims struct for JWT
type Claims struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.StandardClaims
}

var jwtKey = []byte("my_secret_key") // Change this with your own secret key

func initDB() {
	var err error
	db, err = sql.Open("postgres", "postgres://postgres:sathya@localhost/postgres?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
    
	db.AutoMigrate(&User{})
}
func RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Error{Message: "Invalid request payload"})
		return
	}

	if user.Email == "" || user.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Error{Message: "Email and password are required fields"})
		return
	}

	var existingUser User
	err = db.QueryRow("SELECT id FROM users WHERE email = $1", user.Email).Scan(&existingUser.ID)
	if err == nil {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(Error{Message: "User already exists"})
		return
	} else if err != sql.ErrNoRows {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{Message: "Error checking existing user"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{Message: "Failed to hash password"})
		return
	}
	user.Password = string(hashedPassword)

	// Generate a 4-digit OTP
	rand.Seed(time.Now().UnixNano())
	otp := fmt.Sprintf("%04d", rand.Intn(10000))

	_, err = db.Exec("INSERT INTO users (email, password, role, otp) VALUES ($1, $2, $3, $4)", user.Email, user.Password, user.Role, otp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{Message: "Failed to create user"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func UserLoginHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Error{Message: "Invalid request payload"})
		return
	}

	var storedUser User
	err = db.QueryRow("SELECT id, password, role FROM users WHERE email = $1", user.Email).Scan(&storedUser.ID, &storedUser.Password, &storedUser.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(Error{Message: "Invalid credentials"})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{Message: "Error checking user credentials"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(Error{Message: "Invalid credentials"})
		return
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Email: storedUser.Email,
		Role:  storedUser.Role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{Message: "Failed to generate token"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}
func MigrateDB(db *gorm.DB) {
	db.AutoMigrate(&User{})
}

func main() {
	initDB()

	http.HandleFunc("/user/register", RegisterUserHandler)
	http.HandleFunc("/user/login", UserLoginHandler)

	http.ListenAndServe(":8080", nil)
}
