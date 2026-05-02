// internal/pipeline/pipeline.go
package pipeline

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/TheAlok15/video_transcoding/internal/configuration"
	"github.com/TheAlok15/video_transcoding/internal/model"
	"github.com/TheAlok15/video_transcoding/internal/transcoder"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Pipeline struct {
	decoder   *transcoder.Decoder
	processor *transcoder.Processor
	encoder   *transcoder.Encoder
	S3Client  *s3.Client
}

func NewPipeline(cfg *configuration.Configuration,s3Client *s3.Client) *Pipeline {
	return &Pipeline{
		decoder:   transcoder.NewDecoder(),
		processor: transcoder.NewProcessor(),
		encoder:   transcoder.NewEncoder(cfg),
		S3Client:  s3Client,
	}
}

func (p *Pipeline) Process(job *model.Job, cfg *configuration.Configuration) error {


	// create a folder

	tempDir := fmt.Sprintf("/tmp/%s", job.ID)
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		return err
	}


	// create file 


	filePath := fmt.Sprintf("%s/input.mp4", tempDir)
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()


	//  get the video from aws
	
	result, err := p.S3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(cfg.S3InputBucket),
		Key:    aws.String(job.InputURL),
	})
	if err != nil {
		return err
	}
	defer result.Body.Close()


	// stream it 

	_,err = io.Copy(file, result.Body)
	if err != nil {
		return err
	}

	
	log.Printf("Starting full transcoding pipeline for job %s", job.ID)

	// decoding
	log.Println("Step 1: Decoding video...")
	if err := p.decoder.Decode(filePath); err != nil {
		return fmt.Errorf("decode failed: %w", err)
	}

	//  processing
	log.Println("Step 2: Processing (resizing)...")
	if err := p.processor.Process(job, filePath); err != nil {
		return fmt.Errorf("process failed: %w", err)
	}

	// then encode + upload
	log.Println("Step 3: Encoding and uploading to S3...")
	if err := p.encoder.EncodeAndUpload(job, filePath); err != nil {
		return fmt.Errorf("encode failed: %w", err)
	}

	job.Status = "completed"
	log.Printf("Pipeline completed successfully for job %s", job.ID)

	return nil
}
