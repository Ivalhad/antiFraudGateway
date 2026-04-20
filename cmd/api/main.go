package main

import (
	"context"
	"log"
	"os"
	"time"

	"antiFraudGateway/pkg/crypto"
	"antiFraudGateway/pkg/middleware"

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

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		log.Fatalf("FATAL: Gagal terhubung ke Redis: %v", err)
	}
	log.Println("INFO: Koneksi Redis berhasil.")

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Selamat datang di API Gateway")
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("Sistem Anti-Fraud berjalan normal")
	})

	apiKey := os.Getenv("API_KEY")
	api := app.Group("/api/v1",
		middleware.RateLimiter(redisClient, 10, 1*time.Second),
		middleware.APIKeyAuth(apiKey),
	)

	api.Get("/testEncrypt", func(c *fiber.Ctx) error {
		secretKey := os.Getenv("AES_SECRET_KEY")
		dummyData := `{"device_id": "DEV-999", "latitude": -5.3971, "longitude": 105.2668, "is_mock_location": false}`

		encryptedStr, err := crypto.EncryptAESGCM(dummyData, secretKey)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Gagal mengenkripsi: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{"payload": encryptedStr})
	})

	api.Post("/ingest", func(c *fiber.Ctx) error {
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
	log.Printf("INFO: Server berjalan di port %s\n", port)
	log.Fatal(app.Listen(":" + port))
}
