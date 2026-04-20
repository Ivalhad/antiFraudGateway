package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func RateLimiter(rdb *redis.Client, maxReq int, window time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {

		ip := c.IP()
		key := "rate_limit:" + ip

		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {

			c.Locals("redis_error", err.Error())
			return c.Next()
		}

		if count == 1 {
			rdb.Expire(ctx, key, window)
		}

		remaining := int64(maxReq) - count
		if remaining < 0 {
			remaining = 0
		}
		c.Set("X-RateLimit-Limit", fmt.Sprintf("%d", maxReq))
		c.Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		if count > int64(maxReq) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Terlalu banyak request. Coba lagi dalam beberapa saat.",
				"limit": maxReq,
				"IP":    ip,
			})
		}

		return c.Next()
	}
}
