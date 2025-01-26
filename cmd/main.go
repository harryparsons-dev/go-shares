package main

import (
	"github.com/harryparsons-dev/go-shares/handlers"
	"github.com/harryparsons-dev/go-shares/models"
	"github.com/labstack/echo/v4"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func main() {
	app := echo.New()
	initDatabase()

	//HANDLERS
	exportHandler := handlers.NewExportHandler(db)

	// Pages
	app.GET("/", exportHandler.List)
	app.POST("/exports", exportHandler.Create)
	//Endpoints
	app.GET("/upload", exportHandler.ShowUploadPage)
	app.GET("/exports", exportHandler.Get)
	app.Start(":3000")

}

func initDatabase() {
	var err error
	// Initialize database connection
	db, err = gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Auto-migrate the User models
	err = db.AutoMigrate(models.Exports{})
	if err != nil {
		panic("failed to migrate database")
	}
}
