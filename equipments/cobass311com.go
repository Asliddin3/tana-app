package equipments

import (
	"ZettaGroup/Tana-App/tools"
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lenaten/hl7"
	"go.bug.st/serial"
)

type SerialCobass311 struct {
	mType            int
	host             string
	conn             serial.Port
	MessageOrder     int
	MessageLength    int
	Order            []string
	MainBuffer       []byte
	StrArr           []string
	ResArr           []string
	mode             *serial.Mode
	connectionStatus bool
}

func NewAsCobas311Com(name string, host string, conf tools.EquipmentConfig) *SerialCobass311 {
	return &SerialCobass311{
		MessageOrder:  0,
		mode:          CreateMode(conf),
		host:          host,
		MessageLength: 0,
		mType:         0,
	}
}
func (as *SerialCobass311) GetConnectionStatus() bool {
	return as.connectionStatus
}
func (s *SerialCobass311) parseReceived(ch rune) {
	fmt.Println("gotten type", string(ch))
	switch ch {
	case 'Q':
		s.mType = 0
	case 'P':
		s.mType = 1
	default:
		s.mType = 2
	}
}

func (as *SerialCobass311) Connect() {
	conn, err := serial.Open("/dev/ttyUSB0", as.mode)
	fmt.Println("start connect err", err)
	if err != nil {
		fmt.Println("failed to open mode", err)
		return
	}
	as.conn = conn
	as.connectionStatus = true
	as.Start()
	as.connectionStatus = false
}

func (c *SerialCobass311) Start() {
	port := c.conn
	// f, err := os.Create(fmt.Sprintf("logs/%s.txt", name))
	// out := time.Now().Format("2006-01-02T15:04:05")
	// outF, err := os.Create(fmt.Sprintf("out/%s.txt", out))
	// fmt.Println("create file error:", err)
	for {
		buffer, err := hl7.ReadBuf(port)

		if err != nil {
			fmt.Println("failed to read", err)
			return
		}
		if bytes.Contains(buffer, []byte("\u0005")) {
			port.Write([]byte("\u0006"))
			showLogs("<ENQ>")
			c.MainBuffer = []byte{}
		} else if bytes.Contains(buffer, []byte("\u0015")) {
			showLogs("<NAQ>")
			showLogs("DID not get message" + string(buffer))
			port.Write([]byte("\u0006"))

		} else if bytes.Contains(buffer, []byte("\u0002")) {
			showLogs("Append more")
			port.Write([]byte("\u0006"))
			c.MainBuffer = append(c.MainBuffer, buffer...)
		} else if bytes.Contains(buffer, []byte("\u0004")) {
			showLogs("<EOT>")
			fmt.Println("MAIN BUFFER")
			fmt.Println(string(c.MainBuffer))
			strArr := parseString(c.MainBuffer)
			c.StrArr = strArr
			fmt.Println("Main strArr", c.StrArr)
			c.MainBuffer = []byte{}
			if len(strArr) > 0 {
				c.parseReceived(rune([]rune(string(strArr[1]))[0]))
				fmt.Println("setten type", c.mType)
				if strings.Contains(strArr[0], "TSREQ^REAL") {
					c.mType = 0
				}
				switch c.mType {
				case 0:
					c.Order = c.getOrder(nil)
					if len(c.Order) > 0 {
						showLogs("Order is being sent")
						c.MessageOrder = 0
						c.MessageLength = len(c.Order)
						// port.Write([]byte(c.Order[c.MessageOrder]))
						c.conn.Write([]byte(c.Order[c.MessageOrder]))
						// f.Write([]byte(c.Order[c.MessageOrder]))
						c.MessageOrder++
					}
					showLogs("Query for order received")
				case 1:
					showLogs("Result recieved")
					c.saveResults()
				default:
					showLogs("Message type is neither QUERY nor RESULTS")
				}
			}
		} else if bytes.Contains(buffer, []byte("\u0006")) {
			showLogs("<ACK>")
			fmt.Println("write test order")
			if len(c.Order) > 0 {
				if c.MessageLength == c.MessageOrder {
					c.MessageLength = 0
					c.MessageOrder = 0
					c.Order = []string{}
				} else {
					test := c.Order[c.MessageOrder]
					c.MessageOrder++
					// n, err := port.Write([]byte(test))
					c.conn.Write([]byte(test))
					// f.Write([]byte(test))
					showLogs("<STX> sending test order")
				}
			}
		} else {
			c.MainBuffer = append(c.MainBuffer, buffer...)
		}
		// port.Write(buffer)
	}
}

func (c *SerialCobass311) getQueryArray() string {
	for _, val := range c.StrArr {
		if strings.Contains(val, "Q|") && !strings.Contains(val, "TSREQ") {
			return val
		}
	}
	return c.StrArr[1]
}

func (c *SerialCobass311) getOrder(i *int) []string {
	var order []string
	orderFound := false
	var testArr []string
	//I changed []byte("^") to []byte("/")
	// if i != nil {
	// 	fmt.Println("str arr", c.StrArr[*i])
	// 	testArr = c.parser(c.parser(c.StrArr[*i], "|")[2], "^")
	// } else {
	fmt.Println("search arr", c.getQueryArray())
	testArr = c.parser(c.parser(c.getQueryArray(), "|")[2], "^")
	// }
	fmt.Println("testArr", testArr)
	// fmt.Println("barcode", c.StrArr[1])
	// testArr = c.parser(c.parser(c.StrArr[1], "|")[2], "/")
	barcode := strings.Trim(testArr[2], " ")
	showLogs("gotten sample barcode " + string(barcode))
	order = append(order, "\u0005")
	orderText := "H|\\^&|||host^2|||||H7600|TSDWN^REPLY|P|1\r"
	orderText = orderText + "P|1" + "\r"
	selInfo := string(testArr[3]) + "^" + string(testArr[4]) + "^" + string(testArr[5]) + "^^"
	// } else {
	//  selInfo = string("") + "^" + string("") + "^" + string("") + "^^"
	// }
	fmt.Println("self Info", testArr)
	fmt.Println(testArr)
	if orderFound {
		orderText = orderText + "O|1|" + testArr[2] + "|" + selInfo +
			"||R||||||A||||1||||||||||O" + "\r" + "L|I|N" + "\r"
	} else {
		var text string
		// biomaterial := "S1"
		gluCode := "717"
		text += "^^^" + gluCode + "^"
		// text += "\\" if more test add to end
		orderText += "O|1|" + barcode + "|" + selInfo + "S1" + "^" + "SC" + "|" + text + "|R||||||A||||1||||||||||O" + "\r" + "L|I|N" + "\r"
	}
	order = append(order, getETBMessages(orderText)...)

	order = append(order, "\u0004")

	return order
}
func (c *SerialCobass311) Write(data []byte) {
	for i := 0; i < len(data); i = i + 14 {
		var err error
		time.Sleep(time.Microsecond * 100)
		if i > len(data)-14 {
			_, err = c.conn.Write(data[i:])
		} else {
			_, err = c.conn.Write(data[i : i+14])
		}
		if err != nil {
			showError("failed to write message", err)
		}
	}
}

func getETBMessages(input string) []string {
	var inputList []string
	length := len(input)
	fmt.Printf("Full Order Message: %s\n", input)
	fmt.Printf("Length of a Message: %d\n", length)

	i := 0
	frame := 1

	for i <= length-240 {
		text := input[i : i+240]
		fmt.Printf("Chunks of a Message: %s\n", text)

		checksum := getChecksum(fmt.Sprintf("%d%s\u0017", frame, text))
		inputList = append(inputList, fmt.Sprintf("\u0002%d%s\u0017%s\r\n", frame, text, checksum))
		frame++
		i += 240
	}

	remainingText := input[i:]
	checksum := getChecksum(fmt.Sprintf("%d%s\u0003", frame, remainingText))
	inputList = append(inputList, fmt.Sprintf("\u0002%d%s\u0003%s\r\n", frame, remainingText, checksum))

	return inputList
}

func getChecksum(data string) string {
	byteArr := []byte(data)
	var hexArr []string

	for _, el := range byteArr {
		hexArr = append(hexArr, fmt.Sprintf("%02x", el))
	}

	sum := 0
	for _, el := range hexArr {
		// var val int
		// _, err := fmt.Sscanf(el, "%x", &val)
		val, err := strconv.ParseInt(el, 16, 0)
		if err == nil {
			sum += int(val)
		}
	}

	data = fmt.Sprintf("%02x", sum)
	return strings.ToUpper(data[len(data)-2:])
}

func (s *SerialCobass311) parser(line string, del string) []string {
	return strings.Split(strings.ReplaceAll(line, del, "!"), "!")
}

func parseString(data []byte) (res []string) {
	temp := strings.Split(string(data), "\u0017")
	var fullMessage strings.Builder

	for i := 0; i < len(temp); i++ {
		part := strings.Split(temp[i], "\u0002")

		// Append the desired substring to fullMessage
		if i == len(temp)-1 {
			fullMessage.WriteString(part[1][1:strings.Index(part[1], "\u0003")])
		} else {
			// For other parts, take a substring starting from index 1
			fullMessage.WriteString(part[1][1:])
		}

	}
	res = strings.Split(fullMessage.String(), "\r")
	return res
}
func parseReceived(ch rune, mType *int) {
	switch ch {
	case 'Q':
		*mType = 0
	case 'P':
		*mType = 1
	default:
		*mType = 2
	}
}

func showError(event string, err error) {
	fmt.Println("-------------" + event + err.Error() + "--------------")
}
func (c *SerialCobass311) saveResults() {
	resArr := make([][]string, 0)
	for i := 0; i < len(c.StrArr); i++ {
		runeArr := []rune(string(c.StrArr[i]))
		var letter rune
		if len(runeArr) > 0 {
			letter = []rune(string(c.StrArr[i]))[0]
		}
		var barcode string
		fmt.Println("letter", string(letter))
		fmt.Println("line", string(c.StrArr[i]))
		switch letter {
		case 'O':
			// barcode = c.parser(c.StrArr[i], "|")[2]
			barcode = c.parser(c.StrArr[i], "|")[2]
			// barcode = c.parser(barcode, "^")[1]
			barcode = strings.Trim(barcode, string([]byte{32}))
			fmt.Println("barcode", string(barcode))
			showLogs("gotten test with barcode:" + string(barcode))
		case 'R':
			var comment string
			fmt.Println("Comment", string(c.StrArr[i+1]))
			if []rune(string(c.StrArr[i+1]))[1] == 'C' {
				arr := c.parser(c.StrArr[i+1], "|")
				if len(arr) > 2 {
					comment = arr[3]
				} else {
					comment = ""
				}
				fmt.Println("comment", comment)
			}
			comment += "|"
			comment += c.StrArr[i]
			fmt.Println("append res", comment)
			resArr = append(resArr, c.parser(comment, "|"))
		case 'Q':
			// c.Order = c.getOrder(&i)
			// c.StrArr = []string{c.StrArr[i]}
			// if len(c.Order) > 0 {
			// 	showLogs("Order is being sent")
			// 	c.MessageOrder = 0
			// 	c.MessageLength = len(c.Order)

			// 	c.conn.Write([]byte(c.Order[c.MessageOrder]))
			// 	c.MessageOrder++
			// }
			showLogs("Received get order from save result")

		}
	}
	fmt.Println("resultArr", resArr)
	// c.ResArr = resArr
	results := make([]OBXResult, 0)
	for _, val := range resArr {
		code := strings.ReplaceAll(c.parser(val[3], "^")[3], "/", "")
		results = append(results, OBXResult{
			Code:  string(code),
			Res:   string(val[4]),
			Name:  string(val[3]),
			Unit:  string(val[5]),
			Norms: string(val[6]),
			Flag:  string(val[7]),
		})
	}
	str, err := json.MarshalIndent(&results, "", " ")
	if err != nil {
		showError("failed marshal indent", err)
	}
	fmt.Println("gotten results\n", string(str))
}

// type OBXResult struct {
// 	Code  string `json:"code"`
// 	Res   string `json:"res"`
// 	Name  string `json:"name"`
// 	Val   string `json:"val"`
// 	Unit  string `json:"unit"`
// 	Norms string `json:"norms"`
// 	Flag  string `json:"flag"`
// 	Error string `json:"error"`
// }
