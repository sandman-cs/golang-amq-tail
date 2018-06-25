package main

import (
	"fmt"
	"os"
)

func main() {

	fmt.Println("Starting Program...")

	go OpenChannel(srcRabbitConn)

	forever := make(chan bool)
	<-forever
	os.Exit(0)
}
func processPayload(payload []byte, route string) {

	s := route + "\n" + string(payload[0:]) + "\n"
	fmt.Println(s)
}
