package main

import (
	"log"
	"os"

	"github.com/TheAlok15/video_transcoding/internal/configuration"
	"github.com/TheAlok15/video_transcoding/internal/database"
	"github.com/TheAlok15/video_transcoding/internal/handler"
	"github.com/TheAlok15/video_transcoding/internal/rabbitmq"
	"github.com/TheAlok15/video_transcoding/internal/sqs"
	"github.com/TheAlok15/video_transcoding/internal/storage"
	"github.com/TheAlok15/video_transcoding/internal/worker"

	"github.com/gin-gonic/gin"
)

func main() {

	mode := os.Getenv("APP_MODE")

	if mode != "ci" {

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

		sqsConsumer, err := sqs.NewSQSClient(cfg.SQSQueueURL, cfg)
		if err != nil {
			log.Fatalf("Failed to init SQS: %v", err)
		}

		go sqsConsumer.GetMessages(cfg, rabbitMQ)

		worker := worker.NewWorker(rabbitMQ, database.DB, cfg, s3client)
		worker.Start(5, cfg)
		if mode == "ci" {
			log.Println("Running in CI mode - skipping server start")
			return
		}
		router := gin.Default()
		h := &handler.Handler{
			RabbitMQ:  rabbitMQ,
			S3:        s3client,
			Configure: cfg,
		}

		router.POST("/upload", h.CreateUpload)
		router.GET("/status/:job_id", handler.GetJobStatus)

		router.GET("/health", func(ctx *gin.Context) {
			ctx.JSON(200, gin.H{"message": "Server is healthy"})
		})

		log.Printf("Server starting on : %s", cfg.Port)
		if err := router.Run(":" + cfg.Port); err != nil {
			log.Fatal(err)
		}

	}

}
