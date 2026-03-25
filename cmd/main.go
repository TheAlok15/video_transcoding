package main

import (
	"log"

	"github.com/TheAlok15/video_transcoding/internal/config"
	"github.com/TheAlok15/video_transcoding/internal/handler"
	"github.com/gin-gonic/gin"
	// "honnef.co/go/tools/config"
)

func main() {

	cfg := config.Load()

	router := gin.Default()
	router.POST("/upload", file.UploadHandler)

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"message": "Setver is healthy"})
	})

	Port := "8080"

	router.Run("localhost:8080")
	log.Printf("Server starting on :%s", Port)
	if err := router.Run(":" + Port); err != nil {
		log.Fatal(err)
	}



}
