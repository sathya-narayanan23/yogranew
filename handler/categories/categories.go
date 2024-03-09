package categories

import (
	"encoding/json"
	"fmt"
    
	// "encoding/json"
	"sort"
	"net/http"
	// "os"
	// "path/filepath"
	"sathya-narayanan23/crudapp/database"
	"sathya-narayanan23/crudapp/users/user"

    "strconv"
    
	"github.com/gorilla/mux"

)



type Error struct {
	IsError bool   `json:"isError"`
	Message string `json:"message"`
}
func SetError(err Error, message string) Error {
	err.IsError = true
	err.Message = message
	return err
}

func CreateCategory(w http.ResponseWriter, r *http.Request) {
	connection :=  database.GetDatabase()     // Assuming you have a GetDatabase() function
	defer  database.CloseDatabase(connection) // Assuming you have a CloseDatabase() function

	vars := mux.Vars(r)
	clientId := vars["clientId"]

	var category user1.Category
	err := json.NewDecoder(r.Body).Decode(&category)
	if err != nil {
		errResponse := Error{IsError: true, Message: "Error in reading payload."}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(errResponse)
		return
	}

	// Check if the provided ClientID is valid
	var client user1.Client
	if err := connection.First(&client, clientId).Error; err != nil {
		errResponse := Error{IsError: true, Message: "Invalid ClientID"}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errResponse)
		return
	}

	if category.Image == "" || category.CategoryName == "" {
		errResponse := Error{IsError: true, Message: "Image and categoryName are required fields."}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(errResponse)
		return
	}

	// Check if category name is already in use for the given client
	var existingCategory user1.Category
	result := connection.Where("category_name = ? AND client_id = ?", category.CategoryName, clientId).First(&existingCategory)
	if result.Error == nil {
		errResponse := Error{IsError: true, Message: "CategoryName already in use for the given client"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(errResponse)
		return
	}

	// Update the client ID and create the category
	category.ClientID = client.ID
	connection.Create(&category)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}



func GetAllCategories(w http.ResponseWriter, r *http.Request) {
	connection :=  database.GetDatabase()
	defer  database.CloseDatabase(connection)

	var categories []user1.Category
	connection.Find(&categories)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}



func UpdateCategoryName(w http.ResponseWriter, r *http.Request) {
	connection :=  database.GetDatabase()
	defer  database.CloseDatabase(connection)

	vars := mux.Vars(r)
	categoryID := vars["id"]

	var updatedCategory user1.Category
	err := json.NewDecoder(r.Body).Decode(&updatedCategory)
	if err != nil {
		var err Error
		err = SetError(err, "Error in reading payload.")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(err)
		return
	}

	var existingCategory user1.Category
	if err := connection.First(&existingCategory, categoryID).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(Error{IsError: true, Message: "Category not found"})
		return
	}

	// Check if the updated category name is already in use
	var dbCategory user1.Category
	if err := connection.Where("category_name = ?", updatedCategory.CategoryName).First(&dbCategory).Error; err == nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Error{IsError: true, Message: "Category name already in use"})
		return
	}

	existingCategory.CategoryName = updatedCategory.CategoryName // Update the categoryName
	existingCategory.Image = updatedCategory.Image
	connection.Save(&existingCategory)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingCategory)
}

func DeleteCategory(w http.ResponseWriter, r *http.Request) {
	connection :=  database.GetDatabase()
	defer  database.CloseDatabase(connection)

	vars := mux.Vars(r)
	categoryID := vars["id"]

	var category user1.Category
	if err := connection.First(&category, categoryID).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(Error{IsError: true, Message: "Category not found"})
		return
	}

	// Delete associated menu items
	var menuItems []user1.MenuItem
	if err := connection.Where("category_id = ?", category.ID).Find(&menuItems).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{IsError: true, Message: "Failed to retrieve menu items"})
		return
	}

	// Delete each menu item
	for _, menuItem := range menuItems {
		if err := connection.Delete(&menuItem).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Error{IsError: true, Message: "Failed to delete menu item"})
			return
		}
	}

	if err := connection.Delete(&category).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{IsError: true, Message: "Failed to delete category"})
		return
	}

	responseMessage := fmt.Sprintf("Category with ID %s and its associated menu items have been deleted", categoryID)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": responseMessage,
	})

	w.WriteHeader(http.StatusNoContent)
}



func UpdateCategoriesStatusForClient(w http.ResponseWriter, r *http.Request) {
	// Parse URL parameters
	params := mux.Vars(r)
	clientIDStr := params["clientId"]
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	// Parse category IDs and status from the request body
	var requestPayload struct {
		CategoryIDs []uint `json:"categoryIds"`
		Status      bool   `json:"status"`
	}

	err = json.NewDecoder(r.Body).Decode(&requestPayload)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	connection := database.GetDatabase()
	defer  database.CloseDatabase(connection)

	// Retrieve categories based on the provided IDs and client ID
	var categories []user1.Category
	result := connection.Where("id IN (?) AND client_id = ?", requestPayload.CategoryIDs, clientID).Find(&categories)
	if result.Error != nil {
		http.Error(w, "Error retrieving categories", http.StatusInternalServerError)
		return
	}

	// Update 'Status' for each category
	for i := range categories {
		categories[i].Status = requestPayload.Status
		result := connection.Save(&categories[i])
		if result.Error != nil {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Respond with the updated categories in the JSON format
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(categories)
}



func UpdateCategoriesStatusForClientactive(w http.ResponseWriter, r *http.Request) {
	// Parse URL parameters
	params := mux.Vars(r)
	clientIDStr := params["clientId"]
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}

	// Parse category IDs and status from the request body
	var requestPayload struct {
		CategoryIDs []uint `json:"categoryIds"`
	}

	err = json.NewDecoder(r.Body).Decode(&requestPayload)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	// Update 'Status' for all categories associated with the client
	result := connection.Model(&user1.Category{}).Where("client_id = ?", clientID).Update("status", false)
	if result.Error != nil {
		http.Error(w, "Error updating categories status", http.StatusInternalServerError)
		return
	}

	// Update 'Status' for the specified category IDs
	result = connection.Model(&user1.Category{}).Where("id IN (?) AND client_id = ?", requestPayload.CategoryIDs, clientID).Update("status", true)
	if result.Error != nil {
		http.Error(w, "Error updating specified categories status", http.StatusInternalServerError)
		return
	}

	// Respond with the updated categories in the JSON format
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Fetch and return all updated categories
	var updatedCategories []user1.Category
	connection.Where("client_id = ?", clientID).Find(&updatedCategories)
	json.NewEncoder(w).Encode(updatedCategories)
}


func UpdateCategorys(w http.ResponseWriter, r *http.Request) {
	// Parse URL parameters
	params := mux.Vars(r)
	categoryIDStr := params["id"]
	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	// Retrieve existing client record
	var existingCategory user1.Category
	result := connection.First(&existingCategory, categoryID)
	if result.Error != nil {
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}


	statusStr := r.FormValue("status")
	if statusStr != "" {
		status, err := strconv.ParseBool(statusStr)
		if err != nil {
			http.Error(w, "Invalid status value", http.StatusBadRequest)
			return
		}
		existingCategory.Status = status
	}

	
	result = connection.Save(&existingCategory)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(existingCategory)

}



func GetActiveCategoriesForClientId(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	vars := mux.Vars(r)
	clientID := vars["clientId"]

	var categories []user1.Category
	if err := connection.Where("client_id = ? AND active_status = ?  ", clientID,true).Find(&categories).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{IsError: true, Message: "Failed to fetch active categories"})
		return
	}

	if len(categories) == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(Error{IsError: true, Message: "No active categories found for the client"})
		return
	}

	// Fetch and include menu items for each category
	for i := range categories {
		var menuItems []user1.MenuItem
		if err := connection.Where("category_id = ?", categories[i].ID).Find(&menuItems).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Error{IsError: true, Message: "Failed to fetch menu items for category"})
			return
		}
		categories[i].MenuItems = menuItems
		categories[i].TotalMenuItems = len(menuItems)
	}
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].ID < categories[j].ID
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}


func GetCategoriesForClientId(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	vars := mux.Vars(r)
	clientID := vars["clientId"]

	var categories []user1.Category
	if err := connection.Where("client_id = ? ", clientID,).Find(&categories).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{IsError: true, Message: "Failed to fetch categories"})
		return
	}

	if len(categories) == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(Error{IsError: true, Message: "No categories found for the client"})
		return
	}
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].ID < categories[j].ID
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}



func GetActiveCategoriesForClientIdStatus(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	vars := mux.Vars(r)
	clientID := vars["clientId"]

	var categories []user1.Category
	if err := connection.Where("client_id = ? AND active_status = ?  AND status = ?  ", clientID,true,true).Find(&categories).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{IsError: true, Message: "Failed to fetch active categories"})
		return
	}

	if len(categories) == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(Error{IsError: true, Message: "No active categories found for the client"})
		return
	}

	// Fetch and include menu items for each category
	for i := range categories {
		var menuItems []user1.MenuItem
		if err := connection.Where("category_id = ?", categories[i].ID).Find(&menuItems).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Error{IsError: true, Message: "Failed to fetch menu items for category"})
			return
		}
		categories[i].MenuItems = menuItems
		categories[i].TotalMenuItems = len(menuItems)
	}
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].ID < categories[j].ID
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}


func GetCategoriesForClientName(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	vars := mux.Vars(r)
	clientName := vars["clientName"]

	var categories []user1.Category
	if err := connection.Where("client_name = ?", clientName).Find(&categories).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{IsError: true, Message: "Failed to fetch categories"})
		return
	}

	if len(categories) == 0 {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(Error{IsError: true, Message: "No categories found for the client"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}



