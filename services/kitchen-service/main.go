package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"time"

	"restaurant-system/messaging"
)

type OrderEvent struct {
	OrderID string   `json:"orderId"`
	TableID string   `json:"tableId"`
	Items   []string `json:"items"`
}

type DishPreparedEvent struct {
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

	orderQueue := "orders"
	preparedQueue := "prepared"

	if err := messaging.DeclareQueue(ch, "orders"); err != nil {
		log.Fatalf("declare orders queue failed: %v", err)
	}

	if err := messaging.DeclareQueue(ch, "prepared"); err != nil {
		log.Fatalf("declare prepared queue failed: %v", err)
	}
	msgs, err := messaging.Consume(ch, "orders")
	if err != nil {
		log.Fatalf("consume failed: %v", err)
	}

	log.Println("Kitchen service waiting for orders...")

	for msg := range msgs {
		var order OrderEvent

		if err := json.Unmarshal(msg.Body, &order); err != nil {
			log.Printf("invalid message: %v", err)
			msg.Nack(false, false) // drop bad message
			continue
		}

		log.Printf("Received order %s with items %v", order.OrderID, order.Items)

		// Simulate cooking time
		cookTime := time.Duration(rand.Intn(3)+1) * time.Second
		log.Printf("Cooking order %s (takes %v)...", order.OrderID, cookTime)
		time.Sleep(cookTime)

		event := DishPreparedEvent{
			OrderID: order.OrderID,
			TableID: order.TableID,
			Items:   order.Items,
			Status:  "prepared",
		}

		body, err := json.Marshal(event)
		if err != nil {
			log.Printf("failed to serialize: %v", err)
			msg.Nack(false, true)
			continue
		}

		if err := messaging.PublishJSON(ch, "prepared", event); err != nil {
			log.Printf("failed to publish: %v", err)
			msg.Nack(false, true)
			continue
		}

		log.Printf("Order %s prepared!", order.OrderID)

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
