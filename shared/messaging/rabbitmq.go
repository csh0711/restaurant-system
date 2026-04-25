package messaging

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func Connect(url string) (*amqp.Connection, *amqp.Channel, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, nil, fmt.Errorf("connect: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, fmt.Errorf("channel: %w", err)
	}

	return conn, ch, nil
}

func DeclareQueue(ch *amqp.Channel, name string) error {
	_, err := ch.QueueDeclare(name, true, false, false, false, nil)
	return err
}

func PublishJSON(ch *amqp.Channel, queue string, v interface{}) error {
	body, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	return ch.PublishWithContext(
		context.Background(),
		"",
		queue,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func Consume(ch *amqp.Channel, queue string) (<-chan amqp.Delivery, error) {
	return ch.Consume(queue, "", false, false, false, false, nil)
}
