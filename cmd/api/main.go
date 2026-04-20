package main

import (
	"context"
	"log"
	"os"

	"antiFraudGateway/pkg/crypto"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type IngestRequest struct {
	EncryptedPayload string `json:"payload"`
}

func main() {
	godotenv.Load()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	redisClient.Ping(ctx)

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Selamat datang di API Gateway")
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("Sistem Anti-Fraud berjalan normal")
	})

	app.Get("/api/v1/test-encrypt", func(c *fiber.Ctx) error {
		secretKey := os.Getenv("AES_SECRET_KEY")
		dummyData := `{"device_id": "DEV-999", "latitude": -5.3971, "longitude": 105.2668, "is_mock_location": false}`

		encryptedStr, err := crypto.EncryptAESGCM(dummyData, secretKey)
		if err != nil {
			return c.Status(500).SendString("Gagal mengenkripsi: " + err.Error())
		}

		return c.JSON(fiber.Map{"payload": encryptedStr})
	})

	app.Post("/api/v1/ingest", func(c *fiber.Ctx) error {
		secretKey := os.Getenv("AES_SECRET_KEY")

		var body IngestRequest
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Format JSON tidak valid",
			})
		}

		decryptedData, err := crypto.DecryptAESGCM(body.EncryptedPayload, secretKey)
		if err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Data ditolak. Gagal mendekripsi payload (Potensi Fraud).",
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":   "success",
			"message":  "Data berhasil didekripsi",
			"raw_data": decryptedData,
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Fatal(app.Listen(":" + port))
}
