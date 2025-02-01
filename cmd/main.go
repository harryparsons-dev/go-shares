package main

import (
	"github.com/harryparsons-dev/go-shares/handlers"
	"github.com/harryparsons-dev/go-shares/models"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"net/http"
)

var db *gorm.DB

func main() {
	app := echo.New()
	initDatabase()

	//HANDLERS
	exportHandler := handlers.NewExportHandler(db)
	userHandler := handlers.NewUserHandler(db)
	//MIDDLEWHERE
	app.Use(middleware.Logger())
	app.Use(middleware.Recover())
	app.Use(middleware.CORS())

	// Pages
	app.GET("/login", userHandler.Home)
	app.GET("/logout", userHandler.Logout)

	app.GET("/exports", exportHandler.List)
	app.POST("/exports", exportHandler.Create, authMiddleware)
	//Endpoints
	app.GET("/upload", exportHandler.ShowUploadPage, authMiddleware)
	app.GET("/exports/download/:id", exportHandler.Get, authMiddleware)

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
	err = db.AutoMigrate(models.Exports{}, models.User{})
	if err != nil {
		panic("failed to migrate database")
	}
}

func authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie("session")
		log.Println("cookie", cookie.Value)
		if err != nil || cookie.Value == "" {
			return c.Redirect(http.StatusSeeOther, "/login")
		}

		// Validate session
		username := cookie.Value
		user := &models.User{}

		db.Where("username = ?", username).Find(&user)
		if err != nil || user.ID == 0 {
			return c.Redirect(http.StatusSeeOther, "/login")
		}

		// Store username in context
		c.Set("user", user)

		return next(c)
	}
}
