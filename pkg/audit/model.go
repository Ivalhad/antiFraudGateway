package audit

import "time"

type AuditLog struct {
	RequestID  string      `bson:"request_id" json:"request_id"`
	DeviceID   string      `bson:"device_id" json:"device_id"`
	IP         string      `bson:"ip" json:"ip"`
	Endpoint   string      `bson:"endpoint" json:"endpoint"`
	Status     string      `bson:"status" json:"status"`
	Violations []string    `bson:"violations,omitempty" json:"violations,omitempty"`
	Payload    interface{} `bson:"payload,omitempty" json:"payload,omitempty"`
	CreatedAt  time.Time   `bson:"created_at" json:"created_at"`
}
