package fraud

type DevicePayload struct {
	DeviceID       string  `json:"device_id"`
	Latitude       float64 `json:"latitude"`
	Longitude      float64 `json:"longitude"`
	IsMockLocation bool    `json:"is_mock_location"`
	Timestamp      int64   `json:"timestamp"`
}

type RuleResult struct {
	RuleName string `json:"rule_name"`
	IsFraud  bool   `json:"is_fraud"`
	Reason   string `json:"reason,omitempty"`
}
