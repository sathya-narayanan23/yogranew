package user1

import (
    "time"
    "gorm.io/gorm"
    "sathya-narayanan23/crudapp/database"
)


type Admin struct {
	gorm.Model
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type Client struct {
    gorm.Model
    Name                  string    `json:"name"`
    DisplayName           string    `json:"displayName"`
    MobileNumber          string    `json:"mobileNumber"`
    SecondaryMobileNumber string    `json:"secondaryMobileNumber"`
    Email                 string    `json:"email"`
    SecondaryEmail        string    `json:"secondaryEmail"`
    Country               string    `json:"country"`
    State                 string    `json:"state"`
    District              string    `json:"district"`
    Logo                  string    `json:"logo"`
    Logoimagepath         string    `json:"logoimagepath"`
    Specialist            string    `json:"specialist"`
    PrimaryColour         string    `json:"primaryColour"`
    Password              string    `json:"password"`
    Plan                  string    `json:"plan"`
    Upi                   string    `json:"upi"`
    UpiName               string    `json:"upiName"`
    FilePath              string    `json:"filePath"`
    PaymentMethod         string    `json:"paymentMethod"`
    Status                bool      `json:"status"`
    Currency              string    `json:"currency"`
    PlanCreateTime        time.Time `json:"planCreateTime"` // Add PlanCreateTime field
    PlanUpdateTime        time.Time `json:"planUpdateTime"`
    NotificationDate      time.Time `json:"notificationDate"`
    PlanExpiration        time.Time `json:"plan_expiration"`
    SelectedCategories    []string  `json:"selectedCategories" gorm:"-"`
    Categories            []Category `json:"categories" gorm:"foreignKey:ClientID"`
}

type Category struct {
	gorm.Model
	ClientID       uint       `json:"clientId"`
	ImagePath      string     `json:"imagePath"`
	ImagefilePath      string     `json:"imagefilePath"`
	CategoryName   string     `json:"categoryName"`
	Status         bool     `json:"status"`
	Active_status         bool     `json:"active_status"`
	Image          string     `json:"image"`
	MenuItems      []MenuItem `json:"menuItems" gorm:"foreignKey:CategoryID"`
	TotalMenuItems int        `json:"totalMenuItems" gorm:"-"`
}

type Banner struct {
	gorm.Model
	ClientID       uint       `json:"clientId"`
	
	Name   string     `json:"name"`
	Special_banner          string     `json:"special_banner"`
	
	Bannerpath          string     `json:"bannerpath"`
	MenuItems      []MenuItem `json:"menuItems" gorm:"foreignKey:ClientID "`
	TotalMenuItems int        `json:"totalMenuItems" gorm:"-"`
}

type MenuItem struct {
	gorm.Model
	ClientID     uint    `json:"clientId"`
	Image        string  `json:"image"`
	Imagepath       string  `json:"imagepath"`
	Name     string  `json:"name"`
	ItemName     string  `json:"itemName"`
	Currency     string  `json:"currency"`
    Food_type     string  `json:"food_type"`
	Time         int     `json:"time"`
	
	Offer         float64     `json:"offer"`
	OfferRate        float64     `json:"offerRate"`
	Status       bool    `json:"status"`
	Banner       bool    `json:"banner"`
	Recommendation  bool    `json:"recommendation"`
	Temporary_status       bool    `json:"temporary_status"`
	Price        float64 `json:"price"`
	Sub_category string  `json:"sub_category"`
	Description string  `json:"description"`
	CategoryID   uint    `json:"categoryID"`
	CategoryName string  `json:"categoryName"`
	Quantity int
}

type Resource struct {
	gorm.Model
	ClientID     uint   `json:"clientId"`
	Name         string `json:"name"`
	Role         string `json:"role"`
	MobileNumber string `json:"mobileNumber"`
	Passwords    string `json:"passwords"`
	Password     string `json:"password"`
}

type Order struct {
	gorm.Model
	Number           int
	ClientID         uint      `json:"clientID"`
	TableNoID        uint      `json:"tableNoID"`
	TableNoTableName string    `json:"tableNoTableName"`
	TotalPrice       float64   `json:"totalPrice"`
	OrderStatus      string    `json:"orderStatus"`
	UpdateDate       time.Time `json:"updateDate"`
	OldUpdateDate    time.Time `json:"oldUpdateDate"`
	OldUpdateTime    int
	UpdateTime       int 
	Balprice         float64 `json:"balprice"`
	NowUpdateDate time.Time `json:"nowUpdateDate"`
	// OldUpdateDate   time.Time  `json:"oldUpdateDate"`
	TotalQuantity int
	OrderTime     int
	// OrderItems    []OrderItem
	Status    string
	TableName string `json:"tableName"` // Add this field
	
	OrderItems []OrderItem `json:"orderItems" gorm:"foreignKey:OrderID"`
	
}


type TableNo struct {
	gorm.Model
	TableName      string   `json:"tableName"`
	ClientID       uint     `json:"clientID"`
	QRCodeFilePath string   `json:"qrcodeFilePath"`
	
	QRCode string   `json:"qrcode"`
	// QRCode         []QRCode `json:"QRCode" gorm:"foreignKey:ClientID"`
}

type OrderItem struct {
	gorm.Model
	ClientID uint
	// OrderID uint `gorm:"foreignkey:ID"`
	UniqueID         int `json:"uniqueID"`
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
	Successdate      time.Time `json:"successdate"`
	UpdateTime       int
	OldUpdateDate    time.Time `json:"oldUpdateDate"`

	NowUpdateDate time.Time `json:"nowUpdateDate"`
	OldUpdateTime int
	OrderTime     int
	Time          int
}

func InitialMigration() {
	connection := database.GetDatabase()
	defer database.CloseDatabase(connection)
	
	connection.AutoMigrate(&Client{})
	connection.AutoMigrate(&Resource{})
	connection.AutoMigrate(&Category{})
	connection.AutoMigrate(&MenuItem{})
	connection.AutoMigrate(&TableNo{})
	connection.AutoMigrate(&Order{})
	connection.AutoMigrate(&OrderItem{})

	// connection.AutoMigrate(&QRCode{})
	
	connection.AutoMigrate(&Admin{})
	connection.AutoMigrate(&Banner{})

}


