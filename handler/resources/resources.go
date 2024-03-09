package resources

import(
	// "io"
	"net/http"
	"time"
	"fmt"
	// "path/filepath"
	"sathya-narayanan23/crudapp/database"
	"sathya-narayanan23/crudapp/users/user"
	
	"sathya-narayanan23/handler/mail"
	"github.com/gorilla/mux"
	"encoding/json"
	"log"
	"golang.org/x/crypto/bcrypt"
	
	"github.com/golang-jwt/jwt"
	"strconv"

)
var (
	
	secretkey string = "secretkeyjwt"
)

type Error struct {
	IsError bool   `json:"isError"`
	Message string `json:"message"`
}


type Authentication struct {
	MobileNumber string `json:"mobileNumber"`
	Password     string `json:"password"`
}

func SetError(err Error, message string) Error {
	err.IsError = true
	err.Message = message
	return err
}

func GetChefsForClient(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	vars := mux.Vars(r)
	clientID := vars["id"]

	var chefs []user1.Resource
	connection.Where("client_id = ? AND role = ?", clientID, "chef").Find(&chefs)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chefs)
}

func GetCashiersForClient(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	vars := mux.Vars(r)
	clientID := vars["id"]

	var cashiers []user1.Resource
	connection.Where("client_id = ? AND role = ?", clientID, "cashier").Find(&cashiers)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cashiers)
}

func GeneratehashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 4)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}



func GenerateJWT(mobileNumber string) (string, error) {
	
	var mySigningKey = []byte(secretkey)
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["authorized"] = true
	claims["mobileNumber"] = mobileNumber
	claims["exp"] = time.Now().Add(time.Minute * 1000).Unix()

	tokenString, err := token.SignedString(mySigningKey)
	if err != nil {
		fmt.Printf("Something went Wrong: %s", err.Error())
		return "", err
	}

	return tokenString, nil
}



func CreateResource(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	var resource user1.Resource
	err := json.NewDecoder(r.Body).Decode(&resource)
	if err != nil {
		var err Error
		err = SetError(err, "Error in reading payload.")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}

	if resource.MobileNumber == "" || resource.Passwords == "" || resource.Name == "" || resource.ClientID == 0 {
		var err Error
		err = SetError(err, "name ,mobileNumber, and passwords are required fields.")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}

	if resource.Role != "chef" && resource.Role != "cashier" {
		var err Error
		err = SetError(err, "Invalid Role. Allowed roles are 'chef' or 'cashier'.")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}
	var dbresource user1.Resource
	connection.Where("mobile_number = ?", resource.MobileNumber).First(&dbresource)

	if dbresource.MobileNumber != "" {
		var err Error
		err = SetError(err, "Mobile number  already in use")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}

	resource.Password, err = GeneratehashPassword(resource.Passwords)
	if err != nil {
		log.Fatalln("Error in passwords hashing.")
	}

	connection.Create(&resource)

	var client user1.Client
	connection.First(&client, resource.ClientID)

	// Send email notification
	err = mail.SendEmailNotification(resource.Role, resource.Name, client.Email, resource.MobileNumber, resource.Passwords)
	if err != nil {
		// Handle error if needed
		log.Println("Error sending email notification:", err)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resource)
}

func ResourceLogin(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	var authDetails Authentication

	err := json.NewDecoder(r.Body).Decode(&authDetails)
	if err != nil {
		var err Error
		err = SetError(err, "Error in reading payload.")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}

	var authResource user1.Resource
	connection.Where("mobile_number = ?", authDetails.MobileNumber).First(&authResource)

	if authResource.MobileNumber == "" {
		var err Error
		err = SetError(err, "Mobile number or Password is incorrect")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}

	check := CheckPasswordHash(authDetails.Password, authResource.Password)

	if !check {
		var err Error
		err = SetError(err, "Mobile number or Password is incorrect")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}

	// Fetch the client associated with the authenticated resource
	var client user1.Client
	connection.First(&client, authResource.ClientID)

	validToken, err := GenerateJWT(authResource.MobileNumber)
	if err != nil {
		var err Error
		err = SetError(err, "Failed to generate token")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}

	authResource.Password = ""

	// Set the response status code to 200 for successful login
	w.WriteHeader(http.StatusOK)

	response := struct {
		User         user1.Resource `json:"user"`
		Token        string   `json:"token"`
		ClientID     int      `json:"clientID"`
		PrimaryColor string   `json:"primaryColor"`

		ClientName string `json:"clientName"`
		Logo       string `json:"logo"`
		Status     bool   `json:"status"`
	}{
		User:         authResource,
		Token:        validToken,
		ClientName:   client.Name,
		PrimaryColor: client.PrimaryColour,
		Logo:         client.Logo,
		Status:       client.Status,
	}

	w.Header().Set("Content-type", "application/json")
	json.NewEncoder(w).Encode(response)
}



func GetAllResources(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	var resources []user1.Resource
	connection.Find(&resources)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resources)
}

func GetResource(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	vars := mux.Vars(r)
	resourceID := vars["id"]

	var resource user1.Resource
	connection.First(&resource, resourceID) // Preload Categories relationship
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resource)
}

func DeleteResource(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	vars := mux.Vars(r)
	resourceID := vars["id"]

	var resource user1.Resource
	if err := connection.First(&resource, resourceID).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(Error{IsError: true, Message: "Resource not found"})
		return
	}

	if err := connection.Delete(&resource).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{IsError: true, Message: "Failed to delete resource"})
		return
	}

	responseMessage := fmt.Sprintf(" ResourceID %s has been deleted", resourceID)
	json.NewEncoder(w).Encode(responseMessage)
	w.WriteHeader(http.StatusNoContent)
}

func UpdateResource(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	vars := mux.Vars(r)
	resourceID := vars["id"]

	var updatedResource user1.Resource
	err := json.NewDecoder(r.Body).Decode(&updatedResource)
	if err != nil {
		var err Error
		err = SetError(err, "Error in reading payload.")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}
	var existingResource user1.Resource
	connection.First(&existingResource, resourceID)
	if existingResource.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if updatedResource.Name != "" {
		existingResource.Name = updatedResource.Name
	}

	if updatedResource.Password != "" {
		existingResource.Password = updatedResource.Password
	}
	if updatedResource.Role != "" {
		existingResource.Role = updatedResource.Role
	}

	// Check if the new mobile number is already in use for the same client
	if existingResource.MobileNumber != updatedResource.MobileNumber || existingResource.ClientID != updatedResource.ClientID {
		var resourceWithSameMobile user1.Resource
		connection.Where("mobile_number = ? AND client_id = ?", updatedResource.MobileNumber, updatedResource.ClientID).First(&resourceWithSameMobile)
		if resourceWithSameMobile.ID != 0 {
			w.WriteHeader(http.StatusConflict) // 409 Conflict (already in use)
			err := SetError(Error{}, "Mobile number is already in use for the same client.")
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(err)
			return
		}
	}
	if updatedResource.MobileNumber != "" {
		existingResource.MobileNumber = updatedResource.MobileNumber
	}
	connection.Save(&existingResource)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingResource)
}


func GetResourcesForClient(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	vars := mux.Vars(r)
	clientID := vars["client_id"]


	type SearchCriteria struct {
		
		Name         string `json:"name"`
		Role         string `json:"role"`
		MobileNumber string `json:"mobile_number"`
    }

    // Parse the query parameters to extract the search criteria
    queryValues := r.URL.Query()
    searchCriteria := SearchCriteria{
        Name:    queryValues.Get("name"),
        Role: queryValues.Get("role"),
        MobileNumber:    queryValues.Get("mobile_number"),
    
	}

    // Build the query based on the provided criteria
    query := connection.Where("client_id = ?", clientID)
	
    if searchCriteria.Role != "" {
        query = query.Where("role ILIKE  ?","%"+ searchCriteria.Role +"%")
    }
	if searchCriteria.Name != "" {
        query = query.Where("name ILIKE ?", "%"+searchCriteria.Name+"%")
    }
	if searchCriteria.MobileNumber != "" {
        query = query.Where("mobile_number  ?", "%"+searchCriteria.MobileNumber+"%")
    }
    // Execute the query
    var resources []user1.Resource
    result := query.Find(&resources)
    if result.Error != nil {
        http.Error(w, result.Error.Error(), http.StatusInternalServerError)
        return
    }

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resources)
}



func GetResourceByID(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	// Parse URL parameters
	params := mux.Vars(r)
	resourceIDStr := params["id"]
	resourceID, err := strconv.Atoi(resourceIDStr)
	if err != nil {
		http.Error(w, "Invalid resource ID", http.StatusBadRequest)
		return
	}

	// Retrieve the resource by ID
	var resource user1.Resource
	result := connection.First(&resource, resourceID)
	if result.Error != nil {
		http.Error(w, "Resource not found", http.StatusNotFound)
		return
	}

	// Respond with the retrieved resource
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resource)
}
