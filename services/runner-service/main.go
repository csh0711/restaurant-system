package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"restaurant-system/shared/events"
	"restaurant-system/shared/infrastructure"
	"syscall"
	"time"
)

const preparedQueue = "prepared"
const servedQueue = "served"

func main() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	rabbitURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")

	conn, ch, err := infrastructure.Connect(rabbitURL)
	if err != nil {
		log.Fatalf("rabbitmq connect failed: %v", err)
	}
	// Limit consumer to process one message at a time to avoid overload
	if err := ch.Qos(1, 0, false); err != nil {
		log.Fatalf("failed to set qos: %v", err)
	}
	defer conn.Close()
	defer ch.Close()

	if err := infrastructure.DeclareQueue(ch, preparedQueue); err != nil {
		log.Fatalf("declare prepared queue failed: %v", err)
	}

	if err := infrastructure.DeclareQueue(ch, servedQueue); err != nil {
		log.Fatalf("declare served queue failed: %v", err)
	}

	msgs, err := infrastructure.Consume(ch, preparedQueue)
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

			var prepared events.DishPreparedEvent

			if err := json.Unmarshal(msg.Body, &prepared); err != nil {
				log.Printf("invalid message: %v", err)
				msg.Nack(false, false)
				continue
			}

			log.Printf("[runner] serving order=%s table=%s", prepared.OrderID, prepared.TableID)

			serveTime := time.Duration(r.Intn(2)+1) * time.Second
			time.Sleep(serveTime)

			event := events.OrderServedEvent{
				OrderID: prepared.OrderID,
				TableID: prepared.TableID,
				Items:   prepared.Items,
				Status:  "served",
			}

			if err := infrastructure.PublishJSON(ch, servedQueue, event); err != nil {
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
