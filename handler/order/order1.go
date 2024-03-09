package order

import (
	"encoding/json"
	"fmt"
	"strconv"
	// "bytes"
    
	// // "encoding/json"
	"time"
	"sort"
	"net/http"
	
	// "github.com/jung-kurt/gofpdf"
	
	"sathya-narayanan23/crudapp/database"
	"sathya-narayanan23/crudapp/users/user"

    // "strconv"
	"github.com/jinzhu/gorm"
	"github.com/gorilla/mux"
)


func GetMenuItemByID(connection *gorm.DB, menuItemID uint) *user1.MenuItem {
	var menuItem user1.MenuItem
	result := connection.First(&menuItem, menuItemID)
	if result.Error != nil {
		return nil
	}
	return &menuItem
}




func GetOrderStatusForClientAndOrderID(w http.ResponseWriter, r *http.Request) {
	// Parse URL parameters
	params := mux.Vars(r)

	clientIDStr := params["clientID"]
	orderIDStr := params["orderID"]

	// Parse the client ID from the URL parameter
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		http.Error(w, "Invalid clientID", http.StatusBadRequest)
		return
	}

	// Parse the order ID from the URL parameter
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		http.Error(w, "Invalid orderID", http.StatusBadRequest)
		return
	}

	// Get the order based on client ID and order ID
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	var order user1.Order
	result := connection.Where("client_id = ? AND id = ? AND status = ?", clientID, orderID, "Pending").First(&order)
	if result.Error != nil {
		http.Error(w, "Order not found or not in 'Pending' status", http.StatusNotFound)
		return
	}

	// Send the response with the order status
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(struct {
		OrderStatus string `json:"orderStatus"`
	}{
		OrderStatus: order.Status,
	})
}


func UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	// Parse URL parameters
	params := mux.Vars(r)

	clientIDStr := params["clientID"]
	orderIDStr := params["orderID"]

	// Parse the client ID from the URL parameter
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		http.Error(w, "Invalid clientID", http.StatusBadRequest)
		return
	}

	// Parse the order ID from the URL parameter
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		http.Error(w, "Invalid orderID", http.StatusBadRequest)
		return
	}

	// Parse JSON request to get the new status
	var statusUpdate struct {
		Status string `json:"status"`
	}
	err = json.NewDecoder(r.Body).Decode(&statusUpdate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the order based on client ID and order ID
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	var order user1.Order
	result := connection.Where("client_id = ? AND id = ?", clientID, orderID).First(&order)
	if result.Error != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	// Update the order status
	order.Status = statusUpdate.Status
	result = connection.Save(&order)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Send a success responsefunc
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(order)
}


func GetPendingOrdersForClientnew(w http.ResponseWriter, r *http.Request) {
	// Parse URL parameters
	params := mux.Vars(r)
	clientIDStr := params["clientID"]

	// Parse the client ID from the URL parameter
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		http.Error(w, "Invalid clientID", http.StatusBadRequest)
		return
	}

	// Get the orders with the specified client ID and status "Pending" or "Prepare"
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	// Query orders that match the criteria and join with order_items
	var orders []user1.Order
	result := connection.Joins("JOIN order_items ON orders.id = order_items.order_id").
		Where("orders.client_id = ? AND order_items.order_item_status = ?", clientID, "OrderPending").
		Select("DISTINCT orders.*, order_items.created_at").
		Find(&orders)

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Create a response structure for orders and order items
	var responseOrders []struct {
		ID               uint      `json:"id"`
		Status           string    `json:"status"`
		TableNoID        uint      `json:"tableNoID"`
		TotalPrice       float64   `json:"totalPrice"`
		TableNoTableName string    `json:"tableNoTableName"`
		OrderTime        int       `json:"orderTime"`
		CreatedAt        time.Time `json:"createdAt"`
		OrderItems       []struct {
			ID               uint      `json:"id"`
			MenuItemID       uint      `json:"menuItemID"`
			MenuItemItemName string    `json:"menuItemItemName"`
			OrderItemStatus  string    `json:"orderItemStatus"`
			OrderPayStatus   string    `json:"orderPayStatus"`
			Quantity         int       `json:"quantity"`
			MenuItemPrice    float64   `json:"menuItemPrice"`
			MenuItemTime     int       `json:"menuItemTime"`
			CreatedAt        time.Time `json:"createdAt"`
		} `json:"orderItems"`
	}

	// Loop through the retrieved orders and fetch their associated order items
	for _, order := range orders {
		var orderItems []user1.OrderItem
		result := connection.Model(&order).Related(&orderItems)
		if result.Error != nil {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
			return
		}

		// Find the earliest timestamp among order items
		var earliestTimestamp time.Time
		if len(orderItems) > 0 {
			earliestTimestamp = orderItems[0].CreatedAt
			for _, item := range orderItems {
				if item.CreatedAt.Before(earliestTimestamp) {
					earliestTimestamp = item.CreatedAt
				}
			}
		}

		// Declare responseOrder here
		var responseOrder struct {
			ID               uint      `json:"id"`
			Status           string    `json:"status"`
			TableNoID        uint      `json:"tableNoID"`
			TotalPrice       float64   `json:"totalPrice"`
			TableNoTableName string    `json:"tableNoTableName"`
			OrderTime        int       `json:"orderTime"`
			CreatedAt        time.Time `json:"createdAt"`
			OrderItems       []struct {
				ID               uint      `json:"id"`
				MenuItemID       uint      `json:"menuItemID"`
				MenuItemItemName string    `json:"menuItemItemName"`
				OrderItemStatus  string    `json:"orderItemStatus"`
				OrderPayStatus   string    `json:"orderPayStatus"`
				Quantity         int       `json:"quantity"`
				MenuItemPrice    float64   `json:"menuItemPrice"`
				MenuItemTime     int       `json:"menuItemTime"`
				CreatedAt        time.Time `json:"createdAt"`
			} `json:"orderItems"`
		}

		// Populate the responseOrder structure
		responseOrder.ID = order.ID
		responseOrder.Status = order.Status
		responseOrder.TableNoID = order.TableNoID
		responseOrder.TableNoTableName = order.TableNoTableName
		responseOrder.TotalPrice = order.TotalPrice
		responseOrder.CreatedAt = earliestTimestamp // Set to the earliest timestamp among order items

		// Convert each OrderItem to the desired struct type and append to responseOrder.OrderItems
		for _, orderItem := range orderItems {
			// Ensure that MenuItemItemName is correctly populated
			menuItem := GetMenuItemByID(connection, orderItem.MenuItemID)
			if menuItem != nil {
				orderItem.MenuItemItemName = menuItem.ItemName
			}

			responseOrderItem := struct {
				ID               uint      `json:"id"`
				MenuItemID       uint      `json:"menuItemID"`
				MenuItemItemName string    `json:"menuItemItemName"`
				OrderItemStatus  string    `json:"orderItemStatus"`
				OrderPayStatus   string    `json:"orderPayStatus"`
				Quantity         int       `json:"quantity"`
				MenuItemPrice    float64   `json:"menuItemPrice"`
				MenuItemTime     int       `json:"menuItemTime"`
				CreatedAt        time.Time `json:"createdAt"`
			}{
				ID:               orderItem.ID,
				MenuItemID:       orderItem.MenuItemID,
				MenuItemItemName: orderItem.MenuItemItemName, // Make sure it's correctly populated
				OrderItemStatus:  orderItem.OrderItemStatus,
				OrderPayStatus:   orderItem.OrderPayStatus,
				Quantity:         orderItem.Quantity,
				MenuItemPrice:    orderItem.MenuItemPrice,
				MenuItemTime:     orderItem.MenuItemTime,
				CreatedAt:        orderItem.CreatedAt, // Preserve the original CreatedAt timestamp from the order item
			}

			responseOrder.OrderItems = append(responseOrder.OrderItems, responseOrderItem)
		}

		// Append the responseOrder to the responseOrders
		responseOrders = append(responseOrders, responseOrder)
	}

	// Send the response with the pending orders and their associated menu items
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseOrders)
}


func GetPendingOrdersForClient123(w http.ResponseWriter, r *http.Request) {
	// Parse URL parameters
	params := mux.Vars(r)

	clientIDStr := params["clientID"]

	// Parse the client ID from the URL parameter
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		http.Error(w, "Invalid clientID", http.StatusBadRequest)
		return
	}

	// Get the orders with the specified client ID and status "Pending" or "Prepare"
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	var orders []user1.Order
	result := connection.Where("client_id = ? AND status IN (?)", clientID, []string{"Pending", "Prepare"}).Find(&orders)
	if result.Error != nil {
		http.Error(w, "No pending orders found", http.StatusNotFound)
		return
	}

	// Fetch order items and their associated menu items
	var responseOrders []struct {
		ID               uint    `json:"id"`
		Status           string  `json:"status"`
		TableNoID        uint    `json:"tableNoID"`
		TotalPrice       float64 `json:"totalPrice"`
		TableNoTableName string  `json:"tableNoTableName"`
		OrderTime        int     `json:"orderTime"`
		OrderItems       []struct {
			ID              uint    `json:"id"`
			MenuItemID      uint    `json:"menuItemID"`
			MenuItemName    string  `json:"menuItemName"`
			OrderItemStatus string  `json:"orderItemStatus"`
			OrderPayStatus  string  `json:"orderPayStatus"`
			Quantity        int     `json:"quantity"`
			MenuItemPrice   float64 `json:"menuItemPrice"`
			MenuItemTime    int     `json:"menuItemTime"`
		} `json:"orderItems"`
	}

	for _, order := range orders {
		var orderItems []user1.OrderItem
		result := connection.Model(&order).Related(&orderItems)
		if result.Error != nil {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
			return
		}

		var responseOrder struct {
			ID               uint    `json:"id"`
			Status           string  `json:"status"`
			TableNoID        uint    `json:"tableNoID"`
			TotalPrice       float64 `json:"totalPrice"`
			TableNoTableName string  `json:"tableNoTableName"`
			OrderTime        int     `json:"orderTime"`
			OrderItems       []struct {
				ID              uint    `json:"id"`
				MenuItemID      uint    `json:"menuItemID"`
				MenuItemName    string  `json:"menuItemName"`
				OrderItemStatus string  `json:"orderItemStatus"`
				OrderPayStatus  string  `json:"orderPayStatus"`
				Quantity        int     `json:"quantity"`
				MenuItemPrice   float64 `json:"menuItemPrice"`
				MenuItemTime    int     `json:"menuItemTime"`
			} `json:"orderItems"`
		}

		responseOrder.ID = order.ID
		responseOrder.Status = order.Status
		responseOrder.TableNoID = order.TableNoID
		responseOrder.TableNoTableName = order.TableNoTableName
		responseOrder.TotalPrice = order.TotalPrice

		for _, orderItem := range orderItems {
			var menuItem user1.MenuItem
			result := connection.First(&menuItem, orderItem.MenuItemID)
			if result.Error != nil {
				http.Error(w, result.Error.Error(), http.StatusInternalServerError)
				return
			}

			responseOrder.OrderItems = append(responseOrder.OrderItems, struct {
				ID              uint    `json:"id"`
				MenuItemID      uint    `json:"menuItemID"`
				MenuItemName    string  `json:"menuItemName"`
				OrderItemStatus string  `json:"orderItemStatus"`
				OrderPayStatus  string  `json:"orderPayStatus"`
				Quantity        int     `json:"quantity"`
				MenuItemPrice   float64 `json:"menuItemPrice"`
				MenuItemTime    int     `json:"menuItemTime"`
			}{
				ID:              orderItem.ID,
				MenuItemID:      orderItem.MenuItemID,
				MenuItemName:    menuItem.ItemName,
				OrderItemStatus: orderItem.OrderItemStatus,
				OrderPayStatus:  orderItem.OrderPayStatus,
				Quantity:        orderItem.Quantity,
				MenuItemPrice:   orderItem.MenuItemPrice,
				MenuItemTime:    menuItem.Time,
			})
		}

		responseOrders = append(responseOrders, responseOrder)
	}

	// Send the response with the pending orders and their associated menu items, TableNoID, and TotalPrice
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseOrders)
}


func GetPendingOrdersForClient456(w http.ResponseWriter, r *http.Request) {
	// Parse URL parameters
	params := mux.Vars(r)
	clientIDStr := params["clientID"]

	// Parse the client ID from the URL parameter
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		http.Error(w, "Invalid clientID", http.StatusBadRequest)
		return
	}

	// Get the orders with the specified client ID and status "Pending" or "Prepare"
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	// Query orders that match the criteria and join with order_items
	var orders []user1.Order
	result := connection.Joins("JOIN order_items ON orders.id = order_items.order_id").
		Where("orders.client_id = ? AND order_items.order_item_status = ?", clientID, "OrderPending").
		Select("DISTINCT orders.*, order_items.created_at").
		Find(&orders)

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Create a response structure for orders and order items
	var responseOrders []struct {
		ID               uint      `json:"id"`
		Status           string    `json:"status"`
		TableNoID        uint      `json:"tableNoID"`
		TotalPrice       float64   `json:"totalPrice"`
		TableNoTableName string    `json:"tableNoTableName"`
		OrderTime        int       `json:"orderTime"`
		CreatedAt        time.Time `json:"createdAt"`
		OrderItems       []struct {
			ID              uint      `json:"id"`
			MenuItemID      uint      `json:"menuItemID"`
			MenuItemName    string    `json:"menuItemName"`
			OrderItemStatus string    `json:"orderItemStatus"`
			OrderPayStatus  string    `json:"orderPayStatus"`
			Quantity        int       `json:"quantity"`
			MenuItemPrice   float64   `json:"menuItemPrice"`
			MenuItemTime    int       `json:"menuItemTime"`
			CreatedAt       time.Time `json:"createdAt"`
		} `json:"orderItems"`
	}

	// Loop through the retrieved orders and fetch their associated order items
	for _, order := range orders {
		var orderItems []user1.OrderItem
		result := connection.Model(&order).Related(&orderItems)
		if result.Error != nil {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
			return
		}

		// Declare responseOrder here
		var responseOrder struct {
			ID               uint      `json:"id"`
			Status           string    `json:"status"`
			TableNoID        uint      `json:"tableNoID"`
			TotalPrice       float64   `json:"totalPrice"`
			TableNoTableName string    `json:"tableNoTableName"`
			OrderTime        int       `json:"orderTime"`
			CreatedAt        time.Time `json:"createdAt"`
			OrderItems       []struct {
				ID              uint      `json:"id"`
				MenuItemID      uint      `json:"menuItemID"`
				MenuItemName    string    `json:"menuItemName"`
				OrderItemStatus string    `json:"orderItemStatus"`
				OrderPayStatus  string    `json:"orderPayStatus"`
				Quantity        int       `json:"quantity"`
				MenuItemPrice   float64   `json:"menuItemPrice"`
				MenuItemTime    int       `json:"menuItemTime"`
				CreatedAt       time.Time `json:"createdAt"`
			} `json:"orderItems"`
		}

		// Populate the responseOrder structure
		responseOrder.ID = order.ID
		responseOrder.Status = order.Status
		responseOrder.TableNoID = order.TableNoID
		responseOrder.TableNoTableName = order.TableNoTableName
		responseOrder.TotalPrice = order.TotalPrice

		// Convert each OrderItem to the desired struct type and append to responseOrder.OrderItems
		for _, orderItem := range orderItems {
			responseOrderItem := struct {
				ID              uint      `json:"id"`
				MenuItemID      uint      `json:"menuItemID"`
				MenuItemName    string    `json:"menuItemName"`
				OrderItemStatus string    `json:"orderItemStatus"`
				OrderPayStatus  string    `json:"orderPayStatus"`
				Quantity        int       `json:"quantity"`
				MenuItemPrice   float64   `json:"menuItemPrice"`
				MenuItemTime    int       `json:"menuItemTime"`
				CreatedAt       time.Time `json:"createdAt"`
			}{
				ID:         orderItem.ID,
				MenuItemID: orderItem.MenuItemID,
				// MenuItemName:  orderItem.MenuItemName,
				OrderItemStatus: orderItem.OrderItemStatus,
				OrderPayStatus:  orderItem.OrderPayStatus,
				Quantity:        orderItem.Quantity,
				MenuItemPrice:   orderItem.MenuItemPrice,
				MenuItemTime:    orderItem.MenuItemTime,
				CreatedAt:       orderItem.CreatedAt, // Preserve the original CreatedAt timestamp from the order item
			}
			responseOrder.CreatedAt = orderItem.CreatedAt // Assign the original order's CreatedAt timestamp

			responseOrder.OrderItems = append(responseOrder.OrderItems, responseOrderItem)
		}

		// Append the responseOrder to the responseOrders
		responseOrders = append(responseOrders, responseOrder)
	}

	// Send the response with the pending orders and their associated menu items
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseOrders)
}


func GetPendingOrdersForClient(w http.ResponseWriter, r *http.Request) {
	// Parse URL parameters
	params := mux.Vars(r)
	clientIDStr := params["clientID"]

	// Parse the client ID from the URL parameter
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		http.Error(w, "Invalid clientID", http.StatusBadRequest)
		return
	}

	// Get the orders with the specified client ID and status "Pending" or "Prepare"
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	// Query orders that match the criteria and join with order_items
	var orders []user1.Order
	result := connection.Joins("JOIN order_items ON orders.id = order_items.order_id").
		Where("orders.client_id = ? AND order_items.order_item_status = ?", clientID, "OrderPending").
		Select("DISTINCT orders.*, order_items.created_at").
		Find(&orders)

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Create a response structure for orders and order items
	var responseOrders []struct {
		ID               uint      `json:"id"`
		Status           string    `json:"status"`
		TableNoID        uint      `json:"tableNoID"`
		TotalPrice       float64   `json:"totalPrice"`
		TableNoTableName string    `json:"tableNoTableName"`
		OrderTime        int       `json:"orderTime"`
		CreatedAt        time.Time `json:"createdAt"`
		OrderItems       []struct {
			ID               uint      `json:"id"`
			MenuItemID       uint      `json:"menuItemID"`
			MenuItemItemName string    `json:"menuItemItemName"`
			OrderItemStatus  string    `json:"orderItemStatus"`
			OrderPayStatus   string    `json:"orderPayStatus"`
			Quantity         int       `json:"quantity"`
			MenuItemPrice    float64   `json:"menuItemPrice"`
			MenuItemTime     int       `json:"menuItemTime"`
			CreatedAt        time.Time `json:"createdAt"`
		} `json:"orderItems"`
	}

	// Loop through the retrieved orders and fetch their associated order items
	for _, order := range orders {
		var orderItems []user1.OrderItem
		result := connection.Model(&order).Where("order_item_status = ?", "OrderPending").Related(&orderItems) // Filter by order_item_status
		if result.Error != nil {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
			return
		}

		for i := 1; i < len(orderItems)-1; i++ {
			order.CreatedAt = orderItems[i+1].CreatedAt // Set CreatedAt to the orderItem's CreatedAt
		}

		// Declare responseOrder here
		var responseOrder struct {
			ID               uint      `json:"id"`
			Status           string    `json:"status"`
			TableNoID        uint      `json:"tableNoID"`
			TotalPrice       float64   `json:"totalPrice"`
			TableNoTableName string    `json:"tableNoTableName"`
			OrderTime        int       `json:"orderTime"`
			CreatedAt        time.Time `json:"createdAt"`
			OrderItems       []struct {
				ID               uint      `json:"id"`
				MenuItemID       uint      `json:"menuItemID"`
				MenuItemItemName string    `json:"menuItemItemName"`
				OrderItemStatus  string    `json:"orderItemStatus"`
				OrderPayStatus   string    `json:"orderPayStatus"`
				Quantity         int       `json:"quantity"`
				MenuItemPrice    float64   `json:"menuItemPrice"`
				MenuItemTime     int       `json:"menuItemTime"`
				CreatedAt        time.Time `json:"createdAt"`
			} `json:"orderItems"`
		}

		// Populate the responseOrder structure
		responseOrder.ID = order.ID
		responseOrder.Status = order.Status
		responseOrder.TableNoID = order.TableNoID
		responseOrder.TableNoTableName = order.TableNoTableName
		responseOrder.TotalPrice = order.TotalPrice
		responseOrder.CreatedAt = order.CreatedAt // Set CreatedAt to the order's CreatedAt

		// Convert each OrderItem to the desired struct type and append to responseOrder.OrderItems
		for _, orderItem := range orderItems {
			// Ensure that MenuItemItemName is correctly populated
			menuItem := GetMenuItemByID(connection, orderItem.MenuItemID)
			if menuItem != nil {
				orderItem.MenuItemItemName = menuItem.ItemName
			}

			responseOrderItem := struct {
				ID               uint      `json:"id"`
				MenuItemID       uint      `json:"menuItemID"`
				MenuItemItemName string    `json:"menuItemItemName"`
				OrderItemStatus  string    `json:"orderItemStatus"`
				OrderPayStatus   string    `json:"orderPayStatus"`
				Quantity         int       `json:"quantity"`
				MenuItemPrice    float64   `json:"menuItemPrice"`
				MenuItemTime     int       `json:"menuItemTime"`
				CreatedAt        time.Time `json:"createdAt"`
			}{
				ID:               orderItem.ID,
				MenuItemID:       orderItem.MenuItemID,
				MenuItemItemName: orderItem.MenuItemItemName, // Make sure it's correctly populated
				OrderItemStatus:  orderItem.OrderItemStatus,
				OrderPayStatus:   orderItem.OrderPayStatus,
				Quantity:         orderItem.Quantity,
				MenuItemPrice:    orderItem.MenuItemPrice,
				MenuItemTime:     orderItem.MenuItemTime,
				CreatedAt:        orderItem.CreatedAt, // Preserve the original CreatedAt timestamp from the order item
			}

			responseOrder.OrderItems = append(responseOrder.OrderItems, responseOrderItem)

		}

		// Append the responseOrder to the responseOrders
		responseOrders = append(responseOrders, responseOrder)
	}

	// Send the response with the pending orders and their associated menu items
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseOrders)
}




func UpdateOrder(w http.ResponseWriter, r *http.Request) {
	// Parse URL parameters
	params := mux.Vars(r)

	orderIDStr := params["orderID"]

	// Parse the order ID from the URL parameter
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		http.Error(w, "Invalid orderID", http.StatusBadRequest)
		return
	}

	// Parse JSON request for the updated order details
	var updatedOrder user1.Order
	err = json.NewDecoder(r.Body).Decode(&updatedOrder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the existing order from the database
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	var existingOrder user1.Order
	result := connection.First(&existingOrder, orderID)
	if result.Error != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	// Update the "orderPayStatus" field
	existingOrder.OrderItems[0].OrderPayStatus = "success" // Assuming you want to update the first order item

	// Save the updated order back to the database
	result = connection.Save(&existingOrder)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Return the updated order as a JSON response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(existingOrder)
}

func GetPaymentPendingOrdersForTables(w http.ResponseWriter, r *http.Request) {
	// Parse URL parameters
	params := mux.Vars(r)

	tableIDStr := params["tableID"]
	tableID, err := strconv.Atoi(tableIDStr)
	if err != nil {
		http.Error(w, "Invalid tableID", http.StatusBadRequest)
		return
	}

	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	// Retrieve all orders with OrderStatus "paymentPending" for the given TableNoID
	var paymentPendingOrders []user1.Order
	result := connection.Where("table_no_id = ? AND order_status = ?", tableID, "paymentPending").Find(&paymentPendingOrders)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Retrieve the highest order based on tableID
	var highestOrder user1.Order
	result = connection.Where("table_no_id = ?", tableID).Order("id DESC").First(&highestOrder)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}
	result = connection.Model(&highestOrder).Related(&highestOrder.OrderItems)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Create a struct to represent an order with order items
	type OrderWithItems struct {
		Order      user1.Order       `json:"order"`
		OrderItems []user1.OrderItem `json:"orderItems"`
	}

	// Create a slice to hold the orders with order items
	var ordersWithItems []OrderWithItems

	// Populate OrderItems for each paymentPendingOrder
	for _, order := range paymentPendingOrders {
		// Fetch the associated OrderItems for the current order
		result := connection.Model(&order).Related(&order.OrderItems)
		if result.Error != nil {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
			return
		}

		// Append the order and its order items to the slice
		ordersWithItems = append(ordersWithItems, OrderWithItems{
			Order:      order,
			OrderItems: order.OrderItems,
		})
	}

	// Create the response struct with the list of orders and their order items
	response := struct {
		OrdersWithItems []OrderWithItems `json:"ordersWithItems"`
		HighestOrder    user1.Order            `json:"highestOrder"`
	}{
		OrdersWithItems: ordersWithItems,
		HighestOrder:    highestOrder,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func AddToCartIDtry(w http.ResponseWriter, r *http.Request) {
	// Parse URL parameters
	params := mux.Vars(r)

	tableIDStr := params["tableID"]
	tableID, err := strconv.Atoi(tableIDStr)
	if err != nil {
		http.Error(w, "Invalid tableID", http.StatusBadRequest)
		return
	}
	clientIDStr := params["clientID"]
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		http.Error(w, "Invalid clientID", http.StatusBadRequest)
		return
	}

	// Parse JSON request containing items to be added to the cart/order
	var items []struct {
		MenuItemID       int     `json:"menuItemID"`
		Quantity         int     `json:"quantity"`
		MenuItemPrice    float64 `json:"menuItemPrice"`
		MenuItemTime     int     `json:"menuItemTime"`
		MenuItemItemName string  `json:"menuItemItemName"`
	}
	err = json.NewDecoder(r.Body).Decode(&items)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the table based on the table ID
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	var table user1.TableNo
	result := connection.Where("id = ? AND client_id =?", tableID, clientID).First(&table)
	if result.Error != nil {
		http.Error(w, "Table not found", http.StatusNotFound)
		return
	}

	// Create a new order
	var order user1.Order
	order.TableNoID = table.ID
	order.TableNoTableName = table.TableName
	order.ClientID = table.ClientID
	order.OrderStatus = "PaymentPending"
	order.Status = "Pending"

	result = connection.Create(&order)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate the total price and create order items
	var totalPrice float64
	var totalQuantity int
	var highestMenuItemTime int
	var pendingPaymentTotal float64

	for _, item := range items {
		var menuItem user1.MenuItem
		result := connection.First(&menuItem, item.MenuItemID)
		if result.Error != nil {
			http.Error(w, "Menu item not found", http.StatusNotFound)
			return
		}

		menuItemTime := item.Quantity * item.MenuItemTime
		orderItemPrice := item.MenuItemPrice * float64(item.Quantity)
		totalPrice += orderItemPrice
		totalQuantity += item.Quantity
		orderItemStatus := "OrderPending"
		orderPayStatus := "PaymentPending"
		uniqueID := 1

		// Create the order item
		orderItem := user1.OrderItem{
			OrderID:          order.ID,
			ClientID:         table.ClientID,
			MenuItemID:       menuItem.ID,
			UniqueID:         uniqueID,
			MenuItemItemName: item.MenuItemItemName,
			MenuItemPrice:    orderItemPrice,
			Quantity:         item.Quantity,
			OrderItemStatus:  orderItemStatus,
			OrderPayStatus:   orderPayStatus,
			MenuItemTime:     menuItemTime,
		}

		// Save the order item
		result = connection.Create(&orderItem)
		if result.Error != nil {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
			return
		}

		if menuItemTime > highestMenuItemTime {
			highestMenuItemTime = menuItemTime
		}

		if orderPayStatus == "PaymentPending" {
			pendingPaymentTotal += orderItemPrice
		}
	}

	order.TotalPrice = totalPrice
	order.TotalQuantity = totalQuantity
	order.Balprice = pendingPaymentTotal
	order.Number = +1

	// Save the updated order
	result = connection.Save(&order)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}
	var orderItems []user1.OrderItem
	connection.Where("order_id = ?", order.ID).Find(&orderItems)

	// Combine order and orderItems into a single response
	response := struct {
		Order      user1.Order       `json:"order"`
		OrderItems []user1.OrderItem `json:"orderItems"`
	}{
		Order:      order,
		OrderItems: orderItems,
	}

	
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}



func AddOrderItemsByID(w http.ResponseWriter, r *http.Request) {
	// Parse URL parameters
	params := mux.Vars(r)

	orderIDStr := params["orderID"]
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		http.Error(w, "Invalid orderID", http.StatusBadRequest)
		return
	}

	// Parse JSON request to get the new order items
	var newOrderItems []struct {
		MenuItemID       int     `json:"menuItemID"`
		Quantity         int     `json:"quantity"`
		MenuItemPrice    float64 `json:"menuItemPrice"`
		MenuItemTime     int     `json:"menuItemTime"`
		MenuItemItemName string  `json:"menuItemItemName"`
	}
	err = json.NewDecoder(r.Body).Decode(&newOrderItems)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the order based on the order ID
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	var order user1.Order
	result := connection.Where("id = ?", orderID).First(&order)
	if result.Error != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	var totalPrice float64
	var totalQuantity int
	var highestMenuItemTime int
	var pendingPaymentTotal float64

	// Find the highest existing unique ID for OrderItem within the same order
	var highestUniqueID int
	var orderItems []user1.OrderItem
	result = connection.Model(&user1.OrderItem{}).Where("order_id = ?", order.ID).Find(&orderItems)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate the highest unique ID
	for _, item := range orderItems {
		if item.UniqueID > highestUniqueID {
			highestUniqueID = item.UniqueID
		}
	}

	// Increment the highest unique ID to generate a new unique ID
	newUniqueID := highestUniqueID + 1

	for _, item := range newOrderItems {
		var menuItem user1.MenuItem
		result := connection.First(&menuItem, item.MenuItemID)
		if result.Error != nil {
			http.Error(w, "Menu item not found", http.StatusNotFound)
			return
		}

		// Calculate menuItemTime based on quantity
		menuItemTime := item.Quantity * item.MenuItemTime

		// Calculate the Price for the orderItem
		orderItemPrice := item.MenuItemPrice * float64(item.Quantity)

		totalQuantity += item.Quantity
		totalPrice += orderItemPrice
		orderItemStatus := "OrderPending"
		orderPayStatus := "PaymentPending"

		orderItem := user1.OrderItem{
			OrderID:    order.ID,
			ClientID:   order.ClientID,
			MenuItemID: menuItem.ID,

			UniqueID:         newUniqueID, // Use the new unique ID
			MenuItemItemName: item.MenuItemItemName,
			MenuItemPrice:    orderItemPrice, // Calculate Price here
			Quantity:         item.Quantity,
			OrderPayStatus:   orderPayStatus,
			OrderItemStatus:  orderItemStatus,
			MenuItemTime:     menuItemTime,
		}

		// Update highestMenuItemTime if needed
		if menuItemTime > highestMenuItemTime {
			highestMenuItemTime = menuItemTime
		}

		if orderPayStatus == "PaymentPending" {
			pendingPaymentTotal += orderItemPrice
		}

		// Create and save the new order item
		result = connection.Create(&orderItem)
		if result.Error != nil {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Update the total price, order time, total quantity, and pending payment total for the existing order
	order.TotalPrice += totalPrice

	order.TotalQuantity += totalQuantity
	order.Balprice = pendingPaymentTotal

	result = connection.Save(&order)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}
	// var orderItems []OrderItem
	connection.Where("order_id = ?", order.ID).Find(&orderItems)

	// Combine order and orderItems into a single response
	response := struct {
		Order      user1.Order       `json:"order"`
		OrderItems []user1.OrderItem `json:"orderItems"`
	}{
		Order:      order,
		OrderItems: orderItems,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
func UpdateOrderStatusAndPayStatus(w http.ResponseWriter, r *http.Request) {
    // Parse JSON request
	params := mux.Vars(r)
    orderIDStr := params["orderID"]
    orderID, errs := strconv.Atoi(orderIDStr)
    if errs != nil {
        http.Error(w, "Invalid orderID", http.StatusBadRequest)
        return
    }

    var updateRequest struct {
        // OrderID         uint   `json:"orderID"`
        NewOrderStatus  string `json:"newOrderStatus"`
    }

    err := json.NewDecoder(r.Body).Decode(&updateRequest)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    newOrderStatus := updateRequest.NewOrderStatus

    if orderID == 0 || newOrderStatus == "" {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Get the order based on the provided orderID
    connection := database.GetDatabase()
    defer database.CloseDatabase(connection)

    var order user1.Order
    result := connection.First(&order, orderID)
    if result.Error != nil {
        http.Error(w, "Order not found for the given orderID", http.StatusNotFound)
        return
    }

    // Update the order status
    order.OrderStatus = newOrderStatus
    result = connection.Save(&order)
    if result.Error != nil {
        http.Error(w, result.Error.Error(), http.StatusInternalServerError)
        return
    }

    // Get the order items based on the provided orderID
    var orderItems []user1.OrderItem
    result = connection.Where("order_id = ?", orderID).Find(&orderItems)
    if result.Error != nil {
        http.Error(w, "Order items not found for the given orderID", http.StatusNotFound)
        return
    }

    // Update the order pay status for all order items
    for _, orderItem := range orderItems {
        orderItem.OrderPayStatus = newOrderStatus // assuming orderPayStatus should be updated to the same value as orderStatus
        result = connection.Save(&orderItem)
        if result.Error != nil {
            http.Error(w, result.Error.Error(), http.StatusInternalServerError)
            return
        }
    }
	// var orderItems []OrderItem
    result = connection.Where("order_id = ?", orderID).Find(&orderItems)
    if result.Error != nil {
        http.Error(w, "Order items not found for the given orderID", http.StatusNotFound)
        return
    }

    // Update the order pay status for all order items
    for _, orderItem := range orderItems {
        orderItem.OrderPayStatus = newOrderStatus // assuming orderPayStatus should be updated to the same value as orderStatus
        result = connection.Save(&orderItem)
        if result.Error != nil {
            http.Error(w, result.Error.Error(), http.StatusInternalServerError)
            return
        }
    }

    // Include order items in the response
    order.OrderItems = orderItems

	response := struct {
        Order          user1.Order       `json:"order"`
        SuccessMessage string      `json:"successMessage"`
    }{
        Order:          order,
        SuccessMessage: "Order status and order pay status updated to '" + newOrderStatus + "' for orderID " + strconv.Itoa(orderID) + ".",
    }

    // Return a success response
    w.WriteHeader(http.StatusOK)
    // w.Write([]byte("Order status and order pay status updated to '" + newOrderStatus + "' for orderID " + strconv.FormatUint(uint64(orderID), 10) + "."))
	
    json.NewEncoder(w).Encode(response)
}

func UpdateOrderItemsStatusToPendings(w http.ResponseWriter, r *http.Request) {
	// Parse JSON request

	var updateRequest struct {
		OrderID        uint   `json:"orderID"`
		OrderItemIDs   []int  `json:"orderItemIDs"`
		NewOrderStatus string `json:"newOrderStatus"`
	}

	err := json.NewDecoder(r.Body).Decode(&updateRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Extract order ID, order item IDs, and new order status from the request
	orderID := updateRequest.OrderID
	orderItemIDs := updateRequest.OrderItemIDs
	newOrderStatus := updateRequest.NewOrderStatus

	if orderID == 0 || len(orderItemIDs) == 0 || newOrderStatus == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get the order items based on the provided orderItemIDs and orderID
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	var orderItems []user1.OrderItem
	result := connection.Where("id IN (?) AND order_id = ?", orderItemIDs, orderID).Find(&orderItems)
	if result.Error != nil {
		http.Error(w, "Order items not found for the given orderID and orderItemIDs", http.StatusNotFound)
		return
	}

	// Update the order item statuses
	for _, orderItem := range orderItems {
		// Update only the order item status, not the entire order
		orderItem.OrderItemStatus = newOrderStatus
		result = connection.Save(&orderItem)
		if result.Error != nil {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Check if any order items were updated
	if len(orderItems) > 0 {
		// Return a success response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Order items updated to '" + newOrderStatus + "' for orderID " + strconv.FormatUint(uint64(orderID), 10) + "."))
	} else {
		// No order items were updated
		http.Error(w, "No order items were updated for the given orderID and orderItemIDs", http.StatusNotFound)
	}
}


func UpdateOrderItemsStatusToPreparingnew(w http.ResponseWriter, r *http.Request) {
	// if r.Method != http.MethodPut {
    //     http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    //     return
    // }
	vars := mux.Vars(r)
	orderIDStr := vars["orderID"]
	uniqueIDStr := vars["uniqueID"]

	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		http.Error(w, "Invalid orderID", http.StatusBadRequest)
		return
	}
	uniqueID, err := strconv.Atoi(uniqueIDStr)
	if err != nil {
		http.Error(w, "Invalid uniqueID", http.StatusBadRequest)
		return
	}

	var updateRequest struct {
		OrderItemIDs   []int  `json:"orderItemIDs"`
		NewOrderStatus string `json:"newOrderStatus"`
	}


	err = json.NewDecoder(r.Body).Decode(&updateRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	orderItemIDs := updateRequest.OrderItemIDs
	newOrderStatus := updateRequest.NewOrderStatus

	if newOrderStatus != "OrderPending" && newOrderStatus != "Preparing" && newOrderStatus != "Success" {
		http.Error(w, "Invalid newOrderStatus", http.StatusBadRequest)
		return
	}

	if len(orderItemIDs) == 0 || newOrderStatus == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Assuming you have a function GetDatabase() that returns a database connection
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	var orderItems []user1.OrderItem
	result := connection.Where("id IN (?) AND order_id = ? AND unique_id = ?", orderItemIDs, orderID, uniqueID).Find(&orderItems)
	if result.Error != nil {
		http.Error(w, "Order items not found for the given orderID and createAt", http.StatusNotFound)
		return
	}

	var highestMenuItemTime int

	for _, orderItem := range orderItems {
		orderItem.OrderItemStatus = newOrderStatus

		if newOrderStatus == "Preparing" {
			currentTime := time.Now()

			orderItem.Preparedate = currentTime
			orderItem.OldUpdateDate = orderItem.UpdateDate
			orderItem.UpdateDate = currentTime

			// Calculate the highestMenuItemTime
			if orderItem.MenuItemTime > highestMenuItemTime {
				highestMenuItemTime = orderItem.MenuItemTime
			}

		} else if newOrderStatus == "Success" {
			currentTime := time.Now()
			orderItem.Successdate = currentTime
			orderItem.NowUpdateDate = currentTime

			// Calculate the highestMenuItemTime
			if orderItem.MenuItemTime > highestMenuItemTime {
				highestMenuItemTime = orderItem.MenuItemTime
			}
		} else {
			fmt.Println(err)
		}

		result := connection.Save(&orderItem)
		if result.Error != nil {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
			return
		}
	}

	if newOrderStatus == "Preparing" {
		var order user1.Order
		connection.First(&order, orderID)

		order.OldUpdateTime = order.UpdateTime

		currentTime := time.Now()

		order.OldUpdateDate = order.UpdateDate
		order.UpdateDate = currentTime

		UpdatedAt := order.OldUpdateDate.Format("2006-01-02T15:04:05.999999-07:00")
		NewUpdatedAt := order.UpdateDate.Format("2006-01-02T15:04:05.999999-07:00")

		timeFormat := "2006-01-02T15:04:05.999999-07:00"
		a, _ := time.Parse(timeFormat, UpdatedAt)
		b, _ := time.Parse(timeFormat, NewUpdatedAt)
		diff := b.Sub(a)
		// Calculate the new order.UpdateTime
		d := int(diff.Minutes())

		c := order.OldUpdateTime

		e := c - d
		if e < 0 {
			e = 0
		}
		g := e + highestMenuItemTime
		order.UpdateTime = g

		result := connection.Save(&order)
		if result.Error != nil {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
			return
		}
	} else if newOrderStatus == "Success" {
		var order user1.Order
		connection.First(&order, orderID)

		// order.OldUpdateTime = order.UpdateTime

		// currentTime := time.Now()

		UpdatedAt := order.OldUpdateDate.Format("2006-01-02T15:04:05.999999-07:00")
		NewUpdatedAt := order.UpdateDate.Format("2006-01-02T15:04:05.999999-07:00")

		NowUpdatedAt := order.NowUpdateDate.Format("2006-01-02T15:04:05.999999-07:00")

		timeFormat := "2006-01-02T15:04:05.999999-07:00"
		a, _ := time.Parse(timeFormat, UpdatedAt)
		b, _ := time.Parse(timeFormat, NewUpdatedAt)

		g, _ := time.Parse(timeFormat, NowUpdatedAt)
		diff := b.Sub(a)
		h := g.Sub(b)

		// Calculate the new order.UpdateTime
		d := int(diff.Minutes())
		i := int(h.Minutes())
		fmt.Println(d)

		fmt.Println(i)

		// c := order.OldUpdateTime

		// e:= c -d
		// if e < 0 {
		// 	e=0
		// }
		// g:= e + highestMenuItemTime
		// order.UpdateTime =g

		result := connection.Save(&order)
		if result.Error != nil {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
			return
		}
	}

	if len(orderItems) > 0 {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Order items updated to '" + newOrderStatus + "' for orderID " + strconv.Itoa(orderID) + "."))
	} else {
		http.Error(w, "No order items were updated for the given orderID, orderItemIDs, and uniqueID", http.StatusNotFound)
	}
}

func UpdateOrderItemsStatusnews(w http.ResponseWriter, r *http.Request) {
	// Parse URL parameters
	params := mux.Vars(r)

	clientIDStr := params["clientID"]
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		http.Error(w, "Invalid clientID", http.StatusBadRequest)
		return
	}

	uniqueIDStr := params["uniqueID"]
	uniqueID, err := strconv.Atoi(uniqueIDStr)
	if err != nil {
		http.Error(w, "Invalid uniqueID", http.StatusBadRequest)
		return
	}

	orderIDStr := params["orderID"]
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		http.Error(w, "Invalid orderID", http.StatusBadRequest)
		return
	}

	// Parse JSON request containing the new OrderItemStatus
	var requestBody struct {
		NewOrderStatus string `json:"newOrderStatus"`
	}

	err = json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate the new OrderItemStatus
	allowedStatusValues := map[string]bool{
		"OrderPending": true,
		"Preparing":    true,
		"Success":      true,
	}

	if !allowedStatusValues[requestBody.NewOrderStatus] {
		http.Error(w, "Invalid newOrderStatus", http.StatusBadRequest)
		return
	}

	// Get the order items with the specified clientID, uniqueID, and orderID
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	var orderItems []user1.OrderItem
	result := connection.Where("order_id = ? AND client_id = ? AND unique_id = ?", orderID, clientID, uniqueID).Find(&orderItems)
	if result.Error != nil {
		http.Error(w, "Order items not found", http.StatusNotFound)
		return
	}

	// Begin a transaction
	tx := connection.Begin()
	if tx.Error != nil {
		http.Error(w, "Failed to begin transaction", http.StatusInternalServerError)
		return
	}

	// Update the order item status for each item
	for _, orderItem := range orderItems {
		// Update the OrderItemStatus
		orderItem.OrderItemStatus = requestBody.NewOrderStatus

		// Update Preparedate if status changes to Preparing
		if requestBody.NewOrderStatus == "Preparing" {
			orderItem.Preparedate = time.Now()
			orderItem.Successdate = time.Now().Add(time.Duration(orderItem.MenuItemTime) * time.Minute)
		
		}

		// Update Successdate if status changes to Success
		if requestBody.NewOrderStatus == "Success" {
			orderItem.Successdate = time.Now()
			orderItem.Preparedate = time.Now()
		}

		// Save the updated order item
		result := tx.Save(&orderItem)
		if result.Error != nil {
			tx.Rollback()
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orderItems)
}




func UpdateOrderItemsStatusToPendingsnewsssorg(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderIDStr := vars["orderID"] // Update the variable name to match the URL parameter

	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		http.Error(w, "Invalid orderID", http.StatusBadRequest)
		return
	}

	// Parse createAt from the URL path
	// Use the timestamp format from the URL

	var updateRequest struct {
		OrderItemIDs   []int  `json:"orderItemIDs"`
		NewOrderStatus string `json:"newOrderStatus"`
	}

	err = json.NewDecoder(r.Body).Decode(&updateRequest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	orderItemIDs := updateRequest.OrderItemIDs
	newOrderStatus := updateRequest.NewOrderStatus

	if len(orderItemIDs) == 0 || newOrderStatus == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	var orderItems []user1.OrderItem
	// Assuming OrderItem model matches your database structure
	result := connection.Where("id IN (?) AND order_id = ? ", orderItemIDs, orderID).Find(&orderItems)
	if result.Error != nil {
		http.Error(w, "Order items not found for the given orderID ", http.StatusNotFound)
		return
	}

	for _, orderItem := range orderItems {
		orderItem.OrderItemStatus = newOrderStatus
		result := connection.Save(&orderItem)
		if result.Error != nil {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
			return
		}
	}

	if len(orderItems) > 0 {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Order items updated to '" + newOrderStatus + "' for orderID " + strconv.Itoa(orderID) + "."))
	} else {
		http.Error(w, "No order items were updated for the given orderID, orderItemIDs, and createAt", http.StatusNotFound)
	}
}





func GetOrdersByClientIDAndStatusnew(w http.ResponseWriter, r *http.Request) {
	// Parse URL parameters
	params := mux.Vars(r)

	clientIDStr := params["clientID"]
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		http.Error(w, "Invalid clientID", http.StatusBadRequest)
		return
	}

	// Get orders from the database based on clientID and today's date
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	// Calculate the start and end of today
	today := time.Now()
	startOfDay := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	endOfDay := time.Date(today.Year(), today.Month(), today.Day()+1, 0, 0, 0, 0, today.Location())

	var orders []user1.Order
	connection.Preload("OrderItems").Where("client_id = ? AND created_at BETWEEN ? AND ?", clientID, startOfDay, endOfDay).Find(&orders)

	// Organize orders by orderStatus and ID-wise arrangement
	ordersByStatus := make(map[string][]user1.Order)

	for _, order := range orders {
		ordersByStatus[order.OrderStatus] = append(ordersByStatus[order.OrderStatus], order)
	}

	// Sort orders within each orderStatus by ID
	for _, statusOrders := range ordersByStatus {
		sort.Slice(statusOrders, func(i, j int) bool {
			return statusOrders[i].ID < statusOrders[j].ID
		})
	}

	// Calculate total order length and orderStatus wise length
	var totalOrderLen int
	orderStatusLen := make(map[string]int)
	var nontotalOrderLen int

	for status, statusOrders := range ordersByStatus {
		orderStatusLen[status] = len(statusOrders)
		totalOrderLen += len(statusOrders)
		if status != "Success" {
			nontotalOrderLen += len(statusOrders)
		}
	}

	// Create a response struct to include the required information
	response := struct {
		TotalOrders     int                  `json:"totalOrders"`
		NontotalOrderLen int                  `json:"nontotalOrderLen"`
		OrderStatusLen  map[string]int       `json:"orderStatusLen"`
		// OrdersByStatus  map[string][]Order   `json:"ordersByStatus"`
	}{
		TotalOrders:     totalOrderLen,
		NontotalOrderLen: nontotalOrderLen,
		OrderStatusLen:  orderStatusLen,
		// OrdersByStatus:  ordersByStatus,
	}

	// Respond with the structured response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}


func GetOrdersByClientIDAndStatus(w http.ResponseWriter, r *http.Request) {
	// Parse URL parameters
	params := mux.Vars(r)

	clientIDStr := params["clientID"]
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		http.Error(w, "Invalid clientID", http.StatusBadRequest)
		return
	}

	// Get orders from the database based on clientID and today's date
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	// Calculate the start and end of today
	today := time.Now()
	startOfDay := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	endOfDay := time.Date(today.Year(), today.Month(), today.Day()+1, 0, 0, 0, 0, today.Location())

	var orders []user1.Order
	connection.Preload("OrderItems").Where("client_id = ? AND created_at BETWEEN ? AND ?", clientID, startOfDay, endOfDay).Find(&orders)

	// Organize orders by orderStatus and ID-wise arrangement
	ordersByStatus := make(map[string][]user1.Order)

	for _, order := range orders {
		ordersByStatus[order.OrderStatus] = append(ordersByStatus[order.OrderStatus], order)
	}

	// Sort orders within each orderStatus by ID
	for _, statusOrders := range ordersByStatus {
		sort.Slice(statusOrders, func(i, j int) bool {
			return statusOrders[i].ID < statusOrders[j].ID
		})
	}

	// Calculate total order length and orderStatus wise length
	var totalOrderLen int
	orderStatusLen := make(map[string]int)
	var nontotalOrderLen int

	for status, statusOrders := range ordersByStatus {
		orderStatusLen[status] = len(statusOrders)
		totalOrderLen += len(statusOrders)
		if status != "Success" {
			nontotalOrderLen += len(statusOrders)
		}
	}

	// Create a response struct to include the required information
	response := struct {
		TotalOrders     int                  `json:"totalOrders"`
		NontotalOrderLen int                 `json:"nontotalOrderLen"`
		OrderStatusLen  map[string]int       `json:"orderStatusLen"`
		OrdersByStatus  map[string][]user1.Order   `json:"ordersByStatus"`
	}{
		TotalOrders:     totalOrderLen,
		NontotalOrderLen: nontotalOrderLen,
		OrderStatusLen:  orderStatusLen,
		OrdersByStatus:  ordersByStatus,
	}

	// Respond with the structured response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetOrdersWithOrderItemStatusOrderPendingssnew(w http.ResponseWriter, r *http.Request) {
	// Parse URL parameters
	params := mux.Vars(r)

	clientIDStr := params["clientID"]

	// Parse the client ID from the URL parameter
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		http.Error(w, "Invalid clientID", http.StatusBadRequest)
		return
	}

	// Get distinct orders with the specified client ID and status "OrderPending" or "Preparing"
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	var orders []user1.Order
	result := connection.Joins("JOIN order_items ON orders.id = order_items.order_id").
		Where("orders.client_id = ? AND order_items.order_item_status IN (?)", clientID, []string{"OrderPending", "Preparing"}).
		Select("DISTINCT orders.*").
		Find(&orders)
	if result.Error != nil {
		http.Error(w, "No orders with OrderPending or Preparing status found", http.StatusNotFound)
		return
	}

	// Create the response structure
	type ResponseOrderItem struct {
		CreatedTime string      `json:"createdTime"`
		OrderItems  []user1.OrderItem `json:"orderItems"`
	}

	type ResponseOrder struct {
		ID           uint             `json:"id"`
		Status       string           `json:"status"`
		TableNoID    uint             `json:"tableNoID"`
		TotalPrice   float64          `json:"totalPrice"`
		BalPrice     float64          `json:"balPrice"`
		TableName    string           `json:"tableName"`
		OrderTime    int              `json:"orderTime"`
		Created      string           `json:"created"`
		BalTime      int              `json:"balTime"`
		UniqueID     int              `json:"uniqueID"`
		BalQuantity  int              `json:"balQuantity"`
		CreatedAt    string           `json:"createdAt"` // Include the created date
		OrderItems   []user1.OrderItem      `json:"orderItems"`
	}

	var responseOrders []ResponseOrder

	for _, order := range orders {
		var orderItems []user1.OrderItem
		result := connection.Model(&order).Where("order_items.order_item_status IN (?)", []string{"OrderPending", "Preparing"}).Related(&orderItems)
		if result.Error != nil {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
			return
		}

		// Only include orders that have at least one "OrderPending" or "Preparing" item
		if len(orderItems) > 0 {
			// Create a map to group order items by date and time
			itemGroups := make(map[string]ResponseOrderItem)
			for _, item := range orderItems {
				// Convert the item's CreatedAt timestamp to a string
				createdAtStr := item.CreatedAt.Format(time.RFC3339)

				// Format the timestamp in the desired format
				createdAtTime, err := time.Parse(time.RFC3339, createdAtStr)
				if err != nil {
					http.Error(w, "Error parsing timestamp", http.StatusInternalServerError)
					return
				}

				createdTime := createdAtTime.Format("2006-01-02T15:04:05.999999-07:00")

				if group, ok := itemGroups[createdTime]; !ok {
					itemGroups[createdTime] = ResponseOrderItem{
						CreatedTime: createdTime,
						OrderItems:  []user1.OrderItem{item},
					}

				} else {
					group.OrderItems = append(group.OrderItems, item)
					itemGroups[createdTime] = group
				}
			}

			// Create a response order for each group of order items
			for createdAt, responseItem := range itemGroups {

				var highestMenuItemTime int    // Variable to store the highest MenuItemTime
				var totalMenuItemPrice float64 // Variable to store the total MenuItemPrice
				var totalMenuItemQuantity int  // Variable to store the total MenuItemQuantity

				// Iterate through order items to calculate the total MenuItemPrice, quantity, and highestMenuItemTime
				for _, item := range responseItem.OrderItems {
					totalMenuItemPrice += item.MenuItemPrice
					totalMenuItemQuantity += item.Quantity
					// Calculate the difference between Successdate and Preparedate for each order item
					diff := item.Successdate.Sub(item.Preparedate)
					d := int(diff.Minutes())
					if d > highestMenuItemTime {
						highestMenuItemTime = d
					}
				}

				responseOrder := ResponseOrder{
					ID:           order.ID,
					Status:       order.OrderStatus,
					TableNoID:    order.TableNoID,
					TotalPrice:   order.TotalPrice,
					BalPrice:     totalMenuItemPrice,
					TableName:    order.TableNoTableName,
					Created:      createdAt,
					BalTime:      highestMenuItemTime,
					BalQuantity:  totalMenuItemQuantity,
					UniqueID:     responseItem.OrderItems[0].UniqueID,
					CreatedAt:    responseItem.OrderItems[0].CreatedAt.Format("2006-01-02T15:04:05.999999-07:00"),
					OrderItems:   responseItem.OrderItems,
				}

				responseOrders = append(responseOrders, responseOrder)
			}
		}
	}
	sort.Slice(responseOrders, func(i, j int) bool {
		timeI, _ := time.Parse(time.RFC3339, responseOrders[i].CreatedAt)
		timeJ, _ := time.Parse(time.RFC3339, responseOrders[j].CreatedAt)
		return timeI.Before(timeJ)
	})

	// Send the response with distinct orders and their filtered and grouped order items
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseOrders)
}
func GetHighestMenuItemTimeForUniqueID(w http.ResponseWriter, r *http.Request) {
    // Parse URL parameters
    params := mux.Vars(r)

    clientIDStr := params["clientID"]
    orderIDStr := params["orderID"]

    // Parse the client ID and order ID from the URL parameters
    clientID, err := strconv.Atoi(clientIDStr)
    if err != nil {
        http.Error(w, "Invalid clientID", http.StatusBadRequest)
        return
    }

    orderID, err := strconv.Atoi(orderIDStr)
    if err != nil {
        http.Error(w, "Invalid orderID", http.StatusBadRequest)
        return
    }

    // Get order details with the specified clientID and orderID
    connection := database.GetDatabase()
    defer database.CloseDatabase(connection)

    var order user1.Order
    result := connection.Where("client_id = ? AND id = ?", clientID, orderID).First(&order)
    if result.Error != nil {
        http.Error(w, "Order not found", http.StatusNotFound)
        return
    }

    // Get order items associated with the order
    var orderItems []user1.OrderItem
    connection.Model(&order).Related(&orderItems)

    // Group order items by UniqueID and calculate the highest MenuItemTime for each group
    highestMenuItemTimes := make(map[int]int)
    for _, item := range orderItems {
        // Calculate the difference between Successdate and the current time
        now := time.Now()
        diff := item.Successdate.Sub(now)
        c:= int(diff.Minutes())
		d:= (c+1)
        if d < 0 {
            // If the difference is negative, it means Successdate is in the future, so use 0
            d = 0
        }
        // Update or set the highest menu item time for the unique ID
        if existingTime, ok := highestMenuItemTimes[item.UniqueID]; ok {
            if d > existingTime {
                highestMenuItemTimes[item.UniqueID] = d 
            }
        } else {
            highestMenuItemTimes[item.UniqueID] = d 
        }
    }

    // Construct response JSON
    type OrderItemResponse struct {
        ID        uint   `json:"id"`
        UniqueID  int    `json:"uniqueID"`
        SuccessDate time.Time `json:"successdate"`
		OrderID          uint
		MenuItemID       uint
		Quantity         int
		OrderItemStatus  string    `json:"orderItemStatus"`
		OrderPayStatus   string    `json:"orderPayStatus"`
		MenuItemItemName string    `json:"menuItemItemName"`
		MenuItemPrice    float64   `json:"menuItemPrice"`
		MenuItemTime     int       `json:"menuItemTime"`
		UpdateDate       time.Time `json:"updateDate"`
		Preparedate      time.Time `json:"preparedate"`
		UpdateTime       int
		OldUpdateDate    time.Time `json:"oldUpdateDate"`
		NowUpdateDate time.Time `json:"nowUpdateDate"`
		OldUpdateTime int
		OrderTime     int
		Time          int
	
    }


	type OrderResponse struct {
		ID         uint                `json:"id"`
		ClientID   uint                `json:"clientID"`
		TableNoID  uint                `json:"tableNoID"`
		TableNoTableName string        `json:"tableNoTableName"`
		TotalPrice float64             `json:"totalPrice"`
		OrderStatus string             `json:"orderStatus"`
		UpdateDate time.Time           `json:"updateDate"`
		OldUpdateTime    int
		UpdateTime       int 
		Balprice         float64 `json:"balprice"`
		TotalQuantity int
		OrderTime     int
		Status    string
		TableName string `json:"tableName"` // Add this field
		OrderItems []OrderItemResponse `json:"orderItems"`
	}
	
    var orderItemsResponse []OrderItemResponse
    for _, item := range orderItems {
        orderItemsResponse = append(orderItemsResponse, OrderItemResponse{
            ID:         item.ID,
            UniqueID:   item.UniqueID,
            SuccessDate: item.Successdate,
			OrderID : item.OrderID      ,  
			MenuItemID:item.MenuItemID ,
			Quantity: item.Quantity,
			OrderItemStatus : item.OrderItemStatus, 
			OrderPayStatus   : item.OrderPayStatus,
			MenuItemItemName : item.MenuItemItemName,
			MenuItemPrice   : item.MenuItemPrice,
			MenuItemTime    : item.MenuItemTime,
			UpdateDate      : item.UpdateDate,
			Preparedate    : item.Preparedate,
			UpdateTime      : item.UpdateTime,
			OldUpdateDate   : item.OldUpdateDate,
			NowUpdateDate : item.NowUpdateDate,
			OldUpdateTime : item.OldUpdateTime,
			OrderTime     : item.OrderTime,
			Time          : item.Time,
			
		
        })
    }
	
	orderResponse := OrderResponse{
		ID:               order.ID,
		ClientID:         order.ClientID,
		TableNoID:        order.TableNoID,
		TableNoTableName: order.TableNoTableName,
		TotalPrice:       order.TotalPrice,
		OrderStatus:      order.OrderStatus,
		UpdateDate:       order.UpdateDate,	
		OldUpdateTime:       order.OldUpdateTime,
		UpdateTime:      order.UpdateTime,
		Balprice:       order.Balprice,
		TotalQuantity:       order.TotalQuantity,
		OrderTime:      order.OrderTime,
		Status:       order.Status,
		TableName:       order.TableName,
		OrderItems:       orderItemsResponse,
	}
    // Include total HighestMenuItemTime in the response
    var response struct {
        Order     OrderResponse `json:"order"`
        TotalHighestMenuItemTime int `json:"totalHighestMenuItemTime"`
    }
    response.Order = orderResponse

    totalHighestMenuItemTime := 0 // Initialize total highest menu item time
    for _, highestTime := range highestMenuItemTimes {
        // Accumulate total highest menu item time
        totalHighestMenuItemTime += highestTime
    }
    response.TotalHighestMenuItemTime = totalHighestMenuItemTime

    // Send the response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}


func GetOrdersWithOrderItemStatusOrderPendingssnewtimenow(w http.ResponseWriter, r *http.Request) {
	// Parse URL parameters
	params := mux.Vars(r)

	clientIDStr := params["clientID"]

	// Parse the client ID from the URL parameter
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		http.Error(w, "Invalid clientID", http.StatusBadRequest)
		return
	}

	// Get distinct orders with the specified client ID and status "OrderPending" or "Preparing"
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	var orders []user1.Order
	result := connection.Joins("JOIN order_items ON orders.id = order_items.order_id").
		Where("orders.client_id = ? AND order_items.order_item_status IN (?)", clientID, []string{"OrderPending", "Preparing"}).
		Select("DISTINCT orders.*").
		Find(&orders)
	if result.Error != nil {
		http.Error(w, "No orders with OrderPending or Preparing status found", http.StatusNotFound)
		return
	}

	// Create the response structure
	type ResponseOrderItem struct {
		CreatedTime string            `json:"createdTime"`
		OrderItems  []user1.OrderItem `json:"orderItems"`
	}

	type ResponseOrder struct {
		ID           uint             `json:"id"`
		Status       string           `json:"status"`
		TableNoID    uint             `json:"tableNoID"`
		TotalPrice   float64          `json:"totalPrice"`
		BalPrice     float64          `json:"balPrice"`
		TableName    string           `json:"tableName"`
		OrderTime    int              `json:"orderTime"`
		Created      string           `json:"created"`
		BalTime      int              `json:"balTime"`
		UniqueID     int              `json:"uniqueID"`
		BalQuantity  int              `json:"balQuantity"`
		CreatedAt    string           `json:"createdAt"` // Include the created date
		OrderItems   []user1.OrderItem `json:"orderItems"`
	}

	var responseOrders []ResponseOrder

	for _, order := range orders {
		var orderItems []user1.OrderItem
		result := connection.Model(&order).Where("order_items.order_item_status IN (?)", []string{"OrderPending", "Preparing"}).Related(&orderItems)
		if result.Error != nil {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
			return
		}

		// Only include orders that have at least one "OrderPending" or "Preparing" item
		if len(orderItems) > 0 {
			// Create a map to group order items by date and time
			itemGroups := make(map[string]ResponseOrderItem)
			for _, item := range orderItems {
				// Convert the item's CreatedAt timestamp to a string
				createdAtStr := item.CreatedAt.Format(time.RFC3339)

				// Format the timestamp in the desired format
				createdAtTime, err := time.Parse(time.RFC3339, createdAtStr)
				if err != nil {
					http.Error(w, "Error parsing timestamp", http.StatusInternalServerError)
					return
				}

				createdTime := createdAtTime.Format("2006-01-02T15:04:05.999999-07:00")

				if group, ok := itemGroups[createdTime]; !ok {
					itemGroups[createdTime] = ResponseOrderItem{
						CreatedTime: createdTime,
						OrderItems:  []user1.OrderItem{item},
					}

				} else {
					group.OrderItems = append(group.OrderItems, item)
					itemGroups[createdTime] = group
				}
			}

			// Create a response order for each group of order items
			for createdAt, responseItem := range itemGroups {

				var highestMenuItemTime int    // Variable to store the highest MenuItemTime
				var totalMenuItemPrice float64 // Variable to store the total MenuItemPrice
				var totalMenuItemQuantity int  // Variable to store the total MenuItemQuantity

				// Iterate through order items to calculate the total MenuItemPrice, quantity, and highestMenuItemTime
				for _, item := range responseItem.OrderItems {
					totalMenuItemPrice += item.MenuItemPrice
					totalMenuItemQuantity += item.Quantity
					new := time.Now()
					// Calculate the difference between current time and Successdate for each order item
					diff := new.Sub(item.Successdate)
					d := int(diff.Minutes())
					if d > highestMenuItemTime {
						highestMenuItemTime = d
					}
				}

				responseOrder := ResponseOrder{
					ID:           order.ID,
					Status:       order.OrderStatus,
					TableNoID:    order.TableNoID,
					TotalPrice:   order.TotalPrice,
					BalPrice:     totalMenuItemPrice,
					TableName:    order.TableNoTableName,
					Created:      createdAt,
					BalTime:      highestMenuItemTime,
					BalQuantity:  totalMenuItemQuantity,
					UniqueID:     responseItem.OrderItems[0].UniqueID,
					CreatedAt:    responseItem.OrderItems[0].CreatedAt.Format("2006-01-02T15:04:05.999999-07:00"),
					OrderItems:   responseItem.OrderItems,
				}

				responseOrders = append(responseOrders, responseOrder)
			}
		}
	}
	sort.Slice(responseOrders, func(i, j int) bool {
		timeI, _ := time.Parse(time.RFC3339, responseOrders[i].CreatedAt)
		timeJ, _ := time.Parse(time.RFC3339, responseOrders[j].CreatedAt)
		return timeI.Before(timeJ)
	})

	// Send the response with distinct orders and their filtered and grouped order items
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseOrders)
}


func GetOrdersWithOrderItemStatusOrderPendingssnewtimenows(w http.ResponseWriter, r *http.Request) {
	// Parse URL parameters
	params := mux.Vars(r)

	clientIDStr := params["clientID"]

	// Parse the client ID from the URL parameter
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		http.Error(w, "Invalid clientID", http.StatusBadRequest)
		return
	}

	// Get distinct orders with the specified client ID and status "OrderPending" or "Preparing"
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	var orders []user1.Order
	result := connection.Joins("JOIN order_items ON orders.id = order_items.order_id").
		Where("orders.client_id = ? AND order_items.order_item_status IN (?)", clientID, []string{"OrderPending", "Preparing"}).
		Select("DISTINCT orders.*").
		Find(&orders)
	if result.Error != nil {
		http.Error(w, "No orders with OrderPending or Preparing status found", http.StatusNotFound)
		return
	}

	// Create the response structure
	type ResponseOrderItem struct {
		CreatedTime string      `json:"createdTime"`
		OrderItems  []user1.OrderItem `json:"orderItems"`
	}

	type ResponseOrder struct {
		ID           uint             `json:"id"`
		Status       string           `json:"status"`
		TableNoID    uint             `json:"tableNoID"`
		TotalPrice   float64          `json:"totalPrice"`
		BalPrice     float64          `json:"balPrice"`
		TableName    string           `json:"tableName"`
		OrderTime    int              `json:"orderTime"`
		Created      string           `json:"created"`
		BalTime      int              `json:"balTime"`
		UniqueID     int              `json:"uniqueID"`
		BalQuantity  int              `json:"balQuantity"`
		CreatedAt    string           `json:"createdAt"` // Include the created date
		OrderItems   []user1.OrderItem      `json:"orderItems"`
	}

	var responseOrders []ResponseOrder

	for _, order := range orders {
		var orderItems []user1.OrderItem
		result := connection.Model(&order).Where("order_items.order_item_status IN (?)", []string{"OrderPending", "Preparing"}).Related(&orderItems)
		if result.Error != nil {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
			return
		}

		// Only include orders that have at least one "OrderPending" or "Preparing" item
		if len(orderItems) > 0 {
			// Create a map to group order items by date and time
			itemGroups := make(map[string]ResponseOrderItem)
			for _, item := range orderItems {
				// Convert the item's CreatedAt timestamp to a string
				createdAtStr := item.CreatedAt.Format(time.RFC3339)

				// Format the timestamp in the desired format
				createdAtTime, err := time.Parse(time.RFC3339, createdAtStr)
				if err != nil {
					http.Error(w, "Error parsing timestamp", http.StatusInternalServerError)
					return
				}

				createdTime := createdAtTime.Format("2006-01-02T15:04:05.999999-07:00")

				if group, ok := itemGroups[createdTime]; !ok {
					itemGroups[createdTime] = ResponseOrderItem{
						CreatedTime: createdTime,
						OrderItems:  []user1.OrderItem{item},
					}

				} else {
					group.OrderItems = append(group.OrderItems, item)
					itemGroups[createdTime] = group
				}
			}

			// Create a response order for each group of order items
			for createdAt, responseItem := range itemGroups {

				var highestMenuItemTime int    // Variable to store the highest MenuItemTime
				var totalMenuItemPrice float64 // Variable to store the total MenuItemPrice
				var totalMenuItemQuantity int  // Variable to store the total MenuItemQuantity

				// Iterate through order items to calculate the total MenuItemPrice, quantity, and highestMenuItemTime
				for _, item := range responseItem.OrderItems {
					totalMenuItemPrice += item.MenuItemPrice
					totalMenuItemQuantity += item.Quantity
					new := time.Now()
					// Calculate the difference between current time and Successdate for each order item
					// diff := new.Sub(item.Successdate)
					// Calculate the difference between Successdate and Preparedate for each order item
					diff := item.Successdate.Sub(new)
					c := int(diff.Minutes())
					d := (c+1) 
					if d > highestMenuItemTime {
						highestMenuItemTime = d
					}
				}

				responseOrder := ResponseOrder{
					ID:           order.ID,
					Status:       order.OrderStatus,
					TableNoID:    order.TableNoID,
					TotalPrice:   order.TotalPrice,
					BalPrice:     totalMenuItemPrice,
					TableName:    order.TableNoTableName,
					Created:      createdAt,
					BalTime:      highestMenuItemTime,
					BalQuantity:  totalMenuItemQuantity,
					UniqueID:     responseItem.OrderItems[0].UniqueID,
					CreatedAt:    responseItem.OrderItems[0].CreatedAt.Format("2006-01-02T15:04:05.999999-07:00"),
					OrderItems:   responseItem.OrderItems,
				}

				responseOrders = append(responseOrders, responseOrder)
			}
		}
	}
	sort.Slice(responseOrders, func(i, j int) bool {
		timeI, _ := time.Parse(time.RFC3339, responseOrders[i].CreatedAt)
		timeJ, _ := time.Parse(time.RFC3339, responseOrders[j].CreatedAt)
		return timeI.Before(timeJ)
	})

	// Send the response with distinct orders and their filtered and grouped order items
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseOrders)
}

func GetOrdersWithOrderItemStatusOrderPendingss(w http.ResponseWriter, r *http.Request) {
	// Parse URL parameters
	params := mux.Vars(r)

	clientIDStr := params["clientID"]

	// Parse the client ID from the URL parameter
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		http.Error(w, "Invalid clientID", http.StatusBadRequest)
		return
	}

	// Get distinct orders with the specified client ID and status "OrderPending"
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	var orders []user1.Order
	result := connection.Joins("JOIN order_items ON orders.id = order_items.order_id").
		Where("orders.client_id = ? AND order_items.order_item_status IN (?)", clientID, []string{"OrderPending", "Preparing"}).
		Select("DISTINCT orders.*").
		Find(&orders)
	if result.Error != nil {
		http.Error(w, "No orders with OrderPending status found", http.StatusNotFound)
		return
	}

	// Create the response structure
	type ResponseOrderItem struct {
		CreatedTime string      `json:"createdTime"`
		OrderItems  []user1.OrderItem `json:"orderItems"`
	}

	type ResponseOrder struct {
		ID         uint    `json:"id"`
		Status     string  `json:"status"`
		TableNoID  uint    `json:"tableNoID"`
		TotalPrice float64 `json:"totalPrice"`

		BalPrice    float64     `json:"balPrice"`
		TableName   string      `json:"tableNoTableName"`
		OrderTime   int         `json:"orderTime"`
		Created     string      `json:"created"`
		BalTime     int         `json:"balTime"`
		UniqueID    int         `json:"uniqueID"`
		BalQuantity int         `json:"balQuantity"`
		CreatedAt   string      `json:"createdAt"` // Include the created date
		OrderItems  []user1.OrderItem `json:"orderItems"`
	}

	var responseOrders []ResponseOrder

	for _, order := range orders {
		var orderItems []user1.OrderItem
		result := connection.Model(&order).Where("order_items.order_item_status IN (?)", []string{"OrderPending", "Preparing"}).Related(&orderItems)
		if result.Error != nil {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
			return
		}

		// Only include orders that have at least one "OrderPending" item
		if len(orderItems) > 0 {
			// Create a map to group order items by date and time
			itemGroups := make(map[string]ResponseOrderItem)
			for _, item := range orderItems {
				// Convert the item's CreatedAt timestamp to a string
				createdAtStr := item.CreatedAt.Format(time.RFC3339)

				// Format the timestamp in the desired format
				createdAtTime, err := time.Parse(time.RFC3339, createdAtStr)
				if err != nil {
					http.Error(w, "Error parsing timestamp", http.StatusInternalServerError)
					return
				}

				createdTime := createdAtTime.Format("2006-01-02T15:04:05.999999-07:00")

				if group, ok := itemGroups[createdTime]; !ok {
					itemGroups[createdTime] = ResponseOrderItem{
						CreatedTime: createdTime,
						OrderItems:  []user1.OrderItem{item},
					}

				} else {
					group.OrderItems = append(group.OrderItems, item)
					itemGroups[createdTime] = group
				}
			}

			// Create a response order for each group of order items
			for createdAt, responseItem := range itemGroups {

				var highestMenuItemTime int    // Variable to store the highest MenuItemTime
				var totalMenuItemPrice float64 // Variable to store the total MenuItemPrice
				var totalMenuItemQuantity int  // Variable to store the total MenuItemQuantity

				// Iterate through order items to calculate the total MenuItemPrice and quantity
				for _, item := range responseItem.OrderItems {
					totalMenuItemPrice += item.MenuItemPrice
					totalMenuItemQuantity += item.Quantity
					if item.MenuItemTime > highestMenuItemTime {
						highestMenuItemTime = item.MenuItemTime
					}
				}

				responseOrder := ResponseOrder{
					ID:         order.ID,
					Status:     order.Status,
					TableNoID:  order.TableNoID,
					TotalPrice: order.TotalPrice,

					BalPrice:    totalMenuItemPrice,
					TableName:   order.TableNoTableName,
					Created:     createdAt,
					BalTime:     highestMenuItemTime,
					BalQuantity: totalMenuItemQuantity,
					UniqueID:    responseItem.OrderItems[0].UniqueID,

					CreatedAt:  responseItem.OrderItems[0].CreatedAt.Format("2006-01-02T15:04:05.999999-07:00"),
					OrderItems: responseItem.OrderItems,
				}

				responseOrders = append(responseOrders, responseOrder)
			}
		}
	}
	sort.Slice(responseOrders, func(i, j int) bool {
		timeI, _ := time.Parse(time.RFC3339, responseOrders[i].CreatedAt)
		timeJ, _ := time.Parse(time.RFC3339, responseOrders[j].CreatedAt)
		return timeI.Before(timeJ)
	})

	// Send the response with distinct orders and their filtered and grouped order items
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseOrders)
}

func Orderlistss(w http.ResponseWriter, r *http.Request) {
	// Parse URL parameters
	params := mux.Vars(r)

	clientIDStr := params["clientID"]

	// Parse the client ID from the URL parameter
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		http.Error(w, "Invalid clientID", http.StatusBadRequest)
		return
	}
	orderIDStr := params["orderID"]
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		http.Error(w, "Invalid orderID", http.StatusBadRequest)
		return
	}

	// Get distinct orders with the specified client ID and status "OrderPending"
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	var orders []user1.Order
	result := connection.Joins("JOIN order_items ON orders.id = order_items.order_id").
		Where("orders.client_id = ? AND order_items.order_id = ?", clientID, orderID).
		Select("DISTINCT orders.*").
		Find(&orders)
	if result.Error != nil {
		http.Error(w, "No orders with OrderPending status found", http.StatusNotFound)
		return
	}

	type ResponseOrder struct {
		ID          uint        `json:"id"`
		Status      string      `json:"status"`
		TableNoID   uint        `json:"tableNoID"`
		TotalPrice  float64     `json:"totalPrice"`
		BalPrice    float64     `json:"balPrice"`
		TableName   string      `json:"tableNoTableName"`
		OrderTime   int         `json:"orderTime"`
		Created     string      `json:"created"`
		BalTime     int         `json:"balTime"`
		UniqueID    int         `json:"uniqueID"`
		BalQuantity int         `json:"balQuantity"`
		CreatedAt   string      `json:"createdAt"`
		OrderItems  []user1.OrderItem `json:"orderItems"`
	}

	var responseOrders []ResponseOrder

	for _, order := range orders {
		var orderItems []user1.OrderItem
		result := connection.Model(&order).Where("order_items.order_item_status IN (?)", []string{"OrderPending", "Prepare"}).Related(&orderItems)
		if result.Error != nil {
			http.Error(w, result.Error.Error(), http.StatusInternalServerError)
			return
		}

		// Create a map to group order items by uniqueID
		itemGroups := make(map[int][]user1.OrderItem)

		for _, item := range orderItems {
			uniqueID := item.UniqueID

			if group, ok := itemGroups[uniqueID]; !ok {
				itemGroups[uniqueID] = []user1.OrderItem{item}
			} else {
				group = append(group, item)
				itemGroups[uniqueID] = group
			}
		}

		// Iterate through the map to create response items
		for uniqueID, orderItemList := range itemGroups {
			// Calculate metrics for this group of order items
			var highestMenuItemTime int
			var totalMenuItemPrice float64
			var totalMenuItemQuantity int

			for _, item := range orderItemList {
				totalMenuItemPrice += item.MenuItemPrice
				totalMenuItemQuantity += item.Quantity
				if item.MenuItemTime > highestMenuItemTime {
					highestMenuItemTime = item.MenuItemTime
				}
			}

			responseOrder := ResponseOrder{
				ID:         order.ID,
				Status:     order.Status,
				TableNoID:  order.TableNoID,
				TotalPrice: order.TotalPrice,
				BalPrice:   totalMenuItemPrice,
				TableName:  order.TableNoTableName,
				BalTime:    highestMenuItemTime,
				UniqueID:   uniqueID,
				CreatedAt:  orderItemList[0].CreatedAt.Format("2006-01-02T15:04:05.999999-07:00"),
				OrderItems: orderItemList,
			}

			responseOrders = append(responseOrders, responseOrder)
		}
	}

	// Send the response with distinct orders and their filtered and grouped order items
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(responseOrders)
}



func UpdateOrders(w http.ResponseWriter, r *http.Request) {
	// Parse URL parameters
	params := mux.Vars(r)

	orderIDStr := params["orderID"]

	// Parse the order ID from the URL parameter
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		http.Error(w, "Invalid orderID", http.StatusBadRequest)
		return
	}

	// Parse JSON request for the updated order details
	var updatedOrder user1.Order
	err = json.NewDecoder(r.Body).Decode(&updatedOrder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the existing order from the database
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	var existingOrder user1.Order
	result := connection.First(&existingOrder, orderID)
	if result.Error != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	if updatedOrder.OrderStatus != "" {
		existingOrder.OrderStatus = updatedOrder.OrderStatus
	}
	if updatedOrder.Status != "" {
		existingOrder.Status = updatedOrder.Status
	}

	// Update the order with the new details // Replace with actual field updates

	// Save the updated order back to the database
	result = connection.Save(&existingOrder)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Return the updated order as a JSON response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(existingOrder)
}

