package socket

import (
	"ZettaGroup/Tana-App/monitor"
	"ZettaGroup/Tana-App/tools"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/polling"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"
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
	// server := socketio.NewServer(nil)
	// m := monitor.NewMonitor(conf.MonitorHost, "10000")
	var socket socketio.Server
	opts := &engineio.Options{
		Transports: []transport.Transport{
			&polling.Transport{
				CheckOrigin: allowOriginFunc,
			},
			&websocket.Transport{
				CheckOrigin: allowOriginFunc,
				// Proxy:       nil,
			},
		},
	}
	socket = *socketio.NewServer(opts)
	socket.OnConnect("", func(c socketio.Conn) error {
		log.Printf("on connect\n")
		// isConn <- true
		return nil
	})
	socket.OnError("", func(c socketio.Conn, err error) {
		fmt.Println("get error", err)
	})
	socket.OnEvent("/", "monitorQueue", func(s socketio.Conn, message interface{}) string {
		// log.Printf("on message:%v\n", msg)
		msg := fmt.Sprintf("%v", message)
		fmt.Println("gotten request for queue", msg)
		// for i := 0; i < 3; i++ {
		err := m.SendMessage(msg)
		if err != nil {
			log.Printf("failed to send message %v", err.Error())
			// socket.Emit("error", fmt.Sprintf("failed to send message:%v", err))
		} else {
			fmt.Println("message sended")
			queue <- msg
			return ""
		}
		fmt.Println("reconnect to monitor")
		err = m.Reconnect()
		if err != nil {
			fmt.Println("failed to reconnect", err)
			return "failed to reconnect"
		}
		err = m.SendMessage(msg)
		if err != nil {
			fmt.Println("failed to send message", err)
			return "failed to send"
		}
		// }
		// s.Emit("monitorQueueResult", "sucess")
		return ""
	})
	socket.OnDisconnect("", func(c socketio.Conn, s string) {
		log.Printf("on disconnect\n")
		// isConn <- false
	})

	log.Println("Serving at localhost:12345...")

	ginserver := gin.Default()
	corsConfig := cors.DefaultConfig()
	// corsConfig.AllowOriginFunc = access
	corsConfig.AllowWebSockets = true
	corsConfig.AllowOrigins = []string{"*"}
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = []string{"*"}
	corsConfig.ExposeHeaders = []string{"*"}

	ginserver.Use(
		cors.New(corsConfig),
	)
	ginserver.GET("/socket.io/*any", gin.WrapH(&socket))
	ginserver.POST("/socket.io/*any", gin.WrapH(&socket))
	go func() {
		if err := socket.Serve(); err != nil {
			log.Fatalf("socketio listen error: %s\n", err)
		}
	}()
	defer socket.Close()

	log.Fatal(ginserver.Run(conf.SocketHost))
}

var allowOriginFunc = func(r *http.Request) bool {
	return true
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
