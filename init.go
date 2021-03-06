package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
)

// Configuration File Opjects
type configuration struct {
	SrcBroker         string
	ServerName        string
	SrcBrokerVhost    string
	SrcBrokerUsr      string
	SrcBrokerPwd      string
	SrcBrokerExchange string
	SrcRoute          string
}

var (
	conf configuration
)

func init() {

	conf.ServerName, _ = os.Hostname()
	conf.SrcRoute = "#"

	//Load Configuration Data
	dat, _ := ioutil.ReadFile("conf.json")
	err := json.Unmarshal(dat, &conf)
	if err != nil {
		log.Println(err)
	}

	// create the rabbitmq error channel
	srcRabbitCloseError = make(chan *amqp.Error)

	srcAmqpURI := "amqp://" + conf.SrcBrokerUsr + ":" + conf.SrcBrokerPwd + "@" + conf.SrcBroker + conf.SrcBrokerVhost

	// run the callback in a separate thread
	go srcRabbitConnector(srcAmqpURI)

	// establish the rabbitmq connection by sending
	// an error and thus calling the error callback
	srcRabbitCloseError <- amqp.ErrClosed
	for srcRabbitConn == nil {
		log.Println("Waiting for Source RabbitMQ Connection...")
		time.Sleep(1 * time.Second)
	}

}
