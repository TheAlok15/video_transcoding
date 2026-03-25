package file

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	maxFileSize  = 100 << 20
	allowedTypes = "video/mp4 video/webm video/quicktime"
)

func UploadHandler(c *gin.Context) {

	file, err := c.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"message": "video file is required",
		})
		return
	}
	fmt.Println(file.Filename)
	fmt.Println(file.Size)

	if file.Size > maxFileSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{
			"error": fmt.Sprintf("file too large (max %d MB)", maxFileSize>>20),
		})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer f.Close()
	buffer := make([]byte, 512)
	f.Read(buffer)

	detectedType := http.DetectContentType(buffer)
	if !strings.Contains(allowedTypes, detectedType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported file type"})
		return
	}

	fmt.Printf("Uploaded file: %s, Size: %d bytes, Detected Type: %s\n", file.Filename, file.Size, detectedType)

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".mp4" && ext != ".webm" && ext != ".mov" && ext != ".mkv" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported file extension"})
		return
	}

	jobID := uuid.New().String()

	c.JSON(http.StatusAccepted, gin.H{
		"job_id":  jobID,
		"message": "Processing started. Check /status/" + jobID + " for updates.",
	})

}
