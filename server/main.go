package socket

import (
	"ZettaGroup/Tana-App/monitor"
	"ZettaGroup/Tana-App/tools"
	"fmt"
	"log"
	"net/http"
	"time"

	socketio "github.com/googollee/go-socket.io"
)

// 198.168.1.250
// var SERVER = "localhost:8000"
// var DeviceID = "123"

// var in = bufio.NewReader(os.Stdin)

// func getInput(input chan string) {
// 	result, err := in.ReadString('\n')
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}
// 	input <- result
// }

func EstablishSocketIOServer(conf tools.ConfigFile, m *monitor.Monitor, isConn chan bool, queue chan string) {
	server := socketio.NewServer(nil)
	// m := monitor.NewMonitor(conf.MonitorHost, "10000")

	server.OnConnect("", func(c socketio.Conn) error {
		log.Printf("on connect\n")
		isConn <- true
		time.Sleep(1 * time.Second)
		return nil
	})
	server.OnError("", func(c socketio.Conn, err error) {
		fmt.Println("get error", err)
	})
	server.OnEvent("/", "monitorQueue", func(s socketio.Conn, msg string) string {
		// log.Printf("on message:%v\n", msg)
		fmt.Println("gotten request for queue", msg)
		// for i := 0; i < 3; i++ {
		err := m.SendMessage(msg)
		if err != nil {
			log.Printf("failed to send message %v", err.Error())
			// server.Emit("error", fmt.Sprintf("failed to send message:%v", err))
		} else {
			fmt.Println("message sended")
			queue <- msg
			return ""
		}
		fmt.Println("reconnect to monitor")
		err = m.Reconnect()
		if err != nil {
			fmt.Println("failed to reconnect", err)
			return ""
		}
		err = m.SendMessage(msg)
		if err != nil {
			fmt.Println("failed to send message", err)
			return ""
		}
		// }
		s.Emit("monitorQueue", "sucess")
		return ""
	})
	server.OnDisconnect("", func(c socketio.Conn, s string) {
		log.Printf("on disconnect\n")
		isConn <- false
	})
	go server.Serve()
	defer server.Close()

	http.Handle("/socket.io/", server)
	// http.Handle("/", http.FileServer(http.Dir("./asset")))
	log.Println("Socket io running at ", conf.SocketHost)
	http.ListenAndServe(conf.SocketHost, nil)
}

type MonitorSocket struct {
	Monitor *monitor.Monitor
}

func NewMonitor(m *monitor.Monitor) *MonitorSocket {
	return &MonitorSocket{
		Monitor: m,
	}
}

// func handlerSocketIO()

// func establishConnection(URL url.URL, m *monitor.Monitor, interrupt chan os.Signal) {
// 	input := make(chan string, 1)
// 	c, _, err := websocket.DefaultDialer.Dial(URL.String(), nil)
// 	if err != nil {
// 		log.Println("Error:", err)
// 		return
// 	}
// 	defer c.Close()
// 	done := make(chan struct{})
// 	go func() {
// 		defer close(done)
// 		err = c.WriteMessage(websocket.TextMessage, []byte(DeviceID))
// 		if err != nil {
// 			fmt.Println("error while writing", err)
// 			return
// 		}
// 		for {
// 			_, message, err := c.ReadMessage()
// 			if err != nil {
// 				log.Println("ReadMessage() error:", err)
// 				return
// 			}
// 			m.SendMessage(string(message))
// 			log.Printf("Received: %s", message)

// 		}
// 	}()

// 	for {
// 		select {
// 		case <-time.After(10 * time.Second):
// 			log.Println("Please give me input!", TIMESWAIT)
// 		case <-done:
// 			return
// 		case t := <-input:
// 			err := c.WriteMessage(websocket.TextMessage, []byte(t))
// 			if err != nil {
// 				fmt.Println("error while writing", err)
// 				return
// 			}
// 			if err != nil {
// 				log.Println("Write error:", err)
// 				return
// 			}
// 			TIMESWAIT = 0
// 			// go getInput(input)
// 		case <-interrupt:
// 			log.Println("Caught interrupt signal - quitting!")
// 			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
// 			if err != nil {
// 				log.Println("Write close error:", err)
// 				return
// 			}
// 			select {
// 			case <-done:
// 			case <-time.After(2 * time.Second):
// 			}
// 			return
// 		}
// 	}
// }
