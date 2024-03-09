package client

import (
	"encoding/json"
	"fmt"
    
	// "encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sathya-narayanan23/crudapp/database"
	"sathya-narayanan23/crudapp/users/user"
    
	"sathya-narayanan23/handler/mail"
	"sort"
	"strings"
	"time"
    "strconv"
    
	"github.com/gorilla/mux"
	// Import other necessary packages
)


type Error struct {
	IsError bool   `json:"isError"`
	Message string `json:"message"`
}

func SignUp(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    connection := database.GetDatabase()
    defer database.CloseDatabase(connection)

    err := r.ParseMultipartForm(10 << 20) // 10 MB limit for form data
    if err != nil {
        http.Error(w, "Error in parsing form data.", http.StatusBadRequest)
        return
    }

    // Parse other form fields
    name := r.FormValue("name")
    mobileNumber := r.FormValue("mobileNumber")
    displayName := r.FormValue("displayName")
    secondaryMobileNumber := r.FormValue("secondaryMobileNumber")
    selectedCategories := strings.Split(r.FormValue("selectedCategories"), ",")
    email := r.FormValue("email")
    secondaryEmail := r.FormValue("secondaryEmail")
    country := r.FormValue("country")
    paymentMethod := r.FormValue("paymentMethod")
    password := r.FormValue("password")
    state := r.FormValue("state")
    district := r.FormValue("district")
    plan := r.FormValue("plan")
    upi := r.FormValue("upi")
    upiName := r.FormValue("upiName")
    primaryColour := r.FormValue("primaryColour")
    specialist := r.FormValue("specialist")
    currency := r.FormValue("currency")

    // Validate required fields
    if name == "" || mobileNumber == "" || len(selectedCategories) == 0 || email == "" || plan == "" {
        http.Error(w, "Name, Mobile number, password, email, plan, and selectedCategories are required fields.", http.StatusBadRequest)
        return
    }

    var planDurations = map[string]int{
        "fretrail":3,
        "Monthly": 30, // 30 days
        "6Month":  180,
        "12Month": 360,
    }

    // Check if the plan is valid and get its duration
    duration, valid := planDurations[plan]
    if !valid {
        http.Error(w, "Invalid plan selected.", http.StatusBadRequest)
        return
    }

    // Calculate plan expiration date
    planExpiration := time.Now().AddDate(0, 0, duration)

    // Calculate notification date (7 days before plan expiration)
    notificationDate := planExpiration.AddDate(0, 0, -7)

    // Create the client record
    client := user1.Client{
        Name:                  name,
        MobileNumber:          mobileNumber,
        Email:                 email,
        Country:               country,
        DisplayName:           displayName,
        State:                 state,
        SecondaryMobileNumber: secondaryMobileNumber,
        SecondaryEmail:        secondaryEmail,
        Plan:                  plan,
        Upi:                   upi,
        UpiName:               upiName,
        Currency:              currency,
        Specialist:            specialist,
        PrimaryColour:         primaryColour,
        District:              district,
        PaymentMethod:         paymentMethod,
        Password:              password,
        SelectedCategories:    selectedCategories,
        Status:                true,
        PlanCreateTime:        time.Now(),
        PlanUpdateTime:        time.Now(),
        PlanExpiration:        planExpiration,
        NotificationDate:      notificationDate,
    }

    // Upload the logo image
    logoFile, logoHeader, err := r.FormFile("logo")
    if err != nil {
        http.Error(w, "Error uploading logo.", http.StatusBadRequest)
        return
    }
    defer logoFile.Close()

    // Create directory if it doesn't exist
    os.MkdirAll("uploads/logos", os.ModePerm)

    // Generate logo file path
    logoPath := filepath.Join("uploads", "logos", client.MobileNumber+"_logo"+client.Name+filepath.Ext(logoHeader.Filename))
    logoFileOnDisk, err := os.Create(logoPath)
    if err != nil {
        http.Error(w, "Error creating logo file.", http.StatusInternalServerError)
        return
    }
    defer logoFileOnDisk.Close()

    // Save logo file to disk
    _, err = io.Copy(logoFileOnDisk, logoFile)
    if err != nil {
        http.Error(w, "Error copying logo data.", http.StatusInternalServerError)
        return
    }

    // Construct image URL
    serverAddr := r.Host
    logoURL := "http://" + serverAddr + "/" + logoPath
    client.Logoimagepath = logoPath
    client.Logo = logoURL

    // Create client record in the database
    if err := connection.Create(&client).Error; err != nil {
        http.Error(w, "Error creating client record.", http.StatusInternalServerError)
        return
    }

    // // Send email notification
    // subject := "YOGRA LICENSE"
    // body := "Your account has been created successfully."
    // sendEmailWithAttachment(client.Email, subject, body, client.FilePath)

    // Default categories
    defaultCategories := map[string]string{
        "breakfast":     "http://" + serverAddr + "/image/breakfast.png",
        "dinner":        "http://" + serverAddr + "/image/dinner.png",
        "hot_drinks":    "http://" + serverAddr + "/image/hot_drinks.png",
        "Icecreams":     "http://" + serverAddr + "/image/Icecreams.png",
        "juices_shakes": "http://" + serverAddr + "/image/juices_shakes.png",
        "lunch":         "http://" + serverAddr + "/image/lunch.png",
        "snacks":        "http://" + serverAddr + "/image/snacks.png",
        "water":         "http://" + serverAddr + "/image/water.png",
    }

    // Create and save categories
    categoryStatusActive := true
    categoryStatusInactive := false
    createdCategories := []user1.Category{}

    for _, selectedCategory := range selectedCategories {
        if image, exists := defaultCategories[selectedCategory]; exists {
            category := user1.Category{
                ClientID:     client.ID,
                CategoryName: selectedCategory,
                Image:        image,
                ImagefilePath: "/image/" + selectedCategory + ".png",
                Active_status: categoryStatusActive,
                Status:       false, // Set to "active" for selected categories
            }
            if err := connection.Create(&category).Error; err != nil {
                http.Error(w, "Error creating category record.", http.StatusInternalServerError)
                return
            }
            createdCategories = append(createdCategories, category)
        }
    }

    // Create categories for non-selected items with status "inactive"
    for categoryName, image := range defaultCategories {
        if !contains(selectedCategories, categoryName) {
            category := user1.Category{
                ClientID:     client.ID,
                CategoryName: categoryName,
                Image:        image,
                ImagePath: "/image/" + (categoryName) + ".png",
                Status:       categoryStatusInactive, // Set to "inactive" for non-selected categories
            }
            if err := connection.Create(&category).Error; err != nil {
                http.Error(w, "Error creating category record.", http.StatusInternalServerError)
                return
            }
            createdCategories = append(createdCategories, category)
        }
    }

    // Response
    response := struct {
        Client            user1.Client     `json:"client"`
        CreatedCategories []user1.Category `json:"createdCategories"`
    }{
        Client:            client,
        CreatedCategories: createdCategories,
    }

    // Send response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// Helper function to check if a string exists in a slice of strings
func contains(s []string, str string) bool {
    for _, v := range s {
        if v == str {
            return true
        }
    }
    return false
}




func GetAllClient(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    connection := database.GetDatabase()
    defer database.CloseDatabase(connection)

    // Execute the query
    var clients []user1.Client
    type SearchCriteria struct {
        Name          string `json:"name"`
        Upi           string `json:"upi"`
        MobileNumber  string `json:"mobile_number"`
        Email         string `json:"email"`
        Plan          string `json:"plan"`
        DisplayName   string `json:"display_name"`
        Country       string `json:"country"`
        State         string `json:"state"`
    }

    // Parse the query parameters to extract the search criteria
    queryValues := r.URL.Query()
    searchCriteria := SearchCriteria{
        Name:         queryValues.Get("name"),
        Upi:          queryValues.Get("upi"),
        MobileNumber: queryValues.Get("mobile_number"),
        Email:        queryValues.Get("email"),
        Plan:         queryValues.Get("plan"),
        DisplayName:  queryValues.Get("display_name"),
        Country:      queryValues.Get("country"),
        State:        queryValues.Get("state"),
    }

    // Build the query based on the provided criteria
    query := connection
 	// query := connection.Clients("id")
    if searchCriteria.Upi != "" {
        query = query.Where("upi LIKE ?", "%"+searchCriteria.Upi+"%")
    }
    if searchCriteria.Name != "" {
        query = query.Where("name LIKE ?", "%"+searchCriteria.Name+"%")
    }
    if searchCriteria.MobileNumber != "" {
        query = query.Where("mobile_number LIKE ?", "%"+searchCriteria.MobileNumber+"%")
    }
    if searchCriteria.Email != "" {
        query = query.Where("email LIKE ?", "%"+searchCriteria.Email+"%")
    }
    if searchCriteria.Plan != "" {
        query = query.Where("plan LIKE ?", "%"+searchCriteria.Plan+"%")
    }
    if searchCriteria.DisplayName != "" {
        query = query.Where("display_name LIKE ?", "%"+searchCriteria.DisplayName+"%")
    }
    if searchCriteria.Country != "" {
        query = query.Where("country LIKE ?", "%"+searchCriteria.Country+"%")
    }
    if searchCriteria.State != "" {
        query = query.Where("state LIKE ?", "%"+searchCriteria.State+"%")
    }
	
    // query = query.Order("id")

    result := query.Find(&clients)
    if result.Error != nil {
        http.Error(w, result.Error.Error(), http.StatusInternalServerError)
        return
    }
	// Sort clients by ID
	sort.Slice(clients, func(i, j int) bool {
		return clients[i].ID < clients[j].ID
	})

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(clients)
}



func GetAllClientsWithCategories(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    // Get a database connection
    connection := database.GetDatabase()
    defer database.CloseDatabase(connection)

    // Initialize a map to store clients and their categories
    // clientCategories := make(map[uint]*Client)

    // Initialize a variable to store all clients
    var clients []user1.Client

    // Find all clients and preload their associated categories
    result := connection.Preload("Categories").Find(&clients)
    if result.Error != nil {
        http.Error(w, "Error fetching clients with categories", http.StatusInternalServerError)
        return
    }

	 sort.Slice(clients, func(i, j int) bool {
        return clients[i].ID < clients[j].ID
    })

    // Respond with the assembled client information
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(clients)
}


func GetAllClientplannew(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    connection := database.GetDatabase()
    defer database.CloseDatabase(connection)

    var clients []user1.Client
    connection.Order("id asc").Find(&clients)

    var trueClients, falseClients []user1.Client
    planClients := make(map[string][]user1.Client)
    planCounts := make(map[string]int)

    // Categorize clients based on their status and plans
    for _, client := range clients {
        if client.Status {
            trueClients = append(trueClients, client)
        } else {
            falseClients = append(falseClients, client)
        }

        planName := client.Plan
        planClients[planName] = append(planClients[planName], client)
        planCounts[planName]++
    }

    totalClientCount := len(clients)

    // Create a result map with true and false status lists, and counts and lists for each plan
    result := map[string]interface{}{
        "totalClientCount": totalClientCount,
        "trueClientCount":  len(trueClients),
        // "trueClients":      trueClients,
        "falseClientCount": len(falseClients),
        // "falseClients":     falseClients,
        // "planClients":      planClients,
        "planCounts":       planCounts,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}

func GetAllClientplan(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    connection := database.GetDatabase()
    defer database.CloseDatabase(connection)

    var clients []user1.Client
    connection.Order("id asc").Find(&clients)

    var trueClients, falseClients []user1.Client
    planClients := make(map[string][]user1.Client)
    planCounts := make(map[string]int)

    // Categorize clients based on their status and plans
    for _, client := range clients {
        if client.Status {
            trueClients = append(trueClients, client)
        } else {
            falseClients = append(falseClients, client)
        }

        planName := client.Plan
        planClients[planName] = append(planClients[planName], client)
        planCounts[planName]++
    }

    totalClientCount := len(clients)

    // Create a result map with true and false status lists, and counts and lists for each plan
    result := map[string]interface{}{
        "totalClientCount": totalClientCount,
        "trueClientCount":  len(trueClients),
        "trueClients":      trueClients,
        "falseClientCount": len(falseClients),
        "falseClients":     falseClients,
        "planClients":      planClients,
        "planCounts":       planCounts,
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}





func GetClient(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    vars := mux.Vars(r)
    clientIDStr, ok := vars["clientID"]
    if !ok {
        http.Error(w, "Client ID not provided in URL", http.StatusBadRequest)
        return
    }

    clientID, err := strconv.Atoi(clientIDStr)
    if err != nil {
        http.Error(w, "Invalid client ID", http.StatusBadRequest)
        return
    }
    connection := database.GetDatabase()
	defer database.CloseDatabase(connection)
    
	var client user1.Client
	connection.Preload("Categories").First(&client, clientID) // Preload Categories relationship
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(client)
}


func UpdateClient(w http.ResponseWriter, r *http.Request) {
	
    if r.Method != http.MethodPut {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
	params := mux.Vars(r)
	clientIDStr := params["id"]
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		http.Error(w, "Invalid client ID", http.StatusBadRequest)
		return
	}
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	// Retrieve existing client record
	var existingClient user1.Client
	result := connection.First(&existingClient, clientID)
	if result.Error != nil {
		http.Error(w, "Client not found", http.StatusNotFound)
		return
	}

	email := r.FormValue("email")
	if email != "" {
		existingClient.Email = email
	}
	paymentMethod := r.FormValue("paymentMethod")
	if paymentMethod != "" {
		existingClient.PaymentMethod = paymentMethod
	}
	secondaryEmail := r.FormValue("secondaryEmail")
	if secondaryEmail != "" {
		existingClient.SecondaryEmail = secondaryEmail
	}
	displayName := r.FormValue("displayName")
	if displayName != "" {
		existingClient.DisplayName = displayName
	}
	secondaryMobileNumber := r.FormValue("secondaryMobileNumber")
	if secondaryMobileNumber != "" {
		existingClient.SecondaryMobileNumber = secondaryMobileNumber
	}

	country := r.FormValue("country")
	if country != "" {
		existingClient.Country = country
	}
	upi := r.FormValue("upi")
	if upi != "" {
		existingClient.Upi = upi
	}
	upiName := r.FormValue("upiName")
	if upiName != "" {
		existingClient.UpiName = upiName
	}
	state := r.FormValue("state")
	if state != "" {
		existingClient.State = state
	}
	district := r.FormValue("district")
	if district != "" {
		existingClient.District = district
	}

	primaryColour := r.FormValue("primaryColour")
	if primaryColour != "" {
		existingClient.PrimaryColour = primaryColour
	}
	specialist := r.FormValue("specialist")
	if specialist != "" {
		existingClient.Specialist = specialist
	}

	name := r.FormValue("name")
	if name != "" {
		existingClient.Name = name
	}
	mobileNumber := r.FormValue("mobileNumber")
	if name != "" {
		existingClient.MobileNumber = mobileNumber
	}
	// mobileNumber := r.FormValue("mobileNumber")
	// if mobileNumber != "" && mobileNumber != existingClient.MobileNumber {
	// 	var existingMobileClient Client
	// 	result := connection.Where("mobile_number = ?", mobileNumber).First(&existingMobileClient)
	// 	if result.Error == nil {
	// 		http.Error(w, "Mobile number already in use by another client", http.StatusBadRequest)
	// 		return
	// 	}
	// }

	// var planDurations = map[string]int{
	// 	"free trial":  10,  // 10 days
	// 	"Basic":       30,  // 30 days
	// 	"Premium":     180, // 190 days
	// 	"Pro premium": 360, // 360 days
	// }

	var planDurations = map[string]int{
		"Monthly": 30, // 30 days
		"6Month":  180,
		"12Month": 360,
	}
	plan := r.FormValue("plan")
	if plan != "" && plan != existingClient.Plan {
		// Check if the plan exists in the planDurations map
		duration, valid := planDurations[plan]
		if !valid {
			http.Error(w, "Invalid plan selected", http.StatusBadRequest)
			return
		}
		var b time.Time
		if existingClient.PlanExpiration.After(time.Now()) {
			b = existingClient.PlanExpiration
		} else {
			b = time.Now()
		}

		// Calculate the new plan expiration time by adding the duration of the new plan to the larger time
		newPlanExpiration := b.Add(time.Duration(duration) * 24 * time.Hour)

		existingClient.Plan = plan
		existingClient.PlanUpdateTime = time.Now()

		// Check if the new planExpiration is later than the current planExpiration
		if newPlanExpiration.After(existingClient.PlanExpiration) {
			existingClient.PlanExpiration = newPlanExpiration

			// Calculate notification date (7 days before plan expiration)
			existingClient.NotificationDate = newPlanExpiration.AddDate(0, 0, -7)
		}

		errmail := mail.SendPlanExpirationEmail(existingClient.Email, existingClient.Plan, existingClient.PlanExpiration)
		if errmail != nil {
			// Handle the email sending error, you can log it or take other actions
			fmt.Println("Error sending plan expiration email:", err)
		}

	} else if plan == existingClient.Plan {
		// If the selected plan is the same as the existing plan, add the duration to the current PlanExpiration
		duration, valid := planDurations[plan]
		if !valid {
			http.Error(w, "Invalid plan selected", http.StatusBadRequest)
			return
		}

		var b time.Time
		if existingClient.PlanExpiration.After(time.Now()) {
			b = existingClient.PlanExpiration
		} else {
			b = time.Now()
		}

		// Calculate the new plan expiration time by adding the duration of the new plan to the larger time
		newPlanExpiration := b.Add(time.Duration(duration) * 24 * time.Hour)

		existingClient.Plan = plan
		existingClient.PlanUpdateTime = time.Now()

		// Check if the new planExpiration is later than the current planExpiration
		if newPlanExpiration.After(existingClient.PlanExpiration) {
			existingClient.PlanExpiration = newPlanExpiration

			// Calculate notification date (7 days before plan expiration)
			existingClient.NotificationDate = newPlanExpiration.AddDate(0, 0, -7)
		}

		errmail := mail.SendPlanExpirationEmail(existingClient.Email, existingClient.Plan, existingClient.PlanExpiration)
		if errmail != nil {
			// Handle the email sending error, you can log it or take other actions
			fmt.Println("Error sending plan expiration email:", err)
		}

	}

	// password := r.FormValue("password")
	// if password != "" {
	// 	existingClient.Password = password
	// }

	statusStr := r.FormValue("status")
	if statusStr != "" {
		status, err := strconv.ParseBool(statusStr)
		if err != nil {
			http.Error(w, "Invalid status value", http.StatusBadRequest)
			return
		}
		existingClient.Status = status
	}

	// Upload the new logo image
	newLogoFile, newLogoHeader, err := r.FormFile("logo")
	if err == nil {
		defer newLogoFile.Close()


		// Delete old logo image
		oldLogoPath := strings.TrimPrefix(existingClient.Logo, "http://"+r.Host)
		err = os.Remove(oldLogoPath)
		if err != nil {
			// Handle error if needed
		}

		// Save the new logo image
		newLogoPath := filepath.Join("uploads", "logos", strconv.Itoa(int(existingClient.ID))+"_logo"+existingClient.Name+filepath.Ext(newLogoHeader.Filename))
		newLogoFileOnDisk, err := os.Create(newLogoPath)
		if err != nil {
			http.Error(w, "Error creating new logo file.", http.StatusInternalServerError)
			return
		}
		defer newLogoFileOnDisk.Close()

		_, err = io.Copy(newLogoFileOnDisk, newLogoFile)
		if err != nil {
			http.Error(w, "Error copying new logo data.", http.StatusInternalServerError)
			return
		}

		// Construct new image URL
		serverAddr := r.Host
		
		logoURL := "http://" + serverAddr + "/" + newLogoPath
		// logoURL := "http://" + serverAddr + "/image/" + strings.ReplaceAll(newLogoPath, "/", "%5C")
		existingClient.Logo = logoURL
		existingClient.Logoimagepath = newLogoPath
		
	}

	serverAddr := r.Host
	// var createdCategories []Category

	selectedCategories := strings.Split(r.FormValue("selectedCategories"), ",")
	if len(selectedCategories) > 0 {
		defaultCategories := map[string]string{
		
			"breakfast":      "http://" + serverAddr + "/image/" + "breakfast.png",
			"dinner":         "http://" + serverAddr + "/image/" + "dinner.png",
			"hot_drinks":     "http://" + serverAddr + "/image/" + "hot_drinks.png",
			"Icecreams":      "http://" + serverAddr + "/image/" + "Icecreams.png",
			"juices_shakes":  "http://" + serverAddr + "/image/" + "juices_shakes.png",
			"lunch":          "http://" + serverAddr + "/image/" + "lunch.png",
			"snacks":         "http://" + serverAddr + "/image/" + "snacks.png",
			"water":          "http://" + serverAddr + "/image/" + "water.png",
		}
		

		// Deactivate all existing categories for this client
		result := connection.Model(&user1.Category{}).Where("client_id = ?", existingClient.ID).Update("Active_status", false)
		if result.Error != nil {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
			return
		}

		createdCategories := []user1.Category{}
		for _, selectedCategory := range selectedCategories {
			if image, exists := defaultCategories[selectedCategory]; exists {
				// Find existing category with the same name and update it
				existingCategory := user1.Category{}
				result := connection.Where("client_id = ? AND category_name = ?", existingClient.ID, selectedCategory).First(&existingCategory)
				if result.Error == nil {
					// Update the category image and set status to "active"
					existingCategory.Image = image
					existingCategory.Active_status = true
					connection.Save(&existingCategory)
					createdCategories = append(createdCategories, existingCategory)
				} else {
					// Create a new category if it doesn't exist and set status to "active"
					category := user1.Category{
						ClientID:     existingClient.ID,
						CategoryName: selectedCategory,
						Image:        image,
						
					    // ImagePath: "/image/" + (selectedCategory) + ".png",
						Active_status:       true,
					}
					connection.Create(&category)
					createdCategories = append(createdCategories, category)
				}
			} else {
				fmt.Printf("Selected category '%s' not found in defaultCategories\n", selectedCategory)
			}
		}
	}
	result = connection.Save(&existingClient)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}
	
	// Construct response
	// response := struct {
	// 	Client            Client     `json:"client"`
	// 	Categories []Category `json:"newresults"`

	// 	CreatedCategories []Category `json:"createdCategories"`
	// }{
	// 	Client:            existingClient,
	// 	CreatedCategories: createdCategories,
	// }
	newresults := connection.Preload("Categories").Find(&existingClient)
    if newresults.Error != nil {
        http.Error(w, "Error fetching clients with categories", http.StatusInternalServerError)
        return
    }
	// fmt.Println(createdCategories)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(existingClient)

}


func DeleteClient(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	vars := mux.Vars(r)
	clientID := vars["id"]

	var client user1.Client
	if err := connection.First(&client, clientID).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(Error{IsError: true, Message: "Client not found"})
		return
	}

	// Delete the categories associated with the client ID
	if err := connection.Where("client_id = ?", clientID).Delete(&user1.Category{}).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{IsError: true, Message: "Failed to delete client categories"})
		return
	}

	// Delete the menu items associated with the client's categories
	var categoryIDs []uint
	connection.Table("categories").Select("id").Where("client_id = ?", clientID).Pluck("id", &categoryIDs)
	if len(categoryIDs) > 0 {
		if err := connection.Where("category_id IN (?)", categoryIDs).Delete(&user1.MenuItem{}).Error; err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Error{IsError: true, Message: "Failed to delete menu items"})
			return
		}
	}

	// Delete the TableNo records associated with the client ID
	if err := connection.Where("client_id = ?", clientID).Delete(&user1.TableNo{}).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{IsError: true, Message: "Failed to delete client TableNo records"})
		return
	}

	if err := connection.Delete(&client).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{IsError: true, Message: "Failed to delete client"})
		return
	}

	responseMessage := fmt.Sprintf("Client with ID %s has been deleted", clientID)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": responseMessage,
	})

	w.WriteHeader(http.StatusNoContent)
}



func GetClientsExpiringWithin7Days() ([]user1.Client, error) {
    var clients []user1.Client
    connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

    // Calculate expiration date 7 days from now
    expirationDate := time.Now().Add(7 * 24 * time.Hour)

    // Query clients whose plan expiration falls within the next 7 days
    if err := connection.Where("plan_expiration <= ?", expirationDate).Find(&clients).Error; err != nil {
        return nil, err
    }

    return clients, nil
}
func Getafter7Days(w http.ResponseWriter, r *http.Request) {
    connection := database.GetDatabase()
    defer database.CloseDatabase(connection)

    // Calculate the date 7 days from now
    sevenDays := time.Now().AddDate(0, 0, 7)

    var clients []user1.Client
    result := connection.Where("plan_expiration BETWEEN ? AND ?", time.Now(), sevenDays).Order("plan_expiration desc").Find(&clients)
    if result.Error != nil {
        http.Error(w, result.Error.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(clients)
}


func GetAllClientplannewwww(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    connection := database.GetDatabase()
    defer database.CloseDatabase(connection)
    
    // sevenDaysLater := time.Now().AddDate(0, 0, 7)

    var clients []user1.Client
    // connection.Order("id asc").Find(&clients)
    // result := connection.Where("plan_expiration BETWEEN ? AND ?", time.Now(), sevenDaysLater).Find(&clients)
    
    sevenDaysLaterUTC := time.Now().UTC().AddDate(0, 0, 7)
    result := connection.Where("plan_expiration BETWEEN ? AND ?", time.Now().UTC(), sevenDaysLaterUTC).Find(&clients)

    if result.Error != nil {
        http.Error(w, result.Error.Error(), http.StatusInternalServerError)
        return
    }
    // errmail := mail.SendPlanExpirationEmail(existingClient.Email, existingClient.Plan, existingClient.PlanExpiration)
	// 	if errmail != nil {
	// 		// Handle the email sending error, you can log it or take other actions
	// 		fmt.Println("Error sending plan expiration email:", err)
	// 	}
    for _, client := range clients {
        if err := mail.SendEmailNotificationnew(client.Email,client.Plan,client.PlanExpiration); err != nil {
            // Handle error sending email for this client
            fmt.Println("Error sending email to", client.Email, ":", err)
        }
    }

    expirationWiseClients := make(map[string][]user1.Client)

    // Categorize clients based on their status, plans, and plan expiration
    for _, client := range clients {
        expirationDate := client.PlanExpiration.Format("2006-01-02") // Format expiration date as "YYYY-MM-DD"
        expirationWiseClients[expirationDate] = append(expirationWiseClients[expirationDate], client)
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(expirationWiseClients)
}

func GetAllClientplannewww() {
    connection := database.GetDatabase()
    defer database.CloseDatabase(connection)

    var clients []user1.Client
    sevenDaysLaterUTC := time.Now().UTC().AddDate(0, 0, 7)
    result := connection.Where("plan_expiration BETWEEN ? AND ?", time.Now().UTC(), sevenDaysLaterUTC).Find(&clients)

    if result.Error != nil {
        fmt.Println("Error fetching clients:", result.Error)
        return
    }

    for _, client := range clients {
        if err := mail.SendEmailNotificationnew(client.Email, client.Plan, client.PlanExpiration); err != nil {
            fmt.Println("Error sending email to", client.Email, ":", err)
        }
    }

    expirationWiseClients := make(map[string][]user1.Client)

    for _, client := range clients {
        expirationDate := client.PlanExpiration.Format("2006-01-02")
        expirationWiseClients[expirationDate] = append(expirationWiseClients[expirationDate], client)
    }

    // You might want to handle the errors from JSON encoding here
    // For simplicity, error handling is omitted in this example
    jsonResponse, _ := json.Marshal(expirationWiseClients)

    // Assuming you want to print the JSON response
    fmt.Println(string(jsonResponse))
}

