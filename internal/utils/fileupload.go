package utils

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

// FileType represents supported file types
type FileType string

const (
	FileTypeImage FileType = "image"
	FileTypePDF   FileType = "pdf"
)

// FileUploadConfig contains configuration for file uploads
type FileUploadConfig struct {
	MaxSize       int64    // Maximum file size in bytes
	AllowedTypes  []string // Allowed file extensions
	UploadDir     string   // Upload directory
	FilenameFn    func(originalName, id string) string // Function to generate filename
}

// DefaultImageConfig returns default configuration for image uploads
func DefaultImageConfig(uploadDir string) FileUploadConfig {
	return FileUploadConfig{
		MaxSize:      10 * 1024 * 1024, // 10MB
		AllowedTypes: []string{".png", ".jpg", ".jpeg"},
		UploadDir:    uploadDir,
		FilenameFn: func(originalName, id string) string {
			ext := filepath.Ext(originalName)
			return fmt.Sprintf("company_%s_logo%s", id, ext)
		},
	}
}

// DefaultPDFConfig returns default configuration for PDF uploads
func DefaultPDFConfig(uploadDir string) FileUploadConfig {
	return FileUploadConfig{
		MaxSize:      50 * 1024 * 1024, // 50MB
		AllowedTypes: []string{".pdf"},
		UploadDir:    uploadDir,
		FilenameFn: func(originalName, id string) string {
			return fmt.Sprintf("sukuk_%s_prospectus.pdf", id)
		},
	}
}

// ValidateFile validates file against configuration
func ValidateFile(file *multipart.FileHeader, config FileUploadConfig) error {
	// Check file size
	if file.Size > config.MaxSize {
		maxSizeMB := config.MaxSize / (1024 * 1024)
		return fmt.Errorf("file size too large (max %dMB)", maxSizeMB)
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := false
	for _, allowedExt := range config.AllowedTypes {
		if ext == allowedExt {
			allowed = true
			break
		}
	}
	
	if !allowed {
		return fmt.Errorf("file type not allowed. Allowed types: %v", config.AllowedTypes)
	}

	return nil
}

// EnsureUploadDir creates upload directory if it doesn't exist
func EnsureUploadDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

// SaveFile saves the uploaded file with proper validation and directory creation
func SaveFile(file *multipart.FileHeader, config FileUploadConfig, id string) (string, string, error) {
	// Validate file
	if err := ValidateFile(file, config); err != nil {
		return "", "", err
	}

	// Ensure upload directory exists
	if err := EnsureUploadDir(config.UploadDir); err != nil {
		return "", "", fmt.Errorf("failed to create upload directory: %v", err)
	}

	// Generate filename
	filename := config.FilenameFn(file.Filename, id)
	fullPath := filepath.Join(config.UploadDir, filename)
	
	// Create a temporary file to save the upload
	src, err := file.Open()
	if err != nil {
		return "", "", fmt.Errorf("failed to open uploaded file: %v", err)
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create destination file: %v", err)
	}
	defer dst.Close()

	// Copy file content
	if _, err := dst.ReadFrom(src); err != nil {
		return "", "", fmt.Errorf("failed to save file: %v", err)
	}

	// Return filename and relative URL path
	relativePath := strings.TrimPrefix(fullPath, "./")
	url := "/" + strings.Replace(relativePath, "\\", "/", -1)
	
	return filename, url, nil
}

// DeleteFile removes a file from the filesystem
func DeleteFile(filePath string) error {
	if filePath == "" {
		return nil
	}
	
	// Convert URL path back to file path
	cleanPath := strings.TrimPrefix(filePath, "/")
	fullPath := filepath.Join(".", cleanPath)
	
	if _, err := os.Stat(fullPath); err == nil {
		return os.Remove(fullPath)
	}
	return nil
}