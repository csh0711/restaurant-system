package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"restaurant-system/shared/events"
	"restaurant-system/shared/infrastructure"
	"restaurant-system/shared/model"
	"syscall"
	"time"

	"restaurant-system/order-service/api"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Server struct {
	channel   *amqp.Channel
	queueName string
}

const queueName = "orders"

func (s *Server) PostOrders(w http.ResponseWriter, r *http.Request) {
	var req api.PostOrdersJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	log.Printf("[order] received table=%s items=%v", req.TableId, req.Items)

	if req.TableId == "" || len(req.Items) == 0 {
		http.Error(w, "missing fields", http.StatusBadRequest)
		return
	}

	menuItems := make([]model.MenuItem, 0, len(req.Items))
	for _, item := range req.Items {
		menuItem := model.MenuItem(item)
		if !model.IsValid(menuItem) {
			http.Error(w, "invalid menu item: "+string(item), http.StatusBadRequest)
			return
		}
		menuItems = append(menuItems, menuItem)
	}

	event := events.OrderEvent{
		OrderID: model.OrderID(uuid.New()),
		TableID: model.TableID(req.TableId),
		Items:   menuItems,
	}

	log.Printf("[order] publishing orderId=%s", event.OrderID)
	if err := infrastructure.PublishJSON(s.channel, s.queueName, event); err != nil {
		http.Error(w, "failed to publish event", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w)
}

func (s *Server) GetMenu(w http.ResponseWriter, _ *http.Request) {
	items := model.List()

	resp := make([]string, 0, len(items))
	for _, item := range items {
		resp = append(resp, string(item))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func main() {
	rabbitURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")

	conn, ch, err := infrastructure.Connect(rabbitURL)
	if err != nil {
		log.Fatalf("rabbitmq connect failed: %v", err)
	}
	defer conn.Close()
	defer ch.Close()

	if err := infrastructure.DeclareQueue(ch, queueName); err != nil {
		log.Fatalf("queue declare failed: %v", err)
	}

	server := &Server{
		channel:   ch,
		queueName: queueName,
	}

	router := chi.NewRouter()

	handler := api.HandlerFromMux(server, router)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	go func() {
		log.Println("Order service running on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	// Shutdown handling
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Println("Shutting down order service...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}

func getEnv(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
