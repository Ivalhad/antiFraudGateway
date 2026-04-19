package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: File .env tidak ditemukan, menggunakan environment variable bawaan OS")
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Gagal terhubung ke Redis: %v", err)
	}
	fmt.Println("✅ Berhasil terhubung ke Redis!")

	app := fiber.New()

	app.Post("/api/v1/ingest", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": "Gateway aktif. Request diterima.",
		})
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("Sistem Anti Fraud berjalan normal")
	})
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	fmt.Printf("Menjalankan server di port %s...\n", port)
	log.Fatal(app.Listen(":" + port))
}
