// internal/transcoder/decoder.go
package transcoder

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"

	"github.com/TheAlok15/video_transcoding/internal/model"
)

type Decoder struct{}

func NewDecoder() *Decoder {
	return &Decoder{}
}

func (d *Decoder) Decode(job *model.Job) error {
	log.Printf("Decoding job %s - Running ffprobe", job.ID)

	// here we get video meta data using ffprobe
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", job.InputURL)
	output, err := cmd.CombinedOutput()
	fmt.Println(string(output))
	if err != nil {
		return fmt.Errorf("ffprobe failed: %w", err)
	}

	// justt checking if output is valid json
	var probe struct {
		Streams []struct { CodecType string `json:"codec_type"`} `json:"streams"`
	}
	if err := json.Unmarshal(output, &probe); err != nil {
		return fmt.Errorf("invalid video format: %w", err)
	}

	log.Printf("Probe successful for job %s", job.ID)
	return nil
}
