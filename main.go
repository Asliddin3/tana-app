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
	"fyne.io/fyne/v2/layout"
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
	// if conf.Type == "lis" {
	// lis := t.showEquipments(conf)
	lis := t.showEquipmentsInTable(conf)
	// } else if conf.Type == "monitor" {
	// conn, err := net.Dial("tcp", t.Conf.MonitorHost+":10000")
	// if err != nil {
	// 	fmt.Println("failed to connect monitor", err.Error())
	// 	return
	// }
	// defer conn.Close()

	m := monitor.NewMonitor(t.Conf.MonitorHost, "10000")
	m.IsConnected = true
	isConn := make(chan bool)
	queue := make(chan string)
	go t.keepSocketIoServer(m, isConn, queue)
	monitor := t.showMonitorStatus(m, isConn, queue)
	// } else if conf.Type == "printer" {
	// }
	tabs := container.NewAppTabs(
		container.NewTabItem("Лаборотория", lis),
		container.NewTabItem("Монитор", monitor),
	)

	//tabs.Append(container.NewTabItemWithIcon("Home", theme.HomeIcon(), widget.NewLabel("Home tab")))

	tabs.SetTabLocation(container.TabLocationTop)
	w.Resize(fyne.NewSize(1000, 600))
	w.SetContent(tabs)
	w.ShowAndRun()
}

func (t *TanaApp) keepSocketIoServer(m *monitor.Monitor, isConn chan bool, queue chan string) {
	m.Reconnect()
	for {
		socket.EstablishSocketIOServer(t.Conf, m, isConn, queue)
		time.Sleep(1 * time.Second)
	}
}

func (t *TanaApp) showEquipments(conf tools.ConfigFile) *fyne.Container {
	title := canvas.NewText("	  Подключенные оборудование", color.Black)
	equipments := equipments.ConnectToEquipments(conf)
	title.TextSize = 30
	title.TextStyle.Bold = true
	title.Move(fyne.NewPos(70, 10))
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
		equipment.TextSize = 25
		equipment.Resize(fyne.NewSize(200, 20))
		// equipmentText[i] = equipment
		color := canvas.NewRectangle(t.StartCol)
		color.Resize(fyne.NewSize(200, 20))
		server := eq.Server
		lab := canvas.NewText("Отключен", t.StartCol)
		lab.TextSize = 25
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

	return mainBox
}

const HEIGHT float32 = 100

func (t *TanaApp) showEquipmentsInTable(conf tools.ConfigFile) fyne.CanvasObject {
	title := canvas.NewText("	  Подключенные оборудование", color.Black)
	equipments := equipments.ConnectToEquipments(conf)
	// data := make([][]fyne.CanvasObject, len(equipments))
	title.TextSize = 35
	title.TextStyle.Bold = true
	title.Move(fyne.NewPos(60, 20))
	space := canvas.NewText("   ", color.Black)
	space.TextSize = 35
	// equipmentText := make([]fyne.CanvasObject, len(equipments))
	// equipmentStatus := make([]fyne.CanvasObject, len(equipments))
	// textSize := 40
	btn := widget.NewButton("Reconnect", func() {
		fmt.Println("Gotten reconnect")
	})
	rows := make([]fyne.CanvasObject, len(equipments)+1)
	// box3 := container.NewVBox(space)
	rows[0] = container.NewGridWithColumns(1, title)
	for i, eq := range equipments {
		equipment := canvas.NewText(fmt.Sprintf("   %s", eq.Name), color.Black)
		equipment.TextSize = 25
		equipment.Resize(fyne.NewSize(250, 50))
		// equipmentText[i] = equipment
		color := canvas.NewRectangle(t.StartCol)
		color.Resize(fyne.NewSize(250, 50))
		server := eq.Server
		lab := canvas.NewText("Отключен", t.StartCol)
		lab.TextSize = 30
		lab.Resize(fyne.NewSize(200, 50))
		// btn := widget.NewButton("Start", func() {
		// })
		// rectangle := canvas.NewRectangle(t.StartCol)
		// conBtn := container.NewStack(rectangle)
		btn := widget.NewButton("Подключить", func() {
			stopCh := make(chan struct{})
			// btn.SetText("Идет подключение...")
			// btn.Refresh()
			// btn.Disable()
			// rectangle.FillColor = t.Connected
			// btn.Hidden = true
			btn.Disable()
			btn.Refresh()
			fmt.Println("btn")
			go t.connect(lab, eq.Server, stopCh)
			go t.enableBtn(btn, stopCh)
		})
		// render := btn.CreateRenderer()

		// btn.Importance = widget.LowImportance
		// rectangle.StrokeColor = t.Connecting
		// conBtn.Add(btn)
		// btn.Hidden = false
		// btn.BaseWidget = widgetB
		btn.Resize(fyne.NewSize(100, 60))
		go t.checkAnalizatorStatus(lab, server)
		// go t.startConnection(lab, server)
		// equipmentStatus[i] = lab
		// data[i] = []fyne.CanvasObject{equipment, lab}
		// con.Add(equipment, lab, btn)
		rows[i+1] = container.NewGridWithColumns(3, equipment, lab, btn)
	}

	// tableData := widget.NewTable(
	// 	func() (int, int) {
	// 		return len(data), len(data[0])
	// 	},
	// 	func() fyne.CanvasObject {
	// 		c := container.NewWithoutLayout()
	// 		c.Resize(fyne.NewSize(500, 500))
	// 		r := canvas.NewRectangle(color.White)
	// 		r.SetMinSize(fyne.NewSize(0, 0))
	// 		r.Resize(fyne.NewSize(0, 80))
	// 		c.Add(r)
	// 		return c
	// 	},
	// 	func(cell widget.TableCellID, o fyne.CanvasObject) {
	// 		container := o.(*fyne.Container)
	// 		var obj fyne.CanvasObject = data[cell.Row][cell.Col]
	// 		container.Add(obj)
	// 		container.Refresh()
	// 	})

	// setDefaultColumnsWidth(tableData)
	// tableData.Resize(fyne.NewSize(1000, 1000))
	// container.NewGridWrap(fyne.NewSize(400, 1000))
	// mainBox := container.NewVBox(title, tableData)

	// box := container.NewVBox(title, tableData)
	// content := container.NewBorder(title, nil, nil, nil, tableData)
	// grid := container.New(layout.NewGridWrapLayout(fyne.NewSize(1000, 200)), title)
	// parentGrid := container.NewGridWithRows(len(equipments), rows...)
	// con := layout.NewGridLayoutWithRows(len(rows))
	// con.Layout(rows, fyne.NewSize(900, 200))
	tab := container.New(layout.NewGridWrapLayout(fyne.NewSize(800, 50)), rows...)
	// con := container.New(layout.NewGridLayoutWithRows(len(rows)), rows...)
	// con.Resize(fyne.NewSize(1000, 900))
	return tab
}
func setDefaultColumnsWidth(table *widget.Table) {
	colWidths := []float32{300, 300}
	for idx, colWidth := range colWidths {
		table.SetColumnWidth(idx, colWidth)
	}
}
func (t *TanaApp) startConnection(text *canvas.Text, server equipments.EquipmentMethods) {
	for {
		text.Text = "Отключен"
		text.Refresh()
		server.Connect()
		time.Sleep(time.Second * 2)
	}
}
func (t *TanaApp) enableBtn(btn *widget.Button, stopCh chan struct{}) {
	<-stopCh
	btn.Disable()
	btn.SetText("Подключить")
	btn.Refresh()
}
func (t *TanaApp) connect(text *canvas.Text, server equipments.EquipmentMethods, stopCh chan struct{}) {
	// text.Text = "Подключен"
	// text.Color = t.Connected
	// text.Refresh()
	server.Connect()
	text.Text = "Отключен"
	text.Color = t.StartCol
	text.Refresh()
	stopCh <- struct{}{}
}
func (t *TanaApp) checkSocketStatus(isConn chan bool, text *canvas.Text) {
	// Create a channel to receive signals for stopping the ticker
	for {

		val, ok := <-isConn
		if ok == true && val == true {
			text.Text = "Подключен"
			text.Color = t.Connected
			text.Refresh()
		} else {
			text.Text = "Отключен"
			text.Color = t.StartCol
			text.Refresh()
		}
		// time.Sleep(time.Second * 4)
	}
}

func (t *TanaApp) checkAnalizatorStatus(text *canvas.Text, server equipments.EquipmentMethods) {
	time.Sleep(time.Second * 3)
	// Create a channel to receive signals for stopping the ticker
	for {
		status := server.GetConnectionStatus()
		if status {
			text.Text = "Подключен"
			text.Color = t.Connected
			text.Refresh()
		} else if !status {
			text.Text = "Отключен"
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
			text.Text = "Подключен"
			text.Color = t.Connected
			conn.Close()
		} else {
			text.Text = "Отключен"
			text.Color = t.StartCol
		}
		text.Refresh()
	}
}

func (t *TanaApp) showMonitorStatus(m *monitor.Monitor, isConnected chan bool, queue chan string) *fyne.Container {
	title := canvas.NewText("	 Статус монитора", color.Black)

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
	equipment := canvas.NewText("Сервер", color.Black)
	equipment.TextSize = 25
	equipment.Resize(fyne.NewSize(200, 40))
	// equipmentText[i] = equipment
	lab := canvas.NewText("Отключен", t.StartCol)
	lab.TextSize = 25

	hostServer := canvas.NewText("Монитор", color.Black)
	hostServer.TextSize = 25
	hostServer.Resize(fyne.NewSize(200, 40))
	numberServer := canvas.NewText("Очередь: ", color.Black)
	numberServer.TextSize = 25
	numberServer.Resize(fyne.NewSize(200, 40))

	// equipmentText[i] = equipment
	// color := canvas.NewRectangle(t.StartCol)
	// color.Resize(fyne.NewSize(300, 30))
	serverLab := canvas.NewText("Отключен", t.StartCol)
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
	// grid := container.NewGridWithRows(3, Подключен, Отключен)
	// box2.Refresh()
	// box1 := container.NewVBox(equipmentText...)
	// horbox := container.NewVBox(space, box1, space, space, box2)
	// horbox.Move(fyne.NewPos(-300, 20))

	return mainBox
}

func updateNumber(queue chan string, text *canvas.Text) {
	for {
		number := <-queue
		text.Text = fmt.Sprintf("Очередь: %s", number)
		text.Refresh()
	}
}
