package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"math/rand"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

// User struct
type User struct {
	ID               uint   `gorm:"primaryKey"`
	Email            string `json:"email"`
	Password         string `json:"password"`
	MobileNumber     string `json:"mobile_number"`
	Role             string `json:"role"`
	ClientID    uint   `json:"client_id"`
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
	dsn := "host=localhost user=postgres password=sathya dbname=postgres port=5432 sslmode=disable"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	// Migrate the database
	MigrateDB()
}

func MigrateDB() {
	err := db.AutoMigrate(&User{})
	if err != nil {
		log.Fatal(err)
	}
}

func RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Error{Message: "Invalid request payload"})
		return
	}

	if user.Email == "" || user.Password == "" || user.ClientID == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Error{Message: "Email and password are required fields"})
		return
	}

	// var existingUser User
	// result := db.Where("email = ?", user.Email).First(&existingUser)
	// if result.RowsAffected != 0 {
	// 	w.WriteHeader(http.StatusConflict)
	// 	json.NewEncoder(w).Encode(Error{Message: "User already exists"})
	// 	return
	// } else if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	json.NewEncoder(w).Encode(Error{Message: "Error checking existing user"})
	// 	return
	// }

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
	user.OTP = otp

	err = db.Create(&user).Error
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
	result := db.Where("email = ?", user.Email).First(&storedUser)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
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

func main() {
	initDB()

	http.HandleFunc("/user/register", RegisterUserHandler)
	http.HandleFunc("/user/login", UserLoginHandler)

	http.ListenAndServe(":8080", nil)
}

