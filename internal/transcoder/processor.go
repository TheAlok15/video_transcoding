// internal/transcoder/processor.go
package transcoder

import (
	"fmt"
	"log"
	"os"

	"github.com/TheAlok15/video_transcoding/internal/model"

)

type Processor struct{}

func NewProcessor() *Processor {
	return &Processor{}
}

func (p *Processor) Process(job *model.Job) error {
	log.Printf("processing (resizing) job %s", job.ID)

	// Create temp directory for intermediate files
	tempDir := fmt.Sprintf("./temp/%s", job.ID)
	os.MkdirAll(tempDir, 0755) 

	// We will directly encode with different scales (no raw decode for speed now)
	return nil
}