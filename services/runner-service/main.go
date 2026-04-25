package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"restaurant-system/messaging"
)

type DishPreparedEvent struct {
	OrderID string   `json:"orderId"`
	TableID string   `json:"tableId"`
	Items   []string `json:"items"`
	Status  string   `json:"status"`
}

type OrderServedEvent struct {
	OrderID string   `json:"orderId"`
	TableID string   `json:"tableId"`
	Items   []string `json:"items"`
	Status  string   `json:"status"`
}

const preparedQueue = "prepared"
const servedQueue = "served"

func main() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	rabbitURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")

	conn, ch, err := messaging.Connect(rabbitURL)
	if err != nil {
		log.Fatalf("rabbitmq connect failed: %v", err)
	}
	// Limit consumer to process one message at a time to avoid overload
	if err := ch.Qos(1, 0, false); err != nil {
		log.Fatalf("failed to set qos: %v", err)
	}
	defer conn.Close()
	defer ch.Close()

	if err := messaging.DeclareQueue(ch, preparedQueue); err != nil {
		log.Fatalf("declare prepared queue failed: %v", err)
	}

	if err := messaging.DeclareQueue(ch, servedQueue); err != nil {
		log.Fatalf("declare served queue failed: %v", err)
	}

	msgs, err := messaging.Consume(ch, preparedQueue)
	if err != nil {
		log.Fatalf("consume failed: %v", err)
	}

	// graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Println("Runner service waiting for prepared dishes...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down runner service...")
			return

		case msg, ok := <-msgs:
			if !ok {
				log.Println("Message channel closed")
				return
			}

			var prepared DishPreparedEvent

			if err := json.Unmarshal(msg.Body, &prepared); err != nil {
				log.Printf("invalid message: %v", err)
				msg.Nack(false, false)
				continue
			}

			log.Printf("[runner] serving order=%s table=%s", prepared.OrderID, prepared.TableID)

			serveTime := time.Duration(r.Intn(2)+1) * time.Second
			time.Sleep(serveTime)

			event := OrderServedEvent{
				OrderID: prepared.OrderID,
				TableID: prepared.TableID,
				Items:   prepared.Items,
				Status:  "served",
			}

			if err := messaging.PublishJSON(ch, servedQueue, event); err != nil {
				log.Printf("failed to publish: %v", err)
				msg.Nack(false, true)
				continue
			}

			log.Printf("[runner] served order=%s", prepared.OrderID)

			msg.Ack(false)
		}
	}
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
