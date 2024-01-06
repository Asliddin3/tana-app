package equipments

import (
	"ZettaGroup/Tana-App/tools"
	"fmt"

	"go.bug.st/serial"
)

type Equipment struct {
	Name     string
	IsOnline bool
	Server   EquipmentMethods
}

// func Register() []Equipment {
// 	file, err := os.Open("./config.txt")
// 	if err != nil {
// 		fmt.Println("Error opening file:", err)
// 		return []Equipment{}
// 	}
// 	defer file.Close()
// 	num := 123.123123123
// 	fmt.Println("num", math.Round(num*100)/100)
// 	// Decode the JSON data into a struct
// 	// var conf map[string]interface{}
// 	conf := ConfigFile{}
// 	decoder := json.NewDecoder(file)
// 	err = decoder.Decode(&conf)
// 	if err != nil {
// 		fmt.Println("Error decoding JSON:", err)
// 		return []Equipment{}
// 	}

//		return ConnectToEquipments(conf)
//	}
type ConfigFile struct {
	Host      string            `json:"host"`
	Equipment []EquipmentConfig `json:"equipments"`
}
type EquipmentConfig struct {
	Name      string `json:"name"`
	Equipment string `json:"equipment"`
	Type      string `json:"type"`
	BaudRate  int    `json:"baudRate"`
	Parity    string `json:"parity"`
	StopBit   int    `json:"stopBit"`
	DataBits  int    `json:"dataBits"`
	Host      string `json:"host"`
}

type EquipmentMethods interface {
	Connect()
	GetConnectionStatus() bool
}

func ConnectToEquipments(config tools.ConfigFile) []Equipment {
	apiHost := config.Host
	equipments := config.Equipment
	var equipmentArr []Equipment
	for _, val := range equipments {
		// eq := val.(map[string]interface{})
		equipment := val.Equipment
		name := val.Name
		connT := val.Type
		var server EquipmentMethods
		if equipment == "Maglumi800" || equipment == "Biossays240" {
			if connT == "comport" {
				server = NewAsServerMaglumi800Com(name, apiHost, val)
			} else {
				server = NewAsServerMaglumi800Lan(name, apiHost, val)
			}
		} else if equipment == "Iflash1200" {
			if connT == "comport" {
				server = NewAsServerIflash1200Com(name, apiHost, val)
			} else {
				server = NewAsServerIflash1200Lan(name, apiHost, val)

			}
		} else if equipment == "DymindDF50" {
			server = NewDymindDF50AsServer(name, apiHost, val.Host)
		} else if equipment == "Cobas-c311" {
			server = NewAsCobas311Com(name, apiHost, val)
		} else if equipment == "Cobas-e411" {
			server = NewAsCobas411Com(name, apiHost, val)
		} else {
			fmt.Println("unexpected equipment")
			continue
		}
		equipmentArr = append(equipmentArr, Equipment{
			Name:   name,
			Server: server,
		})
	}
	return equipmentArr
}

func CreateMode(conf tools.EquipmentConfig) *serial.Mode {
	mode := &serial.Mode{
		BaudRate: conf.BaudRate,
		DataBits: conf.DataBits,
	}
	stopBit := conf.StopBit
	if stopBit == 1 {
		mode.StopBits = serial.OneStopBit
	} else if stopBit == 2 {
		mode.StopBits = serial.TwoStopBits
	}
	parity := conf.Parity
	if parity == "odd" {
		mode.Parity = serial.OddParity
	} else if parity == "even" {
		mode.Parity = serial.EvenParity
	} else if parity == "spaced" {
		mode.Parity = serial.SpaceParity
	} else if parity == "mark" {
		mode.Parity = serial.MarkParity
	} else {
		mode.Parity = serial.NoParity
	}
	return mode
}
