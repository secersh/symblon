package messaging

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

const exchange = "symblon"

// RabbitMQService implements MessagingService backed by a RabbitMQ topic
// exchange.  Both the trigger consumer and the REST API use this client.
type RabbitMQService struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

// NewRabbitMQService connects to RabbitMQ and idempotently declares the
// shared "symblon" topic exchange.
func NewRabbitMQService(amqpURL string) (*RabbitMQService, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	if err = ch.ExchangeDeclare(
		exchange,
		"topic",
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &RabbitMQService{connection: conn, channel: ch}, nil
}

// PublishTo sends message to the shared exchange with the given routing key.
// Overseer uses this to emit "agent.resolved.<agent_id>" events.
func (r *RabbitMQService) PublishTo(routingKey, message string) error {
	return r.channel.Publish(exchange, routingKey, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        []byte(message),
	})
}

// BindQueue declares a durable queue and binds it to the exchange.
func (r *RabbitMQService) BindQueue(queueName, routingKey string) error {
	q, err := r.channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		return err
	}
	return r.channel.QueueBind(q.Name, routingKey, exchange, false, nil)
}

// Subscribe registers a message handler on queueName in a background goroutine.
func (r *RabbitMQService) Subscribe(queueName string, handler func(string)) error {
	msgs, err := r.channel.Consume(
		queueName,
		"",    // consumer tag
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			handler(string(msg.Body))
		}
	}()

	return nil
}

// Close releases the AMQP channel and connection.
func (r *RabbitMQService) Close() error {
	if r.channel != nil {
		if err := r.channel.Close(); err != nil {
			return err
		}
	}
	if r.connection != nil {
		if err := r.connection.Close(); err != nil {
			return err
		}
	}
	return nil
}
