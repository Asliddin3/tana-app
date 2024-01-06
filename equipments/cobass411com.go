package equipments

import (
	"ZettaGroup/Tana-App/tools"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lenaten/hl7"
	"go.bug.st/serial"
)

type SerialCobass411 struct {
	mType            int
	conn             serial.Port
	MOrder           int
	mode             *serial.Mode
	MLength          int
	Order            []string
	MainBuffer       []byte
	StrArr           []string
	ResArr           []string
	host             string
	connectionStatus bool
}

func NewAsCobas411Com(name string, host string, conf tools.EquipmentConfig) *SerialCobass411 {
	return &SerialCobass411{
		mode:       CreateMode(conf),
		host:       host,
		mType:      0,
		MainBuffer: []byte{},
	}
}
func (as *SerialCobass411) GetConnectionStatus() bool {
	return as.connectionStatus
}
func (s *SerialCobass411) parseReceived(ch rune) {
	fmt.Println("gotten type", string(ch))
	switch ch {
	case 'Q':
		s.mType = 0
	case 'P':
		s.mType = 1
	default:
		s.mType = 1
	}
}

func (s *SerialCobass411) Connect() {
	port, err := serial.Open("/dev/ttyUSB0", s.mode)
	if err != nil {
		fmt.Println("failed to open mode", err)
		return
	}
	s.connectionStatus = true
	s.handler(port)
	s.connectionStatus = false
}
func (c *SerialCobass411) handler(port serial.Port) {

	// name := time.Now().Format("2006-01-02T15:04:05")
	// f, err := os.Create(fmt.Sprintf("logs/%s.txt", name))
	// name = time.Now().Format("2006-01-02T15:04:05")
	// out, err := os.Create(fmt.Sprintf("logt/%s.txt", name))
	// fmt.Println("create file error:", err)
	for {
		buffer, err := hl7.ReadBuf(port)
		fmt.Println("gotten buffer", string(buffer))
		// out.Write(buffer)
		if err != nil {
			fmt.Println("failed to read", err)
			return
		}
		if bytes.Contains(buffer, []byte("\u0005")) {
			port.Write([]byte("\u0006"))
			// f.Write([]byte("\u0006"))
			showLogs("<ENQ>")
			c.MainBuffer = []byte{}
		} else if bytes.Contains(buffer, []byte("\u0002")) {
			showLogs("Append more")
			port.Write([]byte("\u0006"))
			// f.Write([]byte("\u0006"))
			c.MainBuffer = append(c.MainBuffer, buffer...)
		} else if bytes.Contains(buffer, []byte("\u0015")) {
			showLogs("<NAK>")
			port.Write([]byte("\u0006"))
		} else if bytes.Contains(buffer, []byte("\u0004")) {
			showLogs("<EOT>")
			fmt.Println("MAIN BUFFER")
			fmt.Println(string(string(c.MainBuffer)))
			strArr := c.parseString(c.MainBuffer)
			c.StrArr = strArr
			fmt.Println("len", len(strArr))
			for _, item := range strArr {
				fmt.Println("item", item)
			}
			c.MainBuffer = []byte{}
			if len(strArr) > 0 {
				fmt.Println("str", strArr[1])
				c.parseReceived(rune([]rune(string(strArr[1]))[0]))
				switch c.mType {
				case 0:
					showLogs("Query for order received")
					c.Order = c.getOrder()
					if len(c.Order) > 0 {
						showLogs("Order is being sent")
						c.MOrder = 0
						c.MLength = len(c.Order)
						// port.Write([]byte(c.Order[c.MOrder]))
						c.conn.Write([]byte(c.Order[c.MOrder]))
						// f.Write([]byte(c.Order[c.MOrder]))
						c.MOrder++
						break
					}
				case 1:
					showLogs("Result recieved")
					c.saveResults()

				}
			}
		} else if bytes.Contains(buffer, []byte("\u0006")) {
			showLogs("<ACK>")
			fmt.Println("write test order")
			if len(c.Order) > 0 {
				if c.MOrder == c.MLength {
					c.MOrder = 0
					c.MLength = 0
					c.Order = []string{}
				} else {
					test := c.Order[c.MOrder] + "\r\n"
					// f.Write([]byte(test))
					c.conn.Write([]byte(test))
					c.MOrder++
					// showLogs("send message" + test)
					// _, err = port.Write([]byte(test))
					// if err != nil {
					// 	showError("failed to send message", err)
					// 	return
					// }
				}
			}
		} else {
			c.MainBuffer = append(c.MainBuffer, buffer...)
			port.Write([]byte("\u0006"))
			// f.Write([]byte("\u0006"))
		}
		// port.Write(buffer)
	}
}

//	type Results struct {
//		Code        string
//		Biomaterial string
//	}
//
//	func (c *SerialCobass411) writeMessage(data []byte) {
//		for i := 0; i < len(data); i = i + 14 {
//			var err error
//			if i > len(data)-14 {
//				_, err = c.conn.Write(data[i:])
//			} else {
//				_, err = c.conn.Write(data[i : i+14])
//			}
//			if err != nil {
//				showError("failed to write message", err)
//			}
//			time.Sleep(time.Microsecond * 150)
//		}
//	}
func (c *SerialCobass411) getOrder() []string {
	var order []string
	orderFound := false
	fmt.Println("strArr", c.StrArr)
	testArr := c.parser(c.parser(c.StrArr[1], "|")[2], "^")
	fmt.Println("testArr", testArr)
	// barcode := testArr[1]
	selectionInfo := string(testArr[2]) + "^" + string(testArr[3]) + "^" + string(testArr[4]) + "^" + string(testArr[6]) + "^" + string(testArr[7])
	trimedBarcode := strings.ReplaceAll(testArr[1], " ", "")
	showLogs("gotten sample barcode " + trimedBarcode)
	//cobass protocol
	// selectionInfo := string(testArr[3]) + "^" + string(testArr[4]) + "^" + string(testArr[5]) + "^" + string(testArr[7]) + "^" + string(testArr[8])
	// trimedBarcode := strings.Trim(testArr[2], " ")
	// showLogs("gotten sample barcode " + trimedBarcode)
	order = append(order, "\u0005")
	order = append(order, "\u00021H|\\^&||||||||||P||\r\u0003"+getChecksum("1H|\\^&||||||||||P||\r\u0003"))
	order = append(order, "\u00022P|1\r\u0003"+getChecksum("2P|1\r\u0003"))
	// layout := "20060102150405"

	// Format the time according to the layout
	// formattedTime := time.Now().Format(layout)
	if orderFound {
		test := "3O|1|" + trimedBarcode + "|" + selectionInfo +
			"||R||||||N||||||||||||||Z" + string('\r')
		order = append(order, "\u0002", test, "\u0003", getChecksum(test+"\u0003"))
	} else {
		text := "3O|1|" + trimedBarcode + "|" + selectionInfo + "|"
		// patients := patientsMap[string(barcode)]
		// for i, patient := range patients {
		//ttg code
		text += "^^^" + "108"
		dilution := ""
		// if patient.Pregnant == 1 {
		// 	dilution = "9"
		// }
		text += "^" + dilution
		// if i == len(patients)-1 {
		// text += "\\"
		// }
		// }
		// text += "|R||||||N||||||||||||||Q\r"
		text += "|R||||||N||||||||||||||Q\r"

		fmt.Println("text", text)
		order = append(order, "\u0002"+text+"\u0003"+getChecksum(text+"\u0003"))
	}
	order = append(order, "\u00024L|1\r\u0003"+getChecksum("4L|1\r\u0003"))
	order = append(order, "\u0004")
	return order
}

func (s *SerialCobass411) parser(line string, del string) []string {
	return strings.Split(strings.ReplaceAll(line, del, "!"), "!")
}

func (s *SerialCobass411) parseString(data []byte) []string {
	temp := strings.Split(string(data), "\u0003")
	var fullMessage strings.Builder

	for i := 0; i < len(temp); i++ {
		if i != 0 {
			fullMessage.WriteString(strings.Split(temp[i], "\u0002")[1])
		} else {
			fullMessage.WriteString(strings.Split(temp[i], "\u0002")[1])
		}
	}
	return strings.Split(fullMessage.String(), "\r")
}

// func parseString(data []byte) []string {
// 	defer recover()
// 	temp := strings.Split(string(data), "\u0017")
// 	var fullMessage strings.Builder

// 	for i := 0; i < len(temp); i++ {
// 		if i == len(temp)-1 {
// 			fullMessage.WriteString(strings.Split(temp[i], "\u0002")[1][1:])
// 			fullMessage.WriteString(strings.Split(fullMessage.String(), "\u0003")[0])
// 		} else {
// 			fullMessage.WriteString(strings.Split(temp[i], "\u0002")[1][1:])
// 		}
// 	}
// 	return strings.Split(fullMessage.String(), "\r")
// }

func (c *SerialCobass411) saveResults() {
	resArr := make([][]string, 0)
	for i := 0; i < len(c.StrArr); i++ {
		runeArr := []rune(string(c.StrArr[i]))
		var letter rune
		if len(runeArr) > 0 {
			letter = []rune(string(c.StrArr[i]))[1]
		}
		var barcode string
		fmt.Println("letter", string(letter))
		fmt.Println("line", string(c.StrArr[i]))
		switch letter {
		case 'O':
			barcode = c.parser(c.StrArr[i], "|")[2]
			// barcode = c.parser(barcode, []byte("^"))[1]
			barcode = strings.Trim(barcode, string([]byte{32}))
			fmt.Println("barcode", string(barcode))
			showLogs("gotten test with barcode:" + string(barcode))
		case 'R':
			var comment string
			if []rune(string(c.StrArr[i+1]))[1] == 'C' {
				comment = c.parser(c.StrArr[i+1], "|")[3]
			}
			comment += "|"
			comment += c.StrArr[i]
			resArr = append(resArr, c.parser(comment, "|"))
		}
	}
	fmt.Println("resultArr", resArr)
	// c.ResArr = resArr
	results := make([]OBXResult, 0)
	for _, val := range resArr {
		code := c.parser(c.parser(val[3], "^")[3], "/")[0]
		results = append(results, OBXResult{
			Code: string(code),
			Name: string(val[3]),
			Res:  string(val[4]),
			// Error: string(val[0]),
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
func trim(data []byte) []byte {
	STX := "\u0002"
	ETX := "\u0003"

	pos := bytes.Index(data, []byte(ETX))
	if pos != -1 {
		data = data[:pos]
		data = bytes.ReplaceAll(data, []byte(STX), []byte(""))
	}

	return data
}

// func getChecksum(data string) string {
// 	byteArr := []byte(data)
// 	var hexArr []string

// 	for _, el := range byteArr {
// 		hexArr = append(hexArr, fmt.Sprintf("%02x", el))
// 	}

// 	sum := 0
// 	for _, el := range hexArr {
// 		// var val int
// 		// _, err := fmt.Sscanf(el, "%x", &val)
// 		val, err := strconv.ParseInt(el, 16, 0)
// 		if err == nil {
// 			sum += int(val)
// 		}
// 	}

// 	data = fmt.Sprintf("%02x", sum)
// 	return strings.ToUpper(data[len(data)-2:])
// }

// func getChecksum(data string) string {
// 	byteArr := []byte(data)

// 	sum := 0
// 	for _, el := range byteArr {
// 		sum += int(el)
// 	}

// 	// Convert the sum to hexadecimal format
// 	hexSum := fmt.Sprintf("%X", sum)

// 	// Take the last two digits of the hexadecimal sum
// 	if len(hexSum) >= 2 {
// 		hexSum = hexSum[len(hexSum)-2:]
// 	}

// 	// Convert these two digits into ASCII code using the range "0" to "9" and "A" to "F"
// 	var checksumASCII string
// 	_, err := fmt.Sscanf(hexSum, "%X", &checksumASCII)
// 	if err != nil {
// 		fmt.Println("error while converting checksum to ASCII:", err)
// 		return ""
// 	}

// 	return checksumASCII
// }

// func getChecksum(data string) string {
// 	byteArr := []byte(data)
// 	var hexArr []string

// 	for _, el := range byteArr {
// 		hexArr = append(hexArr, fmt.Sprintf("%02X", el))
// 	}

// 	sum := 0
// 	for _, el2 := range hexArr {
// 		var val int
// 		_, err := fmt.Sscanf(el2, "%X", &val)
// 		if err != nil {
// 			fmt.Println("gotten error while check sum", err)
// 			return ""
// 		}
// 		sum += val
// 		fmt.Println("sum", sum)
// 	}

// 	data = fmt.Sprintf("%02X", sum)
// 	return strings.ToUpper(data)
// }

// type OBXResult struct {
// 	Code  string `json:"code"`
// 	Name  string `json:"name"`
// 	Val   string `json:"val"`
// 	Unit  string `json:"unit"`
// 	Norms string `json:"norms"`
// 	Flag  string `json:"flag"`
// 	Error string `json:"error"`
// }
