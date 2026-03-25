package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Port           string `mapstructure:"PORT"`
	AWSRegion      string `mapstructure:"AWS_REGION"`
	AWSAccessKey   string `mapstructure:"AWS_ACCESS_KEY_ID"`
	AWSSecretKey   string `mapstructure:"AWS_SECRET_ACCESS_KEY"`
	S3InputBucket  string `mapstructure:"S3_INPUT_BUCKET"`
	S3OutputBucket string `mapstructure:"S3_OUTPUT_BUCKET"`
	RabbitMQURL    string `mapstructure:"RABBITMQ_URL"`
	DBURL          string `mapstructure:"DB_URL"`
	MaxFileSizeMB  int    `mapstructure:"MAX_FILE_SIZE_MB"`
	LogLevel       string `mapstructure:"LOG_LEVEL"`
	MaxRetries     int    `mapstructure:"MAX_RETRIES"`
}

func Load() *Config {

	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	viper.SetDefault("PORT", "3000")
	viper.SetDefault("MAX_FILE_SIZE_MB", "100")
	viper.SetDefault("MAX_RETRIES", "3")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Fatalf("Fatal config error: %v", err)
		}
	}


	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("unable to decode into struct : %v", err)
	}

	if cfg.AWSAccessKey == "" || cfg.AWSRegion == "" || cfg.AWSSecretKey == "" {
		log.Fatal("missing required aws credentials")
	} 

	if cfg.RabbitMQURL == ""{
		log.Fatal("missing rabbit mq url")
	}

	return &cfg



}
