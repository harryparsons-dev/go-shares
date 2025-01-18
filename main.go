package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/labstack/gommon/log"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

type Exports struct {
	ID             uint       `json:"id" gorm:"primaryKey"`
	Title          string     `json:"title"`
	CreatedAt      *time.Time `json:"created_at"`
	SourceFilePath string     `json:"source_file_path"`
	ExportFilePath string     `json:"export_file_path"`
	FileSize       uint       `json:"file_size"`
	Status         string     `json:"status"`
	Meta           string     `json:"meta"`
}

//type ExportRequest struct {
//	Title    string `json:"title"`
//	FontSize string `json:"font_size"`
//	Padding  string `json:"padding"`
//}

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
	log.Print("starting server")
	// Create an Echo instance
	e := echo.New()

	// Define routes
	e.GET("/exports", ListExport)
	e.GET("/exports/:id", GetExport)
	e.POST("/exports", CreateExport)

	e.Static("/", "html")

	// Start the server
	e.Logger.Fatal(e.Start(":8080"))
}

func ListExport(c echo.Context) error {
	var tmpl = template.Must(template.New("row").Parse(`
		<tr>
			<td>{{.ID}}</td>
			<td>{{.Title}}</td>
			<td>{{.FileSize}}</td>
			<td>{{.Status}}</td>
			<td>{{.SourceFilePath}}</td>
			<td>{{.ExportFilePath}}</td>
			<td><a href="/exports/{{.ID}}" download="{{.">Download</a></td>
			<td>{{.Meta}}</td>
		</tr>
	`))

	var exports []Exports
	result := db.Find(&exports)
	if result.Error != nil {
		return c.String(http.StatusInternalServerError, "Error fetching data")
	}
	//log.Print(exports)
	c.Response().Header().Set("Content-Type", "text/html")
	var buf bytes.Buffer
	for _, export := range exports {
		if err := tmpl.Execute(&buf, export); err != nil {
			log.Print("Template execution error:", err)
			return c.String(http.StatusInternalServerError, "Error rendering data")
		}
	}
	return c.HTML(http.StatusOK, buf.String())
}

func GetExport(c echo.Context) error {
	id := c.Param("id")
	log.Print("hello")
	export := &Exports{}
	db.Find(export, id)
	if export.ID == 0 {
		return c.HTML(http.StatusBadRequest, "<p>File not found</p>")
	}
	log.Print(export.ExportFilePath)
	return c.Attachment(fmt.Sprintf("%v.txt", export.ExportFilePath), export.Title)
}

func CreateExport(c echo.Context) error {
	//request := &ExportRequest{}
	//err := c.Bind(request)
	//if err != nil {
	//	return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	//}
	log.Print("CreateExport")
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

	// Destination
	dir := "assets/csv"
	// Ensure the directory exists
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	filePath := fmt.Sprintf("%v/%v", dir, file.Filename)
	dst, err := os.Create(filePath)
	if err != nil {
		log.Error(err)
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		log.Error(err)
	}
	now := time.Now()

	meta := make(map[string]interface{})

	meta["font_size"] = fontSize
	meta["padding"] = padding

	metaString, err := json.Marshal(meta)
	if err != nil {
		log.Error(err)
	}
	// create object
	export := &Exports{
		Title:          title,
		CreatedAt:      &now,
		SourceFilePath: filePath,
		FileSize:       uint(file.Size),
		Status:         "processing",
		Meta:           string(metaString),
	}
	db.Create(export)
	fontSizeInt, _ := strconv.Atoi(fontSize)
	paddingInt, _ := strconv.Atoi(padding)
	GeneratePdf(export, fontSizeInt, paddingInt)

	return c.HTML(http.StatusOK, fmt.Sprintf("<p>File %s uploaded successfully.</p>", file.Filename))
}

func GeneratePdf(export *Exports, fontSize, padding int) {
	pdfPath := fmt.Sprintf("assets/pdf/%v", export.ID)
	cmd := exec.Command(
		"bash", "-c",
		fmt.Sprintf("source venv/bin/activate && python3 shares_script.py %v %v %v %v", export.SourceFilePath, fontSize, padding, pdfPath),
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Command execution failed with exit status %v: %v\nOutput: %s", err, string(output))
	}
	log.Print(string(output))

	export.Status = "Complete"
	export.ExportFilePath = pdfPath
	db.Save(export)

}
