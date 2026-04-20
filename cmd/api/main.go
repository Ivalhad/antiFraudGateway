package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"antiFraudGateway/pkg/crypto"
	"antiFraudGateway/pkg/fraud"
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
		timestamp := time.Now().Unix()
		dummyData := fmt.Sprintf(`{"device_id":"DEV-999","latitude":-5.3971,"longitude":105.2668,"is_mock_location":false,"timestamp":%d}`, timestamp)

		encryptedStr, err := crypto.EncryptAESGCM(dummyData, secretKey)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Gagal mengenkripsi: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"payload":   encryptedStr,
			"raw_debug": dummyData,
		})
	})

	api.Post("/testEncrypt", func(c *fiber.Ctx) error {
		secretKey := os.Getenv("AES_SECRET_KEY")

		rawBody := string(c.Body())
		if rawBody == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Kirim JSON body yang ingin dienkripsi",
			})
		}

		encryptedStr, err := crypto.EncryptAESGCM(rawBody, secretKey)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Gagal mengenkripsi: " + err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"payload":   encryptedStr,
			"raw_debug": rawBody,
		})
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

		var payload fraud.DevicePayload
		if err := json.Unmarshal([]byte(decryptedData), &payload); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Data terdekripsi tapi format JSON payload tidak valid",
			})
		}

		results := fraud.Evaluate(payload, redisClient)

		if fraud.HasFraud(results) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"status":     "fraud_detected",
				"message":    "Request ditolak karena terdeteksi aktivitas mencurigakan",
				"violations": fraud.GetFraudResults(results),
				"all_checks": results,
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":     "clean",
			"message":    "Data lolos semua pengecekan fraud",
			"payload":    payload,
			"all_checks": results,
		})
	})

	api.Post("/blacklist", func(c *fiber.Ctx) error {
		var body struct {
			DeviceID string `json:"device_id"`
		}
		if err := c.BodyParser(&body); err != nil || body.DeviceID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Kirim device_id yang ingin di-blacklist",
			})
		}

		redisClient.SAdd(ctx, "blacklist:devices", body.DeviceID)

		return c.JSON(fiber.Map{
			"status":  "success",
			"message": fmt.Sprintf("Device '%s' berhasil ditambahkan ke blacklist", body.DeviceID),
		})
	})

	api.Delete("/blacklist", func(c *fiber.Ctx) error {
		var body struct {
			DeviceID string `json:"device_id"`
		}
		if err := c.BodyParser(&body); err != nil || body.DeviceID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Kirim device_id yang ingin dihapus dari blacklist",
			})
		}

		redisClient.SRem(ctx, "blacklist:devices", body.DeviceID)

		return c.JSON(fiber.Map{
			"status":  "success",
			"message": fmt.Sprintf("Device '%s' berhasil dihapus dari blacklist", body.DeviceID),
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Printf("INFO: Server berjalan di port %s\n", port)
	log.Fatal(app.Listen(":" + port))
}
