package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
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

func main() {
	rand.Seed(time.Now().UnixNano())

	rabbitURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")

	conn, ch, err := messaging.Connect(rabbitURL)
	if err != nil {
		log.Fatalf("rabbitmq connect failed: %v", err)
	}
	defer conn.Close()
	defer ch.Close()

	preparedQueue := "prepared"
	servedQueue := "served"

	messaging.DeclareQueue(ch, "prepared")
	messaging.DeclareQueue(ch, "served")

	msgs, err := messaging.Consume(ch, "prepared")
	if err != nil {
		log.Fatalf("consume failed: %v", err)
	}

	log.Println("Runner service waiting for prepared dishes...")

	for msg := range msgs {
		var prepared DishPreparedEvent

		if err := json.Unmarshal(msg.Body, &prepared); err != nil {
			log.Printf("invalid message: %v", err)
			msg.Nack(false, false) // drop invalid
			continue
		}

		log.Printf("Serving order %s for table %s", prepared.OrderID, prepared.TableID)

		// Simulate delivery time
		serveTime := time.Duration(rand.Intn(2)+1) * time.Second
		time.Sleep(serveTime)

		event := OrderServedEvent{
			OrderID: prepared.OrderID,
			TableID: prepared.TableID,
			Items:   prepared.Items,
			Status:  "served",
		}

		if err := messaging.PublishJSON(ch, "served", event); err != nil {
			log.Printf("failed to publish: %v", err)
			msg.Nack(false, true)
			continue
		}
		
		log.Printf("Order %s served!", prepared.OrderID)

		msg.Ack(false)
	}
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
