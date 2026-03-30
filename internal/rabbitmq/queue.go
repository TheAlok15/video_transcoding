// internal/rabbitmq/rabbitmq.go
package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/TheAlok15/video_transcoding/internal/configuration"
	"github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn *amqp091.Connection
	ch   *amqp091.Channel
}

func NewRabbitMQ(cfg *configuration.Configuration) (*RabbitMQ, error) {
	conn, err := amqp091.Dial(cfg.RabbitMQURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// i use Quorum queue for reliability
	_, err = ch.QueueDeclare(
		"transcode_queue",
		true,   // durable
		false,  // autoDelete => means it should not be delete automatically when not in use
		false,  // exclusive => 
		false,  // no wait
		amqp091.Table{
			amqp091.QueueTypeArg: amqp091.QueueTypeQuorum,
		},
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	log.Println("rabbitMQ connected successfully and Queue: transcode_queue ready")

	return &RabbitMQ{
		conn: conn,
		ch:   ch,
	}, nil
}

// PublishJob
func (r *RabbitMQ) PublishJob(jobID string) error {
	if r.ch == nil {
		return fmt.Errorf("rabitmq channel is not started")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := r.ch.PublishWithContext(
		ctx,
		"",                // exchange
		"transcode_queue", // routing key
		false,
		false,
		amqp091.Publishing{
			ContentType:  "text/plain",
			Body:         []byte(jobID),
			DeliveryMode: amqp091.Persistent,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish job %s: %w", jobID, err)
	}

	log.Printf("Job %s pushed to the queue", jobID)
	return nil
}

func (r *RabbitMQ) Close() {
	if r.ch != nil {
		r.ch.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}
	log.Println("RabbitMQ connection closed")
}

func (r *RabbitMQ) NewChannel()(*amqp091.Channel, error){
	return r.conn.Channel()
}
