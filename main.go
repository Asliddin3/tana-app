package main

import (
	"ZettaGroup/Tana-App/equipments"
	"ZettaGroup/Tana-App/monitor"
	socket "ZettaGroup/Tana-App/server"
	"ZettaGroup/Tana-App/tools"
	"fmt"
	"image/color"
	"net"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type TanaApp struct {
	StartCol   color.NRGBA
	Connecting color.NRGBA
	Connected  color.NRGBA
	Box        *fyne.Container
	App        fyne.App
	Conf       tools.ConfigFile
}

func NewTanaApp(conf tools.ConfigFile) TanaApp {
	return TanaApp{
		StartCol: color.NRGBA{
			R: 255,
			A: 200,
		},
		Connecting: color.NRGBA{
			R: 100,
			G: 100,
			B: 255,
			A: 255,
		},
		Connected: color.NRGBA{
			R: 30,
			G: 160,
			B: 30,
			A: 230,
		},
		Conf: conf,
	}
}
func main() {
	a := app.New()
	w := a.NewWindow("Tana")
	conf := tools.GetConfig()
	t := NewTanaApp(conf)
	t.App = a
	a.Settings().SetTheme(theme.LightTheme())
	if conf.Type == "lis" {
		w.Resize(fyne.NewSize(800, 500))
		w = t.showEquipments(w, conf)
	} else if conf.Type == "monitor" {
		w.Resize(fyne.NewSize(400, 300))
		conn, err := net.Dial("tcp", t.Conf.MonitorHost+":10000")
		if err != nil {
			fmt.Println("failed to connect monitor", err.Error())
			return
		}
		defer conn.Close()

		m := monitor.NewMonitor(conn, t.Conf.MonitorHost, "10000")
		m.IsConnected = true
		isConn := make(chan bool)
		queue := make(chan string)
		go t.keepSocketIoServer(m, isConn, queue)
		w = t.showMonitorStatus(w, m, isConn, queue)
	} else if conf.Type == "printer" {
		w.Resize(fyne.NewSize(800, 500))
	}

	w.ShowAndRun()
}

func (t *TanaApp) keepSocketIoServer(m *monitor.Monitor, isConn chan bool, queue chan string) {
	for {
		socket.EstablishSocketIOServer(t.Conf, m, isConn, queue)
		time.Sleep(1 * time.Second)
	}
}

func (t *TanaApp) showEquipments(w fyne.Window, conf tools.ConfigFile) fyne.Window {
	title := canvas.NewText("	   Connected Equipments", color.Black)

	title.TextSize = 40
	title.TextStyle.Bold = true
	title.Move(fyne.NewPos(100, 30))
	equipments := equipments.ConnectToEquipments(conf)
	space := canvas.NewText("   ", color.Black)
	space.TextSize = 40
	// equipmentText := make([]fyne.CanvasObject, len(equipments))
	// equipmentStatus := make([]fyne.CanvasObject, len(equipments))
	// textSize := 40
	box1 := container.NewVBox(title)
	box2 := container.NewVBox(space)
	btn := widget.NewButton("Reconnect", func() {
		fmt.Println("Gotten reconnect")
	})

	// box3 := container.NewVBox(space)
	for _, eq := range equipments {
		equipment := canvas.NewText(eq.Name, color.Black)
		equipment.TextSize = 30
		equipment.Resize(fyne.NewSize(300, 40))
		// equipmentText[i] = equipment
		color := canvas.NewRectangle(t.StartCol)
		color.Resize(fyne.NewSize(200, 30))
		server := eq.Server
		lab := canvas.NewText("Offline", t.StartCol)
		lab.TextSize = 30
		// btn := widget.NewButton("Start", func() {
		// })
		go t.startConnection(lab, server)
		go t.checkAnalizatorStatus(lab, server)
		// equipmentStatus[i] = lab
		box1.Add(equipment)
		box2.Add(lab)

	}

	container.NewGridWrap(fyne.NewSize(400, 800))
	mainBox := container.NewHBox(space, box1, space, box2, btn)

	w.SetContent(mainBox)
	return w
}

func (t *TanaApp) startConnection(text *canvas.Text, server equipments.EquipmentMethods) {
	for {
		text.Text = "Offline"
		text.Refresh()
		server.Connect()
		time.Sleep(time.Second * 2)
	}
}
func (t *TanaApp) checkSocketStatus(isConn chan bool, text *canvas.Text) {
	// Create a channel to receive signals for stopping the ticker
	for {

		val, ok := <-isConn
		if ok == true && val == true {
			text.Text = "Online"
			text.Color = t.Connected
			text.Refresh()
		} else {
			text.Text = "Offline"
			text.Color = t.StartCol
			text.Refresh()
		}
		// time.Sleep(time.Second * 4)
	}
}

func (t *TanaApp) checkAnalizatorStatus(text *canvas.Text, server equipments.EquipmentMethods) {
	time.Sleep(time.Second * 5)
	// Create a channel to receive signals for stopping the ticker
	for {
		status := server.GetConnectionStatus()
		if status {
			text.Text = "Online"
			text.Color = t.Connected
			text.Refresh()
		} else if !status {
			text.Text = "Offline"
			text.Color = t.StartCol
			text.Refresh()
		}
		time.Sleep(time.Second * 4)
	}
}
func (t *TanaApp) checkMonitorConn(text *canvas.Text, m *monitor.Monitor) {
	for {
		time.Sleep(time.Second * 2)
		conn, err := net.DialTimeout("tcp", m.Host+":"+m.Port, time.Second*2)
		if err == nil {
			text.Text = "Online"
			text.Color = t.Connected
			conn.Close()
		} else {
			text.Text = "Offline"
			text.Color = t.StartCol
		}
		text.Refresh()
	}
}

func (t *TanaApp) showMonitorStatus(w fyne.Window, m *monitor.Monitor, isConnected chan bool, queue chan string) fyne.Window {
	title := canvas.NewText("	 Monitor Connection", color.Black)

	title.TextSize = 25
	title.TextStyle.Bold = true
	title.Move(fyne.NewPos(20, 10))
	space := canvas.NewText("   ", color.Black)
	space.TextSize = 25
	// equipmentText := make([]fyne.CanvasObject, len(equipments))
	// equipmentStatus := make([]fyne.CanvasObject, len(equipments))
	// textSize := 40
	box1 := container.NewVBox(title)
	box2 := container.NewVBox(space)
	// for _, eq := range equipments {
	equipment := canvas.NewText("Server", color.Black)
	equipment.TextSize = 25
	equipment.Resize(fyne.NewSize(200, 40))
	// equipmentText[i] = equipment
	lab := canvas.NewText("Offline", t.StartCol)
	lab.TextSize = 25

	hostServer := canvas.NewText("Monitor", color.Black)
	hostServer.TextSize = 25
	hostServer.Resize(fyne.NewSize(200, 40))
	numberServer := canvas.NewText("Navbat: ", color.Black)
	numberServer.TextSize = 25
	numberServer.Resize(fyne.NewSize(200, 40))

	// equipmentText[i] = equipment
	// color := canvas.NewRectangle(t.StartCol)
	// color.Resize(fyne.NewSize(300, 30))
	serverLab := canvas.NewText("Offline", t.StartCol)
	serverLab.TextSize = 25
	// btn := widget.NewButton("Start", func() {
	// })
	// go t.checkServerConn(lab, isConnected)
	go t.checkMonitorConn(serverLab, m)
	go t.checkSocketStatus(isConnected, lab)
	go updateNumber(queue, numberServer)
	// equipmentStatus[i] = lab
	box1.Add(equipment)
	box2.Add(lab)
	box1.Add(hostServer)
	box2.Add(serverLab)
	box1.Add(numberServer)

	container.NewGridWrap(fyne.NewSize(400, 800))
	mainBox := container.NewHBox(space, box1, space, box2)
	// grid := container.NewGridWithRows(3, online, offline)
	// box2.Refresh()
	// box1 := container.NewVBox(equipmentText...)
	// horbox := container.NewVBox(space, box1, space, space, box2)
	// horbox.Move(fyne.NewPos(-300, 20))

	w.SetContent(mainBox)
	return w
}

func updateNumber(queue chan string, text *canvas.Text) {
	for {
		number := <-queue
		text.Text = fmt.Sprintf("Navbat: %s", number)
		text.Refresh()
	}
}
