package mocks

import (
	"context"

	"github.com/mainflux/mainflux/pkg/messaging"
)

type MockPublisher struct{}

// NewPublisher returns mock message publisher.
func NewPublisher() messaging.Publisher {
	return MockPublisher{}
}

func (pub MockPublisher) Publish(ctx context.Context, topic string, msg *messaging.Message) error {
	return nil
}

func (pub MockPublisher) Close() error {
	return nil
}
