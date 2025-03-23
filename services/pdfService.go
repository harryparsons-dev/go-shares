package services

import (
	"fmt"
	"github.com/harryparsons-dev/go-shares/models"
	"gorm.io/gorm"
	"log"
	"os/exec"
	"strings"
)

type PdfService struct {
	db *gorm.DB
}

func NewPdfService(db *gorm.DB) *PdfService {
	return &PdfService{db: db}
}

func (s *PdfService) GeneratePdf(export *models.Exports, fontSize, padding int) {

	title := strings.ReplaceAll(export.Title, " ", "-")
	pdfPath := fmt.Sprintf("assets/exports/%v/pdf/%v", export.ID, title)
	pieChartPath := fmt.Sprintf("assets/exports/%v/pie/%v", export.ID, title)

	cmd := exec.Command(
		"bash", "-c",
		fmt.Sprintf("source scripts/venv/bin/activate && python3 scripts/shares_script.py '%v' %v %v %v %v", export.SourceFilePath, fontSize, padding, pdfPath, pieChartPath),
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error creating pdf: %v : %v", err.Error(), string(output))
		export.Status = "Failed"
		export.ErrorMessages = fmt.Sprintf("%v", err)
		s.db.Save(export)
		return
	}

	export.Status = "Completed"
	export.ExportFilePath = pdfPath
	export.PieChartFilePath = pieChartPath

	s.db.Save(export)
	return
}
