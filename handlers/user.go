package handlers

import (
	"encoding/base64"
	"github.com/harryparsons-dev/go-shares/models"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strings"
	"time"
)

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

func (h *UserHandler) Login(c echo.Context) error {
	return nil
}

func (h *UserHandler) Home(c echo.Context) error {

	session, err := c.Cookie("session")
	if err != nil || session.Value == "" {

		auth := c.Request().Header.Get("Authorization")
		if auth == "" {
			c.Response().Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			return c.String(http.StatusUnauthorized, "401 - Unauthorized\n")
		}

		// Check if the header starts with "Basic "
		if !strings.HasPrefix(auth, "Basic ") {
			return c.String(http.StatusBadRequest, "Invalid Authorization header format\n")
		}

		// Decode the base64 credentials
		encodedCredentials := strings.TrimPrefix(auth, "Basic ")
		decodedBytes, err := base64.StdEncoding.DecodeString(encodedCredentials)
		if err != nil {
			return c.String(http.StatusBadRequest, "Failed to decode Authorization header\n")
		}

		// Split username and password
		credentials := strings.SplitN(string(decodedBytes), ":", 2)
		if len(credentials) != 2 {
			return c.String(http.StatusBadRequest, "Invalid Authorization header format\n")
		}

		username, password := credentials[0], credentials[1]

		user := &models.User{}
		h.db.Where("username = ?", username).First(&user)

		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err != nil {
			return c.String(http.StatusUnauthorized, "Invalid username or password")
		}

		if user.ID == 0 {
			c.Response().Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			return c.String(http.StatusUnauthorized, "Invalid credentials")
		}

		// Set session cookie
		cookie := &http.Cookie{
			Name:     "session",
			Value:    username,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
		}
		c.SetCookie(cookie)

		return c.Redirect(http.StatusSeeOther, "/exports")
	}
	return c.String(http.StatusForbidden, "no")

}

func (h *UserHandler) Logout(c echo.Context) error {
	log.Print("hello")
	cookie := &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0), // Expire it
		HttpOnly: true,
	}
	c.SetCookie(cookie)
	c.Set("user", &models.User{})
	return c.Redirect(http.StatusSeeOther, "/")
}
