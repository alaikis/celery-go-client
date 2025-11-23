package celery

import (
	"encoding/base64"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// TaskMessage represents the Celery task message (protocol v1)
// This is the inner message that gets base64 encoded in the body
type TaskMessage struct {
	ID      string                 `json:"id"`
	Task    string                 `json:"task"`
	Args    []interface{}          `json:"args,omitempty"`
	Kwargs  map[string]interface{} `json:"kwargs,omitempty"`
	Retries int                    `json:"retries,omitempty"`
	ETA     *string                `json:"eta,omitempty"`
	Expires *time.Time             `json:"expires,omitempty"`
}

// CeleryMessage is the outer message envelope sent to the broker
type CeleryMessage struct {
	Body            string                 `json:"body"`
	Headers         map[string]interface{} `json:"headers,omitempty"`
	ContentType     string                 `json:"content-type"`
	Properties      CeleryProperties       `json:"properties"`
	ContentEncoding string                 `json:"content-encoding"`
}

// CeleryProperties contains message properties
type CeleryProperties struct {
	BodyEncoding  string             `json:"body_encoding"`
	CorrelationID string             `json:"correlation_id"`
	ReplyTo       string             `json:"reply_to"`
	DeliveryInfo  CeleryDeliveryInfo `json:"delivery_info"`
	DeliveryMode  int                `json:"delivery_mode"`
	DeliveryTag   string             `json:"delivery_tag"`
}

// CeleryDeliveryInfo contains routing information
type CeleryDeliveryInfo struct {
	Priority   int    `json:"priority"`
	RoutingKey string `json:"routing_key"`
	Exchange   string `json:"exchange"`
}

// NewTaskMessage creates a new task message with default values
func NewTaskMessage(taskName string, args []interface{}, kwargs map[string]interface{}) *TaskMessage {
	if args == nil {
		args = make([]interface{}, 0)
	}
	if kwargs == nil {
		kwargs = make(map[string]interface{})
	}

	return &TaskMessage{
		ID:      uuid.New().String(),
		Task:    taskName,
		Args:    args,
		Kwargs:  kwargs,
		Retries: 0,
	}
}

// SetETA sets the estimated time of arrival for the task
func (tm *TaskMessage) SetETA(eta time.Time) {
	etaStr := eta.Format(time.RFC3339)
	tm.ETA = &etaStr
}

// SetExpires sets the expiration time for the task
func (tm *TaskMessage) SetExpires(expires time.Time) {
	tm.Expires = &expires
}

// Encode serializes the task message to base64-encoded JSON
func (tm *TaskMessage) Encode() (string, error) {
	jsonData, err := json.Marshal(tm)
	if err != nil {
		return "", err
	}
	encodedData := base64.StdEncoding.EncodeToString(jsonData)
	return encodedData, nil
}

// EncodeJSON serializes the task message to raw JSON string
func (tm *TaskMessage) EncodeJSON() (string, error) {
	jsonData, err := json.Marshal(tm)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// NewCeleryMessage creates a new Celery message envelope
func NewCeleryMessage(encodedBody, queue, exchange string) *CeleryMessage {
	return NewCeleryMessageWithEncoding(encodedBody, queue, exchange, "application/json", "base64", "utf-8")
}

// NewCeleryMessageWithEncoding creates a new Celery message envelope with custom encoding
func NewCeleryMessageWithEncoding(body, queue, exchange, contentType, bodyEncoding, contentEncoding string) *CeleryMessage {
	if queue == "" {
		queue = "celery"
	}
	if exchange == "" {
		exchange = "celery"
	}

	return &CeleryMessage{
		Body:        body,
		ContentType: contentType,
		Properties: CeleryProperties{
			BodyEncoding:  bodyEncoding,
			CorrelationID: uuid.New().String(),
			ReplyTo:       uuid.New().String(),
			DeliveryInfo: CeleryDeliveryInfo{
				Priority:   0,
				RoutingKey: queue,
				Exchange:   exchange,
			},
			DeliveryMode: 2, // persistent
			DeliveryTag:  uuid.New().String(),
		},
		ContentEncoding: contentEncoding,
	}
}

// Encode serializes the Celery message to JSON
func (cm *CeleryMessage) Encode() ([]byte, error) {
	return json.Marshal(cm)
}



