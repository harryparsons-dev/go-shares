package main

import (
	"github.com/labstack/echo/v4"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

type Exports struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	FilePath string `json:"filePath"`
}

func initDatabase() {
	var err error
	// Initialize database connection
	db, err = gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Auto-migrate the User model
	err = db.AutoMigrate(&Exports{})
	if err != nil {
		panic("failed to migrate database")
	}
}

func main() {
	// Initialize database
	initDatabase()

	// Create an Echo instance
	e := echo.New()

	// Define routes

	e.POST("/export", CreateExport)

	// Start the server
	e.Logger.Fatal(e.Start(":8080"))
}

func CreateExport(c echo.Context) error {

}
