package fraud

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func CheckBlacklist(payload DevicePayload, rdb *redis.Client) RuleResult {
	result := RuleResult{RuleName: "device_blacklist"}

	isMember, err := rdb.SIsMember(ctx, "blacklist:devices", payload.DeviceID).Result()
	if err != nil {
		return result
	}

	if isMember {
		result.IsFraud = true
		result.Reason = fmt.Sprintf("Device '%s' terdaftar dalam blacklist", payload.DeviceID)
	}

	return result
}

func CheckMockGPS(payload DevicePayload) RuleResult {
	result := RuleResult{RuleName: "mock_gps"}

	if payload.IsMockLocation {
		result.IsFraud = true
		result.Reason = "Terdeteksi penggunaan lokasi palsu (Mock GPS)"
	}

	return result
}

func CheckTimestamp(payload DevicePayload) RuleResult {
	result := RuleResult{RuleName: "abnormal_timestamp"}

	now := time.Now().Unix()
	diff := now - payload.Timestamp

	if diff > 300 {
		result.IsFraud = true
		result.Reason = fmt.Sprintf("Timestamp terlalu lama (%d detik lalu), potensi replay attack", diff)
		return result
	}

	if diff < -30 {
		result.IsFraud = true
		result.Reason = "Timestamp dari masa depan, potensi manipulasi waktu perangkat"
	}

	return result
}
