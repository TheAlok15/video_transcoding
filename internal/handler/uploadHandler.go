package handler

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/TheAlok15/video_transcoding/internal/configuration"
	"github.com/TheAlok15/video_transcoding/internal/database"
	"github.com/TheAlok15/video_transcoding/internal/model"
	"github.com/TheAlok15/video_transcoding/internal/rabbitmq"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	maxFileSize  = 100 << 20
	allowedTypes = "video/mp4 video/webm video/quicktime"
)

type Handler struct {
	RabbitMQ *rabbitmq.RabbitMQ
	S3 *s3.Client
	Configure *configuration.Configuration
}

func (h *Handler) UploadHandler(c *gin.Context) {

	file, err := c.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"message": "video file is required",
		})
		return
	}
	// fmt.Println(file.Filename)
	// fmt.Println(file.Size)

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

	fileForUpload, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file for upload"})
		return
	}
	defer fileForUpload.Close()

	//  from here s3 work start, we load config then we setup client but now we move that logic to separate constructor for better separation of concern
	cfg := h.Configure

	// awsCfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(cfg.AWSRegion), 
	// 		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
	// 		cfg.AWSAccessKey,
	// 		cfg.AWSSecretKey,
	// 		"",
	// 	)),
	// )
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load AWS configuration"})
	// 	return
	// }
    // s3Client := s3.NewFromConfig(awsCfg)

	key := fmt.Sprintf("originals/%s%s", jobID, filepath.Ext(file.Filename))

	_, err = h.S3.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(cfg.S3InputBucket),  // we can move this also to struct 
		Key:    aws.String(key),
		Body:   fileForUpload,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload to S3: " + err.Error()})
		return
	}

	// S3 URL generate (public bucket assume kr re h )
	inputURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s",cfg.S3InputBucket, cfg.AWSRegion, key)

	fmt.Printf("Video uploaded to S3: %s\n", inputURL)

	job := model.Job{
		ID:        jobID,
		InputURL:  inputURL,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = database.DB.Create(&job).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to save in db",
		})
		return
	}

	fmt.Println("job created and saved to postgres")

	err = h.RabbitMQ.PublishJob(jobID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to queue job"})
		return
	}

}
