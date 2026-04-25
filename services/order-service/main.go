package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"order-service/api"
	"restaurant-system/messaging"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Server struct {
	channel   *amqp.Channel
	queueName string
}

func (s *Server) PostOrders(w http.ResponseWriter, r *http.Request) {
	var req api.PostOrdersJSONRequestBody

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if req.TableId == "" || len(req.Items) == 0 {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}

	event := map[string]interface{}{
		"orderId": uuid.New().String(),
		"tableId": req.TableId,
		"items":   req.Items,
	}
	
	if err := messaging.PublishJSON(s.channel, s.queueName, event); err != nil {
		http.Error(w, "failed to publish event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(event)
}

func main() {
	rabbitURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")

	conn, ch, err := messaging.Connect(rabbitURL)
	if err != nil {
		log.Fatalf("rabbitmq connect failed: %v", err)
	}
	defer conn.Close()
	defer ch.Close()

	queueName := "orders"

	if err := messaging.DeclareQueue(ch, "orders"); err != nil {
		log.Fatalf("queue declare failed: %v", err)
	}

	server := &Server{
		channel:   ch,
		queueName: queueName,
	}

	router := chi.NewRouter()

	handler := api.HandlerFromMux(server, router)

	log.Println("Order service running on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
