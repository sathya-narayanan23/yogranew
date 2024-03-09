package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	// "math/rand"
	"time"

    _ "github.com/lib/pq"
	"github.com/jinzhu/gorm"
	// "fmt"
)


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


func GetDatabase() *gorm.DB {

	databasename := "yogra"
	database := "postgres"
	databasepassword := "sathya"
	databaseurl := "postgres://postgres:" + databasepassword + "@localhost/" + databasename + "?sslmode=disable"

	connection, err := gorm.Open(database, databaseurl)
	if err != nil {
		log.Fatalln(err)
	}
	sqldb := connection.DB()

	err = sqldb.Ping()
	if err != nil {
		log.Fatal("Database connec/ted")
	}
	fmt.Println("Database connection successful.")
	return connection
}

func CloseDatabase(connection *gorm.DB) {
	sqldb := connection.DB()
	sqldb.Close()
}

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

func main() {
	initDB()

	http.HandleFunc("/user/register", RegisterUserHandler)
	http.HandleFunc("/user/login", UserLoginHandler)

	http.ListenAndServe(":8080", nil)
}

