package database

import (
	"log"
	"time"

	"github.com/TheAlok15/video_transcoding/internal/configuration"
	"github.com/TheAlok15/video_transcoding/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(cfg *configuration.Configuration) {

	dsn := cfg.DBURL

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	// Connection pool setting
	sqlDB, _ := DB.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Auto migrate Job table
	err = DB.AutoMigrate(&model.Job{})
	if err != nil {
		log.Fatalf("Failed to migrate Job model: %v", err)
	}

	log.Println("connected to postgresql and migrated Job table")

}
