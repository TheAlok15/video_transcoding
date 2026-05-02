package transcoder

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	// "path/filepath"
	"strconv"
)

type Decoder struct{}

func NewDecoder() *Decoder {
	return &Decoder{}
}

type VideoDetails struct{
	Streams []struct{
		CodecType string `json:"codec_type"`
	}`json:"streams"`
	Format struct{
		Duration string  `json:"duration"`
		Size string `json:"size"`
	}`json:"format"`

}

func (d *Decoder) Decode(file string) error {
	log.Printf("Decoding job %s - Running ffprobe", file)

	info, err := os.Stat(file)
	if err != nil {
		return fmt.Errorf("file not found : %w ", err)
	}

	if info.Size() == 0 {
		return fmt.Errorf("File is empty : %w ", err)

	}

	// here we get video meta data using ffprobe
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", file)
	output, err := cmd.CombinedOutput()
	fmt.Println(string(output))
	if err != nil {
		return fmt.Errorf("ffprobe failed: %w", err)
	}

	var video VideoDetails

	if err = json.Unmarshal(output, &video); err != nil {
		return fmt.Errorf("invalid video format: %w", err)
	}

	hasVideoStream := false

	for _, st := range video.Streams{
		if st.CodecType == "video"{
			hasVideoStream = true
			break
		}
	}
	if !hasVideoStream{
		return fmt.Errorf("no video stream found")
	}

	duration, _ := strconv.ParseFloat(video.Format.Duration, 64)
	size, _ := strconv.ParseInt(video.Format.Size, 10, 64)

	if duration > 120{
		return fmt.Errorf("video too long")
	}

	if size > 100*1024*1024 { // 100MB
	return fmt.Errorf("file too large")
}

	
	return nil
}
