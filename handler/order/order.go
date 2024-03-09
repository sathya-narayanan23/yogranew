package order

import (
	"encoding/json"
	"fmt"
	// "errors"
	"bytes"
    
	// "encoding/json"
	"time"
	"io"
	"net/http"
	
	"github.com/jung-kurt/gofpdf"
	
	"sathya-narayanan23/crudapp/database"
	"sathya-narayanan23/crudapp/users/user"

    "strconv"
    
	"github.com/gorilla/mux"
)

type OrderResponses struct {
	Orders         []user1.Order `json:"orders"`
	TodayTotalSale float64 `json:"todayTotalSale"`
	// If you want to include TodayTotalSaless, add it here.
}
	

func generatePDF(orderResponses OrderResponses, pdfWriter io.Writer) *gofpdf.Fpdf {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Add content to the PDF
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "Orders Report")

	pdf.Ln(10)

	// Add order details to the PDF
	for _, order := range orderResponses.Orders {
		pdf.Cell(40, 10, fmt.Sprintf("Order ID: %d", order.ID))
		pdf.Ln(5)

		// Add order items to the PDF
		for _, item := range order.OrderItems {
			pdf.Cell(40, 10, fmt.Sprintf("Item: %s, Quantity: %d, Price: %.2f", item.MenuItemItemName, item.Quantity, item.MenuItemPrice))
			pdf.Ln(5)
		}

		pdf.Cell(40, 10, fmt.Sprintf("Total Price: %.2f", order.TotalPrice))
		pdf.Ln(10)
	}
	

	pdf.Cell(40, 10, fmt.Sprintf("Total Sale: %.2f", orderResponses.TodayTotalSale))

	// Write PDF content to the provided writer
	pdf.Output(pdfWriter)

	return pdf
}


func GetOrderDetail(w http.ResponseWriter, r *http.Request) {
	// Parse URL parameters
	params := mux.Vars(r)
	orderIDStr := params["orderID"]
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		http.Error(w, "Invalid orderID", http.StatusBadRequest)
		return
	}

	// Get the order details from the database
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	var order user1.Order
	result := connection.Where("id = ?", orderID).First(&order)
	if result.Error != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	// Get the associated order items
	var orderItems []user1.OrderItem
	result = connection.Where("order_id = ?", order.ID).Find(&orderItems)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate the total BalPrice
	var totalBalPrice float64
	for _, item := range orderItems {
		totalBalPrice += item.MenuItemPrice
	}

	// Update the BalPrice in the order struct
	order.Balprice = totalBalPrice

	// Prepare the response with order details and items
	orderDetails := struct {
		Order      user1.Order       `json:"order"`
		OrderItems []user1.OrderItem `json:"orderItems"`
	}{
		Order:      order,
		OrderItems: orderItems,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orderDetails)
}


func DeleteOrder(w http.ResponseWriter, r *http.Request) {
	// Parse URL parameters
	params := mux.Vars(r)
	orderIDStr := params["orderID"]
	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		http.Error(w, "Invalid orderID", http.StatusBadRequest)
		return
	}

	// Get the order details from the database
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	var order user1.Order
	result := connection.Where("id = ?", orderID).First(&order)
	if result.Error != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	// Delete the order
	result = connection.Delete(&order)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with success message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "Order deleted successfully"}`)
}


func DeleteAllOrders(w http.ResponseWriter, r *http.Request) {
	// Get the database connection
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	// Delete all orders
	result := connection.Delete(&user1.Order{})
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with success message
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"message": "All orders deleted successfully"}`)
}


func GetOrdersByDateTime(w http.ResponseWriter, r *http.Request) {
    connection := database.GetDatabase()
    defer database.CloseDatabase(connection)

    // Get the "day" and "clientID" parameters from the URL
    vars := mux.Vars(r)
    day := vars["day"]
    clientID := vars["clientID"]

    var orders []user1.Order
    result := connection.Where("DATE(created_at) = ? AND client_id = ?", day, clientID).Order("id asc").Find(&orders)
    if result.Error != nil {
        http.Error(w, result.Error.Error(), http.StatusInternalServerError)
        return
    }

    // Fetch and include order items for each order
    for i := range orders {
        var orderItems []user1.OrderItem
        connection.Where("order_id = ?", orders[i].ID).Find(&orderItems)
        orders[i].OrderItems = orderItems
    }

    // Calculate the total sale for today based on the sum of TotalPrice of all orders
    var todayTotalSale float64
    for _, order := range orders {
        todayTotalSale += order.TotalPrice
    }

    // Create an instance of OrderResponses to include in the response
    orderResponses := OrderResponses{
        Orders:         orders,         // Include the list of orders
        TodayTotalSale: todayTotalSale, // Include the total sale for today
        // You can include other fields here if needed.
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(orderResponses)
}


func GetOrdersLast7Days(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	// Get the "clientID" parameter from the URL
	vars := mux.Vars(r)
	clientID := vars["clientID"]

	// Calculate the date 7 days ago from today
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)

	var orders []user1.Order
	result := connection.Where("created_at BETWEEN ? AND ? AND client_id = ?", sevenDaysAgo, time.Now(), clientID).Order("created_at desc").Find(&orders)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch and include order items for each order
	for i := range orders {
		var orderItems []user1.OrderItem
		connection.Where("order_id = ?", orders[i].ID).Find(&orderItems)
		orders[i].OrderItems = orderItems
	}

	// Calculate the total sale for the last 7 days based on the sum of TotalPrice of all orders
	var last7DaysTotalSale float64
	for _, order := range orders {
		last7DaysTotalSale += order.TotalPrice
	}
	type OrderResponses struct {
		
		Last7DaysTotalSale float64 `json:"last7DaysTotalSale"`
		Orders             []user1.Order `json:"orders"`
	
	}

	// Create an instance of OrderResponses to include in the response
	orderResponses := OrderResponses{
		
		Last7DaysTotalSale: last7DaysTotalSale,
		Orders:             orders,                
		}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orderResponses)
}


func GetOrdersLast14Days(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	// Get the "clientID" parameter from the URL
	vars := mux.Vars(r)
	clientID := vars["clientID"]

	// Calculate the date 14 days ago from today
	fourteenDaysAgo := time.Now().AddDate(0, 0, -14)
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)

	// Fetch orders for the last 7 days
	var ordersLast7Days []user1.Order
	result := connection.Where("created_at BETWEEN ? AND ? AND client_id = ?", sevenDaysAgo, time.Now(), clientID).Order("created_at desc").Find(&ordersLast7Days)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch and include order items for each order in the last 7 days
	for i := range ordersLast7Days {
		var orderItems []user1.OrderItem
		connection.Where("order_id = ?", ordersLast7Days[i].ID).Find(&orderItems)
		ordersLast7Days[i].OrderItems = orderItems
	}

	// Calculate the total sale for the last 7 days
	var last7DaysTotalSale float64
	for _, order := range ordersLast7Days {
		last7DaysTotalSale += order.TotalPrice
	}

	// Fetch orders for the week before the last 7 days
	var ordersBeforeLast7Days []user1.Order
	result = connection.Where("created_at BETWEEN ? AND ? AND client_id = ?", fourteenDaysAgo, sevenDaysAgo, clientID).Order("created_at desc").Find(&ordersBeforeLast7Days)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch and include order items for each order before the last 7 days
	for i := range ordersBeforeLast7Days {
		var orderItems []user1.OrderItem
		connection.Where("order_id = ?", ordersBeforeLast7Days[i].ID).Find(&orderItems)
		ordersBeforeLast7Days[i].OrderItems = orderItems
	}

	// Calculate the total sale for the week before the last 7 days
	var beforeLast7DaysTotalSale float64
	for _, order := range ordersBeforeLast7Days {
		beforeLast7DaysTotalSale += order.TotalPrice
	}
	type OrderResponses struct {
		Orders                        []user1.Order `json:"orders"`
		Last7DaysTotalSale            float64 `json:"last7DaysTotalSale"`
		BeforeLast7DaysTotalSale      float64 `json:"beforeLast7DaysTotalSale"`
		AverageTotalSale              float64 `json:"averageTotalSale"`
		PercentageChange              float64 `json:"percentageChange"`
		// Other fields if needed
	}
	

	averageTotalSale := (last7DaysTotalSale + beforeLast7DaysTotalSale) / 2

    // Calculate the percentage change
    var percentageChange float64
    if beforeLast7DaysTotalSale != 0 {
        percentageChange = ((last7DaysTotalSale - beforeLast7DaysTotalSale) / beforeLast7DaysTotalSale) * 100
    }

    // Create an instance of OrderResponses to include in the response
    orderResponses := OrderResponses{
        Orders:                  ordersLast7Days,           // Include the list of orders in the last 7 days
        Last7DaysTotalSale:      last7DaysTotalSale,        // Include the total sale for the last 7 days
        BeforeLast7DaysTotalSale: beforeLast7DaysTotalSale,  // Include the total sale for the week before the last 7 days
        AverageTotalSale:        averageTotalSale,          // Include the average total sale
        PercentageChange:        percentageChange,          // Include the percentage change
        // You can include other fields here if needed.
    }


	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orderResponses)
}


func GetOrdersLast14Daysold(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	// Get the "clientID" parameter from the URL
	vars := mux.Vars(r)
	clientID := vars["clientID"]

	// Calculate the date 14 days ago from today
	fourteenDaysAgo := time.Now().AddDate(0, 0, -14)
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)

	// Fetch orders for the last 7 days
	var ordersLast7Days []user1.Order
	result := connection.Where("created_at BETWEEN ? AND ? AND client_id = ?", sevenDaysAgo, time.Now(), clientID).Order("created_at desc").Find(&ordersLast7Days)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch and include order items for each order in the last 7 days
	for i := range ordersLast7Days {
		var orderItems []user1.OrderItem
		connection.Where("order_id = ?", ordersLast7Days[i].ID).Find(&orderItems)
		ordersLast7Days[i].OrderItems = orderItems
	}

	// Calculate the total sale for the last 7 days
	var last7DaysTotalSale float64
	for _, order := range ordersLast7Days {
		last7DaysTotalSale += order.TotalPrice
	}

	// Fetch orders for the week before the last 7 days
	var ordersBeforeLast7Days []user1.Order
	result = connection.Where("created_at BETWEEN ? AND ? AND client_id = ?", fourteenDaysAgo, sevenDaysAgo, clientID).Order("created_at desc").Find(&ordersBeforeLast7Days)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch and include order items for each order before the last 7 days
	for i := range ordersBeforeLast7Days {
		var orderItems []user1.OrderItem
		connection.Where("order_id = ?", ordersBeforeLast7Days[i].ID).Find(&orderItems)
		ordersBeforeLast7Days[i].OrderItems = orderItems
	}

	// Calculate the total sale for the week before the last 7 days
	var beforeLast7DaysTotalSale float64
	for _, order := range ordersBeforeLast7Days {
		beforeLast7DaysTotalSale += order.TotalPrice
	}
	type OrderResponses struct {
		Orders                        []user1.Order `json:"orders"`
		Last7DaysTotalSale            float64 `json:"last7DaysTotalSale"`
		BeforeLast7DaysTotalSale      float64 `json:"beforeLast7DaysTotalSale"`
		AverageTotalSale              float64 `json:"averageTotalSale"`
		PercentageChange              float64 `json:"percentageChange"`
		// Other fields if needed
	}
	

	averageTotalSale := (last7DaysTotalSale + beforeLast7DaysTotalSale) / 2

    // Calculate the percentage change
    var percentageChange float64
    if beforeLast7DaysTotalSale != 0 {
        percentageChange = ((last7DaysTotalSale - beforeLast7DaysTotalSale) / beforeLast7DaysTotalSale) * 100
    }

    // Create an instance of OrderResponses to include in the response
    orderResponses := OrderResponses{
        // Orders:                  ordersLast7Days,           // Include the list of orders in the last 7 days
        Last7DaysTotalSale:      last7DaysTotalSale,        // Include the total sale for the last 7 days
        BeforeLast7DaysTotalSale: beforeLast7DaysTotalSale,  // Include the total sale for the week before the last 7 days
        AverageTotalSale:        averageTotalSale,          // Include the average total sale
        PercentageChange:        percentageChange,          // Include the percentage change
        // You can include other fields here if needed.
    }


	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orderResponses)
}



func GetLast12MonthsTotalSales(w http.ResponseWriter, r *http.Request) {
    // Get database connection
    connection := database.GetDatabase()
    defer database.CloseDatabase(connection)

    // Get the "clientID" parameter from the URL
    vars := mux.Vars(r)
    clientID := vars["clientID"]

    // Calculate the start date (12 months ago) and end date (today)
    endDate := time.Now()
    startDate := endDate.AddDate(0, -12, 0)

    // Initialize a map to store total sales for each month
    monthlySales := make(map[string]float64)

    // Fetch orders for the last 12 months for the specified client
    var orders []user1.Order
    if err := connection.Where("client_id = ? AND created_at >= ? AND created_at <= ?", clientID, startDate, endDate).Find(&orders).Error; err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Calculate total sales for each month
    for _, order := range orders {
        month := order.CreatedAt.Month()
        year := order.CreatedAt.Year()
        key := fmt.Sprintf("%d-%02d", year, month)
        monthlySales[key] += order.TotalPrice
    }

    // Prepare data for chart assembly
    var chartData []map[string]interface{}

    // Get current year and month
    currentYear, currentMonth, _ := time.Now().Date()

    // Iterate over the last 12 months
    for i := 0; i < 12; i++ {
        // Calculate the year and month for this iteration
        year := currentYear
        month := int(currentMonth) - i
        if month <= 0 {
            month += 12
            year--
        }
        // Create key for the month
        key := fmt.Sprintf("%d-%02d", year, month)
        // Add data to chartData, include month name
        chartData = append(chartData, map[string]interface{}{
            "month":      time.Month(month).String() + " " + strconv.Itoa(year),
            "totalSales": monthlySales[key],
        })
    }

    // Send response
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(chartData); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
}
func GetOrdersLast30DaysAndBefore(w http.ResponseWriter, r *http.Request) {
    connection := database.GetDatabase()
    defer database.CloseDatabase(connection)

    // Get the "clientID" parameter from the URL
    vars := mux.Vars(r)
    clientID := vars["clientID"]

    // Calculate the date 30 days ago from today
    thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
    // Calculate the date 60 days ago from today
    sixtyDaysAgo := time.Now().AddDate(0, 0, -60)

    // Fetch orders for the last 30 days
    var ordersLast30Days []user1.Order
    result := connection.Where("created_at BETWEEN ? AND ? AND client_id = ?", thirtyDaysAgo, time.Now(), clientID).Order("created_at desc").Find(&ordersLast30Days)
    if result.Error != nil {
        http.Error(w, result.Error.Error(), http.StatusInternalServerError)
        return
    }

    // Fetch orders for the period before the last 30 days (30-60 days ago)
    var ordersBeforeLast30Days []user1.Order
    result = connection.Where("created_at BETWEEN ? AND ? AND client_id = ?", sixtyDaysAgo, thirtyDaysAgo, clientID).Order("created_at desc").Find(&ordersBeforeLast30Days)
    if result.Error != nil {
        http.Error(w, result.Error.Error(), http.StatusInternalServerError)
        return
    }

    // Calculate the total sale for the last 30 days
    var last30DaysTotalSale float64
    for _, order := range ordersLast30Days {
        last30DaysTotalSale += order.TotalPrice
    }

    // Calculate the total sale for the period before the last 30 days
    var beforeLast30DaysTotalSale float64
    for _, order := range ordersBeforeLast30Days {
        beforeLast30DaysTotalSale += order.TotalPrice
    }

    // Calculate average sales for the period before the last 30 days
    // averageBefore30DaysTotalSale := beforeLast30DaysTotalSale / 30.0
	var percentageChange float64
    if beforeLast30DaysTotalSale != 0 {
        percentageChange = ((last30DaysTotalSale - beforeLast30DaysTotalSale) / beforeLast30DaysTotalSale) * 100
    }

    // Prepare response
    type OrderResponses struct {
        // Orders                     []Order `json:"orders"`
        Last30DaysTotalSale        float64 `json:"last30DaysTotalSale"`
        BeforeLast30DaysTotalSale  float64 `json:"beforeLast30DaysTotalSale"`
        AverageBefore30DaysTotalSale float64 `json:"percentageChange"`
    }

    // Create an instance of OrderResponses to include in the response
    orderResponses := OrderResponses{
        // Orders:                    ordersLast30Days,            // Include the list of orders in the last 30 days
        Last30DaysTotalSale:       last30DaysTotalSale,        // Include the total sale for the last 30 days
        BeforeLast30DaysTotalSale: beforeLast30DaysTotalSale,  // Include the total sale for the period before the last 30 days
        AverageBefore30DaysTotalSale: percentageChange, // Include the average total sale for the period before the last 30 days
    }

    // Send response
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(orderResponses)
}


func UpdateOrderCreatedAt(w http.ResponseWriter, r *http.Request) {
    // Get database connection
    connection := database.GetDatabase()
    defer database.CloseDatabase(connection)

    // Get the order ID from the URL parameters
    vars := mux.Vars(r)
    orderID := vars["id"]

    // Parse the string into a time.Time value using RFC3339Nano layout
    createdAtStr := r.FormValue("created_at")
    createdAt, err := time.Parse(time.RFC3339Nano, createdAtStr)
    if err != nil {
        // Handle parsing error
        http.Error(w, "Invalid CreatedAt value", http.StatusBadRequest)
        return
    }

    // Fetch the existing order by its ID
    var order user1.Order
    result := connection.Where("id = ?", orderID).First(&order)
    if result.Error != nil {
        http.Error(w, "Order not found", http.StatusNotFound)
        return
    }

    // Update the CreatedAt field of the existing order
    order.CreatedAt = createdAt

    // Save the updated order back to the database
    if err := connection.Save(&order).Error; err != nil {
        // Handle error if failed to save
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Respond with success message
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("CreatedAt updated successfully"))
}




func GetOrdersByMonth(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	// Get the "year," "month," and "clientID" parameters from the URL
	vars := mux.Vars(r)
	year := vars["year"]
	month := vars["month"]
	clientID := vars["clientID"]

	// Convert the year and month parameters to integers
	yearInt, err := strconv.Atoi(year)
	if err != nil {
		http.Error(w, "Invalid year", http.StatusBadRequest)
		return
	}
	monthInt, err := strconv.Atoi(month)
	if err != nil || monthInt < 1 || monthInt > 12 {
		http.Error(w, "Invalid month", http.StatusBadRequest)
		return
	}

	// Calculate the start date for the specified month
	startDate := year + "-" + month + "-01"

	// Calculate the last day of the specified month
	lastDay := -31 // Default to 31 days

	// Check if the month has fewer than 31 days
	if monthInt == 4 || monthInt == 6 || monthInt == 9 || monthInt == 11 {
		lastDay = -30
	} else if monthInt == 2 {
		// Check for February (handle leap years as well)
		if (yearInt%4 == 0 && yearInt%100 != 0) || yearInt%400 == 0 {
			lastDay = -29 // Leap year
		} else {
			lastDay = -28 // Non-leap year
		}
	}

	endDate := year + "-" + month + fmt.Sprintf("%02d", lastDay)

	var orders []user1.Order
	result := connection.Where("created_at >= ? AND created_at <= ? AND client_id = ?", startDate, endDate, clientID).Order("id asc").Find(&orders)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Fetch and include order items for each order
	for i := range orders {
		var orderItems []user1.OrderItem
		connection.Where("order_id = ?", orders[i].ID).Find(&orderItems)
		orders[i].OrderItems = orderItems
	}

	// Calculate the total sale for the month based on the sum of TotalPrice of all orders
	var monthTotalSale float64
	for _, order := range orders {
		monthTotalSale += order.TotalPrice
	}

	// Create an instance of OrderResponses to include in the response
	orderResponses := OrderResponses{
		Orders:         orders,         // Include the list of orders
		TodayTotalSale: monthTotalSale, // Include the total sale for the month
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orderResponses)
}

type OrderPDFResponse struct {
	Orders     []user1.Order
	TotalPrice float64
	PDFFile    []byte
}


func GetOrdersByDateRanges(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	vars := mux.Vars(r)
	startDate := vars["startDate"]
	endDate := vars["endDate"]
	clientID := vars["clientID"]

	var orders []user1.Order
	result := connection.Where("created_at >= ? AND created_at <= ? AND client_id = ?", startDate, endDate, clientID).Order("id asc").Find(&orders)

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	for i := range orders {
		var orderItems []user1.OrderItem
		connection.Where("order_id = ?", orders[i].ID).Find(&orderItems)
		orders[i].OrderItems = orderItems
	}

	var totalPrice float64
	for _, order := range orders {
		totalPrice += order.TotalPrice
	}

	orderResponses := OrderResponses{
		Orders:         orders,
		TodayTotalSale: totalPrice,
	}

	// Generate PDF
	pdfBuffer := new(bytes.Buffer)
	pdf := generatePDF(orderResponses, pdfBuffer)
	fmt.Println(pdf)

	// Set HTTP headers for JSON response
	w.Header().Set("Content-Type", "application/json")

	// Marshal the response to JSON
	responseJSON, err := json.Marshal(OrderPDFResponse{
		Orders:     orderResponses.Orders,
		TotalPrice: orderResponses.TodayTotalSale,
		PDFFile:    pdfBuffer.Bytes(),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write the JSON response to the response writer
	w.WriteHeader(http.StatusOK)
	w.Write(responseJSON)
}



func GetOrdersByDateRangeso(w http.ResponseWriter, r *http.Request) {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	vars := mux.Vars(r)
	startDate := vars["startDate"]
	endDate := vars["endDate"]
	clientID := vars["clientID"]

	var orders []user1.Order
	// result := connection.Where("created_at >= ? AND created_at <= ? AND client_id = ?", startDate, endDate, clientID).Order("created_at desc").Find(&orders)
	result := connection.Where("created_at >= ? AND created_at <= ? AND client_id = ?", startDate, endDate, clientID).Order("id asc").Find(&orders)

	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	for i := range orders {
		var orderItems []user1.OrderItem
		connection.Where("order_id = ?", orders[i].ID).Find(&orderItems)
		orders[i].OrderItems = orderItems
	}

	var totalSale float64
	for _, order := range orders {
		totalSale += order.TotalPrice
	}

	orderResponses := OrderResponses{
		Orders:         orders,
		TodayTotalSale: totalSale,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orderResponses)
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

func GetOrdersByClientId(w http.ResponseWriter, r *http.Request) {
	// Parse URL parameters
	params := mux.Vars(r)
	clientIDStr := params["clientID"]
	clientID, err := strconv.Atoi(clientIDStr)
	if err != nil {
		http.Error(w, "Invalid clientId", http.StatusBadRequest)
		return
	}

	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)

	var orders []user1.Order
	result := connection.Where("client_id = ?", clientID).Find(&orders)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orders)
}


func UpdateOrderPayStatusToPayed(w http.ResponseWriter, r *http.Request) {
	// Parse JSON request
	var updateRequest struct {
		OrderID           uint   `json:"orderID"`
		NewOrderPayStatus string `json:"newOrderPayStatus"`
	}

	err := json.NewDecoder(r.Body).Decode(&updateRequest)
	if err != nil {


		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Extract order ID and new orderPayStatus from the request
	orderID := updateRequest.OrderID
	newOrderPayStatus := updateRequest.NewOrderPayStatus

	if orderID == 0 || newOrderPayStatus == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
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

	// Update the order's orderPayStatus
	order.OrderStatus = newOrderPayStatus
	result = connection.Save(&order)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	// Return a success response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Order pay status updated to '" + newOrderPayStatus + "'."))
}





