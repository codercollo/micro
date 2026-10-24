package event

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

// declareExchange sets up a durable topic exchange foe routing log messages
func declareExchange(ch *amqp.Channel) error {
	return ch.ExchangeDeclare(
		"logs_topic", // name?
		"topic",      // type?
		true,         // durable?
		false,        // auto-deleted?
		false,        // internal?
		false,        //  no-wait?
		nil,          // arguments?
	)
}

// declareRandomQueue create an exclusive, server-named temporary queue for this listener instance
func declareRandomQueue(ch *amqp.Channel) (amqp.Queue, error) {
	return ch.QueueDeclare(
		"",    //name?
		false, //durable?
		false, // delete when unused?
		true,  // exclusive?
		false, // no-wait?
		nil,   // arguments?
	)
}
