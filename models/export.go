package models

import "time"

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
