package audit

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
)

func StartWorkerPool(ctx context.Context, collection *mongo.Collection, logChan <-chan AuditLog, workerCount int) {
	for i := 0; i < workerCount; i++ {
		go func(id int) {
			for entry := range logChan {
				_, err := collection.InsertOne(ctx, entry)
				if err != nil {
					log.Printf("WARN: Worker %d gagal menyimpan audit log: %v", id, err)
				}
			}
		}(i)
	}
	log.Printf("INFO: %d audit worker berhasil dijalankan.", workerCount)
}
