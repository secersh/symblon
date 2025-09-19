package messaging

type MessagingService interface {
	Publish(message string) error
	Subscribe(queueName string, handler func(message string)) error
}
