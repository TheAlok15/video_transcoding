package configuration

import (
	"log"

	"github.com/spf13/viper"
)

type Configuration struct {
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
	MaxMessages    int    `mapstructure:"MAX_MESSAGES"`
	WaitTime       int    `mapstructure:"WAIT_TIME"`
	SQSQueueURL string `mapstructure:"SQS_QUEUE_URL"`
}

func Load() *Configuration {

	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	viper.SetDefault("PORT", "3000")
	viper.SetDefault("MAX_FILE_SIZE_MB", 100)
	viper.SetDefault("MAX_RETRIES", 3)

	// log.Println("DB_URL:", viper.GetString("DB_URL"))

	viper.SetConfigFile(".env")

	if err := viper.ReadInConfig(); err != nil {
		log.Println("No .env file found, using defaults")
	}

	var cfg Configuration
	log.Println("DB_URL AFTER READ:", viper.GetString("DB_URL"))
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("unable to decode into struct : %v", err)
	}

	if cfg.AWSAccessKey == "" || cfg.AWSRegion == "" || cfg.AWSSecretKey == "" {
		log.Fatal("missing required aws credentials")
	}

	if cfg.RabbitMQURL == "" {
		log.Fatal("missing rabbit mq url")
	}

	return &cfg

}
