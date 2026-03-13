package messaging

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

const exchange = "symblon"

type RabbitMQService struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

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

	// Declare the topic exchange once on startup.  Both the ingestor (publisher)
	// and paquetier (consumer) call this constructor, so the declaration is
	// idempotent – RabbitMQ silently accepts a re-declaration with identical
	// parameters.
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

	return &RabbitMQService{
		connection: conn,
		channel:    ch,
	}, nil
}

func (r *RabbitMQService) Publish(message string) error {
	return r.PublishTo("activity.github", message)
}

// PublishTo sends message to the shared exchange with the given routing key.
func (r *RabbitMQService) PublishTo(routingKey, message string) error {
	return r.channel.Publish(exchange, routingKey, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        []byte(message),
	})
}

// BindQueue declares a durable queue and binds it to the shared exchange with
// the given routing key pattern (AMQP topic syntax, e.g. "activity.#").
// It must be called before Subscribe when setting up a consumer.
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

func (r *RabbitMQService) Subscribe(queueName string, handler func(string)) error {
	msgs, err := r.channel.Consume(
		queueName,
		"",    // consumer tag
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
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
