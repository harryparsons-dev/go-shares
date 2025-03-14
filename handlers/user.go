package handlers

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

func (h *UserHandler) Login(c echo.Context) error {
	type Request struct {
		Code int `json:"code"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request"})
	}

	intCode, err := strconv.Atoi(os.Getenv("AUTH_CODE"))
	if err != nil {
		log.Printf("Error converting code %v", err)
		return err
	}

	if req.Code != intCode {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid code")
	}

	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Hour * 1).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		log.Printf("Error %v", err)
		return c.JSON(http.StatusInternalServerError, "Error")
	}
	fmt.Println("Generated Token:", tokenString)
	log.Printf("JWT_SECRET used for signing: %s", os.Getenv("JWT_SECRET"))

	return c.JSON(http.StatusOK, echo.Map{
		"token": tokenString,
	})
}
