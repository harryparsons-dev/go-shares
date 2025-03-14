package models

import "time"

type Exports struct {
	ID               uint       `json:"id" gorm:"primaryKey"`
	Title            string     `json:"title"`
	CreatedAt        *time.Time `json:"created_at"`
	SourceFilePath   string     `json:"source_file_path"`
	ExportFilePath   string     `json:"export_file_path"`
	PieChartFilePath string     `json:"pie_chart_file_path"`
	FileSize         uint       `json:"file_size"`
	Status           string     `json:"status"`
	ErrorMessages    string     `json:"error_messages"`
	FontSize         string     `json:"font_size"`
	Padding          string     `json:"padding"`
	Meta             string     `json:"meta"`
}
