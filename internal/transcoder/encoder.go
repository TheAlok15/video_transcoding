// internal/transcoder/encoder.go
package transcoder

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/TheAlok15/video_transcoding/internal/configuration"
	"github.com/TheAlok15/video_transcoding/internal/model"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Encoder struct {
	s3Client *s3.Client
	cfg      *configuration.Configuration
}

//consturtor just because we dont want to load aws confi every timw
func NewEncoder(cfg *configuration.Configuration) *Encoder {
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.AWSRegion),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AWSAccessKey,
			cfg.AWSSecretKey,
			"",
		)),
	)
	if err != nil {
		log.Fatalf("Failed to create S3 client for encoder: %v", err)
	}

	return &Encoder{
		s3Client: s3.NewFromConfig(awsCfg),
		cfg:      cfg,
	}
}

func (e *Encoder) EncodeAndUpload(job *model.Job) error {
	log.Printf(" encoding 3 resolutions for job %s", job.ID)

	outputDir := "./outputs"
	os.MkdirAll(outputDir, 0755)
	baseName := job.ID

	resolutions := []struct {
		name   string
		width  int
		height int
	}{
		{"360p", 640, 360},
		{"480p", 854, 480},
		{"720p", 1280, 720},
	}

	for _, res := range resolutions {
		outputPath := fmt.Sprintf("%s/%s_%s.mp4", outputDir, baseName, res.name)

		// we run ffpeg
		err := exec.Command(
			"ffmpeg",
			"-y",
			"-i", job.InputURL,
			"-vf", fmt.Sprintf(
				"scale=w=%d:h=%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2",
				res.width, res.height, res.width, res.height,
			),
			"-c:v", "libx264",
			"-preset", "slow",
			"-crf", "18", 
			"-profile:v", "high",
			"-level", "4.1",
			"-c:a", "aac",
			"-b:a", "128k",
			"-movflags", "+faststart",

			outputPath,
		).Run()

		if err != nil {
			return fmt.Errorf("%s encoding failed: %w", res.name, err)
		}

		// upload to s3 output bucket
		file, err := os.Open(outputPath)
		if err != nil {
			return fmt.Errorf("failed to open %s file: %w", res.name, err)
		}

		key := fmt.Sprintf("outputs/%s_%s.mp4", baseName, res.name)

		_, err = e.s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
			Bucket: aws.String(e.cfg.S3OutputBucket),
			Key:    aws.String(key),
			Body:   file,
			// ACL:    "public-read",
		})
		file.Close()

		if err != nil {
			return fmt.Errorf("failed to upload %s to S3: %w", res.name, err)
		}

		// Generate public URL
		url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s",
			e.cfg.S3OutputBucket, e.cfg.AWSRegion, key)

		switch res.name {
		case "360p":
			job.Output360URL = url
		case "480p":
			job.Output480URL = url
		case "720p":
			job.Output720URL = url
		}

		log.Printf("%s uploaded to S3: %s", res.name, url)

		// Optional: Delete local file to save disk space
		os.Remove(outputPath)
	}

	log.Printf("all 3 resolutions encoded and uploaded for job %s", job.ID)
	return nil
}
