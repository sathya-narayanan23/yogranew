package database

import (
    // "database/sql"
    "log"
	// "strconv"

    _ "github.com/lib/pq"
	// "gopkg.in/gomail.v2"
	"github.com/jinzhu/gorm"
	"fmt"
)

// Database connection string
// const (
//     host     = "localhost"
//     port     = 5432
//     user     = "postgres"
//     password = "sathya"
//     dbname   = "postgres"
// )

// var db *sql.DB

// // InitDB initializes the database connection
// func InitDB() {
//     var err error
//     dbinfo := "host=" + host + " port=" + strconv.Itoa(port) + " user=" + user +
//         " password=" + password + " dbname=" + dbname + " sslmode=disable"
//     db, err = sql.Open("postgres", dbinfo)
//     if err != nil {
//         log.Fatal(err)
//     }

//     // Check if the connection is successful
//     err = db.Ping()
//     if err != nil {
//         log.Fatal(err)
//     }

//     log.Println("Database connected!")
// }

// GetDB returns a connection to the database
// func GetDB() *sql.DB {
//     return db
// }
func GetDatabase() *gorm.DB {

	databasename := "yogra"
	database := "postgres"
	databasepassword := "sathya"
	databaseurl := "postgres://postgres:" + databasepassword + "@localhost/" + databasename + "?sslmode=disable"

	connection, err := gorm.Open(database, databaseurl)
	if err != nil {
		log.Fatalln(err)
	}
	sqldb := connection.DB()

	err = sqldb.Ping()
	if err != nil {
		log.Fatal("Database connec/ted")
	}
	fmt.Println("Database connection successful.")
	return connection
}


func CloseDatabase(connection *gorm.DB) {
	sqldb := connection.DB()
	sqldb.Close()
}
