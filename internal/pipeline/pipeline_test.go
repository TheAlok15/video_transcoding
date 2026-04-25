package pipeline

import (
	"testing"

	"github.com/TheAlok15/video_transcoding/internal/configuration"
	"github.com/TheAlok15/video_transcoding/internal/model"
)

func TestPipelineStructure(t *testing.T) {
	cfg := &configuration.Configuration{}

	p := NewPipeline(cfg)

	if p.decoder == nil || p.processor == nil || p.encoder == nil {
		t.Fatal("pipeline not initialized properly")
	}
}

func TestJobInitialization(t *testing.T) {
	job := model.Job{
		ID:     "test-id",
		Status: "pending",
	}

	if job.ID == "" {
		t.Fatal("job ID should not be empty")
	}
}