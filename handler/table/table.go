package table

import (
	"encoding/json"
	"fmt"
	"errors"
    
	// "encoding/json"
	// "sort"
	"net/http"
	
	"github.com/jinzhu/gorm"
	"github.com/skip2/go-qrcode"
	"os"
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

func CreateTable(w http.ResponseWriter, r *http.Request) {
	// Parse JSON request
	var tableNo user1.TableNo
	err := json.NewDecoder(r.Body).Decode(&tableNo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the table name already exists for the given client
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	var existingTable user1.TableNo
	result := connection.Where("table_name = ? AND client_id = ?", tableNo.TableName, tableNo.ClientID).First(&existingTable)
	if result.Error == nil {
		http.Error(w, "Table name already exists for the given client", http.StatusBadRequest)
		return
	}

	// Validate the provided ClientID
	var client user1.Client
	if err := connection.First(&client, tableNo.ClientID).Error; err != nil {
		http.Error(w, "Invalid ClientID", http.StatusBadRequest)
		return
	}

	// Validate the table name (you can add more validation if needed)
	if tableNo.TableName == "" {
		http.Error(w, "tableName is a required field.", http.StatusBadRequest)
		return
	}

	// Create the new table entry
	result = connection.Create(&tableNo)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}
	
	os.MkdirAll("qrcodes", os.ModePerm)

	

	// Generate QR code content including client.Logo and client.PrimaryColour
	qrCodeContent := fmt.Sprintf("TableID: %d, TableName: %s, ClientId: %d ",
		tableNo.ID, tableNo.TableName, tableNo.ClientID)

	// Generate QR code and save to file
	qrCodeFileName := fmt.Sprintf("%s_%d_qr.png", tableNo.TableName, tableNo.ClientID)
	qrCodeFilePath := fmt.Sprintf("qrcodes/%s", qrCodeFileName)
	err = qrcode.WriteFile(qrCodeContent, qrcode.Medium, 512, qrCodeFilePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	serverAddr := r.Host
	imageServeURL := "http://" + serverAddr + "/" + qrCodeFilePath

	// Set the QRCodeFilePath in the table struct and update the table entry
	tableNo.QRCodeFilePath = qrCodeFilePath
	tableNo.QRCode = imageServeURL
	connection.Save(&tableNo)

	// Prepare the response
	responses := map[string]interface{}{
		"QRCodeFilePath": imageServeURL,
		
		"QRCode": imageServeURL,
		
		"TableDetails":   qrCodeContent,
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(responses)
}

func DeleteTable(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	vars := mux.Vars(r)
	tableNoID := vars["id"]

	var tableNo user1.TableNo
	if err := connection.First(&tableNo, tableNoID).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(Error{IsError: true, Message: "Resource not found"})
		return
	}
	if err := os.Remove(tableNo.QRCodeFilePath); err != nil {
        // Handle error (e.g., log it, but don't interrupt the response)
        fmt.Println("Error deleting image file:", err)
    }

	if err := connection.Delete(&tableNo).Error; err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(Error{IsError: true, Message: "Failed to delete resource"})
		return
	} 

	responseMessage := fmt.Sprintf("Table  ID %s has been deleted", tableNoID)
	json.NewEncoder(w).Encode(responseMessage)
	w.WriteHeader(http.StatusNoContent)
}


func GetTableNoByID(w http.ResponseWriter, r *http.Request) {
	// Extract the table ID from the request parameters or wherever it is specified
	// For example, if the ID is passed as a query parameter: /get-table?id=123
	tableIDStr := r.URL.Query().Get("id")
	tableID, err := strconv.ParseUint(tableIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid table ID", http.StatusBadRequest)
		return
	}

	// Retrieve the table by ID
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	var tableNo user1.TableNo
	result := connection.First(&tableNo, tableID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			http.Error(w, "Table not found", http.StatusNotFound)
		} else {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Respond with the retrieved table details
	response := map[string]interface{}{
		"ID":              tableNo.ID,
		"TableName":       tableNo.TableName,
		"ClientID":        tableNo.ClientID,
		"QRCodeFilePath":  tableNo.QRCodeFilePath,
		// Add other fields as needed
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}


func GetTable(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	vars := mux.Vars(r)
	tableNoID := vars["id"]

	var tableNo user1.TableNo
	result := connection.First(&tableNo, tableNoID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			http.Error(w, "Table not found", http.StatusNotFound)
		} else {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tableNo)
}


func GetTablesByClientIDs(w http.ResponseWriter, r *http.Request) {
	// Extract the clientID parameter from the URL
	vars := mux.Vars(r)
	clientIDStr := vars["clientID"]
	clientID, err := strconv.ParseUint(clientIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ClientID", http.StatusBadRequest)
		return
	}

	// Get tables associated with the provided ClientID
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	var tables []user1.TableNo
	result := connection.Where("client_id = ?", clientID).Find(&tables)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}
	// imagepathin := TableNo.QRCodeFilePath


	// Set the response content type and send the JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tables)
}
