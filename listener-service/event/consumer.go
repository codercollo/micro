package event

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	conn      *amqp.Connection
	queueName string
}

// NewConsumer initializes a new event consumer
func NewConsumer(conn *amqp.Connection) (Consumer, error) {
	consumer := Consumer{
		conn: conn,
	}

	err := consumer.setup()
	if err != nil {
		return Consumer{}, err
	}

	return consumer, nil

}

// setup prepares the RabbitMQ channel and exchange
func (consumer *Consumer) setup() error {
	channel, err := consumer.conn.Channel()
	if err != nil {
		return err
	}

	return declareExchange(channel)
}

type Payload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

// Listen starts consuming messages from  bound topics
func (consumer *Consumer) Listen(topics []string) error {
	ch, err := consumer.conn.Channel()
	if err != nil {
		return err
	}

	defer ch.Close()

	q, err := declareRandomQueue(ch)
	if err != nil {
		return err
	}

	// Bind the random queue to each topic in the exchange
	for _, s := range topics {
		err = ch.QueueBind(
			q.Name,
			s,
			"logs_topic",
			false,
			nil,
		)

		if err != nil {
			return err
		}
	}

	messages, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	forever := make(chan bool)

	// Consume incoming messages concurrently
	go func() {
		for d := range messages {
			var payload Payload

			_ = json.Unmarshal(d.Body, &payload)

			go handlePayload(payload)

		}
	}()

	log.Println("Listening for and consuming RabbitMQ messages")
	fmt.Printf("Waiting  for message [Exchange, Queue] [logs_topic, %s]\n", q.Name)
	<-forever

	return nil
}

// handlePayload routes incoming events to the appropriate handler
func handlePayload(payload Payload) {
	switch payload.Name {
	case "log", "events":
		err := logEvent(payload)
		if err != nil {
			log.Println(err)
		}

	case "auth":

	default:
		err := logEvent(payload)
		if err != nil {
			log.Println(err)
		}
	}
}

// logEvent sends the payload to the logger microservice via HTTP POST
func logEvent(entry Payload) error {
	jsonData, _ := json.MarshalIndent(entry, "", "\t")

	logServiceURL := "http://logger-service/log"

	request, err := http.NewRequest("POST", logServiceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	// Return an error if the logger service did not accept the request
	if response.StatusCode != http.StatusAccepted {
		return fmt.Errorf("logger service returned non-accepted status: %d", response.StatusCode)
	}

	return nil
}
