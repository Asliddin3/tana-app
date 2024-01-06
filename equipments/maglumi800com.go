package equipments

import (
	"ZettaGroup/Tana-App/tools"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"

	"go.bug.st/serial"

	"time"
)

type AsServerMaglumi800Com struct {
	MESSAGE_ORDER    int
	MESSAGE_LENGTH   int
	MESSAGE_TYPE     int
	QUERY            int
	RESULTS          int
	dataReceived     []string
	MainBuffer       []byte
	testOrder        []string
	analyzerMessage  []string
	tools            *tools.LISTools
	socket           serial.Port
	connectionStatus bool
	eqName           string
	host             string
	server           net.Listener
	mode             *serial.Mode
}

func NewAsServerMaglumi800Com(name string, host string, conf tools.EquipmentConfig) *AsServerMaglumi800Com {
	return &AsServerMaglumi800Com{
		MESSAGE_ORDER:  0,
		mode:           CreateMode(conf),
		host:           host,
		MESSAGE_LENGTH: 0,
		MESSAGE_TYPE:   2,
		QUERY:          0,
		RESULTS:        1,
		tools:          tools.NewLISTools(name),
		socket:         nil,
		server:         nil,
	}
}

func (as *AsServerMaglumi800Com) Connect() {
	conn, err := serial.Open("/dev/ttyUSB0", as.mode)
	fmt.Println("start connect err", err)
	if err != nil {
		fmt.Println("failed to open mode", err)
		return
	}
	as.socket = conn
	as.connectionStatus = true
	as.tools.LogAndDisplayMessage("\n--------------------------------------------------------------------------------------------------------------------------\n", 2)
	as.Start()
	as.connectionStatus = false
}
func (as *AsServerMaglumi800Com) GetConnectionStatus() bool {
	return as.connectionStatus
}

func (as *AsServerMaglumi800Com) Start() {
	for {
		resArr := make([]byte, 60)
		n, err := as.socket.Read(resArr)
		if err != nil {
			as.tools.ErrorLog(err.Error(), err)
			return
			// break
		}
		// fmt.Println("gotten buffer", resArr)
		log.Println("gotten request from maglumi", string(resArr))
		result := resArr[:n]
		var charBuffer strings.Builder
		for _, b := range result {
			charBuffer.WriteByte(b)
		}
		buffer := charBuffer.String()

		for len(buffer) > 0 {
			switch {
			case strings.Contains(buffer, "\u0005"):
				log.Println("---->ENQ")
				log.Println("<----ACK")
				as.tools.LogAndDisplayMessage("<ENQ>", 0)
				_, err = as.socket.Write([]byte{'\u0006'})
				if err != nil {
					fmt.Println("gotten err", err)
				}
				as.tools.LogAndDisplayMessage("<ACK>", 1)
			case strings.Contains(buffer, "\u0002"):
				log.Println("---->STX")
				log.Println("<----ACK")
				as.tools.LogAndDisplayMessage("<STX>", 0)
				as.socket.Write([]byte{'\u0006'})
				as.tools.LogAndDisplayMessage("<ACK>", 1)

			case strings.Contains(buffer, "\u0004"):
				log.Println("---->EOT")
				log.Println("<----ACK")
				as.tools.LogAndDisplayMessage("<EOT>", 0)
				as.socket.Write([]byte{'\u0006'})
				as.tools.LogAndDisplayMessage("<ACK>", 1)
				t := bytes.Split(as.MainBuffer, []byte{'\r'})
				for _, buf := range t {
					fmt.Println("buf", string(buf))
					as.dataReceived = append(as.dataReceived, string(buf))
				}
				as.MainBuffer = []byte{}
				if len(as.dataReceived) > 0 {
					log.Println("data received ", as.dataReceived)
					as.parseReceived()
					switch as.MESSAGE_TYPE {
					case 0:
						as.testOrder = as.getOrder()
						if len(as.testOrder) > 0 {
							log.Println("Start analyzer query ")
							as.tools.LogAndDisplayMessage("Analyzer query was processed and test order was prepared", 2)
							as.MESSAGE_ORDER = 0
							// as.MESSAGE_LENGTH = len(as.testOrder)
							_, err = as.socket.Write([]byte(as.testOrder[as.MESSAGE_ORDER]))
							if err != nil {
								fmt.Println("gotten err", err)
							}
							fmt.Println("writen order", as.testOrder[as.MESSAGE_ORDER])
							as.MESSAGE_ORDER++
							// _, err = as.socket.Write([]byte(as.testOrder[as.MESSAGE_ORDER]))
							// if err != nil {
							// 	fmt.Println("gotten err", err)
							// }
							// as.MESSAGE_ORDER++
							as.tools.LogAndDisplayMessage("<ENQ>", 1)
						}
					case 1:
						log.Println("Save analyzer results ")
						as.saveResults()
						as.tools.LogAndDisplayMessage("Analyzer test results were processed and saved", 2)
					default:
						as.tools.LogAndDisplayMessage("The message was empty", 2)
					}
					// as.testOrder = nil
				}
				as.dataReceived = nil

			case strings.Contains(buffer, "\u0006"):
				log.Println("----->ACK")
				fmt.Println("gotten get order")
				as.tools.LogAndDisplayMessage("<ACK>", 0)
				if len(as.testOrder) > 0 {
					fmt.Println("as.MessageOrder", as.MESSAGE_ORDER)
					fmt.Println("as.testOrder")
					if as.MESSAGE_ORDER == len(as.testOrder) {
						as.MESSAGE_ORDER = 0
						as.MESSAGE_LENGTH = 0
						as.testOrder = nil
					} else {
						test := as.testOrder[as.MESSAGE_ORDER] + "\r\n"
						as.MESSAGE_ORDER++
						as.socket.Write([]byte(test))
						fmt.Println("sendingtest")
						as.tools.LogAndDisplayMessage(test, 1)
					}
				}

			case strings.Contains(buffer, "\u0015"):
				log.Println("---->NAK")
				log.Println("<----ACK")
				as.tools.LogAndDisplayMessage("<NAK>", 0)
				as.socket.Write([]byte("\u0006"))
				as.tools.LogAndDisplayMessage("<ACK>", 1)

			default:
				as.tools.LogAndDisplayMessage(buffer, 0)
				// as.dataReceived = append(as.dataReceived, buffer)
				as.MainBuffer = append(as.MainBuffer, []byte(buffer)...)
			}

			buffer = ""
		}
	}
}

func (as *AsServerMaglumi800Com) parseReceived() {
	defer func() {
		as.dataReceived = nil
	}()
	fmt.Println("gotten type", string(as.dataReceived[1]))
	if len(as.dataReceived) > 1 {
		switch as.dataReceived[1][0] {
		case 'Q':
			as.MESSAGE_TYPE = 0
		case 'P':
			as.MESSAGE_TYPE = 1
		default:
			as.MESSAGE_TYPE = 2
		}
		as.analyzerMessage = as.dataReceived
	} else {
		as.MESSAGE_TYPE = 2
		as.tools.ErrorLog("Error in parsing the received message and the message is not saved: "+as.eqName, nil)
	}
}

type Results struct {
	Res  string `json:"res"`
	Code string `json:"code"`
	Unit string `json:"unit"`
	Norm string `json:"norm"`
	Flag string `json:"flag"`
}

func (as *AsServerMaglumi800Com) saveResults() {
	barcode := ""
	resArr := make([][]string, 0)
	for _, el := range as.analyzerMessage {
		switch el[0] {
		case 'H', 'P':
			continue
		case 'O':
			parts := bytes.Split([]byte(el), []byte("|"))
			barcode = string(parts[2])
			as.tools.LogAndDisplayMessage("Sample Barcode: "+barcode, 2)
		case 'R':
			resArr = append(resArr, strings.Split(el, "|"))
		case 'L':
			continue
		}
	}

	// headBarcode := "method=apiResultSave&lisResult[name]=" + url.QueryEscape(as.eqName) +
	// 	"&lisResult[host]=" + url.QueryEscape(as.host) +
	// 	"&lisResult[barcode]=" + url.QueryEscape(barcode)
	results := make([]Results, len(resArr))
	for i, el := range resArr {
		codeParts := strings.Split(el[2], "^")
		code := codeParts[3]
		fmt.Println("el", el)
		results[i] = Results{
			Code: code,
			Res:  el[3],
			Unit: el[4],
			Norm: el[5],
			Flag: el[6],
		}
		// params := "&lisResult[code]=" + url.QueryEscape(code) +
		// 	"&lisResult[R][res]=" + url.QueryEscape(el[3]) +
		// 	"&lisResult[R][unit]=" + url.QueryEscape(el[4]) +
		// 	"&lisResult[R][norms]=" + url.QueryEscape(el[5]) +
		// 	"&lisResult[R][flag]=" + url.QueryEscape(el[6])
		// paramsArr = append(paramsArr, headBarcode+params)
	}
	res, err := json.MarshalIndent(&results, "", " ")
	if err != nil {
		fmt.Println("failed to parse")
	}
	fmt.Println("Gotten results", string(res))
}
func (as *AsServerMaglumi800Com) getOrder() []string {
	var order []string
	barcodeArr := as.tools.Parser(as.tools.Parser(as.analyzerMessage[1], "|")[2], "^")
	fmt.Println("barcode Arr", barcodeArr)
	barcode := "00000000"
	if len(barcodeArr) > 1 {
		barcode = barcodeArr[1]
	}
	exists := true
	as.tools.LogAndDisplayMessage("Sample Barcode: "+barcode, 2)
	// as.tools.LogAndDisplayMessage("Getting orders from API...: "+as.eqName, 2)

	date := time.Now().Format("20060102")
	order = append(order, "\u0005", "\u0002", "H|\\^&||PSWD|MaglumiBio|||||Lis||P|E1394-97|"+date, "P|1")
	//Write there result codes
	codes := []string{"ALB h"}
	if exists {
		for i, obj := range codes {
			order = append(order, fmt.Sprintf("O|%d|%s||^^^%s|R", i+1, barcode, obj))
		}
	}
	order = append(order, "L|1|N", "\u0003", "\u0004")

	return order
}

func (as *AsServerMaglumi800Com) getOrderList() [][]string {
	return nil
}

func (as *AsServerMaglumi800Com) GetType() string {
	return "server"
}

// func handleConnection(server *AsServerMaglumi800Com) {

// 	fmt.Println("server start")

// 	isConnect := server.Connect()
// 	if isConnect {
// 		server.Start()
// 	} else {
// 		fmt.Println("Not connected")
// 		time.Sleep(time.Second * 3)
// 	}
// 	// Start processing data
// }

// func Maglumi800Lan() AsServerMaglumi800Com {
// 	server := AsServerMaglumi800Com{
// 		// Initialize your AsServerMaglumi800Com properties here
// 		tools: tools.NewLISTools("test"),
// 		// server: listener,

// 	}

// 	// for {
// 	handleConnection(&server)
// 	// }
// 	return server
// }

// Implement other methods of BaseEquipment interface if needed
