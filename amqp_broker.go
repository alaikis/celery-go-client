package celery

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// AMQPBroker implements Broker interface for RabbitMQ/AMQP
type AMQPBroker struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	exchange string
	queue    string
}

// AMQPBrokerConfig contains configuration for AMQP broker
type AMQPBrokerConfig struct {
	URL      string // AMQP connection URL (e.g., "amqp://guest:guest@localhost:5672/")
	Exchange string // Exchange name (default: "celery")
	Queue    string // Queue name (default: "celery")
}

// NewAMQPBroker creates a new AMQP broker instance
func NewAMQPBroker(config AMQPBrokerConfig) (*AMQPBroker, error) {
	if config.Exchange == "" {
		config.Exchange = "celery"
	}
	if config.Queue == "" {
		config.Queue = "celery"
	}

	// Connect to RabbitMQ
	conn, err := amqp.Dial(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	// Create channel
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	// Declare exchange
	err = channel.ExchangeDeclare(
		config.Exchange, // name
		"direct",        // type
		true,            // durable
		false,           // auto-deleted
		false,           // internal
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare queue
	_, err = channel.QueueDeclare(
		config.Queue, // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange
	err = channel.QueueBind(
		config.Queue,    // queue name
		config.Queue,    // routing key
		config.Exchange, // exchange
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	return &AMQPBroker{
		conn:     conn,
		channel:  channel,
		exchange: config.Exchange,
		queue:    config.Queue,
	}, nil
}

// SendTask sends a task message to RabbitMQ
func (ab *AMQPBroker) SendTask(ctx context.Context, message *CeleryMessage) error {
	// Encode the message to JSON
	data, err := message.Encode()
	if err != nil {
		return fmt.Errorf("failed to encode message: %w", err)
	}

	// Determine routing key
	routingKey := ab.queue
	if message.Properties.DeliveryInfo.RoutingKey != "" {
		routingKey = message.Properties.DeliveryInfo.RoutingKey
	}

	// Determine exchange
	exchange := ab.exchange
	if message.Properties.DeliveryInfo.Exchange != "" {
		exchange = message.Properties.DeliveryInfo.Exchange
	}

	// Publish message
	err = ab.channel.PublishWithContext(
		ctx,
		exchange,   // exchange
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         data,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

// Close closes the AMQP connection
func (ab *AMQPBroker) Close() error {
	if ab.channel != nil {
		ab.channel.Close()
	}
	if ab.conn != nil {
		return ab.conn.Close()
	}
	return nil
}
