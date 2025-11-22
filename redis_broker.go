package celery

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// RedisBroker implements Broker interface for Redis
type RedisBroker struct {
	client *redis.Client
	queue  string
}

// RedisBrokerConfig contains configuration for Redis broker
type RedisBrokerConfig struct {
	Addr     string // Redis server address (e.g., "localhost:6379")
	Password string // Redis password (empty if no password)
	DB       int    // Redis database number
	Queue    string // Default queue name (default: "celery")
}

// NewRedisBroker creates a new Redis broker instance
func NewRedisBroker(config RedisBrokerConfig) (*RedisBroker, error) {
	if config.Queue == "" {
		config.Queue = "celery"
	}

	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisBroker{
		client: client,
		queue:  config.Queue,
	}, nil
}

// SendTask sends a task message to Redis
func (rb *RedisBroker) SendTask(ctx context.Context, message *CeleryMessage) error {
	// Encode the message to JSON
	data, err := message.Encode()
	if err != nil {
		return fmt.Errorf("failed to encode message: %w", err)
	}

	// Push to Redis list
	queue := rb.queue
	if message.Properties.DeliveryInfo.RoutingKey != "" {
		queue = message.Properties.DeliveryInfo.RoutingKey
	}

	if err := rb.client.LPush(ctx, queue, data).Err(); err != nil {
		return fmt.Errorf("failed to push message to Redis: %w", err)
	}

	return nil
}

// Close closes the Redis connection
func (rb *RedisBroker) Close() error {
	return rb.client.Close()
}
