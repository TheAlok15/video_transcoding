package sqs

import (
	"context"
	"encoding/json"
	"log"

	"github.com/TheAlok15/video_transcoding/internal/configuration"
	"github.com/TheAlok15/video_transcoding/internal/database"
	"github.com/TheAlok15/video_transcoding/internal/model"
	"github.com/TheAlok15/video_transcoding/internal/rabbitmq"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type Consumer struct {
	SQSClient *sqs.Client
	QueueURL  string
}

type S3Event struct {
	Records []struct {
		S3 struct {
			Object struct {
				Key string `json:"key"`
			} `json:"object"`
		} `json:"s3"`
	} `json:"Records"`
}

func NewSQSClient(queueURL string, cfg *configuration.Configuration) (*Consumer, error) {

	sqsClient, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(cfg.AWSRegion))
	if err != nil {
		return nil, err
	}
	return &Consumer{
		SQSClient: sqs.NewFromConfig(sqsClient),
		QueueURL:  queueURL,
	}, nil

}

func (c *Consumer) GetMessages(cf *configuration.Configuration, rabbit *rabbitmq.RabbitMQ) {

	// var messages []types.Message
	for {

		result, err := c.SQSClient.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
			QueueUrl:            &c.QueueURL,
			MaxNumberOfMessages: int32(cf.MaxMessages),
			WaitTimeSeconds:     int32(cf.WaitTime),
		})
		if err != nil {
			log.Printf("Couldn't get messages from queue %v. Here's why: %v\n", c.QueueURL, err)
		}

		for _, msg := range result.Messages {
			go c.processMessage(msg, rabbit)

		}

	}

}

func (c *Consumer) processMessage(msg types.Message, rabbit *rabbitmq.RabbitMQ) {

	var event S3Event

	err := json.Unmarshal([]byte(*msg.Body), &event)
	if err != nil {
		log.Println("invalid message:", err)
		return
	}

	if len(event.Records) == 0 {
		return
	}

	key := event.Records[0].S3.Object.Key
	log.Println("received key:", key)

	// find job
	var job model.Job
	err = database.DB.First(&job, "input_url = ?", key).Error
	if err != nil {
		log.Println("job not found:", key)
		return
	}

	// avoid duplicate
	if job.Status != "uploading" {
		log.Println("already processed:", job.ID)
		return
	}

	// update status
	job.Status = "pending"
	database.DB.Save(&job)

	err = rabbit.PublishJob(job.ID)
	if err != nil {
		log.Println("rabbit publish failed:", err)
		return
	}

	// delete message
	_, err = c.SQSClient.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
		QueueUrl:      &c.QueueURL,
		ReceiptHandle: msg.ReceiptHandle,
	})
	if err != nil {
		log.Println("delete failed:", err)
	}
}
