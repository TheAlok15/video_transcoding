package storage

import (
	"context"

	"github.com/TheAlok15/video_transcoding/internal/configuration"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)


func NewS3Client(cfg *configuration.Configuration) (*s3.Client, error){

	awsCfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(cfg.AWSRegion), config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
		cfg.AWSAccessKey,
		cfg.AWSSecretKey,
		"",
	)))

	if err != nil {
		return nil, err
	}

	return s3.NewFromConfig(awsCfg), nil



}