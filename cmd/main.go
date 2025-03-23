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
	"net/http"
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

	app.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"https://shares-converter.netlify.app", "http://localhost:5173"}, // Replace with your actual Netlify URL
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowHeaders: []string{
			"Origin", "Content-Type", "Accept", "Authorization", "ngrok-skip-browser-warning", // Add ngrok-skip-browser-warning
		},
	}))

	app.GET("/exports", exportHandler.List, echojwt.JWT([]byte(os.Getenv("JWT_SECRET"))))
	app.POST("/exports", exportHandler.Create, echojwt.JWT([]byte(os.Getenv("JWT_SECRET"))))
	app.GET("/exports/download/:id", exportHandler.Get, echojwt.JWT([]byte(os.Getenv("JWT_SECRET"))))
	app.GET("/exports/pie/download/:id", exportHandler.GetPie, echojwt.JWT([]byte(os.Getenv("JWT_SECRET"))))
	app.DELETE("/exports/:id", exportHandler.Delete, echojwt.JWT([]byte(os.Getenv("JWT_SECRET"))))

	app.POST("/login", userHandler.Login)

	err = app.Start(":4000")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

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
