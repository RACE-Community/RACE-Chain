package pub

import (
	"context"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/ethereum-optimism/optimism/l2geth/log"
)

const messageOrderingKey = "o"

type Config struct {
	ProjectID string
	TopicID   string
	Timeout   time.Duration
}

type GooglePublisher struct {
	client          *pubsub.Client
	topic           *pubsub.Topic
	publishSettings pubsub.PublishSettings
	Timeout         time.Duration
}

func NewGooglePublisher(ctx context.Context, config Config) (*GooglePublisher, error) {
	client, err := pubsub.NewClient(ctx, config.ProjectID)
	if err != nil {
		return nil, err
	}
	topic := client.Topic(config.TopicID)
	topic.EnableMessageOrdering = true

	// Publish messages immediately
	publishSettings := pubsub.PublishSettings{
		DelayThreshold: 0,
		CountThreshold: 0,
	}
	timeout := config.Timeout
	if timeout == 0 {
		log.Info("Sanitizing publisher timeout to 2 seconds")
		timeout = time.Second * 2
	}
	return &GooglePublisher{client, topic, publishSettings, timeout}, nil
}

func (p *GooglePublisher) Publish(ctx context.Context, msg []byte) error {
	ctx, cancel := context.WithTimeout(ctx, p.Timeout)
	defer cancel()
	pmsg := pubsub.Message{
		Data:        msg,
		OrderingKey: messageOrderingKey,
	}
	// If there was an error previously, clear it out to allow publishing to work again
	p.topic.ResumePublish(messageOrderingKey)
	result := p.topic.Publish(ctx, &pmsg)
	_, err := result.Get(ctx)
	return err
}
