package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"github.com/robfig/cron/v3"
	// "encoding/json"
	// "sathya-narayanan23/crudapp/database"
	"sathya-narayanan23/crudapp/middleware"
	// "sathya-narayanan23/crudapp/users"
	"sathya-narayanan23/crudapp/users/user"
	"sathya-narayanan23/handler/client"
	"sathya-narayanan23/handler/resources"
	"sathya-narayanan23/handler/categories"
	"sathya-narayanan23/handler/menuitem"
	"sathya-narayanan23/handler/image"
	"sathya-narayanan23/handler/table"
	"sathya-narayanan23/handler/order"
	"sathya-narayanan23/handler/banner"
	
	"sathya-narayanan23/crudapp/database"
	// "sathya-narayanan23/crudapp/users/user"
    
	"sathya-narayanan23/handler/mail"
	
	// "sathya-narayanan23/handler/user"
	"sathya-narayanan23/handler/admin"
	"github.com/gorilla/mux"
	// "github.com/gorilla/handlers"
)

var (
	router    *mux.Router
	
	secretkey string = "secretkeyjwt"
)

func codemainss() {
    // Start a goroutine to run the function every second
    go func() {
        ticker := time.Tick(1 * time.Second)
        for range ticker {
            GetAllClientplannewww()
        }
    }()

    // This select statement blocks the main goroutine
    // Keeping the program running indefinitely
    select {}
}

func urlmain(){
	router = mux.NewRouter()
	// GetAllClientplannewww()
	
	
	// router := http.NewServe/Mux()

	router.HandleFunc("/register",  admin.RegisterHandler).Methods("POST")
	router.HandleFunc("/login", admin.LoginHandler).Methods("POST")
	router.HandleFunc("/hi",  admin.ClientPageHandler).Methods("GET")
	
	router.HandleFunc("/clientadmin",  admin.ClientAdminPageHandler).Methods("GET")
	router.HandleFunc("/clientsuperadmin",  admin.ClientSuperAdminPageHandler).Methods("GET")
	

    //CLIENT
	router.HandleFunc("/signup", client.SignUp).Methods("POST")
	router.HandleFunc("/clients", client.GetAllClient).Methods("GET")
	router.HandleFunc("/client", client.GetAllClientsWithCategories).Methods("GET")
	router.HandleFunc("/clientcountlist", client.GetAllClientplannew).Methods("GET")
	

	router.HandleFunc("/clientcountlistss", client.GetAllClientplannewwww).Methods("GET")
	router.HandleFunc("/planandclient", client.GetAllClientplan).Methods("GET")
	router.HandleFunc("/clients/{clientID}", client.GetClient).Methods("GET")
	router.HandleFunc("/clients/{id}",  client.UpdateClient).Methods("PUT")
	router.HandleFunc("/clients/{id}",  client.DeleteClient).Methods("DELETE")

    
	router.HandleFunc("/clients/{id}/chef",  resources.GetChefsForClient).Methods("GET") 
	router.HandleFunc("/clients/{id}/cashier",  resources.GetCashiersForClient).Methods("GET")
	router.HandleFunc("/resources/signup",  resources.CreateResource).Methods("POST")
	router.HandleFunc("/resources/login",  resources.ResourceLogin).Methods("POST")
    
	router.HandleFunc("/resources",  resources.GetAllResources).Methods("GET")
	router.HandleFunc("/resource/{id}",  resources.GetResource).Methods("GET")
	router.HandleFunc("/resources/{client_id}",  resources.GetResourcesForClient).Methods("GET")
	router.HandleFunc("/resource/{id}",  resources.DeleteResource).Methods("DELETE")
	router.HandleFunc("/resource/{id}",  resources.UpdateResource).Methods("PUT")
	router.HandleFunc("/resourcess/{resource_id}",  resources.GetResourceByID).Methods("GET")

	
	router.HandleFunc("/categories/{clientId}",  categories.CreateCategory).Methods("POST")
	router.HandleFunc("/categories",  categories.GetAllCategories).Methods("GET")
	router.HandleFunc("/categories/{id}",  categories.DeleteCategory).Methods("DELETE")
	router.HandleFunc("/categories/{id}",  categories.UpdateCategoryName).Methods("PUT")
	router.HandleFunc("/client/{clientId:[0-9]+}/categories/update/status",  categories.UpdateCategoriesStatusForClient).Methods("PUT")
	router.HandleFunc("/client/{clientId:[0-9]+}/categories/update",  categories.UpdateCategoriesStatusForClientactive).Methods("PUT")
	router.HandleFunc("/categories/id/{id}",  categories.UpdateCategorys).Methods("PUT")
	router.HandleFunc("/categories/clientID/{clientId}",  categories.GetCategoriesForClientId).Methods("GET")
	router.HandleFunc("/categories/{clientId}",  categories.GetActiveCategoriesForClientId).Methods("GET")
	router.HandleFunc("/categories/status/{clientId}",  categories.GetActiveCategoriesForClientIdStatus).Methods("GET")
	router.HandleFunc("/categories/{clientName}",  categories.GetCategoriesForClientName).Methods("GET")


	router.HandleFunc("/menuitems/banner/{clientId}",  menuitem.CreateMenuItemBanner).Methods("POST")
	router.HandleFunc("/menuitems/{clientId}",  menuitem.CreateMenuItem).Methods("POST")
	router.HandleFunc("/menuitems",  menuitem.GetAllMenuItem).Methods("GET")
	router.HandleFunc("/menuitems/id/{id}",  menuitem.GetMenuItem).Methods("GET")
	router.HandleFunc("/menuitems/{id}",  menuitem.DeleteMenuItem).Methods("DELETE")
	router.HandleFunc("/getMenuItems/clientId/{clientId}",  menuitem.GetMenuItemsearch).Methods("GET")
	router.HandleFunc("/getMenuItems/clientId/{clientId}/id/{categoryID}", menuitem.GetMenuItemsearchnew).Methods("GET")
	router.HandleFunc("/clients/{clientId}/menuitems/update/{menuItemId}",  menuitem.UpdateMenuItemBanner).Methods("PUT")
	
	router.HandleFunc("/clients/{clientId}/menuitems/{menuItemId}",  menuitem.UpdateMenuItem).Methods("PUT")
	router.HandleFunc("/clients/{clientId}/menuitems/{categoryid}",  menuitem.GetMenuItemsByCategory).Methods("GET")

	
	router.HandleFunc("/clients/getBanner/{clientID}",  menuitem.GetMenuItemsBybanner).Methods("GET")
	router.HandleFunc("/clients/getBannersAndMenuItems/{clientID}",  menuitem.GetMenuItemsByCategoryclientactivelike).Methods("GET")
	router.HandleFunc("/clients/{clientId}/menuitems/{categoryid}/true",  menuitem.GetMenuItemsByCategoryclientactive).Methods("GET")
	router.HandleFunc("/clients/{clientId}/menuitems/{categoryid}/veg",  menuitem.GetMenuItemsByCategoryclientveg).Methods("GET")
	router.HandleFunc("/clients/{clientId}/menuitems/{categoryid}/non-veg",  menuitem.GetMenuItemsByCategoryclientnonveg).Methods("GET")
	


	
	// router.HandleFunc("/image/{imagePath}",  ServeImage).Methods("GET")
	// router.HandleFunc("/images/{imagePath}",  ServeImages).Methods("GET")
	
	router.HandleFunc("/banner/{filename}",  serveImage.ServeImagenews).Methods("GET")
	router.HandleFunc("/uploads/logos/{filename}",  serveImage.ServeImagelogos).Methods("GET")
	router.HandleFunc("/image/{filename}",  serveImage.ServeImageimage).Methods("GET")
	router.HandleFunc("/uploads/{filename}",  serveImage.ServeImageuploads).Methods("GET")
	router.HandleFunc("/qrcodes/{filename}",  serveImage.ServeImageqrcode).Methods("GET")
	router.HandleFunc("/pdfFilesends/{filename}",  serveImage.ServeImagepdfFilesends).Methods("GET")
	router.HandleFunc("/{filePath}",  serveImage.ServeImagenew).Methods("GET")


	
	router.HandleFunc("/table",  table.CreateTable).Methods("POST")
	router.HandleFunc("/tables/tableid/{id}",  table.GetTable).Methods("GET")
	router.HandleFunc("/tables/tableid/id/{id}",  table.GetTableNoByID).Methods("GET")
	router.HandleFunc("/table/{clientID:[0-9]+}",  table.GetTablesByClientIDs).Methods("GET")
	router.HandleFunc("/tables/tableid/{id}",  table.DeleteTable).Methods("DELETE")


	
	router.HandleFunc("/order/{orderID}",  order.GetOrderDetail).Methods("GET")
	// Add a new route for handling DELETE requests to delete an order
	router.HandleFunc("/ordersss/{orderID:[0-9]+}",  order.DeleteOrder).Methods("DELETE")
	router.HandleFunc("/ordersss",  order.DeleteAllOrders).Methods("DELETE")
	router.HandleFunc("/client/{clientID}/orders/day/{day}",  order.GetOrdersByDateTime).Methods("GET")
	router.HandleFunc("/client/{clientID}/orders/7day/{day}",  order.GetOrdersLast7Days).Methods("GET")
	router.HandleFunc("/client/{clientID}/orders/14day/{day}",  order.GetOrdersLast14Days).Methods("GET")
	router.HandleFunc("/client/{clientID}/orders/14day",  order.GetOrdersLast14Daysold).Methods("GET")
	router.HandleFunc("/client/{clientID}/monthlist",  order.GetLast12MonthsTotalSales).Methods("GET")
	router.HandleFunc("/client/{clientID}/lastmonthaverage",  order.GetOrdersLast30DaysAndBefore).Methods("GET")	
	router.HandleFunc("/clients/{id}/createat",  order.UpdateOrderCreatedAt).Methods("PUT")
	router.HandleFunc("/client/{clientID}/orders/{year}/{month}",  order.GetOrdersByMonth).Methods("GET")
	router.HandleFunc("/orders/{startDate}/{endDate}/clientID/{clientID}",  order.GetOrdersByDateRanges).Methods("GET")
	router.HandleFunc("/orders/{startDate}/{endDate}/s0/{clientID}",  order.GetOrdersByDateRangeso).Methods("GET")
	//new
	router.HandleFunc("/table/{clientID:[0-9]+}",  order.GetTablesByClientIDs).Methods("GET")//ok
	//oldr  ).Methods("GET")
	router.HandleFunc("/ordersclient/{clientID:[0-9]+}",  order.GetOrdersByClientId).Methods("GET")
	router.HandleFunc("/update-status",  order.UpdateOrderPayStatusToPayed).Methods("POST")



	
	router.HandleFunc("/clients/{clientID}/order/{orderID}",  order.GetOrderStatusForClientAndOrderID).Methods("GET")
	router.HandleFunc("/clients/{clientID}/order/{orderID}",  order.UpdateOrderStatus).Methods("PUT", "PATCH")
	router.HandleFunc("/clients/{clientID}/orders/order-pending", order.GetPendingOrdersForClientnew).Methods("GET")
	router.HandleFunc("/clients/{clientID}/orderstatusold",  order.GetPendingOrdersForClient123).Methods("GET") //pendingorder
	router.HandleFunc("/clients/{clientID}/orderss",  order.GetPendingOrdersForClient456).Methods("GET") //pendingorder
	router.HandleFunc("/clients/{clientID}/orderssss",  order.GetPendingOrdersForClient).Methods("GET")  //pendingorder
	router.HandleFunc("/orders/update/{orderID}",  order.UpdateOrder).Methods("PUT")
//time
	router.HandleFunc("/clients/{clientID}/{orderID}/statustime",  order.GetHighestMenuItemTimeForUniqueID).Methods("GET")
	
	router.HandleFunc("/clients/{clientID}/statustime",  order.GetOrdersWithOrderItemStatusOrderPendingssnewtimenows).Methods("GET")
	router.HandleFunc("/clients/{clientID}/status",  order.GetOrdersWithOrderItemStatusOrderPendingssnew).Methods("GET") //pendingorder
	router.HandleFunc("/clients/{clientID}/orderstatus",  order.GetOrdersWithOrderItemStatusOrderPendingss).Methods("GET") //pendingorder
	router.HandleFunc("/clients/{clientID}/orderlist/{orderID}",  order.Orderlistss).Methods("GET")
	router.HandleFunc("/order/update/{orderID}",  order.UpdateOrders).Methods("PUT")
	router.HandleFunc("/table/{tableID}/highestorder",  order.GetPaymentPendingOrdersForTables).Methods("GET")
	//1 order
	router.HandleFunc("/table/{tableID}/addcart/{clientID}",  order.AddToCartIDtry).Methods("POST")
	//2  order extra add
	router.HandleFunc("/add-order-items/{orderID}",  order.AddOrderItemsByID).Methods("POST")
	router.HandleFunc("/orders/paymetupdate/{orderID}",  order.UpdateOrderStatusAndPayStatus).Methods("PUT")
	router.HandleFunc("/orders/{orderID}/updates/{uniqueID}",  order.UpdateOrderItemsStatusToPreparingnew).Methods("PUT")
	router.HandleFunc("/orders/updateOrderItemStatus/{orderID}/{clientID}/{uniqueID}",  order.UpdateOrderItemsStatusnews).Methods("PUT")
	router.HandleFunc("/orders/{orderID}/updatess",  order.UpdateOrderItemsStatusToPendings).Methods("PUT")
	router.HandleFunc("/orders/{orderID}/update",  order.UpdateOrderItemsStatusToPendingsnewsssorg).Methods("PUT")
	// Assuming you are using the Gorilla Mux router
	// admin
	router.HandleFunc("/client/orderstatuS/{clientID}",  order.GetOrdersByClientIDAndStatus).Methods("GET")
	router.HandleFunc("/client/paymentlist/{clientID}",  order.GetOrdersByClientIDAndStatusnew).Methods("GET")



	router.HandleFunc("/createBanner/{clientId}",  banner.CreateBanner).Methods("POST")	
	router.HandleFunc("/CreateOrUpdateBanner/{clientId}",  banner.CreateOrUpdateBanner).Methods("POST")
	router.HandleFunc("/banners/{clientID:[0-9]+}",  banner.GetBannersByClientID).Methods("GET")
	router.HandleFunc("/updateSpecialBanner/{clientID:[0-9]+}",  banner.UpdateSpecialBanner).Methods("POST")
	router.HandleFunc("/updateSpecialBanners/{clientID:[0-9]+}",  banner.UpdateOrCreateBanner).Methods("POST")
	router.HandleFunc("/highestBanner/{clientID:[0-9]+}",  banner.GetHighestBannerForClientID).Methods("GET")
	router.HandleFunc("/getBannersAndMenuItems/{clientID}",  banner.GetBannersAndMenuItems).Methods("GET")
	router.HandleFunc("/banner/{bannerID}",  banner.DeleteBanner).Methods("DELETE")
	


	// router.HandleFunc("/clientsneww",client.GetAllClientplannewSS).Methods("GET")


	// // Attach middleware
	loggedRouter := middleware.LogRequest(router)

	// Start the server
	serverAddress := "192.168.1.2:8080" // Change to your server's IP address

	fmt.Printf("Server listening on port %s... ", serverAddress)
	log.Fatal(http.ListenAndServe(serverAddress, loggedRouter))
	// log.Fatal(http.ListenAndServe(":8080", loggedRouter))


	router.Methods("OPTIONS").HandlerFunc(handleOptions)

	
}
func handleOptions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, ...") // Add necessary headers
}

// func main() {
// 	// Initialize database connection
// 	// database.GetDatabase()
// 	// defer database.CloseDatabase(connection)
// 	// Initialize middleware
//     user1.InitialMigration() 
// 	urlmain()
// 	go func() {
//         ticker := time.Tick(1 * time.Minute)
// 		 // Adjust the time duration here
//         for range ticker {
//             GetAllClientplannewww()
//         }
//     }()
// 	// urlmain()
	
// 	select {}
// 	// urlmain()
// }
 func timerun(){
	
    // Create a new cron scheduler
    c := cron.New()

    // Add the job to run every day at 7 am
    _, err := c.AddFunc("45 18 * * *", func() {
		fmt.Println("code run")
        client.GetAllClientplannewww()
    })
    if err != nil {
        // Handle error
        panic(err)
    }

    // Start the scheduler
    c.Start()
	
    defer c.Stop()
 }

func main () {
	connection := database.GetDatabase()
    defer database.CloseDatabase(connection)

    // Initialize middleware
    user1.InitialMigration()
	
	// GetAllClientplannewww()

    // Create a new cron scheduler
    c := cron.New()

    // Add the job to run every day at 7 am
    _, err := c.AddFunc("29 10 * * *", func() {
        client.GetAllClientplannewww()
    })
	fmt.Println(err)
    // if err != nil {
    //     // Handle error
    //     panic(err)
    // }
    c.Start()
	
    defer c.Stop()
	
	urlmain()
	// timerun()


	
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
	 fmt.Println("mail send")

}
