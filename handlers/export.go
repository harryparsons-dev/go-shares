package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/harryparsons-dev/go-shares/models"
	"github.com/harryparsons-dev/go-shares/services"
	"github.com/harryparsons-dev/go-shares/views/uploadView"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"
)

type ExportHandler struct {
	db         *gorm.DB
	pdfService *services.PdfService
}

func NewExportHandler(db *gorm.DB) *ExportHandler {
	return &ExportHandler{
		db:         db,
		pdfService: services.NewPdfService(db),
	}
}

func (h *ExportHandler) List(c echo.Context) error {
	//user := c.Get("user").(*models.User)
	var exports []models.Exports
	result := h.db.Find(&exports)

	if result.Error != nil {
		return c.String(http.StatusInternalServerError, "Error fetching data")
	}

	return c.JSON(http.StatusOK, exports)
}

func (h *ExportHandler) Get(c echo.Context) error {
	id := c.Param("id")
	log.Print("hello")

	export := &models.Exports{}
	h.db.Find(export, id)

	if export.ID == 0 {
		return c.HTML(http.StatusBadRequest, "<p>File not found</p>")
	}
	log.Print(export.ExportFilePath)
	return c.Attachment(fmt.Sprintf("%v.pdf", export.ExportFilePath), export.Title)
}

func (h *ExportHandler) Create(c echo.Context) error {
	title := c.FormValue("title")
	fontSize := c.FormValue("font_size")
	padding := c.FormValue("padding")

	file, err := c.FormFile("file")
	if err != nil {
		log.Print(err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	meta := make(map[string]interface{})

	meta["font_size"] = fontSize
	meta["padding"] = padding

	metaString, err := json.Marshal(meta)
	if err != nil {
		log.Error(err)
	}
	now := time.Now()
	// create object
	export := &models.Exports{
		Title:     title,
		CreatedAt: &now,
		FileSize:  uint(file.Size),
		Status:    "processing",
		Meta:      string(metaString),
	}

	h.db.Create(export)
	// Destination
	dir := fmt.Sprintf("assets/csv/%v", export.ID)
	// Ensure the directory exists
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	export.SourceFilePath = fmt.Sprintf("%v/%v", dir, file.Filename)
	h.db.Save(export)

	src, err := file.Open()
	if err != nil {
		log.Error(err)
	}
	defer func(src multipart.File) {
		err := src.Close()
		if err != nil {
			log.Error(err)
		}
	}(src)

	dst, err := os.Create(export.SourceFilePath)
	if err != nil {
		log.Error(err)
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		log.Error(err)
	}

	fontSizeInt, _ := strconv.Atoi(fontSize)
	paddingInt, _ := strconv.Atoi(padding)

	_ = h.pdfService.GeneratePdf(export, fontSizeInt, paddingInt)

	return c.HTML(http.StatusOK, fmt.Sprintf("<p>File %s uploaded successfully.</p>", file.Filename))
}

func (h *ExportHandler) ShowUploadPage(c echo.Context) error {
	user := c.Get("user").(*models.User)
	return render(c, uploadView.Show(*user))
}
