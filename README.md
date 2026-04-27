# 🧾 Restaurant Ordering System

This project implements a simple **event-driven ordering system** for a restaurant using [Go](https://go.dev/) and [RabbitMQ](https://www.rabbitmq.com/).

The system models the full **order-to-table flow**:
1. Guests place an order
2. Kitchen prepares the dishes
3. Runner serves the order

---

# 🧩 Architecture

The system consists of three services:

- **Order Service**
    - Exposes HTTP API
    - Validates orders
    - Publishes `OrderPlaced` events

- **Kitchen Service**
    - Consumes orders
    - Simulates cooking
    - Publishes `DishPrepared` events

- **Runner Service**
    - Consumes prepared dishes
    - Simulates delivery
    - Publishes `OrderServed` events

Communication between services is handled via **RabbitMQ**.

---

# ⚙️ Tech Stack

- [Go](https://go.dev/)
- [RabbitMQ](https://www.rabbitmq.com/)
- [OpenAPI (oapi-codegen)](https://github.com/oapi-codegen/oapi-codegen)
- [Event-driven architecture](https://en.wikipedia.org/wiki/Event-driven_architecture)

---

# 🚀 Running the System

## 1. Start RabbitMQ

```bash
docker-compose up
```

RabbitMQ UI:
http://localhost:15672  
user: guest  
pass: guest

---

## 2. Start Services

Run each service in a separate terminal:

### Order Service

```bash
cd services/order-service
go run main.go
```

### Kitchen Service

```bash
cd services/kitchen-service
go run main.go
```

### Runner Service

```bash
cd services/runner-service
go run main.go
```

---

## 3. Get Menu

```bash
curl http://localhost:8080/menu
```

---

## 4. Create an Order

```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "tableId": "1",
    "items": ["Caesar Salad"]
  }'
```

---

# 🧠 Design Decisions

### Event-Driven Architecture
Services are loosely coupled via RabbitMQ to allow asynchronous processing and scalability.

### Shared Domain Types
Common types (e.g. `MenuItem`, `OrderID`) are defined in a shared module to ensure consistency across services.

### Input Validation
Menu items are basically validated against a predefined set to avoid invalid input.

### Failure Handling
- Manual message acknowledgements
- Retry mechanism
- Dead Letter Queue (DLQ)

### Graceful Shutdown
All services handle shutdown signals to ensure clean termination.

### Simplicity First
The system intentionally avoids overengineering and focuses on clarity and correctness.

---

# 💬 Notes

- Docker is used only for RabbitMQ to keep the setup simple and environment-independent.
- Services are run locally for easier debugging and development.

---

# 🎯 Future Improvements

- Address unhandled errors
- Add comprehensive unit and integration tests using Mocks and/or Testcontainers
- Event versioning
- Observability (tracing / metrics)
- Persistent storage
- Authentication / authorization 
