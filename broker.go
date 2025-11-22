package celery

import "context"

// Broker defines the interface for message brokers
type Broker interface {
	// SendTask sends a task message to the broker
	SendTask(ctx context.Context, message *CeleryMessage) error
	
	// Close closes the broker connection
	Close() error
}
