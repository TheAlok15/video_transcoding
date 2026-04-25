package worker

import (
	"log"

	"github.com/TheAlok15/video_transcoding/internal/configuration"
	"github.com/TheAlok15/video_transcoding/internal/model"
	"github.com/TheAlok15/video_transcoding/internal/pipeline"
	"github.com/TheAlok15/video_transcoding/internal/rabbitmq"
	"gorm.io/gorm"
)

type Worker struct {
	rabbit   *rabbitmq.RabbitMQ
	pipeline *pipeline.Pipeline
	DB       *gorm.DB
}

func NewWorker(rabbit *rabbitmq.RabbitMQ, db *gorm.DB, cfg *configuration.Configuration) *Worker {
	return &Worker{
		rabbit:   rabbit,
		pipeline: pipeline.NewPipeline(cfg),
		DB:       db,
	}
}

func (w *Worker) run(workerID int) {

	ch, err := w.rabbit.NewChannel()
	if err != nil {
		log.Printf("Worker %d: failed to create channel: %v", workerID, err)
		return
	}
	defer ch.Close()

	ch.Qos(1, 0, false)

	msgs, err := ch.Consume(
		"transcode_queue",
		"",
		false, // manual ack => job is removed only after successfuly processing, worker must manually confirm 
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("worker %d: consume error: %v", workerID, err)
		return
	}

	log.Printf("worker %d started", workerID)

	for msg := range msgs {

		jobID := string(msg.Body)
		log.Printf("Worker %d: received job %s", workerID, jobID)

		// load job from db
		var job model.Job
		if err := w.DB.First(&job, "id = ?", jobID).Error; err != nil {
			log.Printf("Worker %d: job not found %s", workerID, jobID)
			msg.Nack(false, true)
			continue
		}

		// then process pipleine
		err := w.pipeline.Process(&job)

		if err != nil {
			log.Printf("Worker %d: job %s failed: %v", workerID, jobID, err)

			job.Status = "failed"
			job.ErrorMessage = err.Error()
			w.DB.Save(&job)

			msg.Nack(false, true) // retry
		} else {

			job.Status = "completed"
			w.DB.Save(&job)

			msg.Ack(false)
		}
	}
}

func (w *Worker) Start(numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		go w.run(i + 1)
	}
	log.Printf("%d workers started", numWorkers)
}