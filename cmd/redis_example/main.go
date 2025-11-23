package main

import (
	"context"
	"fmt"
	"log"
	"time"

	celery "github.com/alaikis/celery-go-client"
)

func main() {
	// Create Redis broker
	broker, err := celery.NewRedisBroker(celery.RedisBrokerConfig{
		Addr:     "localhost:6379",
		Password: "", // no password
		DB:       0,  // default DB
		Queue:    "celery",
	})
	if err != nil {
		log.Fatalf("Failed to create Redis broker: %v", err)
	}
	defer broker.Close()

	// Create Celery client
	client := celery.NewClient(celery.ClientConfig{
		Broker: broker,
		Queue:  "celery",
	})
	defer client.Close()

	ctx := context.Background()

	// Example 1: Send task with positional arguments
	fmt.Println("Example 1: Sending task with args...")
	taskID1, err := client.SendTaskWithArgs(ctx, "tasks.add", 10, 20)
	if err != nil {
		log.Fatalf("Failed to send task: %v", err)
	}
	fmt.Printf("Task sent successfully! Task ID: %s\n", taskID1)

	// Example 2: Send task with keyword arguments
	fmt.Println("\nExample 2: Sending task with kwargs...")
	taskID2, err := client.SendTaskWithKwargs(ctx, "tasks.process_data", map[string]interface{}{
		"name":   "John Doe",
		"age":    30,
		"active": true,
	})
	if err != nil {
		log.Fatalf("Failed to send task: %v", err)
	}
	fmt.Printf("Task sent successfully! Task ID: %s\n", taskID2)

	// Example 3: Send task with both args and kwargs
	fmt.Println("\nExample 3: Sending task with args and kwargs...")
	taskID3, err := client.SendTask(ctx, "tasks.complex_task", &celery.TaskOptions{
		Args: []interface{}{"arg1", "arg2"},
		Kwargs: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		},
	})
	if err != nil {
		log.Fatalf("Failed to send task: %v", err)
	}
	fmt.Printf("Task sent successfully! Task ID: %s\n", taskID3)

	// Example 4: Send task with ETA (scheduled execution)
	fmt.Println("\nExample 4: Sending task with ETA...")
	eta := time.Now().Add(5 * time.Minute)
	taskID4, err := client.SendTask(ctx, "tasks.scheduled_task", &celery.TaskOptions{
		Args: []interface{}{"scheduled"},
		ETA:  &eta,
	})
	if err != nil {
		log.Fatalf("Failed to send task: %v", err)
	}
	fmt.Printf("Task sent successfully! Task ID: %s (scheduled for %s)\n", taskID4, eta.Format(time.RFC3339))

	// Example 5: Send task to a specific queue
	fmt.Println("\nExample 5: Sending task to specific queue...")
	taskID5, err := client.SendTaskToQueue(ctx, "tasks.priority_task", "high_priority",
		[]interface{}{"urgent"},
		map[string]interface{}{"priority": "high"},
	)
	if err != nil {
		log.Fatalf("Failed to send task: %v", err)
	}
	fmt.Printf("Task sent successfully! Task ID: %s (queue: high_priority)\n", taskID5)

	fmt.Println("\nAll tasks sent successfully!")
}
