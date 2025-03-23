package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/harryparsons-dev/go-shares/models"
	"github.com/harryparsons-dev/go-shares/services"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
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
		return c.HTML(http.StatusBadRequest, "File not found")
	}
	log.Print(export.ExportFilePath)
	return c.Attachment(fmt.Sprintf("%v.pdf", export.ExportFilePath), export.Title)
}

func (h *ExportHandler) GetPie(c echo.Context) error {
	id := c.Param("id")

	export := &models.Exports{}
	h.db.Find(export, id)

	if export.ID == 0 {
		return c.HTML(http.StatusBadRequest, "File not found")
	}
	return c.Attachment(fmt.Sprintf("%v.pdf", export.PieChartFilePath), export.Title)
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

	if filepath.Ext(file.Filename) != ".csv" {
		log.Error("Invalid file extension:", file.Filename)
		return c.JSON(http.StatusBadRequest, "Invalid file type. Please submit a .csv file")
	}

	meta := make(map[string]interface{})

	meta["font_size"] = fontSize
	meta["padding"] = padding

	metaString, err := json.Marshal(meta)
	if err != nil {
		log.Errorf("Error marshal json %v ", err)
	}
	now := time.Now()
	// create object
	export := &models.Exports{
		Title:     title,
		CreatedAt: &now,
		FileSize:  uint(file.Size),
		Status:    "Processing",
		Meta:      string(metaString),
		FontSize:  fontSize,
		Padding:   padding,
	}

	h.db.Create(export)
	// Destination
	dir := fmt.Sprintf("assets/csv/%v", export.ID)
	// Ensure the directory exists
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error making file")
	}
	export.SourceFilePath = fmt.Sprintf("%v/%v", dir, file.Filename)
	err = h.db.Save(export).Error
	if err != nil {
		log.Printf("Error saving to db: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error saving to the db")
	}

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

	go func() {
		log.Printf("Starting export")
		h.pdfService.GeneratePdf(export, fontSizeInt, paddingInt)
		log.Printf("Finished generating pdf")
	}()

	return c.JSON(http.StatusOK, "File upload successfully")
}

func (h *ExportHandler) Delete(c echo.Context) error {
	id := c.Param("id")

	export := &models.Exports{}
	h.db.Where("id = ?", id).Find(export)

	if export.ID == 0 {
		return c.JSON(http.StatusNotFound, "Export not found")
	}

	err := h.db.Delete(export).Error
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Unexpected error occurred deleting export")
	}

	return c.JSON(http.StatusOK, "Export deleted successfully")
}
