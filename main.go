package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/a-h/templ"
	"github.com/harryparsons-dev/go-shares/views"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
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

func initDatabase() {
	var err error
	// Initialize database connection
	db, err = gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Auto-migrate the User models
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
	// load templates

	// Define routes
	e.GET("/exportView", ListExport)
	e.GET("/exportView/:id", GetExport)
	e.POST("/exportView", CreateExport)

	e.GET("/", func(c echo.Context) error {
		return c.File("html/index.html")
	})
	e.GET("/upload", func(c echo.Context) error {
		return c.File("html/upload.html")
	})
	e.GET("/home", homeHandler)
	// Start the server
	e.Logger.Fatal(e.Start(":8080"))
}

func homeHandler(c echo.Context) error {
	page := views.Layout("Home page")
	return render(c, page)
}
func render(ctx echo.Context, cmp templ.Component) error {
	return cmp.Render(ctx.Request().Context(), ctx.Response())
}

func ListExport(c echo.Context) error {
	var tmpl = template.Must(template.New("row").Parse(`
		<tr>
			<td>{{.ID}}</td>
			<td>{{.Title}}</td>
			<td>{{.CreatedAt}}</td>
			<td>{{.FileSize}}</td>
			<td>{{.Status}}</td>
			<td>{{.SourceFilePath}}</td>
			<td>{{.ExportFilePath}}</td>
			<td><button><a href="/exportView/{{.ID}}" download="{{.ID}}">Download</a></button></td>
			<td>{{.Meta}}</td>
		</tr>
	`))

	var exports []Exports
	result := db.Find(&exports)
	if result.Error != nil {
		return c.String(http.StatusInternalServerError, "Error fetching data")
	}
	//log.Print(exportView)
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
	return c.Attachment(fmt.Sprintf("%v.pdf", export.ExportFilePath), export.Title)
}

func CreateExport(c echo.Context) error {
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
	export := &Exports{
		Title:     title,
		CreatedAt: &now,
		FileSize:  uint(file.Size),
		Status:    "processing",
		Meta:      string(metaString),
	}

	db.Create(export)
	// Destination
	dir := fmt.Sprintf("assets/csv/%v", export.ID)
	// Ensure the directory exists
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	export.SourceFilePath = fmt.Sprintf("%v/%v", dir, file.Filename)
	db.Save(export)

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
	GeneratePdf(export, fontSizeInt, paddingInt)

	return c.HTML(http.StatusOK, fmt.Sprintf("<p>File %s uploaded successfully.</p>", file.Filename))
}

func GeneratePdf(export *Exports, fontSize, padding int) {

	title := strings.ReplaceAll(export.Title, " ", "-")
	pdfPath := fmt.Sprintf("assets/pdf/%v/%v", export.ID, title)

	log.Print(export.SourceFilePath)
	log.Print(pdfPath)

	log.Print(padding)
	log.Print(fontSize)

	cmd := exec.Command(
		"bash", "-c",
		fmt.Sprintf("source scripts/venv/bin/activate && python3 scripts/shares_script.py %v %v %v %v", export.SourceFilePath, fontSize, padding, pdfPath),
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
