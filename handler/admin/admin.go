package admin

import (
	"encoding/json"
	"fmt"

	// "github.com/golang-jwt/jwt"
	"net/http"
	// "os"
	// "path/filepath"
	// "strconv"
	"time"
	"github.com/golang-jwt/jwt"
	// "github.com/gorilla/handlers"
	// "github.com/gorilla/mux"
	// "gopkg.in/gomail.v2"
	// "github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
	"sathya-narayanan23/crudapp/database"
	
	// "sathya-narayanan23/crudapp/database"
	"sathya-narayanan23/crudapp/users/user"


)




type Claims struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.StandardClaims
}

type Error struct {
	IsError bool   `json:"isError"`
	Message string `json:"message"`
}
func SetError(err Error, message string) Error {
	err.IsError = true
	err.Message = message
	return err
}
var (
	// connection     
	jwtKey = []byte("your-secret-key")
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	
	connection := database.GetDatabase()
	var admin user1.Admin
	json.NewDecoder(r.Body).Decode(&admin)
	
	if admin.Email == "" || admin.Password == "" || admin.Role == ""  {
		var err Error
		err = SetError(err, "email,role and password are required fields.")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}

	if admin.Role != "admin" && admin.Role != "superadmin" {
		var err Error
		err = SetError(err, "Invalid Role. Allowed roles are 'admin' or 'superadmin'.")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}
	var dbadmin user1.Admin
	connection.Where("email = ?", admin.Email).First(&dbadmin)

	if dbadmin.Email != "" {
		var err Error
		err = SetError(err, "Email already in use")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}


	// Hash the password before saving it to the database
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(admin.Password), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	admin.Password = string(hashedPassword)
	
	connection.Create(&admin)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode( admin.Role + "   registered successfully")
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var admin user1.Admin
	
	connection := database.GetDatabase()
	json.NewDecoder(r.Body).Decode(&admin)

	// Retrieve the admin from the database using the provided email
	var storedAdmin user1.Admin
	if err := connection.Where("email = ?", admin.Email).First(&storedAdmin).Error; err != nil {
		// Admin not found in the database
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode("Invalid credentials")
		return
	}

	// Compare the stored hashed password with the provided password
	if err := bcrypt.CompareHashAndPassword([]byte(storedAdmin.Password), []byte(admin.Password)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode("Invalid credentials")
		return
	}

	// Generate JWT token with admin's role
	expirationTime := time.Now().Add(15 * time.Minute)
	claims := &Claims{
		Email: storedAdmin.Email,
		Role:  storedAdmin.Role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Include the token in the response with "Bearer " prefix
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token":tokenString})
}


func ClientPageHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve the JWT token from the request header
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Remove the "Bearer " prefix
	tokenString = tokenString[7:]

	// Parse the JWT token
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Check if the  has the required role to access the resource
	switch claims.Role {
	case "admin":
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(fmt.Sprintf("Hello, %s! You have access to the admin page.", claims.Email))
	case "superadmin":
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(fmt.Sprintf("Hello, %s! You have access to the superadmin page.", claims.Email))
	default:
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode("Insufficient privileges")
	}
}




func ClientAdminPageHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve the JWT token from the request header
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Remove the "Bearer " prefix
	tokenString = tokenString[7:]

	// Parse the JWT token
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Check if the admin has the required role to access the resource
	if claims.Role != "admin" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode("Insufficient privileges")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(fmt.Sprintf("Hello, %s! You have access to the clientadmin page.", claims.Email))
}

func ClientSuperAdminPageHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve the JWT token from the request header
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Remove the "Bearer " prefix
	tokenString = tokenString[7:]

	// Parse the JWT token
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !token.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Check if the user has the required role to access the resource
	if claims.Role != "superadmin" {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode("Insufficient privileges")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(fmt.Sprintf("Hello, %s! You have access to the clientsuperadmin page.", claims.Email))
}



