package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"time"

	"github.com/codercollo/micro/listener-service/event"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	// Initialize connection to RabbitMQ with retry logic
	rabbitConn, err := connect()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	defer rabbitConn.Close()

	log.Println("Connected to RabbitMQ!")

	// Initialize the event consumer using the active connection
	consumer, err := event.NewConsumer(rabbitConn)
	if err != nil {
		panic(err)
	}

	// Start consuming messages for the specified routing keys
	err = consumer.Listen([]string{"log.INFO", "log.WARNING", "log.ERROR"})
	if err != nil {
		log.Println(err)
	}

}

func connect() (*amqp.Connection, error) {
	var counts int64
	var backOff = 1 * time.Second
	var connection *amqp.Connection

	for {

		// Attempt to establish connection to RabbitMQ
		c, err := amqp.Dial("amqp://guest:guest@rabbitmq")
		if err != nil {
			fmt.Println("RabbitMQ not yet ready...")
			counts++
		} else {
			connection = c
			break
		}

		// Terminate if retry count exceeds limits
		if counts > 5 {
			fmt.Println(err)
			return nil, err
		}

		// Apply exponential backoff delay before next retry
		backOff = time.Duration(math.Pow(float64(counts), 2)) * time.Second
		log.Println("backing off ...")
		time.Sleep(backOff)
		continue
	}

	return connection, nil
}
