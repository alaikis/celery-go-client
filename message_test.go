package celery

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"
)

func TestNewTaskMessage(t *testing.T) {
	taskName := "test.task"
	args := []interface{}{1, 2, 3}
	kwargs := map[string]interface{}{"key": "value"}

	msg := NewTaskMessage(taskName, args, kwargs)

	if msg.Task != taskName {
		t.Errorf("Expected task name %s, got %s", taskName, msg.Task)
	}

	if len(msg.Args) != 3 {
		t.Errorf("Expected 3 args, got %d", len(msg.Args))
	}

	if msg.Kwargs["key"] != "value" {
		t.Errorf("Expected kwargs key=value, got %v", msg.Kwargs["key"])
	}

	if msg.ID == "" {
		t.Error("Expected non-empty task ID")
	}

	if msg.Retries != 0 {
		t.Errorf("Expected retries=0, got %d", msg.Retries)
	}
}

func TestNewTaskMessageWithNilArgs(t *testing.T) {
	msg := NewTaskMessage("test.task", nil, nil)

	if msg.Args == nil {
		t.Error("Expected non-nil args")
	}

	if msg.Kwargs == nil {
		t.Error("Expected non-nil kwargs")
	}

	if len(msg.Args) != 0 {
		t.Errorf("Expected empty args, got %d items", len(msg.Args))
	}

	if len(msg.Kwargs) != 0 {
		t.Errorf("Expected empty kwargs, got %d items", len(msg.Kwargs))
	}
}

func TestTaskMessageSetETA(t *testing.T) {
	msg := NewTaskMessage("test.task", nil, nil)
	eta := time.Now().Add(1 * time.Hour)

	msg.SetETA(eta)

	if msg.ETA == nil {
		t.Error("Expected non-nil ETA")
	}

	// Parse the ETA string back to time
	parsedETA, err := time.Parse(time.RFC3339, *msg.ETA)
	if err != nil {
		t.Errorf("Failed to parse ETA: %v", err)
	}

	// Compare with some tolerance (1 second)
	diff := parsedETA.Sub(eta)
	if diff > time.Second || diff < -time.Second {
		t.Errorf("ETA mismatch: expected %v, got %v", eta, parsedETA)
	}
}

func TestTaskMessageSetExpires(t *testing.T) {
	msg := NewTaskMessage("test.task", nil, nil)
	expires := time.Now().Add(2 * time.Hour)

	msg.SetExpires(expires)

	if msg.Expires == nil {
		t.Error("Expected non-nil Expires")
	}

	if !msg.Expires.Equal(expires) {
		t.Errorf("Expires mismatch: expected %v, got %v", expires, msg.Expires)
	}
}

func TestTaskMessageEncode(t *testing.T) {
	msg := NewTaskMessage("test.task", []interface{}{1, 2}, map[string]interface{}{"key": "value"})

	encoded, err := msg.Encode()
	if err != nil {
		t.Fatalf("Failed to encode message: %v", err)
	}

	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("Failed to decode base64: %v", err)
	}

	// Unmarshal JSON
	var result TaskMessage
	if err := json.Unmarshal(decoded, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if result.Task != "test.task" {
		t.Errorf("Expected task name test.task, got %s", result.Task)
	}

	if len(result.Args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(result.Args))
	}
}

func TestNewCeleryMessage(t *testing.T) {
	encodedBody := []byte(`"dGVzdA=="`)
	queue := "test_queue"
	exchange := "test_exchange"

	msg := NewCeleryMessage(encodedBody, queue, exchange)

	if string(msg.Body) != string(encodedBody) {
		t.Errorf("Expected body %s, got %s", string(encodedBody), string(msg.Body))
	}

	if msg.ContentType != "application/json" {
		t.Errorf("Expected content-type application/json, got %s", msg.ContentType)
	}

	if msg.ContentEncoding != "utf-8" {
		t.Errorf("Expected content-encoding utf-8, got %s", msg.ContentEncoding)
	}

	if msg.Properties.BodyEncoding != "base64" {
		t.Errorf("Expected body_encoding base64, got %s", msg.Properties.BodyEncoding)
	}

	if msg.Properties.DeliveryMode != 2 {
		t.Errorf("Expected delivery_mode 2, got %d", msg.Properties.DeliveryMode)
	}

	if msg.Properties.DeliveryInfo.RoutingKey != queue {
		t.Errorf("Expected routing_key %s, got %s", queue, msg.Properties.DeliveryInfo.RoutingKey)
	}

	if msg.Properties.DeliveryInfo.Exchange != exchange {
		t.Errorf("Expected exchange %s, got %s", exchange, msg.Properties.DeliveryInfo.Exchange)
	}
}

func TestNewCeleryMessageDefaults(t *testing.T) {
	msg := NewCeleryMessage([]byte(`"test"`), "", "")

	if msg.Properties.DeliveryInfo.RoutingKey != "celery" {
		t.Errorf("Expected default routing_key celery, got %s", msg.Properties.DeliveryInfo.RoutingKey)
	}

	if msg.Properties.DeliveryInfo.Exchange != "celery" {
		t.Errorf("Expected default exchange celery, got %s", msg.Properties.DeliveryInfo.Exchange)
	}
}

func TestCeleryMessageEncode(t *testing.T) {
	msg := NewCeleryMessage([]byte(`"test_body"`), "test_queue", "test_exchange")

	encoded, err := msg.Encode()
	if err != nil {
		t.Fatalf("Failed to encode message: %v", err)
	}

	// Unmarshal to verify structure
	var result CeleryMessage
	if err := json.Unmarshal(encoded, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if string(result.Body) != `"test_body"` {
		t.Errorf("Expected body \"test_body\", got %s", string(result.Body))
	}

	if result.Properties.DeliveryInfo.RoutingKey != "test_queue" {
		t.Errorf("Expected routing_key test_queue, got %s", result.Properties.DeliveryInfo.RoutingKey)
	}
}

func TestFullMessageFlow(t *testing.T) {
	// Create task message
	taskMsg := NewTaskMessage("tasks.add", []interface{}{5, 10}, nil)

	// Encode task message
	encodedTaskStr, err := taskMsg.Encode()
	if err != nil {
		t.Fatalf("Failed to encode task: %v", err)
	}
	encodedTask := []byte(`"` + encodedTaskStr + `"`)

	// Create Celery message
	celeryMsg := NewCeleryMessage(encodedTask, "celery", "celery")

	// Encode Celery message
	encodedCelery, err := celeryMsg.Encode()
	if err != nil {
		t.Fatalf("Failed to encode Celery message: %v", err)
	}

	// Verify it's valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal(encodedCelery, &result); err != nil {
		t.Fatalf("Failed to unmarshal final message: %v", err)
	}

	// Verify structure
	if result["content-type"] != "application/json" {
		t.Error("Missing or incorrect content-type")
	}

	if result["body"] == nil {
		t.Error("Missing body")
	}

	if result["properties"] == nil {
		t.Error("Missing properties")
	}
}
