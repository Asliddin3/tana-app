package equipments

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/lenaten/hl7"
)

// var (
// 	VT = string([]byte{0x0b})
// 	FS = string([]byte{0x1c})
// 	CR = string([]byte{0x0d})
// )

// func main() {
// 	// clientOptions := "192.168.1.201:5100"
// 	// clientOptions := "localhost:5600"
// 	mType := 0
// 	c := NewDymindDF50AsServer("Ma")
// 	c.Connect()
// }

func NewDymindDF50AsServer(name string, apiHost string, host string) *DymindDF50 {
	mType := 0
	return &DymindDF50{
		eqName:  name,
		apiHost: apiHost,
		host:    host,
		MType:   &mType,
	}
}

type DymindDF50 struct {
	MainBuffer       []byte
	MainBuffLength   int
	Conn             net.Conn
	AnalyzedData     [][]byte
	MType            *int
	qMessageID       string
	Order            []string
	MOrder           int
	MLength          int
	eqName           string
	apiHost          string
	host             string
	connectionStatus bool
}

func (d *DymindDF50) Connect() {
	listener, err := net.Listen("tcp", d.host)
	if err != nil {
		fmt.Println("Error listening:", err)
		return
	}
	defer func() {
		d.connectionStatus = false
	}()
	fmt.Println("start server")
	for {
		d.connectionStatus = false
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("failed to accept connection:", err)
			return
		}
		d.connectionStatus = true
		d.Start(conn)
	}
	// Handle the connection concurrently
	// d.ConnectToMachine()
	// }
}

func (c DymindDF50) GetConnectionStatus() bool {
	return c.connectionStatus
}

func (c DymindDF50) Start(conn net.Conn) {
	fmt.Println("connected to client")
	// name := time.Now().Format("2006-01-02T15:04:05")
	// f, _ := os.Create(fmt.Sprintf("test/%s.txt", name))
	// fmt.Println("create file error:", f.Name())

	c.Conn = conn
	for {
		defer c.Conn.Close()

		for {
			// buffer := make([]byte, 4098)
			// _, err := c.Conn.Read(buffer)
			buffer, err := hl7.ReadBuf(c.Conn)
			if err == io.EOF {
				fmt.Println("dconnectionStatus")
				return
			} else if err != nil {
				fmt.Println("failed read buffer", err)
				return
			}
			if bytes.ContainsRune(buffer, '\u0002') {
				showLogs("STX")
				c.Conn.Write([]byte{'\u0006'})
			}
			if bytes.ContainsRune(buffer, '\013') {
				showLogs("ACK")
				buffer = bytes.ReplaceAll(buffer, []byte{'\u0002'}, []byte(""))
				// c.MainBuffer = append(c.MainBuffer, buffer...)
				c.MainBuffer = append(c.MainBuffer, buffer...)
				// c.MainBuffer = append(c.MainBuffer, []byte("\r")...)
				// c.Conn.Write(buffer)
				// f.Write(buffer)
			} else if bytes.ContainsRune(buffer, '\u001c') {
				// buffer = bytes.Replace(buffer, []byte{'\u000b'}, []byte(""), 1)
				fmt.Println("order length: ", len(c.MainBuffer))
				c.MainBuffer = append(c.MainBuffer, '\r')
				// arr :=
				// time.Sleep(time.Microsecond * 400)
				c.AnalyzedData = bytes.Split(c.MainBuffer, []byte{'\r'})

				time.Sleep(time.Second * 1)
				c.qMessageID = string(parser(c.AnalyzedData[0], []byte("|"))[9])
				c.parseReceived()
				if *c.MType == 1 {
					showLogs("Results were received. Started processing...")
					c.saveResults()
					continue
				} else if *c.MType == 0 {
					fmt.Println("gotten get order")
				}

			} else {
				c.MainBuffer = append(c.MainBuffer, buffer...)
				// f.Write(buffer)
			}

			// msgs, err := hl7.NewDecoder(conn).Messages()
			// fmt.Println("err", err)
			// for i, msg := range msgs {
			// 	msg
			// }
		}

		// responseBuffer = ""
	}
}
func convertToString(req [][]byte) []string {
	res := []string{}
	for _, item := range req {
		// fmt.Println("item", string(item))
		res = append(res, string(item))
	}
	return res
}

func (c DymindDF50) AddBuf(req []byte) {
	for i := c.MainBuffLength; i < len(req)+c.MainBuffLength; i++ {
		c.MainBuffer[i] = req[i-c.MainBuffLength]
	}
}
func (c DymindDF50) parseReceived() {
	messageType := parser(c.AnalyzedData[0], []byte{'|'})[8]
	if bytes.Contains(messageType, []byte("ORU^R01")) {
		*c.MType = 1
	} else if bytes.Contains(messageType, []byte("QRY^Q02")) {
		*c.MType = 0
	} else if bytes.Contains(messageType, []byte("ACK^Q03")) {
		*c.MType = 2
	}
	// }
}

type OBXResult struct {
	Code  string `json:"code"`
	Res   string `json:"res"`
	Name  string `json:"name"`
	Unit  string `json:"unit"`
	Norms string `json:"norms"`
	Flag  string `json:"flag"`
}



func (c DymindDF50) getOrder() []string {
	barcode := string(parser(c.AnalyzedData[3], []byte{'|'})[3])
	fmt.Println("gotten barcode", barcode)
	gender := "Male"
	// messageID := 555
	timeNow := time.Now().Format("20060102150405")
	// barcode, err := msg.Find("PID.3")
	orderFound := true
	// if err != nil {
	// 	showErrors("failed to find barcode", err)
	// }
	sex := 1
	gender = ""
	if sex == 1 {
		gender = "M"
	} else if sex == 0 {
		gender = "F"
	}
	//Maybe be there error
	birthDate := "200601020101"
	clientId := "666"
	// surname := "Dehqonov"
	patientName := "Asliddin"
	res := make([]string, 0)
	if !orderFound {
		res = c.resAcknowledgeMessage("QCK^Q02", "NF")
		return res
	}
	res = append(res, c.resAcknowledgeMessage("QCK^Q02", "OK")...)
	res = append(res, c.resAcknowledgeMessage("DSR^Q03", "OK")...)

	res = append(res, "QRD|20070723170707|R|D|1|||RD|34567743|OTH|||T|")
	res = append(res, "QRF||20070723170749|20070723170749|||RCT|COR|ALL||")
	res = append(res, "DSP|1||"+clientId+"|||")
	res = append(res, "DSP|2||256|||")
	res = append(res, "DSP|3||"+patientName+"|||")
	res = append(res, "DSP|4||"+birthDate+"|||")
	res = append(res, "DSP|5||"+gender+"|||")
	res = append(res, "DSP|6|||||")
	res = append(res, "DSP|7|||||")
	res = append(res, "DSP|8|||||")
	res = append(res, "DSP|9|||||")
	res = append(res, "DSP|10|||||")
	res = append(res, "DSP|11|||||")
	res = append(res, "DSP|12|||||")
	res = append(res, "DSP|13|||||")
	res = append(res, "DSP|14|||||")
	res = append(res, "DSP|15|||||")
	res = append(res, "DSP|16|||||")
	res = append(res, "DSP|17|||||")
	res = append(res, "DSP|18|||||")
	res = append(res, "DSP|19|||||")
	res = append(res, "DSP|20|||||")
	res = append(res, "DSP|21||"+barcode+"|||")
	res = append(res, "DSP|22|||||")
	res = append(res, "DSP|23||"+timeNow+"|||")
	res = append(res, "DSP|24||N|||")
	res = append(res, "DSP|25|||||")
	res = append(res, "DSP|26||serum|||")
	res = append(res, "DSP|27|||||")
	res = append(res, "DSP|28|||||")

	tests := []string{""}
	for i, test := range tests {
		res = append(res, "DSP|"+fmt.Sprintf("%d", i+29)+"||"+test+"^^^|||")
	}
	res = append(res, "DSC||")
	res = append(res, "\u001c")
	return res
}

func (c DymindDF50) saveResults() {
	barcode := parser(c.AnalyzedData[3], []byte{'|'})[3]
	// keys, err := msg.FindAll("OBX.3")
	// if err != nil {
	// 	showErrors("failed to find obx key", err)
	// }
	showLogs("Gotten result with barcode " + string(barcode))
	res := make([]OBXResult, 0)
	// vals, err := msg.FindAll("OBX.5")
	// if err != nil {
	// 	showErrors("failed to find obx key", err)
	// }
	obxResult := c.parseOBX()
	for _, el := range obxResult {
		codes := bytes.Split(el[3], []byte{'^'})
		res = append(res, OBXResult{
			Code:  string(codes[1]),
			Name:  string(el[3]),
			Res:   string(el[5]),
			Unit:  string(el[6]),
			Norms: string(el[7]),
			Flag:  string(el[8]),
		})
	}
	_, err := json.MarshalIndent(&res, "", "  ")
	if err != nil {
		showErrors("failed marshal indent", err)
	}

	fmt.Println("gotten results")

}

func (c DymindDF50) parseOBX() [][][]byte {
	// var OBXSections []string
	var OBXList [][][]byte
	for _, row := range c.AnalyzedData {
		// fmt.Println("row", string(row))
		if bytes.Contains(row, []byte("OBX")) {
			arr := parser(row, []byte{'|'})
			if len(arr) != 9 {
				for i := len(arr); i <= 9; i++ {
					arr = append(arr, []byte(""))
				}
				// arr = append(arr, "")
			}
			OBXList = append(OBXList, arr)
		}
	}

	// for _, row := range OBXSections {
	// 	arr := parser(row, "|")
	// 	// var string(els []strin)g
	// 	// els = append(els, arr...)
	// 	OBXList = append(OBXList, arr)
	// }

	return OBXList
}

func showLogs(str string) {
	fmt.Println("------------gotten ", str, "----------------")
}
func parser(text, delimiter []byte) [][]byte {
	return bytes.Split(bytes.ReplaceAll(text, delimiter, []byte{'!'}), []byte{'!'})
}
func showErrors(desc string, str error) {
	fmt.Println(fmt.Errorf("------------gotten %s %v----------------", desc, str))
}

func (c DymindDF50) resAcknowledgeMessage(messageType string, flag string) []string {
	var list []string
	// header := strings.Split(message[0], "|")
	// qMessageID := header[9]
	date := time.Now().Format("20060102150405")
	list = append(list, fmt.Sprint("\013MSH|^~\\&|||MLis|Lis|"+date+"||"+messageType+"|"+c.qMessageID+"|P|2.3.1||||0||ASCII|||"))
	list = append(list, "MSA|AA|"+c.qMessageID+"|Message accepted|||0|")
	if messageType == "ACK^R01" {
		list = append(list, "ERR|0|")
		list = append(list, "QAK|SR|"+flag+"|")
	}
	if messageType != "DSR^Q03" {
		list = append(list, "\u001c")
	}
	return list
}
