package main

import (
	"fmt"
	"time"

	"github.com/sandman-cs/core"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

var (
	srcRabbitConn       *amqp.Connection
	srcRabbitCloseError chan *amqp.Error
	dstRabbitConn       *amqp.Connection
	dstRabbitCloseError chan *amqp.Error
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
			core.SendMessage("Connecting to Source RabbitMQ")
			srcRabbitConn = connectToRabbitMQ(uri)
			srcRabbitCloseError = make(chan *amqp.Error)
			srcRabbitConn.NotifyClose(srcRabbitCloseError)
		}
	}
}

// re-establish the connection to RabbitMQ in case
// the connection has died
//
func dstRabbitConnector(uri string) {
	var dstRabbitErr *amqp.Error

	for {
		dstRabbitErr = <-dstRabbitCloseError
		if dstRabbitErr != nil {
			core.SendMessage(fmt.Sprintln("Connecting to Destination RabbitMQ: ", uri))
			dstRabbitConn = connectToRabbitMQ(uri)
			dstRabbitCloseError = make(chan *amqp.Error)
			dstRabbitConn.NotifyClose(dstRabbitCloseError)
			core.SendMessage("Connection to dst established...")
		}
	}
}

//OpenChannel Opens communication Channel to Queue
func OpenChannel(conn *amqp.Connection) {

	//Open channel to broker
	ch, err := conn.Channel()
	core.FailOnError(err, "Failed to open a channel")
	ch.Qos(10, 0, true)
	defer ch.Close()

	//Declare Queue

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when usused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)

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
	fmt.Println("Rabbit-listener instance started.")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			processPayload(d.Body, d.RoutingKey)
			if err != nil {
				core.SyslogCheckError(err)
				d.Ack(true)
			} else {
				d.Ack(false)
			}
		}
		core.SyslogSendJSON("Error, Lost AMQP Connection.")
		forever <- false
	}()

	<-forever
}
