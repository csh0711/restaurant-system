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

	amqp "github.com/rabbitmq/amqp091-go"
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

const orderQueue = "orders"
const preparedQueue = "prepared"
const deadLetterQueue = "orders.dlq"

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

	if err := messaging.DeclareQueue(ch, orderQueue); err != nil {
		log.Fatalf("declare orders queue failed: %v", err)
	}
	if err := messaging.DeclareQueue(ch, preparedQueue); err != nil {
		log.Fatalf("declare prepared queue failed: %v", err)
	}
	if err := messaging.DeclareQueue(ch, deadLetterQueue); err != nil {
		log.Fatalf("declare dlq failed: %v", err)
	}

	msgs, err := messaging.Consume(ch, orderQueue)
	if err != nil {
		log.Fatalf("consume failed: %v", err)
	}

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Println("Kitchen service waiting for orders...")

	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down kitchen service...")
			return

		case msg, ok := <-msgs:
			if !ok {
				log.Println("Message channel closed")
				return
			}

			var order OrderEvent

			if err := json.Unmarshal(msg.Body, &order); err != nil {
				log.Printf("invalid message: %v", err)
				msg.Nack(false, false)
				continue
			}

			log.Printf("[kitchen] received order=%s items=%v", order.OrderID, order.Items)

			// Random cooking time
			cookTime := time.Duration(r.Intn(3)+1) * time.Second
			log.Printf("[kitchen] cooking order=%s duration=%v", order.OrderID, cookTime)
			time.Sleep(cookTime)

			event := DishPreparedEvent{
				OrderID: order.OrderID,
				TableID: order.TableID,
				Items:   order.Items,
				Status:  "prepared",
			}

			if err := messaging.PublishJSON(ch, preparedQueue, event); err != nil {
				log.Printf("failed to publish: %v", err)

				headers := msg.Headers
				if headers == nil {
					headers = make(map[string]interface{})
				}

				retryCount := 0
				if val, ok := headers["x-retry-count"]; ok {
					retryCount = int(val.(int32))
				}

				if retryCount < 1 {
					log.Printf("[kitchen] retrying order=%s", order.OrderID)

					headers["x-retry-count"] = retryCount + 1

					err := ch.Publish(
						"",
						orderQueue,
						false,
						false,
						amqp.Publishing{
							Headers:     headers,
							ContentType: "application/json",
							Body:        msg.Body,
						},
					)

					if err != nil {
						log.Printf("retry publish failed: %v", err)
						msg.Nack(false, true)
						continue
					}

					msg.Ack(false)

				} else {
					log.Printf("[kitchen] sending to DLQ order=%s", order.OrderID)

					if err := messaging.PublishJSON(ch, deadLetterQueue, order); err != nil {
						log.Printf("dlq publish failed: %v", err)
						msg.Nack(false, true)
						continue
					}

					msg.Ack(false)
				}

				continue
			}

			log.Printf("[kitchen] prepared order=%s", order.OrderID)

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
