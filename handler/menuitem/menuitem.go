package menuitem 

import (
	"encoding/json"
	"fmt"
    
	"github.com/gofrs/uuid"
	"strings"
	"sort"
	"net/http"
	"os"
	"io"
	"path/filepath"
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


func CreateMenuItemBanner(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	vars := mux.Vars(r)
	clientID := vars["clientId"]

	var menuItem user1.MenuItem
	err := r.ParseMultipartForm(10 << 20) // 10 MB limit for form data
	if err != nil {
		var err Error
		err = SetError(err, "Error in parsing form data.")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err)
		return
	}

	menuItem.Name = r.FormValue("name")

	menuItem.Currency = r.FormValue("currency")
	menuItem.Sub_category = r.FormValue("sub_category")
	menuItem.Description = r.FormValue("description")
	price, _ := strconv.ParseFloat(r.FormValue("price"), 64)
	menuItem.Price = price
	
	
	foodType  := r.FormValue("food_type")
	if foodType  == "veg" || foodType  == "non-veg" {
		menuItem.Food_type = foodType 
	} else {
		http.Error(w, "Invalid food type. Should be 'veg' or 'non-veg'", http.StatusBadRequest)
        return
		// menuItem.Food_type = "unknown"
	}

	// Check if the menu item name already exists within the same category and client
	var existingMenuItem user1.MenuItem
	if err := connection.Where("client_id = ? AND  name = ?", clientID,  menuItem.Name).First(&existingMenuItem).Error; err == nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(Error{IsError: true, Message: "Menu item name already exists for this category"})
		return
	}

	var imageFilePath string
	if imageFile, header, err := r.FormFile("image"); err == nil {
		defer imageFile.Close()

		// Generate a unique filename for the image
		uniqueFileName, err := uuid.NewV4()// Generate a UUID
		if err != nil {
			var err Error
			err = SetError(err, "Error generating UUID.")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(err)
			return
		}
		os.MkdirAll("uploads", os.ModePerm)

		// Create the complete file path
		ext := filepath.Ext(header.Filename)
		imageFilePath = filepath.Join("uploads", uniqueFileName.String()+ext)

		// Create the file on the server
		outputFile, err := os.Create(imageFilePath)
		if err != nil {
			var err Error
			err = SetError(err, "Error creating image file.")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(err)
			return
		}
		defer outputFile.Close()
		// Copy the image data to the file
		_, err = io.Copy(outputFile, imageFile)
		if err != nil {
			var err Error
			err = SetError(err, "Error copying image data.")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(err)
			return
		}
	}
	

	if menuItem.Name == "" || menuItem.Price == 0 {
		var err Error
		err = SetError(err, "itemName and price are required fields.")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err)
		return
	}

	time, errs := strconv.Atoi(r.FormValue("time"))
	if errs != nil {
		var errs Error
		errs = SetError(errs, "Invalid time format.")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errs)
		return
	}

	offer, err := strconv.Atoi(r.FormValue("offer"))
	if err != nil {
		var err Error
		err = SetError(err, "Invalid offer format.")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err)
		return
	}
	// var discountedPrice float64

	// Calculate discounted price
	discountedPrice := price - (price * float64(offer) / 100)

	// Assign offer and discounted price to the menuItem
	menuItem.Offer = float64(offer)
	menuItem.OfferRate = float64(discountedPrice)
	clientIDUint, err := strconv.ParseUint(clientID, 10, 64)
	if err != nil {
		http.Error(w, "Invalid clientID format.", http.StatusBadRequest)
		return
	}
	menuItem.ClientID = uint(clientIDUint)
	
	menuItem.Time = time
	menuItem.Status = false
	menuItem.Recommendation = false
	menuItem.Temporary_status = false
	menuItem.Banner = true
	serverAddr := r.Host

	if len(imageFilePath) > 0 {
		imageServeURL := "http://" + serverAddr + "/" + imageFilePath
        menuItem.Image = imageServeURL
		menuItem.Imagepath = imageFilePath

		connection.Create(&menuItem)

		// Respond with the image URL and a success message
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"data": menuItem, "imagePath": imageServeURL, "message": "Image uploaded successfully"})
		return
	}

	// If there's no image, respond with the JSON-encoded menuItem
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(menuItem)
}


func CreateMenuItem(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	vars := mux.Vars(r)
	clientID := vars["clientId"]

	var menuItem user1.MenuItem
	err := r.ParseMultipartForm(10 << 20) // 10 MB limit for form data
	if err != nil {
		var err Error
		err = SetError(err, "Error in parsing form data.")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err)
		return
	}

	// menuItem.ItemName = r.FormValue("itemName")
	menuItem.Currency = r.FormValue("currency")
	menuItem.Sub_category = r.FormValue("sub_category")
	menuItem.Description = r.FormValue("description")
	price, _ := strconv.ParseFloat(r.FormValue("price"), 64)
	menuItem.Price = price
	menuItem.CategoryName = r.FormValue("categoryName")
	
	foodType  := r.FormValue("food_type")
	if foodType  == "veg" || foodType  == "non-veg" {
		menuItem.Food_type = foodType 
	} else {
		http.Error(w, "Invalid food type. Should be 'veg' or 'non-veg'", http.StatusBadRequest)
        return
		// menuItem.Food_type = "unknown"
	}

	// Check if the menu item name already exists within the same category and client
	// var existingMenuItem user1.MenuItem
	// if err := connection.Where("client_id = ? AND category_name = ? AND item_name = ?", clientID, menuItem.CategoryName, menuItem.ItemName).First(&existingMenuItem).Error; err == nil {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	json.NewEncoder(w).Encode(Error{IsError: true, Message: "Menu item name already exists for this category"})
	// 	return
	// }
	var itemName string
	var existingMenuItem user1.MenuItem
	var category user1.Category
	
	if itemName = r.FormValue("itemName"); itemName != "" {
		// If "itemName" is provided, use it
		menuItem.ItemName = itemName
	
		// Check if the menu item name already exists within the same category and client
		if err := connection.Where("client_id = ? AND category_name = ? AND item_name = ?", clientID, menuItem.CategoryName, menuItem.ItemName).First(&existingMenuItem).Error; err == nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Error{IsError: true, Message: "Menu item name already exists for this category"})
			return
		}
		if err := connection.Where("client_id = ? AND category_name = ?", clientID, menuItem.CategoryName).First(&category).Error; err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(Error{IsError: true, Message: "Category not found"})
			return
		}
			
		menuItem.Status = true
		menuItem.Recommendation = false
		menuItem.Temporary_status = false
		menuItem.Banner = false
	} else {
		// If "itemName" is not provided, fallback to "name"
		menuItem.ItemName = r.FormValue("name")
			
		menuItem.Status = false
		menuItem.Recommendation = false
		menuItem.Temporary_status = false
		menuItem.Banner = true
	}
	// Find the category by categoryName and clientID
	

	// Handling Image Upload
	var imageFilePath string
	if imageFile, header, err := r.FormFile("image"); err == nil {
		defer imageFile.Close()

		// Generate a unique filename for the image
		uniqueFileName, err := uuid.NewV4()// Generate a UUID
		if err != nil {
			var err Error
			err = SetError(err, "Error generating UUID.")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(err)
			return
		}
		
		os.MkdirAll("uploads", os.ModePerm)

		// Create the complete file path
		ext := filepath.Ext(header.Filename)
		imageFilePath = filepath.Join("uploads", uniqueFileName.String()+ext)

		// Create the file on the server
		outputFile, err := os.Create(imageFilePath)
		if err != nil {
			var err Error
			err = SetError(err, "Error creating image file.")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(err)
			return
		}
		defer outputFile.Close()

		// Copy the image data to the file
		_, err = io.Copy(outputFile, imageFile)
		if err != nil {
			var err Error
			err = SetError(err, "Error copying image data.")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(err)
			return
		}
	}

	if menuItem.ItemName == "" || menuItem.Price == 0 {
		var err Error
		err = SetError(err, "itemName and price are required fields.")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err)
		return
	}

	time, err := strconv.Atoi(r.FormValue("time"))
	if err != nil {
		var err Error
		err = SetError(err, "Invalid time format.")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err)
		return
	}
	menuItem.Time = time

	menuItem.ClientID = category.ClientID
	menuItem.CategoryID = category.ID
	menuItem.CategoryName = category.CategoryName

	// connection.Create(&menuItem)
	serverAddr := r.Host

	if len(imageFilePath) > 0 {
		imageServeURL := "http://" + serverAddr + "/" + imageFilePath
        menuItem.Image = imageServeURL
		menuItem.Imagepath = imageFilePath

		connection.Create(&menuItem)

		// Respond with the image URL and a success message
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"data": menuItem, "imagePath": imageServeURL, "message": "Image uploaded successfully"})
		return
	}

	// If there's no image, respond with the JSON-encoded menuItem
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(menuItem)
}



func GetAllMenuItem(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	var menuItem []user1.MenuItem
	connection.Find(&menuItem)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(menuItem)
}


func GetMenuItemsearch(w http.ResponseWriter, r *http.Request) {
    connection := database.GetDatabase()
    defer database.CloseDatabase(connection)

    vars := mux.Vars(r)
    clientID := vars["clientId"]

    // Define a struct to hold the search criteria
    type SearchCriteria struct {
		
	    ItemName     string  `json:"itemName"`
        FoodType    string `json:"food_type"`
        SubCategory string `json:"sub_category"`
		
        CategoryID    string   `json:"categoryID"`
        CategoryName  string `json:"categoryName"`
    }

    // Parse the query parameters to extract the search criteria
    queryValues := r.URL.Query()
    searchCriteria := SearchCriteria{
        FoodType:    queryValues.Get("food_type"),
        SubCategory: queryValues.Get("sub_category"),
        ItemName:    queryValues.Get("itemName"),
        CategoryID: queryValues.Get("categoryID"),
        CategoryName:  queryValues.Get("categoryName"),
        // Parse other criteria in a similar way
    }


    // Build the query based on the provided criteria
    query := connection.Where("client_id = ?", clientID)
	
    if searchCriteria.FoodType != "" {
        query = query.Where("food_type ILIKE   ?","%"+ searchCriteria.FoodType +"%")
    }
	if searchCriteria.ItemName != "" {
        query = query.Where("item_name ILIKE  ?", "%"+searchCriteria.ItemName+"%")
    }
	if searchCriteria.CategoryName != "" {
        query = query.Where("category_name ILIKE  ?", "%"+searchCriteria.CategoryName+"%")
    }
    if searchCriteria.SubCategory != "" {
        query = query.Where("sub_category ILIKE   ?", "%" + searchCriteria.SubCategory  + "%")
    }
	if searchCriteria.CategoryID != "" {
        query = query.Where("category_id  = ?", searchCriteria.CategoryID)
    }
    // Add conditions for other criteria

    // Execute the query
    var menuItems []user1.MenuItem
    result := query.Find(&menuItems)
    if result.Error != nil {
        http.Error(w, result.Error.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(menuItems)
}



func GetMenuItemsearchnew(w http.ResponseWriter, r *http.Request) {
    connection := database.GetDatabase()
    defer database.CloseDatabase(connection)

    vars := mux.Vars(r)
    clientID := vars["clientId"]
	
    categoryID := vars["categoryID"]


    // Define a struct to hold the search criteria
    type SearchCriteria struct {
		
	    ItemName     string  `json:"itemName"`
        FoodType    string `json:"food_type"`
        SubCategory string `json:"sub_category"`
        CategoryName  string `json:"categoryName"`
    }

    // Parse the query parameters to extract the search criteria
    queryValues := r.URL.Query()
    searchCriteria := SearchCriteria{
        FoodType:    queryValues.Get("food_type"),
        SubCategory: queryValues.Get("sub_category"),
        ItemName:    queryValues.Get("itemName"),
        CategoryName:  queryValues.Get("categoryName"),
        // Parse other criteria in a similar way
    }


    // Build the query based on the provided criteria
    query := connection.Where("client_id = ? AND category_id = ?", clientID,categoryID)
	
    if searchCriteria.FoodType != "" {
        query = query.Where("food_type ILIKE ?","%"+ searchCriteria.FoodType +"%")
    }
	if searchCriteria.ItemName != "" {
        query = query.Where("item_name ILIKE ?", "%"+searchCriteria.ItemName+"%")
    }
	if searchCriteria.CategoryName != "" {
        query = query.Where("category_name ILIKE ?", "%"+searchCriteria.CategoryName+"%")
    }
    if searchCriteria.SubCategory != "" {
        query = query.Where("sub_category ILIKE ?", "%"+ searchCriteria.SubCategory +"%")
    }
	
    // Execute the query
    var menuItems []user1.MenuItem
    result := query.Find(&menuItems)
    if result.Error != nil {
        http.Error(w, result.Error.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(menuItems)
}



func GetMenuItem(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	vars := mux.Vars(r)
	menuItemID := vars["id"]

	var menuItem user1.MenuItem
	connection.First(&menuItem, menuItemID) // Preload Categories relationship
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(menuItem)
}

func DeleteMenuItem(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	vars := mux.Vars(r)
	menuItemID := vars["id"]

	var menuItem user1.MenuItem
	if err := connection.First(&menuItem, menuItemID).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(Error{IsError: true, Message: "Menu item not found"})
		return
	}

	if err := connection.Delete(&menuItem).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{IsError: true, Message: "Failed to delete menu item"})
		return
	}

	responseMessage := fmt.Sprintf("Menu item with ID %s has been deleted", menuItemID)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": responseMessage,
	})

	w.WriteHeader(http.StatusNoContent)
}

func GetMenuItemsByCategoryclientactive(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	vars := mux.Vars(r)

	clientID := vars["clientId"]

	categoryID := vars["categoryid"]

	var menuItems []user1.MenuItem
	if err := connection.Where("category_id = ? AND client_id = ? AND status = ? AND temporary_status = ?", categoryID, clientID, true, false ).Find(&menuItems).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{IsError: true, Message: "Failed to fetch menu items by category"})
		return
	}

	// Fetch the associated category to populate categoryName for each menu item
	for i, item := range menuItems {
		var category user1.Category
		if err := connection.First(&category, item.CategoryID).Error; err == nil {
			menuItems[i].CategoryName = category.CategoryName
		}
	}
	sort.Slice(menuItems, func(i, j int) bool {
		return menuItems[i].ID < menuItems[j].ID
	})
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(menuItems)
}


func GetMenuItemsBybanner(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	vars := mux.Vars(r)

	clientID := vars["clientID"]

	// categoryID := vars["categoryid"]

	var menuItems []user1.MenuItem
	if err := connection.Where(" client_id = ? AND banner = ? ",  clientID, true).Find(&menuItems).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{IsError: true, Message: "Failed to fetch menu items by category"})
		return
	}

	// Fetch the associated category to populate categoryName for each menu item
	for i, item := range menuItems {
		var category user1.Category
		if err := connection.First(&category, item.CategoryID).Error; err == nil {
			menuItems[i].CategoryName = category.CategoryName
		}
	}
	sort.Slice(menuItems, func(i, j int) bool {
		return menuItems[i].ID < menuItems[j].ID
	})
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(menuItems)
}


func GetMenuItemsByCategoryclientactivelike(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	vars := mux.Vars(r)

	clientID := vars["clientID"]

	// categoryID := vars["categoryid"]

	var menuItems []user1.MenuItem
	if err := connection.Where(" client_id = ? AND status = ? AND temporary_status = ? AND recommendation = ? ",  clientID, true, false,true ).Find(&menuItems).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{IsError: true, Message: "Failed to fetch menu items by category"})
		return
	}

	// Fetch the associated category to populate categoryName for each menu item
	for i, item := range menuItems {
		var category user1.Category
		if err := connection.First(&category, item.CategoryID).Error; err == nil {
			menuItems[i].CategoryName = category.CategoryName
		}
	}
	sort.Slice(menuItems, func(i, j int) bool {
		return menuItems[i].ID < menuItems[j].ID
	})
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(menuItems)
}

func GetMenuItemsByCategoryclientveg(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	vars := mux.Vars(r)

	clientID := vars["clientId"]

	categoryID := vars["categoryid"]

	var menuItems []user1.MenuItem
	if err := connection.Where("category_id = ? AND client_id = ? AND food_type = ? AND temporary_status = ?", categoryID, clientID, "veg",false).Find(&menuItems).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{IsError: true, Message: "Failed to fetch menu items by category"})
		return
	}

	// Fetch the associated category to populate categoryName for each menu item
	for i, item := range menuItems {
		var category user1.Category
		if err := connection.First(&category, item.CategoryID).Error; err == nil {
			menuItems[i].CategoryName = category.CategoryName
		}
	}
	sort.Slice(menuItems, func(i, j int) bool {
		return menuItems[i].ID < menuItems[j].ID
	})
	

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(menuItems)
}


func GetMenuItemsByCategoryclientnonveg(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	vars := mux.Vars(r)

	clientID := vars["clientId"]

	categoryID := vars["categoryid"]

	var menuItems []user1.MenuItem
	if err := connection.Where("category_id = ? AND client_id = ? AND food_type = ? AND temporary_status = ? ", categoryID, clientID, "non-veg" ,false).Find(&menuItems).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{IsError: true, Message: "Failed to fetch menu items by category"})
		return
	}

	// Fetch the associated category to populate categoryName for each menu item
	for i, item := range menuItems {
		var category user1.Category
		if err := connection.First(&category, item.CategoryID).Error; err == nil {
			menuItems[i].CategoryName = category.CategoryName
		}
	}
	sort.Slice(menuItems, func(i, j int) bool {
		return menuItems[i].ID < menuItems[j].ID
	})
	

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(menuItems)
}


func GetMenuItemsByCategory(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	vars := mux.Vars(r)

	clientID := vars["clientId"]
	categoryID := vars["categoryid"]
	searchQuery := r.URL.Query().Get("search")

	var menuItems []user1.MenuItem

	// Modify the query to include both clientID, categoryID, and itemName search
	query := connection.Where("client_id = ? AND category_id = ?", clientID, categoryID)
	if searchQuery != "" {
		query = query.Where("item_name LIKE ? OR description LIKE ?", "%"+searchQuery+"%", "%"+searchQuery+"%")
	}

	if err := query.Find(&menuItems).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{IsError: true, Message: "Failed to fetch menu items by category"})
		return
	}

	// Fetch the associated category to populate categoryName for each menu item
	for i, item := range menuItems {
		var category user1.Category
		if err := connection.First(&category, item.CategoryID).Error; err == nil {
			menuItems[i].CategoryName = category.CategoryName
		}
	}
	sort.Slice(menuItems, func(i, j int) bool {
		return menuItems[i].ID < menuItems[j].ID
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(menuItems)
}


func UpdateMenuItem(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	vars := mux.Vars(r)
	clientID := vars["clientId"]
	menuItemIDStr := vars["menuItemId"]
	menuItemID, err := strconv.Atoi(menuItemIDStr)
	if err != nil {
		http.Error(w, "Invalid menuItemID", http.StatusBadRequest)
		return
	}

	err = r.ParseMultipartForm(10 << 20) // 10 MB limit for form data
	if err != nil {
		http.Error(w, "Error in parsing form data.", http.StatusBadRequest)
		return
	}

	// Find the existing menu item
	var existingMenuItem user1.MenuItem
	result := connection.Where("client_id = ? AND id = ?", clientID, menuItemID).First(&existingMenuItem)
	if result.Error != nil {
		http.Error(w, "Menu item not found", http.StatusNotFound)
		return
	}

	// Update menu item fields
	itemName := r.FormValue("itemName")
	if itemName != "" && itemName != existingMenuItem.ItemName {
		// Check if the new item name is already in use by another menu item
		var existingMobileClient user1.MenuItem
		result := connection.Where("client_id = ? AND item_name = ?", clientID, itemName).First(&existingMobileClient)
		if result.Error == nil && existingMobileClient.ID != existingMenuItem.ID {
			http.Error(w, "Item name already in use by another menu item", http.StatusBadRequest)
			return
		}
		// Update the item name
		existingMenuItem.ItemName = itemName
	}

	// Update food_type
	foodType := r.FormValue("food_type")
	if foodType != "" && foodType != existingMenuItem.Food_type {
		// Check if the food type is "veg" or "non-veg"
		if foodType == "veg" || foodType == "non-veg" {
			existingMenuItem.Food_type = foodType
		} else {
			http.Error(w, "Invalid food type. Should be 'veg' or 'non-veg'", http.StatusBadRequest)
			return
		}
	}
	sub_category := r.FormValue("sub_category")
	if sub_category != "" && sub_category != existingMenuItem.Sub_category {
		existingMenuItem.Sub_category = sub_category
	}
	description := r.FormValue("description")
	if description != "" && description != existingMenuItem.Description {
		existingMenuItem.Description = description
	}
	
	newTimeStr := r.FormValue("time")
	if newTimeStr != "" {
		newTime, err := strconv.Atoi(newTimeStr)
		if err != nil {
			http.Error(w, "Invalid time format", http.StatusBadRequest)
			return
		}
		existingMenuItem.Time = newTime
	}

	// Update menu item fields
	newPriceStr := r.FormValue("price")
	if newPriceStr != "" {
		newPrice, err := strconv.ParseFloat(newPriceStr, 64)
		if err != nil {
			http.Error(w, "Invalid price format", http.StatusBadRequest)
			return
		}
		if newPrice != existingMenuItem.Price {
			existingMenuItem.Price = newPrice

		}
	}
	offerStr := r.FormValue("offer")
	if offerStr != "" {
		newOffer, err := strconv.ParseFloat(offerStr, 64)
		if err != nil {
			http.Error(w, "Invalid offer format", http.StatusBadRequest)
			return
		}
		if newOffer != existingMenuItem.Offer {
			existingMenuItem.Offer = newOffer

			// Calculate discounted price using the updated offer
			discountedPrice := existingMenuItem.Price - (existingMenuItem.Price * (existingMenuItem.Offer / 100))
			existingMenuItem.OfferRate = discountedPrice
		}
	}

	statusStr := r.FormValue("status")
	if statusStr != "" {
		status, err := strconv.ParseBool(statusStr)
		if err != nil {
			http.Error(w, "Invalid status value", http.StatusBadRequest)
			return
		}
		existingMenuItem.Status = status
	}
	bannerStr := r.FormValue("banner")
	if bannerStr != "" {
		banner, err := strconv.ParseBool(bannerStr)
		if err != nil {
			http.Error(w, "Invalid status value", http.StatusBadRequest)
			return
		}
		existingMenuItem.Banner = banner
	}


	recommendationStr := r.FormValue("recommendation")
	if recommendationStr != "" {
		recommendation, err := strconv.ParseBool(recommendationStr)
		if err != nil {
			http.Error(w, "Invalid recommendation value", http.StatusBadRequest)
			return
		}
		existingMenuItem.Recommendation = recommendation
	}



	temporaryStatusStr := r.FormValue("temporary_status")
    if temporaryStatusStr != "" {
    temporaryStatus, err := strconv.ParseBool(temporaryStatusStr)
    if err != nil {
        http.Error(w, "Invalid temporary_status value. Should be a boolean", http.StatusBadRequest)
        return
    }
    existingMenuItem.Temporary_status = temporaryStatus
   
    }

	
	// Handle Image Upload
	var imageFilePath string
	if imageFile, header, err := r.FormFile("image"); err == nil {
		defer imageFile.Close()

		// Delete old image
		if existingMenuItem.Image != "" {
			oldImagePath := strings.TrimPrefix(existingMenuItem.Image, "http://"+r.Host+"/")
			err = os.Remove(oldImagePath)
          // Handle error if needed
		}

		// Generate a unique filename for the new image
		uniqueFileName, err := uuid.NewV4() // Generate a UUID
		if err != nil {
			http.Error(w, "Error generating UUID.", http.StatusInternalServerError)
			return
		}

		// Create the complete file path
		ext := filepath.Ext(header.Filename)
		imageFilePath = filepath.Join("uploads", uniqueFileName.String()+ext)

		// Create the file on the server
		outputFile, err := os.Create(imageFilePath)
		if err != nil {
			http.Error(w, "Error creating image file.", http.StatusInternalServerError)
			return
		}
		defer outputFile.Close()

		// Copy the new image data to the file
		_, err = io.Copy(outputFile, imageFile)
		if err != nil {
			http.Error(w, "Error copying new image data.", http.StatusInternalServerError)
			return
		}

		// Construct the URL for the new served image
		serverAddr := r.Host
		imageServeURL := "http://" + serverAddr + "/" + imageFilePath
		
		existingMenuItem.Imagepath = imageFilePath
		existingMenuItem.Image = imageServeURL
	}
	result = connection.Save(&existingMenuItem)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(existingMenuItem)
}

func UpdateMenuItemBanner(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	vars := mux.Vars(r)
	clientID := vars["clientId"]
	menuItemIDStr := vars["menuItemId"]
	menuItemID, err := strconv.Atoi(menuItemIDStr)
	if err != nil {
		http.Error(w, "Invalid menuItemID", http.StatusBadRequest)
		return
	}

	err = r.ParseMultipartForm(10 << 20) // 10 MB limit for form data
	if err != nil {
		http.Error(w, "Error in parsing form data.", http.StatusBadRequest)
		return
	}

	// Find the existing menu item
	var existingMenuItem user1.MenuItem
	result := connection.Where("client_id = ? AND id = ?", clientID, menuItemID).First(&existingMenuItem)
	if result.Error != nil {
		http.Error(w, "Menu item not found", http.StatusNotFound)
		return
	}

	// Update menu item fields
	var itemName string

	if itemName = r.FormValue("itemName"); itemName != "" {
			
		if itemName != "" && itemName != existingMenuItem.ItemName {
			// Check if the new item name is already in use by another menu item
			var existingMobileClient user1.MenuItem
			result := connection.Where("client_id = ? AND item_name = ? AND category_name = ? ", clientID, itemName , existingMenuItem.CategoryName).First(&existingMobileClient)
			if result.Error == nil && existingMobileClient.ID != existingMenuItem.ID {
				http.Error(w, "Item name already in use by another menu item", http.StatusBadRequest)
				return
			}
			// Update the item name
			existingMenuItem.ItemName = itemName
		}
	} else {
		// If "itemName" is not provided, fallback to "name"
		existingMenuItem.ItemName = r.FormValue("name")
	}


	// Update food_type
	foodType := r.FormValue("food_type")
	if foodType != "" && foodType != existingMenuItem.Food_type {
		// Check if the food type is "veg" or "non-veg"
		if foodType == "veg" || foodType == "non-veg" {
			existingMenuItem.Food_type = foodType
		} else {
			http.Error(w, "Invalid food type. Should be 'veg' or 'non-veg'", http.StatusBadRequest)
			return
		}
	}
	sub_category := r.FormValue("sub_category")
	if sub_category != "" && sub_category != existingMenuItem.Sub_category {
		existingMenuItem.Sub_category = sub_category
	}
	description := r.FormValue("description")
	if description != "" && description != existingMenuItem.Description {
		existingMenuItem.Description = description
	}
	
	newTimeStr := r.FormValue("time")
	if newTimeStr != "" {
		newTime, err := strconv.Atoi(newTimeStr)
		if err != nil {
			http.Error(w, "Invalid time format", http.StatusBadRequest)
			return
		}
		existingMenuItem.Time = newTime
	}

	// Update menu item fields
	newPriceStr := r.FormValue("price")
	if newPriceStr != "" {
		newPrice, err := strconv.ParseFloat(newPriceStr, 64)
		if err != nil {
			http.Error(w, "Invalid price format", http.StatusBadRequest)
			return
		}
		if newPrice != existingMenuItem.Price {
			existingMenuItem.Price = newPrice

		}
	}
	offerStr := r.FormValue("offer")
	if offerStr != "" {
		newOffer, err := strconv.ParseFloat(offerStr, 64)
		if err != nil {
			http.Error(w, "Invalid offer format", http.StatusBadRequest)
			return
		}
		if newOffer != existingMenuItem.Offer {
			existingMenuItem.Offer = newOffer

			// Calculate discounted price using the updated offer
			discountedPrice := existingMenuItem.Price - (existingMenuItem.Price * (existingMenuItem.Offer / 100))
			existingMenuItem.OfferRate = discountedPrice
		}
	}

	statusStr := r.FormValue("status")
	if statusStr != "" {
		status, err := strconv.ParseBool(statusStr)
		if err != nil {
			http.Error(w, "Invalid status value", http.StatusBadRequest)
			return
		}
		existingMenuItem.Status = status
	}
	bannerStr := r.FormValue("banner")
	if bannerStr != "" {
		banner, err := strconv.ParseBool(bannerStr)
		if err != nil {
			http.Error(w, "Invalid status value", http.StatusBadRequest)
			return
		}
		existingMenuItem.Banner = banner
	}


	recommendationStr := r.FormValue("recommendation")
	if recommendationStr != "" {
		recommendation, err := strconv.ParseBool(recommendationStr)
		if err != nil {
			http.Error(w, "Invalid recommendation value", http.StatusBadRequest)
			return
		}
		existingMenuItem.Recommendation = recommendation
	}



	temporaryStatusStr := r.FormValue("temporary_status")
    if temporaryStatusStr != "" {
    temporaryStatus, err := strconv.ParseBool(temporaryStatusStr)
    if err != nil {
        http.Error(w, "Invalid temporary_status value. Should be a boolean", http.StatusBadRequest)
        return
    }
    existingMenuItem.Temporary_status = temporaryStatus
   
    }

	
	// Handle Image Upload
	var imageFilePath string
	if imageFile, header, err := r.FormFile("image"); err == nil {
		defer imageFile.Close()

		// Delete old image
		if existingMenuItem.Image != "" {
			oldImagePath := strings.TrimPrefix(existingMenuItem.Image, "http://"+r.Host+"/")
			err = os.Remove(oldImagePath)
          // Handle error if needed
		}

		// Generate a unique filename for the new image
		uniqueFileName, err := uuid.NewV4() // Generate a UUID
		if err != nil {
			http.Error(w, "Error generating UUID.", http.StatusInternalServerError)
			return
		}

		// Create the complete file path
		ext := filepath.Ext(header.Filename)
		imageFilePath = filepath.Join("uploads", uniqueFileName.String()+ext)

		// Create the file on the server
		outputFile, err := os.Create(imageFilePath)
		if err != nil {
			http.Error(w, "Error creating image file.", http.StatusInternalServerError)
			return
		}
		defer outputFile.Close()

		// Copy the new image data to the file
		_, err = io.Copy(outputFile, imageFile)
		if err != nil {
			http.Error(w, "Error copying new image data.", http.StatusInternalServerError)
			return
		}

		// Construct the URL for the new served image
		serverAddr := r.Host
		imageServeURL := "http://" + serverAddr + "/" + imageFilePath
		
		existingMenuItem.Imagepath = imageFilePath
		existingMenuItem.Image = imageServeURL
	}
	result = connection.Save(&existingMenuItem)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(existingMenuItem)
}
