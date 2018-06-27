package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

func main() {

	log.Println("Starting Program...")

	go func() {
		for {
			OpenChannel(srcRabbitConn)
			log.Println("Channel Closed...")
			time.Sleep(5 * time.Second)
		}
	}()

	for {
		time.Sleep(time.Second * 1)
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		fmt.Print(text)
	}

}
func processPayload(payload []byte, route string) {

	//s := route + "\n" + string(payload[0:]) + "\n"
	//fmt.Println(s)
	log.Println(route)
	printJSON(string(payload[0:]))
}

func printJSON(str string) {

	jsonMap := make(map[string]interface{})
	err := json.Unmarshal([]byte(str), &jsonMap)
	if err != nil {
		panic(err)
	}
	dumpMap("", jsonMap)

}

func dumpMap(space string, m map[string]interface{}) {
	for k, v := range m {
		if mv, ok := v.(map[string]interface{}); ok {
			fmt.Printf("{ \"%v\": \n", k)
			dumpMap(space+"\t", mv)
			fmt.Printf("}\n")
		} else {
			fmt.Printf("%v %v : %v\n", space, k, v)
		}
	}
}
