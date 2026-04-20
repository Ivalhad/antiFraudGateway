package fraud

import (
	"sync"

	"github.com/redis/go-redis/v9"
)

func Evaluate(payload DevicePayload, rdb *redis.Client) []RuleResult {
	ch := make(chan RuleResult, 3)
	var wg sync.WaitGroup

	wg.Add(3)

	go func() {
		defer wg.Done()
		ch <- CheckBlacklist(payload, rdb)
	}()

	go func() {
		defer wg.Done()
		ch <- CheckMockGPS(payload)
	}()

	go func() {
		defer wg.Done()
		ch <- CheckTimestamp(payload)
	}()

	go func() {
		wg.Wait()
		close(ch)
	}()

	results := make([]RuleResult, 0, 3)
	for r := range ch {
		results = append(results, r)
	}

	return results
}

func HasFraud(results []RuleResult) bool {
	for _, r := range results {
		if r.IsFraud {
			return true
		}
	}
	return false
}

func GetFraudResults(results []RuleResult) []RuleResult {
	frauds := make([]RuleResult, 0)
	for _, r := range results {
		if r.IsFraud {
			frauds = append(frauds, r)
		}
	}
	return frauds
}
