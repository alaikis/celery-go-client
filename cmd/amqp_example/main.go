package main

import (
	"context"
	"fmt"
	"log"
	"time"

	celery "github.com/alaikis/celery-go-client"
)

func main() {
	// Create AMQP broker (RabbitMQ)
	broker, err := celery.NewAMQPBroker(celery.AMQPBrokerConfig{
		URL:      "amqp://guest:guest@localhost:5672/",
		Exchange: "celery",
		Queue:    "celery",
	})
	if err != nil {
		log.Fatalf("Failed to create AMQP broker: %v", err)
	}
	defer broker.Close()

	// Create Celery client
	// UseRawJSONBody: true 启用原始 JSON 消息体,适用于 RabbitMQ 且 worker 配置为接受 JSON 的情况
	client := celery.NewClient(celery.ClientConfig{
		Broker:         broker,
		Queue:          "celery",
		Exchange:       "celery",
		UseRawJSONBody: true,
	})
	defer client.Close()

	ctx := context.Background()

	// Example 1: Send task with positional arguments
	fmt.Println("Example 1: Sending task with args...")
	taskID1, err := client.SendTaskWithArgs(ctx, "tasks.multiply", 5, 8)
	if err != nil {
		log.Fatalf("Failed to send task: %v", err)
	}
	fmt.Printf("Task sent successfully! Task ID: %s\n", taskID1)

	// Example 2: Send task with keyword arguments
	fmt.Println("\nExample 2: Sending task with kwargs...")
	taskID2, err := client.SendTaskWithKwargs(ctx, "tasks.send_email", map[string]interface{}{
		"to":      "user@example.com",
		"subject": "Hello from Celery Go Client",
		"body":    "This is a test email",
	})
	if err != nil {
		log.Fatalf("Failed to send task: %v", err)
	}
	fmt.Printf("Task sent successfully! Task ID: %s\n", taskID2)

	// Example 3: Send task with expiration
	fmt.Println("\nExample 3: Sending task with expiration...")
	expires := time.Now().Add(10 * time.Minute)
	taskID3, err := client.SendTask(ctx, "tasks.temporary_task", &celery.TaskOptions{
		Args:    []interface{}{"data"},
		Expires: &expires,
	})
	if err != nil {
		log.Fatalf("Failed to send task: %v", err)
	}
	fmt.Printf("Task sent successfully! Task ID: %s (expires at %s)\n", taskID3, expires.Format(time.RFC3339))

	// Example 4: Send task to custom exchange and queue
	fmt.Println("\nExample 4: Sending task to custom exchange and queue...")
	taskID4, err := client.SendTask(ctx, "tasks.custom_task", &celery.TaskOptions{
		Queue:    "custom_queue",
		Exchange: "custom_exchange",
		Args:     []interface{}{"custom"},
	})
	if err != nil {
		log.Fatalf("Failed to send task: %v", err)
	}
	fmt.Printf("Task sent successfully! Task ID: %s\n", taskID4)

	fmt.Println("\nAll tasks sent successfully!")
}
