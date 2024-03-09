package banner


import (
	"encoding/json"
	"fmt"
	"strconv"
	
	"path/filepath"
	// "errors"
	// // "encoding/json"
	"io"
	"os"
	"net/http"
	
	// "github.com/jung-kurt/gofpdf"
	
	"sathya-narayanan23/crudapp/database"
	"sathya-narayanan23/crudapp/users/user"

    // "strconv"
	// "github.com/jinzhu/gorm"
	
	"github.com/gofrs/uuid"
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
func deleteBannerFile(filePath string) error {
	if _, err := os.Stat(filePath); err == nil {
		return os.Remove(filePath)
	}
	return nil
}


func handleError(w http.ResponseWriter, status int, message string) {
	var err Error
	err = SetError(err, message)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(err)
}


func CreateBanner(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	vars := mux.Vars(r)
	clientIDStr := vars["clientId"]

	// Convert clientID from string to uint
	clientID, err := strconv.ParseUint(clientIDStr, 10, 64)
	if err != nil {
		handleError(w, http.StatusBadRequest, "Invalid clientID. Must be a positive integer")
		return
	}

	var banner user1.Banner
	err = r.ParseMultipartForm(10 << 20) // 10 MB limit for form data
	if err != nil {
		handleError(w, http.StatusBadRequest, "Error in parsing form data.")
		return
	}

	banner.Special_banner = r.FormValue("special_banner")
	banner.ClientID = uint(clientID)

	// Handling Image Upload
	var imageFilePath string
	if imageFile, header, err := r.FormFile("special_banner"); err == nil {
		defer imageFile.Close()

		// Generate a unique filename for the image
		uniqueFileName, err := uuid.NewV4()
		if err != nil {
			handleError(w, http.StatusInternalServerError, "Error generating UUID.")
			return
		}

		
		os.MkdirAll("banner", os.ModePerm)


		// Create the complete file path
		ext := filepath.Ext(header.Filename)
		imageFilePath = filepath.Join("banner", uniqueFileName.String()+ext)

		// Create the file on the server
		outputFile, err := os.Create(imageFilePath)
		if err != nil {
			handleError(w, http.StatusInternalServerError, "Error creating image file.")
			return
		}
		defer outputFile.Close()

		// Copy the image data to the file
		_, err = io.Copy(outputFile, imageFile)
		if err != nil {
			handleError(w, http.StatusInternalServerError, "Error copying image data.")
			return
		}
	}

	if len(imageFilePath) > 0 {
		// Construct the URL for the served image
		serverAddr := r.Host
		imageServeURL := "http://" + serverAddr + "/banner/" + filepath.Base(imageFilePath)
		banner.Special_banner = imageServeURL
		banner.Bannerpath = imageFilePath

		// Save the banner to the database
		connection.Create(&banner)

		// Respond with the image URL and a success message
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data":       banner,
			"imagePath":  imageServeURL,
			"message":    "Image uploaded successfully",
		})
		return
	}

	// Respond with the banner data if no image is uploaded
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(banner)
}




func CreateOrUpdateBanner(w http.ResponseWriter, r *http.Request) {
    connection := database.GetDatabase()
    defer database.CloseDatabase(connection)

    vars := mux.Vars(r)
    clientIDStr := vars["clientId"]

    // Convert clientID from string to uint
    clientID, err := strconv.ParseUint(clientIDStr, 10, 64)
    if err != nil {
        handleError(w, http.StatusBadRequest, "Invalid clientID. Must be a positive integer")
        return
    }

    // Check if there is an existing banner for the client
    var existingBanner user1.Banner
    result := connection.Where("client_id = ?", clientID).First(&existingBanner)
    if result.Error == nil {
        // If an existing banner is found, delete its associated file
        if err := os.Remove(existingBanner.Bannerpath); err != nil {
            // Handle error (e.g., log it, but don't interrupt the response)
            fmt.Println("Error deleting existing image file:", err)
        }

        // Delete the existing banner from the database
        connection.Delete(&existingBanner)
    }

    var banner user1.Banner
    err = r.ParseMultipartForm(10 << 20) // 10 MB limit for form data
    if err != nil {
        handleError(w, http.StatusBadRequest, "Error in parsing form data.")
        return
    }

    banner.Special_banner = r.FormValue("special_banner")
    banner.ClientID = uint(clientID)

    // Handling Image Upload
    var imageFilePath string
    if imageFile, header, err := r.FormFile("special_banner"); err == nil {
        defer imageFile.Close()

        // Generate a unique filename for the image
        uniqueFileName, err := uuid.NewV4()
        if err != nil {
            handleError(w, http.StatusInternalServerError, "Error generating UUID.")
            return
        }

        // Create the complete file path
        ext := filepath.Ext(header.Filename)
        imageFilePath = filepath.Join("banner", uniqueFileName.String()+ext)

        // Create the file on the server
        outputFile, err := os.Create(imageFilePath)
        if err != nil {
            handleError(w, http.StatusInternalServerError, "Error creating image file.")
            return
        }
        defer outputFile.Close()

        // Copy the image data to the file
        _, err = io.Copy(outputFile, imageFile)
        if err != nil {
            handleError(w, http.StatusInternalServerError, "Error copying image data.")
            return
        }
    }

    if len(imageFilePath) > 0 {
        // Construct the URL for the served image
        serverAddr := r.Host
        imageServeURL := "http://" + serverAddr + "/banner/" + filepath.Base(imageFilePath)
        banner.Special_banner = imageServeURL

        // Save the file path in the Bannerpath field
        banner.Bannerpath = imageFilePath

        // Save the banner to the database
        connection.Create(&banner)

        // Respond with the image URL and a success message
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "data":       banner,
            "imagePath":  imageServeURL,
            "message":    "Image uploaded successfully",
        })
        return
    }

    // Respond with the banner data if no image is uploaded
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(banner)
}



func GetBannersByClientID(w http.ResponseWriter, r *http.Request) {
    connection := database.GetDatabase()
    defer database.CloseDatabase(connection)

    // Get the client ID from the URL parameters
    vars := mux.Vars(r)
    clientIDStr := vars["clientID"]

    // Convert client ID from string to uint
    clientID, err := strconv.ParseUint(clientIDStr, 10, 64)
    if err != nil {
        handleError(w, http.StatusBadRequest, "Invalid client ID. Must be a positive integer")
        return
    }

    // Fetch banners based on client ID
    var banners []user1.Banner
    result := connection.Where("client_id = ?", clientID).Find(&banners)
    if result.Error != nil {
        handleError(w, http.StatusInternalServerError, result.Error.Error())
        return
    }

    // Respond with the list of banners
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(banners)
}


func GetHighestBannerForClientID(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	// Get the clientID from the URL parameters
	vars := mux.Vars(r)
	clientIDStr := vars["clientID"]

	// Convert clientID from string to uint
	clientID, err := strconv.ParseUint(clientIDStr, 10, 64)
	if err != nil {
		handleError(w, http.StatusBadRequest, "Invalid clientID. Must be a positive integer")
		return
	}

	// Query the database to get the banner with the highest ID for the specified clientID
	var highestBanner user1.Banner
	result := connection.Where("client_id = ?", clientID).Order("id desc").Limit(1).Find(&highestBanner)
	if result.Error != nil {
		handleError(w, http.StatusInternalServerError, "Error fetching highest banner for clientID")
		return
	}

	// Respond with the result
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(highestBanner)
}


func UpdateSpecialBanner(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	// Get the clientID from the URL parameters
	vars := mux.Vars(r)
	clientIDStr := vars["clientID"]

	// Convert clientID from string to uint
	clientID, err := strconv.ParseUint(clientIDStr, 10, 64)
	if err != nil {
		handleError(w, http.StatusBadRequest, "Invalid clientID. Must be a positive integer")
		return
	}

	// Find the banner with the highest ID for the specified clientID
	var highestBanner user1.Banner
	result := connection.Where("client_id = ?", clientID).Order("id desc").Limit(1).Find(&highestBanner)
	if result.Error != nil {
		handleError(w, http.StatusInternalServerError, "Error fetching highest banner for clientID")
		return
	}

	// Check if a banner was found
	if highestBanner.ID == 0 {
		handleError(w, http.StatusNotFound, "No banners found for clientID")
		return
	}

	// Delete the old bannerpath file from the server
	err = deleteBannerFile(highestBanner.Bannerpath)
	if err != nil {
		// Handle the error (e.g., log it)
		fmt.Println("Error deleting old bannerpath file:", err)
	}

	// Update the special_banner field
	highestBanner.Special_banner = r.FormValue("special_banner")

	// Handling Image Upload for special_banner
	var imageFilePath string
	if imageFile, header, err := r.FormFile("special_banner"); err == nil {
		defer imageFile.Close()

		// Generate a unique filename for the image
		uniqueFileName, err := uuid.NewV4()
		if err != nil {
			handleError(w, http.StatusInternalServerError, "Error generating UUID.")
			return
		}

		// Create the complete file path
		ext := filepath.Ext(header.Filename)
		imageFilePath = filepath.Join("banner", uniqueFileName.String()+ext)

		// Create the file on the server
		outputFile, err := os.Create(imageFilePath)
		if err != nil {
			handleError(w, http.StatusInternalServerError, "Error creating image file.")
			return
		}
		defer outputFile.Close()

		// Copy the image data to the file
		_, err = io.Copy(outputFile, imageFile)
		if err != nil {
			handleError(w, http.StatusInternalServerError, "Error copying image data.")
			return
		}
	}

	// Save the new bannerpath
	if len(imageFilePath) > 0 {
		// Construct the URL for the served image
		serverAddr := r.Host
		imageServeURL := "http://" + serverAddr + "/banner/" + filepath.Base(imageFilePath)
		highestBanner.Special_banner = imageServeURL
		highestBanner.Bannerpath = imageFilePath
	}

	// Update the banner in the database
	connection.Save(&highestBanner)

	// Respond with the updated banner
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(highestBanner)
}



func GetBannersAndMenuItems(w http.ResponseWriter, r *http.Request) {
    connection := database.GetDatabase()
    defer database.CloseDatabase(connection)

    vars := mux.Vars(r)
    clientIDStr := vars["clientID"]

    // Convert clientID from string to uint
    clientID, err := strconv.ParseUint(clientIDStr, 10, 64)
    if err != nil {
        handleError(w, http.StatusBadRequest, "Invalid clientID. Must be a positive integer")
        return
    }

    // Get banners
    var banners []user1.Banner
    connection.Where("client_id = ?", clientID).Find(&banners)

    // Get menu items with Recommendation set to true for the given client
    var menuItems []user1.MenuItem
    connection.Where("client_id = ? AND recommendation = ?", clientID, true).Find(&menuItems)

    // Respond with banners and menu items
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "banners":     banners,
        "menuItems":   menuItems,
    })
}



func UpdateOrCreateBanner(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	// Get the clientID from the URL parameters
	vars := mux.Vars(r)
	clientIDStr := vars["clientID"]

	// Convert clientID from string to uint
	clientID, err := strconv.ParseUint(clientIDStr, 10, 64)
	if err != nil {
		handleError(w, http.StatusBadRequest, "Invalid clientID. Must be a positive integer")
		return
	}

	// Find the banner with the highest ID for the specified clientID
	var highestBanner user1.Banner
	result := connection.Where("client_id = ?", clientID).Order("id desc").Limit(1).Find(&highestBanner)
	if result.Error != nil {
		handleError(w, http.StatusInternalServerError, "Error fetching highest banner for clientID")
		createNewBanner(w, r, clientID)
		return
		
	}

	// Delete the old bannerpath file from the server
	err = deleteBannerFile(highestBanner.Bannerpath)
	if err != nil {
		// Handle the error (e.g., log it)
		fmt.Println("Error deleting old bannerpath file:", err)
	}

	// Create a new banner or update the existing one
	if highestBanner.ID == 0 {
		// Create a new banner if no existing banner found
		
		// CreateBanner(w, r, clientID)
		createNewBanner(w, r, clientID)
	} else {
		// Update the existing banner
		updateExistingBanner(w, r, &highestBanner)
	}
}


func DeleteBanner(w http.ResponseWriter, r *http.Request) {
    connection := database.GetDatabase()
    defer database.CloseDatabase(connection)

    // Get the banner ID from the URL parameters
    vars := mux.Vars(r)
    bannerID := vars["bannerID"]

    // Convert bannerID from string to uint
    bannerIDUint, err := strconv.ParseUint(bannerID, 10, 64)
    if err != nil {
        handleError(w, http.StatusBadRequest, "Invalid banner ID. Must be a positive integer")
        return
    }

    // Check if the banner exists
    var existingBanner user1.Banner
    result := connection.First(&existingBanner, bannerIDUint)
    if result.Error != nil {
        handleError(w, http.StatusNotFound, "Banner not found")
        return
    }

    // Delete the associated image file
    if err := os.Remove(existingBanner.Bannerpath); err != nil {
        // Handle error (e.g., log it, but don't interrupt the response)
        fmt.Println("Error deleting image file:", err)
    }

    // Delete the banner from the database
    connection.Delete(&existingBanner)

    // Respond with success message
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "message": "Banner deleted successfully",
    })
}


func createNewBanner(w http.ResponseWriter, r *http.Request, clientID uint64) {
	var banner user1.Banner
	connection := database.GetDatabase()
	// Set other banner fields as needed
	banner.ClientID = uint(clientID)

	// Update the special_banner field
	banner.Special_banner = r.FormValue("special_banner")

	// Handling Image Upload for special_banner
	var imageFilePath string
	if imageFile, header, err := r.FormFile("special_banner"); err == nil {
		defer imageFile.Close()

		// Generate a unique filename for the image
		uniqueFileName, err := uuid.NewV4()
		if err != nil {
			handleError(w, http.StatusInternalServerError, "Error generating UUID.")
			return
		}

		// Create the complete file path
		ext := filepath.Ext(header.Filename)
		imageFilePath = filepath.Join("banner", uniqueFileName.String()+ext)

		// Create the file on the server
		outputFile, err := os.Create(imageFilePath)
		if err != nil {
			handleError(w, http.StatusInternalServerError, "Error creating image file.")
			return
		}
		defer outputFile.Close()

		// Copy the image data to the file
		_, err = io.Copy(outputFile, imageFile)
		if err != nil {
			handleError(w, http.StatusInternalServerError, "Error copying image data.")
			return
		}
	}

	// Save the new bannerpath
	if len(imageFilePath) > 0 {
		// Construct the URL for the served image
		serverAddr := r.Host
		imageServeURL := "http://" + serverAddr + "/banner/" + filepath.Base(imageFilePath)
		banner.Special_banner = imageServeURL
		banner.Bannerpath = imageFilePath
	}

	// Save the new banner to the database
	connection.Create(&banner)

	// Respond with the created banner
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(banner)
}

// Function to update an existing banner
func updateExistingBanner(w http.ResponseWriter, r *http.Request, existingBanner *user1.Banner) {
	// Update the special_banner field
	connection := database.GetDatabase()
	existingBanner.Special_banner = r.FormValue("special_banner")

	// Handling Image Upload for special_banner
	var imageFilePath string
	if imageFile, header, err := r.FormFile("special_banner"); err == nil {
		defer imageFile.Close()

		// Generate a unique filename for the image
		uniqueFileName, err := uuid.NewV4()
		if err != nil {
			handleError(w, http.StatusInternalServerError, "Error generating UUID.")
			return
		}

		// Create the complete file path
		ext := filepath.Ext(header.Filename)
		imageFilePath = filepath.Join("banner", uniqueFileName.String()+ext)

		// Create the file on the server
		outputFile, err := os.Create(imageFilePath)
		if err != nil {
			handleError(w, http.StatusInternalServerError, "Error creating image file.")
			return
		}
		defer outputFile.Close()

		// Copy the image data to the file
		_, err = io.Copy(outputFile, imageFile)
		if err != nil {
			handleError(w, http.StatusInternalServerError, "Error copying image data.")
			return
		}

		// Delete the old bannerpath file from the server
		err = deleteBannerFile(existingBanner.Bannerpath)
		if err != nil {
			// Handle the error (e.g., log it)
			fmt.Println("Error deleting old bannerpath file:", err)
		}
	}

	// Save the new bannerpath
	if len(imageFilePath) > 0 {
		// Construct the URL for the served image
		serverAddr := r.Host
		imageServeURL := "http://" + serverAddr + "/banner/" + filepath.Base(imageFilePath)
		existingBanner.Special_banner = imageServeURL
		existingBanner.Bannerpath = imageFilePath
	}

	// Update the existing banner in the database
	connection.Save(existingBanner)

	// Respond with the updated banner
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(existingBanner)
}








