package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"golang.org/x/crypto/bcrypt"
)

type Resource struct {
	ID               uint   `gorm:"primaryKey" json:"id"`
	Email            string `json:"email"`
	Password         string `json:"password"`
	MobileNumber     string `json:"mobile_number"`
	Role             string `json:"role"`
	ClientID         uint   `json:"clientID"`
	OTP              string `json:"otp"`
	NewPassword      string `json:"new_password"`
	ConfirmPassword  string `json:"confirm_password"`
}

var (
	router    *mux.Router
	secretKey = []byte("secretkeyjwt")
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

func SetError(message string) Error {
	return Error{IsError: true, Message: message}
}

func GetDatabase() (*gorm.DB, error) {
	databasename := "postgres"
	database := "postgres"
	databasepassword := "sathya"
	databaseurl := fmt.Sprintf("host=localhost user=%s dbname=%s sslmode=disable password=%s", database, databasename, databasepassword)

	connection, err := gorm.Open(database, databaseurl)
	if err != nil {
		return nil, err
	}
	connection.AutoMigrate(&Resource{})
	fmt.Println("Database connection successful.")
	return connection, nil
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	connection, err := GetDatabase()
	if err != nil {
		log.Fatalln(err)
	}
	defer connection.Close()

	var admin Resource
	if err := json.NewDecoder(r.Body).Decode(&admin); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if admin.Email == "" || admin.Password == "" || admin.Role == "" || admin.ClientID == 0 {
		err := SetError("email, role, password, and client_id are required fields.")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}

	if admin.Role != "chef" && admin.Role != "cashier" {
		err := SetError("Invalid Role. Allowed roles are 'chef' or 'cashier'.")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}

	var dbAdmin Resource
	if err := connection.Where("email = ?", admin.Email).First(&dbAdmin).Error; err == nil {
		err := SetError("Email already in use")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(admin.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	admin.Password = string(hashedPassword)

	if err := connection.Create(&admin).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(admin.Role + " registered successfully")
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	connection, err := GetDatabase()
	if err != nil {
		log.Fatalln(err)
	}
	defer connection.Close()

	var admin Resource
	if err := json.NewDecoder(r.Body).Decode(&admin); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var storedAdmin Resource
	if err := connection.Where("email = ?", admin.Email).First(&storedAdmin).Error; err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedAdmin.Password), []byte(admin.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(15 * time.Minute)
	claims := &Claims{
		Email: storedAdmin.Email,
		Role:  storedAdmin.Role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}


func GetAllResources(w http.ResponseWriter, r *http.Request) {
	connection, err := GetDatabase()
	if err != nil {
		log.Fatalln(err)
	}
	defer connection.Close()

	var resources []Resource
	connection.Find(&resources)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resources)
}
func UpdateResource(w http.ResponseWriter, r *http.Request) {
	connection, err := GetDatabase()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer connection.Close()

	vars := mux.Vars(r)
	resourceID := vars["id"]

	var updatedResource Resource
	if err := json.NewDecoder(r.Body).Decode(&updatedResource); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var existingResource Resource
	if err := connection.First(&existingResource, resourceID).Error; err != nil {
		http.Error(w, "Resource not found", http.StatusNotFound)
		return
	}

	// Check if the updated role is valid
	if updatedResource.Role != "" && updatedResource.Role != "chef" && updatedResource.Role != "cashier" {
		http.Error(w, "Invalid role. Allowed roles are 'chef' or 'cashier'", http.StatusBadRequest)
		return
	}

	// Update the password and role fields if they are provided in the request payload
	if updatedResource.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(updatedResource.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		existingResource.Password = string(hashedPassword)
	}
	if updatedResource.Role != "" {
		existingResource.Role = updatedResource.Role
	}

	// Save the updated resource back to the database
	if err := connection.Save(&existingResource).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(existingResource)
}
func UpdatePassword(w http.ResponseWriter, r *http.Request) {
    connection, err := GetDatabase()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer connection.Close()

    vars := mux.Vars(r)
    resourceID := vars["id"]

    var updatePayload struct {
        OldPassword     string `json:"old_password"`
        NewPassword     string `json:"new_password"`
        ConfirmPassword string `json:"confirm_password"`
    }
    if err := json.NewDecoder(r.Body).Decode(&updatePayload); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Check if required fields are provided
    if updatePayload.OldPassword == "" || updatePayload.NewPassword == "" || updatePayload.ConfirmPassword == "" {
        http.Error(w, "Old password, new password, and confirm password are required", http.StatusBadRequest)
        return
    }

    // Check if new password matches confirm password
    if updatePayload.NewPassword != updatePayload.ConfirmPassword {
        http.Error(w, "New password and confirm password do not match", http.StatusBadRequest)
        return
    }

    var existingResource Resource
    if err := connection.First(&existingResource, resourceID).Error; err != nil {
        http.Error(w, "Resource not found", http.StatusNotFound)
        return
    }

    // Verify if the old password matches the password stored in the database
    if err := bcrypt.CompareHashAndPassword([]byte(existingResource.Password), []byte(updatePayload.OldPassword)); err != nil {
        http.Error(w, "Old password is incorrect", http.StatusUnauthorized)
        return
    }

    // Hash the new password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(updatePayload.NewPassword), bcrypt.DefaultCost)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Update the password field with the hashed new password
    existingResource.Password = string(hashedPassword)

    // Save the updated password back to the database
    if err := connection.Save(&existingResource).Error; err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(existingResource)
}


func UpdatePasswords(w http.ResponseWriter, r *http.Request) {
	connection, err := GetDatabase()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer connection.Close()

	vars := mux.Vars(r)
	resourceID := vars["id"]

	var updatedResource Resource
	if err := json.NewDecoder(r.Body).Decode(&updatedResource); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if new password and confirm password fields are provided
	if updatedResource.NewPassword == "" || updatedResource.ConfirmPassword == "" {
		http.Error(w, "New password and confirm password are required", http.StatusBadRequest)
		return
	}

	// Check if new password matches confirm password
	if updatedResource.NewPassword != updatedResource.ConfirmPassword {
		http.Error(w, "New password and confirm password do not match", http.StatusBadRequest)
		return
	}

	var existingResource Resource
	if err := connection.First(&existingResource, resourceID).Error; err != nil {
		http.Error(w, "Resource not found", http.StatusNotFound)
		return
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(updatedResource.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the password field with the hashed new password
	existingResource.Password = string(hashedPassword)

	// Save the updated password back to the database
	if err := connection.Save(&existingResource).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(existingResource)
}



func handleOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Authorization")
}

func main() {
	router = mux.NewRouter()
	router.HandleFunc("/register", RegisterHandler).Methods("POST")
	router.HandleFunc("/login", LoginHandler).Methods("POST")
	router.HandleFunc("/resources/{id}", UpdateResource).Methods("PUT")
	router.HandleFunc("/resources", GetAllResources).Methods("GET")
	router.HandleFunc("/resources/{id}/password", UpdatePassword).Methods("PUT")

	
	router.Methods("OPTIONS").HandlerFunc(handleOptions)

	serverAddress := ":8080"
	fmt.Printf("Server listening on port %s... ", serverAddress)
	log.Fatal(http.ListenAndServe(serverAddress, router))
}
