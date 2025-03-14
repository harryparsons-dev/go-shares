package main

import (
	"github.com/harryparsons-dev/go-shares/handlers"
	"github.com/harryparsons-dev/go-shares/models"
	"github.com/joho/godotenv"
	"github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"os"
)

var db *gorm.DB

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	app := echo.New()
	initDatabase()

	//HANDLERS
	exportHandler := handlers.NewExportHandler(db)
	userHandler := handlers.NewUserHandler(db)
	//MIDDLEWHERE
	app.Use(middleware.Logger())
	app.Use(middleware.Recover())
	app.Use(middleware.CORS())

	app.GET("/exports", exportHandler.List, echojwt.JWT([]byte(os.Getenv("JWT_SECRET"))))
	app.POST("/exports", exportHandler.Create, echojwt.JWT([]byte(os.Getenv("JWT_SECRET"))))
	app.GET("/exports/download/:id", exportHandler.Get, echojwt.JWT([]byte(os.Getenv("JWT_SECRET"))))
	app.GET("/exports/pie/download/:id", exportHandler.GetPie, echojwt.JWT([]byte(os.Getenv("JWT_SECRET"))))

	app.POST("/login", userHandler.Login)

	app.Start(":3000")

}

func initDatabase() {
	var err error

	db, err = gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	err = db.AutoMigrate(models.Exports{}, models.User{})
	if err != nil {
		panic("failed to migrate database")
	}
}
