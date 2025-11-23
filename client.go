package celery

import (
	"context"
	"fmt"
	"time"
)

// Client is the main Celery client
type Client struct {
	broker   Broker
	queue    string
	exchange string
	// UseRawJSONBody is a flag to indicate whether to use raw JSON body instead of base64 encoded body
	// This is typically used for AMQP brokers where the worker is configured to accept raw JSON.
	UseRawJSONBody bool
}

// ClientConfig contains configuration for Celery client
type ClientConfig struct {
	Broker   Broker // The broker implementation to use
	Queue    string // Default queue name (default: "celery")
	Exchange string // Default exchange name (default: "celery")
	// UseRawJSONBody is a flag to indicate whether to use raw JSON body instead of base64 encoded body
	UseRawJSONBody bool
}

// NewClient creates a new Celery client
func NewClient(config ClientConfig) *Client {
	if config.Queue == "" {
		config.Queue = "celery"
	}
	if config.Exchange == "" {
		config.Exchange = "celery"
	}

	return &Client{
		broker:         config.Broker,
		queue:          config.Queue,
		exchange:       config.Exchange,
		UseRawJSONBody: config.UseRawJSONBody,
	}
}

// TaskOptions contains options for task execution
type TaskOptions struct {
	Queue    string                 // Override default queue
	Exchange string                 // Override default exchange
	ETA      *time.Time             // Estimated time of arrival
	Expires  *time.Time             // Task expiration time
	Args     []interface{}          // Positional arguments
	Kwargs   map[string]interface{} // Keyword arguments
}

// SendTask sends a task to the Celery worker
func (c *Client) SendTask(ctx context.Context, taskName string, options *TaskOptions) (string, error) {
	if options == nil {
		options = &TaskOptions{}
	}

	// Create task message
	taskMsg := NewTaskMessage(taskName, options.Args, options.Kwargs)

	// Set ETA if provided
	if options.ETA != nil {
		taskMsg.SetETA(*options.ETA)
	}

	// Set expiration if provided
	if options.Expires != nil {
		taskMsg.SetExpires(*options.Expires)
	}

	// Encode task message
	var encodedBody string
	var contentType string = "application/json"
	var bodyEncoding string = "base64"
	var contentEncoding string = "utf-8"

	if c.UseRawJSONBody {
		var err error
		encodedBody, err = taskMsg.EncodeJSON()
		if err != nil {
			return "", fmt.Errorf("failed to encode task message to raw JSON: %w", err)
		}
		// When using raw JSON body, the content-type should be application/json
		// and body_encoding should be set to "json" or similar, but Celery protocol v1
		// expects "base64" in properties.body_encoding for base64 body.
		// For raw JSON body, we set content-type to application/json and body_encoding to utf-8
		// as the body is the raw JSON string itself.
		contentType = "application/json"
		bodyEncoding = "utf-8" // Indicate the body is utf-8 encoded raw JSON
		contentEncoding = "utf-8"
	} else {
		var err error
		encodedBody, err = taskMsg.Encode() // Base64 encoded JSON
		if err != nil {
			return "", fmt.Errorf("failed to encode task message to base64: %w", err)
		}
		contentType = "application/json"
		bodyEncoding = "base64"
		contentEncoding = "utf-8"
	}

	// Determine queue and exchange
	queue := c.queue
	if options.Queue != "" {
		queue = options.Queue
	}

	exchange := c.exchange
	if options.Exchange != "" {
		exchange = options.Exchange
	}

	// Create Celery message envelope
	celeryMsg := NewCeleryMessageWithEncoding(encodedBody, queue, exchange, contentType, bodyEncoding, contentEncoding)

	// Send to broker
	if err := c.broker.SendTask(ctx, celeryMsg); err != nil {
		return "", fmt.Errorf("failed to send task: %w", err)
	}

	return taskMsg.ID, nil
}

// SendTaskWithArgs is a convenience method to send a task with positional arguments
func (c *Client) SendTaskWithArgs(ctx context.Context, taskName string, args ...interface{}) (string, error) {
	return c.SendTask(ctx, taskName, &TaskOptions{
		Args: args,
	})
}

// SendTaskWithKwargs is a convenience method to send a task with keyword arguments
func (c *Client) SendTaskWithKwargs(ctx context.Context, taskName string, kwargs map[string]interface{}) (string, error) {
	return c.SendTask(ctx, taskName, &TaskOptions{
		Kwargs: kwargs,
	})
}

// SendTaskToQueue sends a task to a specific queue
func (c *Client) SendTaskToQueue(ctx context.Context, taskName, queue string, args []interface{}, kwargs map[string]interface{}) (string, error) {
	return c.SendTask(ctx, taskName, &TaskOptions{
		Queue:  queue,
		Args:   args,
		Kwargs: kwargs,
	})
}

// Close closes the client and its broker connection
func (c *Client) Close() error {
	if c.broker != nil {
		return c.broker.Close()
	}
	return nil
}
