// internal/pipeline/pipeline.go
package pipeline

import (
	"fmt"
	"log"

	"github.com/TheAlok15/video_transcoding/internal/configuration"
	"github.com/TheAlok15/video_transcoding/internal/model"
	"github.com/TheAlok15/video_transcoding/internal/transcoder"
)

type Pipeline struct {
	decoder   *transcoder.Decoder
	processor *transcoder.Processor
	encoder   *transcoder.Encoder
}

func NewPipeline(cfg *configuration.Configuration) *Pipeline {
	return &Pipeline{
		decoder:   transcoder.NewDecoder(),
		processor: transcoder.NewProcessor(),
		encoder:   transcoder.NewEncoder(cfg),
	}
}

func (p *Pipeline) Process(job *model.Job) error {
	log.Printf("Starting full transcoding pipeline for job %s", job.ID)

	// decoding 
	log.Println("Step 1: Decoding video...")
	if err := p.decoder.Decode(job); err != nil {
		return fmt.Errorf("decode failed: %w", err)
	}

	//  processing 
	log.Println("Step 2: Processing (resizing)...")
	if err := p.processor.Process(job); err != nil {
		return fmt.Errorf("process failed: %w", err)
	}

	// then encode + upload
	log.Println("Step 3: Encoding and uploading to S3...")
	if err := p.encoder.EncodeAndUpload(job); err != nil {
		return fmt.Errorf("encode failed: %w", err)
	}

	job.Status = "completed"
	log.Printf("Pipeline completed successfully for job %s", job.ID)

	return nil
}