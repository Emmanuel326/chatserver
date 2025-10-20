// Handles file uploads (images, attachments).
// Receives multipart/form-data POST requests,
// stores the uploaded files under /uploads,
// and responds with a public access URL.

package http

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	MaxUploadSize = 20 * 1024 * 1024 // 20 MB limit
	UploadDir     = "./uploads"
)

// UploadHandler handles POST /upload requests.
// It expects a multipart form field named "file".
func UploadHandler(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "no file uploaded — form field 'file' missing",
		})
	}

	// Validate file size
	if file.Size > MaxUploadSize {
		return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
			"error": fmt.Sprintf("file too large — limit is %d MB", MaxUploadSize/(1024*1024)),
		})
	}

	// Validate allowed file extensions
	ext := filepath.Ext(file.Filename)
	allowed := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
	}
	if !allowed[ext] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("unsupported file type: %s — only JPG and PNG are allowed", ext),
		})
	}

	// Ensure uploads directory exists
	if err := os.MkdirAll(UploadDir, os.ModePerm); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to create uploads directory",
		})
	}

	// Generate a unique filename
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))
	savePath := filepath.Join(UploadDir, filename)

	// Save file
	if err := c.SaveFile(file, savePath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to save file to disk",
		})
	}

	// Construct public URL
	fileURL := fmt.Sprintf("%s/uploads/%s", c.BaseURL(), filename)

	// Return response
	return c.JSON(fiber.Map{
		"message": "upload successful",
		"url":     fileURL,
	})
}

