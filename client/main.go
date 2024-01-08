package main

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"

	gosocketio "github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
)

func doSomethingWith(c *gosocketio.Client, wg *sync.WaitGroup) {
	rand.Seed(time.Now().Unix())
	charset := "1234567890asdf"
	res := make([]byte, 4)
	for i, _ := range res {
		res[i] = charset[rand.Intn(len(charset))]
	}
	fmt.Println("send queue", string(res))
	err := c.Emit("monitorQueue", string(res))
	if err != nil {
		fmt.Println("failed to send queue", err)
	}
	// if res, err := c.Ack("join", "This is a client", time.Second*3); err != nil {

	// 	log.Printf("error: %v", err)
	// } else {
	// 	log.Printf("result %q", res)
	// }
	wg.Done()
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	c, err := gosocketio.Dial(
		gosocketio.GetUrl("127.0.0.1", 12345, false),
		transport.GetDefaultWebsocketTransport())
	if err != nil {
		log.Fatal(err)
	}

	err = c.On(gosocketio.OnDisconnection, func(h *gosocketio.Channel) {
		fmt.Println("disconnected")

	})
	if err != nil {
		log.Fatal(err)
	}

	err = c.On(gosocketio.OnConnection, func(h *gosocketio.Channel) {
		log.Println("Connected")
	})
	if err != nil {
		log.Fatal(err)
	}

	err = c.On(gosocketio.OnError, func(h *gosocketio.Channel, err error) {
		log.Println("error", err.Error())
	})
	// if err != nil {
	// 	log.Fatal(err)
	// }

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go doSomethingWith(c, wg)
	time.Sleep(time.Second * 10)
	go doSomethingWith(c, wg)
	wg.Wait()
	c.Close()
	log.Printf("Done")
}
