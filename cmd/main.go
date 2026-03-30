package main

import (
	"log"

	"github.com/TheAlok15/video_transcoding/internal/configuration"
	"github.com/TheAlok15/video_transcoding/internal/database"
	"github.com/TheAlok15/video_transcoding/internal/handler"
	"github.com/TheAlok15/video_transcoding/internal/rabbitmq"
	"github.com/TheAlok15/video_transcoding/internal/storage"
	"github.com/TheAlok15/video_transcoding/internal/worker"

	"github.com/gin-gonic/gin"
)

func main() {

	cfg := configuration.Load()

	database.Init(cfg)

	// rabbitMQ constructor call
	rabbitMQ, err := rabbitmq.NewRabbitMQ(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ: %v", err)
	}
	// Graceful shutdown
	defer rabbitMQ.Close()

	// s3 constructor call
	s3client, err := storage.NewS3Client(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize S3 client: %v", err)
	}

	worker := worker.NewWorker(rabbitMQ, database.DB, cfg)
	worker.Start(5)
	router := gin.Default()
	h := &handler.Handler{
		RabbitMQ:  rabbitMQ,
		S3:        s3client,
		Configure: cfg,
	}

	router.POST("/upload", h.UploadHandler)
	router.GET("/status/:job_id", handler.GetJobStatus)

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"message": "Server is healthy"})
	})

	log.Printf("Server starting on : %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}

}
