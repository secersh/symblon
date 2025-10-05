package messaging

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

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

	return &RabbitMQService{
		connection: conn,
		channel:    ch,
	}, nil
}

func (r *RabbitMQService) Publish(message string) error {
	return r.channel.Publish("symblon", "activity.github", false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        []byte(message),
	})
}

func (r *RabbitMQService) Subscribe(queueName string, handler func(string)) error {
	return nil
}
