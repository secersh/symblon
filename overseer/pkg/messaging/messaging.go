package messaging

// MessagingService is the contract for publish/subscribe operations.
type MessagingService interface {
	// PublishTo publishes message to the shared topic exchange with the given
	// routing key (e.g. "agent.resolved.bug-squasher").
	PublishTo(routingKey, message string) error

	// BindQueue declares a durable queue and binds it to the exchange with the
	// given routing key pattern (AMQP topic syntax, e.g. "activity.#").
	BindQueue(queueName, routingKey string) error

	// Subscribe registers a handler that is called for every message delivered
	// to queueName.  The handler is invoked in a dedicated goroutine.
	Subscribe(queueName string, handler func(message string)) error

	// Close releases the channel and connection.
	Close() error
}
