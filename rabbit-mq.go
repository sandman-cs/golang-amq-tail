package main

import (
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

var (
	srcRabbitConn       *amqp.Connection
	srcRabbitCloseError chan *amqp.Error
	pubError            int
	pubSuccess          int
	recCount            int
)

// Try to connect to the RabbitMQ server as
// long as it takes to establish a connection
//
func connectToRabbitMQ(uri string) *amqp.Connection {
	for {
		conn, err := amqp.Dial(uri)

		if err == nil {
			log.Println("Connected to Broker...")
			return conn
		}

		log.Println(err)
		log.Printf("Trying to reconnect to RabbitMQ at %s\n", uri)
		time.Sleep(500 * time.Millisecond)
	}
}

// re-establish the connection to RabbitMQ in case
// the connection has died
//
func srcRabbitConnector(uri string) {
	var srcRabbitErr *amqp.Error

	for {
		srcRabbitErr = <-srcRabbitCloseError
		if srcRabbitErr != nil {
			log.Println("Connecting to Source RabbitMQ")
			srcRabbitConn = connectToRabbitMQ(uri)
			srcRabbitCloseError = make(chan *amqp.Error)
			srcRabbitConn.NotifyClose(srcRabbitCloseError)
		}
	}
}

//OpenChannel Opens communication Channel to Queue
func OpenChannel(conn *amqp.Connection) {

	log.Println("Opening Channel to Broker...")

	//Open channel to broker
	ch, err := conn.Channel()
	if err != nil {
		log.Println("Failed to open a channel: ", err)
		return
	}
	ch.Qos(10, 0, true)
	defer ch.Close()

	//Declare Queue

	log.Println("Declaring Queue")

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when usused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)

	log.Println("Binding to queue: ", q.Name)

	err = ch.QueueBind(
		q.Name,                 // queue name
		conf.SrcRoute,          // routing key
		conf.SrcBrokerExchange, // exchange
		false,
		nil)

	//Consume messages off the queue
	msgs, err := ch.Consume(
		q.Name,          // queue
		conf.ServerName, // consumer
		false,           // auto-ack
		false,           // exclusive
		false,           // no-local
		false,           // no-wait
		nil,             // args
	)
	if err != nil {
		log.Println("Failed to register a consumer")
		return
	}
	log.Println("Rabbit-listener instance started.\n\n ")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			processPayload(d.Body, d.RoutingKey)
			if err != nil {
				fmt.Println(err)
				d.Ack(true)
			} else {
				d.Ack(false)
			}
		}
		log.Println("Error, Lost AMQP Connection.")
		forever <- false
	}()

	<-forever
}
