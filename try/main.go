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
	
	// "encoding/json"
	// "fmt"
	"github.com/golang-jwt/jwt"
	
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)


type Client struct {
    gorm.Model
    Name                  string    `json:"name"`
}

// Resource struct
type Resource struct {
	ID               uint   `gorm:"primaryKey"`
	Email            string `json:"email"`
	Password         string `json:"password"`
	MobileNumber     string `json:"mobile_number"`
	Role             string `json:"role"`
	ClientID    uint   `json:"client_id"`
	OTP              string `json:"otp"`
	NewPassword      string `json:"new_password"` 
	ConfirmPassword  string `json:"confirm_password"` 
}

var (
	router    *mux.Router
	
	secretkey string = "secretkeyjwt"
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
func InitialMigration() {
	connection := GetDatabase()
	defer CloseDatabase(connection)
	
	connection.AutoMigrate(&Resource{})
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	
	connection := GetDatabase()
	var admin Resource
	
	err := json.NewDecoder(r.Body).Decode(&admin)
	if err != nil {
		var err Error
		err = SetError(err, "Error in reading payload.")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}
	
	if admin.Email == "" || admin.Password == "" || admin.Role == "" || admin.ClientID == 0 {
		var err Error
		err = SetError(err, "email,role and password are required fields.")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}

	if admin.Role != "chef" && admin.Role != "cashier" {
		var err Error
		err = SetError(err, "Invalid Role. Allowed roles are 'admin' or 'superadmin'.")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}
	var dbadmin Resource
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
	var admin Resource
	
	connection := GetDatabase()
	json.NewDecoder(r.Body).Decode(&admin)

	// Retrieve the admin from the database using the provided email
	var storedAdmin Resource
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
	router = mux.NewRouter()
	InitialMigration()
	// GetAllClientplannewww()
	GetDatabase()
	
	
	// router := http.NewServe/Mux()

	router.HandleFunc("/register",RegisterHandler).Methods("POST")
	router.HandleFunc("/login", LoginHandler).Methods("POST")
	// router.HandleFunc("/hi",  admin.ClientPageHandler).Methods("GET")
	
	
	// art the server
	serverAddress := ":8080" // Change to your server's IP address

	fmt.Printf("Server listening on port %s... ", serverAddress)
	log.Fatal(http.ListenAndServe(serverAddress, nil))
	// log.Fatal(http.ListenAndServe(":8080", loggedRouter))


	router.Methods("OPTIONS").HandlerFunc(handleOptions)


}
func handleOptions(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Access-Control-Allow-Origin", "")
w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, ...") // Add necessary headers
}


