package monitor

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/png"
	"log"
	"net"
	"os"
	"time"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type Request struct {
	FrameHeader [4]byte
	Address     byte
	Mark        byte
	Opcode      int16
	Reserve     [2]byte
	FrameNumber int32
	TotalLength int16
	Data        []byte
	EndOfFrame  [4]byte
}

var (
	dpi      = flag.Float64("dpi", 72, "screen resolution in Dots Per Inch")
	fontfile = flag.String("fontfile", "./luxisr.ttf", "filename of the ttf font")
	hinting  = flag.String("hinting", "none", "none | full")
	size     = flag.Float64("size", 50, "font size in points")
	// size     = flag.Float64("size", 50, "font size in points")
	spacing = flag.Float64("spacing", 1.5, "line spacing (e.g. 2 means double spaced)")
	wonb    = flag.Bool("whiteonblack", false, "white text on a black background")
)

var (
	FrameHeader = []byte{0x55, 0xaa, 0x00, 0x00}
	Address     = []byte{0x01}
	Mark        = byte(0x01)
	Retention   = []byte{0x00, 0x00}
	EndOfFrame  = []byte{0x00, 0x00, 0x0d, 0x0a}
)

type Response struct {
	FrameHeader [4]byte
	Address     byte
	Mark        byte
	Opcode      int16
	Reserve     [2]byte
	FrameNumber int32
	TotalLength int16
	Data        []byte
	EndOfFrame  [4]byte
}
type Monitor struct {
	conn        *net.Conn
	Host        string
	Port        string
	IsConnected bool
}

func NewMonitor(host string, port string) *Monitor {
	return &Monitor{
		Host:        host,
		Port:        port,
		IsConnected: false,
	}
}

func (m *Monitor) Reconnect() error {
	m.IsConnected = false
	conn, err := net.Dial("tcp", m.Host+":"+m.Port)
	if err != nil {
		fmt.Println("failed to connect monitor", err.Error())
		return err
	}
	fmt.Println("TCP communication established with the device!")
	m.conn = &conn
	m.IsConnected = true
	return nil
}

func (m *Monitor) SendMessage(message string) error {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered. Error:\n", r)
		}
	}()
	// targetIP := "192.168.0.31"
	// targetPort := "10000"
	// Establish a TCP connection
	// var color = "red"
	// if strings.Contains(message, ":") {
	// 	slice := strings.Split(message, ":")
	// 	message, color = slice[0], slice[1]
	// }
	if m.conn == nil {
		fmt.Println("connection nil")
		return fmt.Errorf("failed to get connection got nil")
	}
	conn := *m.conn
	conn.SetWriteDeadline(time.Now().Add(time.Second * 5))
	conn.SetReadDeadline(time.Now().Add(time.Second * 5))
	err := paramQuery(conn)
	if err != nil {
		return fmt.Errorf("failed to read connection header %v", err)
	}
	res, err := makeImage(message)
	if err != nil {
		fmt.Println("failed to make image ", err)
		return fmt.Errorf("failed make image %v", err)
	}
	err = sendRequest(conn, res)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to send request: %v", err)
	}
	return nil
}

func paramQuery(conn net.Conn) error {
	var header []byte
	header = append(header, FrameHeader...)
	header = append(header, Address...)
	header = append(header, byte(Mark))
	opcodeSlice := make([]byte, 2)
	opcodeSlice[0] = 0x00
	opcodeSlice[1] = 0xa3

	header = append(header, opcodeSlice...)
	header = append(header, Retention...)
	frameNumSlice := make([]byte, 4)
	frameNumSlice[0] = 0x00
	frameNumSlice[1] = 0x00
	frameNumSlice[2] = 0x00
	frameNumSlice[3] = 0x00

	header = append(header, frameNumSlice...)

	totalLengthSlice := make([]byte, 4)

	totalLengthSlice[0] = 0x08
	header = append(header, totalLengthSlice...)
	frameLenSlice := make([]byte, 2)
	frameLenSlice[0] = 0x08
	frameLenSlice[1] = 0x00
	header = append(header, frameLenSlice...)
	header = append(header, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00)
	header = append(header, EndOfFrame...)
	k, err := conn.Write(header)
	if err != nil || k == 0 {
		fmt.Println("failed to write header", err)
	}
	var res []byte
	_, err = conn.Read(res)
	if err != nil {
		fmt.Println("failed to read", err)
		return err
	}
	time.Sleep(time.Microsecond * 50)
	return nil
}

func sendRequest(conn net.Conn, pixels []byte) error {
	var header []byte
	header = append(header, FrameHeader...)
	header = append(header, Address...)
	header = append(header, byte(Mark))
	opcodeSlice := make([]byte, 2)
	opcodeSlice[0] = 0x00
	opcodeSlice[1] = 0xda

	// binary.LittleEndian.PutUint16(opcodeSlice,)

	header = append(header, opcodeSlice...)
	header = append(header, Retention...)
	frameNumSlice := make([]byte, 4)
	// frameNumSlice[0] = 0x0c
	header = append(header, frameNumSlice...)
	// header = append(header, frameNumber...)
	totalLengthSlice := make([]byte, 4)
	// binary.LittleEndian.PutUint32(totalLengthSlice, uint32(len(data)))
	//!!! change legth to 40 if new and 20 for old
	totalLengthSlice[0] = 0x82
	totalLengthSlice[1] = 0x40
	// totalLengthSlice[0] = 0x00
	// totalLengthSlice[0] = 0x00

	header = append(header, totalLengthSlice...)
	frameLenSlice := make([]byte, 2)
	frameLenSlice[0] = 0x00
	frameLenSlice[1] = 0x02
	// binary.LittleEndian.PutUint16(frameLenSlice, uint16(len(pixels)))
	header = append(header, frameLenSlice...)
	space := make([]byte, 16)
	for i, _ := range space {
		space[i] = 0x75
	}
	var data []byte
	data = append(data, header...)
	data = append(data, space...)
	//Set width ,height and color
	//!!! set 0x003 to 0x001 if new and 0x001 if old after 0x03
	data = append(data, 0x80, 0x00, 0x40, 0x00, 0x03, 0x03, 0x01, 0x02)
	freeSpace := make([]byte, 8)
	data = append(data, freeSpace...)
	data = append(data, totalLengthSlice...)
	data = append(data, freeSpace...)
	//!!! set 0x40 to new and 0x20 to old after 0x56
	data = append(data, 0x00, 0x56, 0x40, 0x00, 0x00)
	//Fixed 0X02
	data = append(data, 0x02, 0x01, 0x00, 0x00, 0x01)
	sp := make([]byte, 16)
	data = append(data, sp...)
	//Regional attributes
	//!!!Set 0x40 to new and 0x20 to old after 0x3c
	data = append(data, 0x01, 0x3c, 0x40, 0x00, 0x00)
	//Fixed 0x05
	data = append(data, 0xe5, 0x01, 0x01)
	data = append(data, make([]byte, 4)...)
	//pixels
	data = append(data, 0x00, 0x00, 0x00, 0x00, 0x7f, 0x00, 0x3f, 0x00)
	data = append(data, make([]byte, 12)...)
	data = append(data, 0x01)
	//!!!Set 0x40 after 0x01c if new and old 0x20
	data = append(data, 0x1c, 0x40, 0x00, 0x00)
	aType := 0x01
	speed := 0x01
	data = append(data, 0x01, 0x01, 0x00, byte(aType), 0x00, byte(speed), byte(speed))
	data = append(data, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00)
	//size
	data = append(data, 0x7f, 0x00, 0x3f, 0x00, 0x01, 0x00)
	data = append(data, make([]byte, 4)...)
	lenData := len(data) - len(header)
	//Data body
	headerCp := make([]byte, len(header))
	copy(headerCp, header)
	var sumLen int
	data = append(data, pixels[:512-lenData]...)
	sumLen = sumLen + len(pixels[:512-lenData])
	pixels = pixels[512-lenData:]

	data = append(data, EndOfFrame...)
	k, err := conn.Write(data)
	if err != nil || k == 0 {
		fmt.Println("failed to write first part of data", err)
		return err
	}
	res := make([]byte, 16)
	time.Sleep(time.Millisecond * 50)

	var n int
	attemp := 0

	for n == 0 && attemp < 5 {
		n, err = conn.Read(res)
		if err != nil {
			fmt.Println("got error while reading", err)
			return err
		}
		attemp++
	}
	count := len(pixels) / 512
	var index []int

	for i := 0; i < count; i++ {
		var req []byte
		headerCp[10] = byte(i + 1)
		index = append(index, i+1)
		req = append(req, headerCp...)
		req = append(req, pixels[:512]...)
		sumLen = sumLen + len(pixels[:512])
		req = append(req, EndOfFrame...)
		pixels = pixels[512:]
		// conn.SetWriteDeadline(time.Now().Add(time.Microsecond * 600))
		k, err = conn.Write(req)
		if err != nil || k == 0 {
			fmt.Println("failed to write request", err)
			return err
		}

		res := make([]byte, 100)
		var n int
		attemp := 0
		for n == 0 && attemp < 2 {
			n, err = conn.Read(res)
			if err != nil {
				fmt.Println("got error while reading", err)
				return err
			}
			attemp++
			time.Sleep(time.Millisecond * 50)
		}

	}
	var req []byte
	headerCp[10] = byte(count + 1)
	index = append(index, count+1)

	headerCp[18] = 0x82
	headerCp[19] = 0x00
	req = append(req, headerCp...)
	req = append(req, pixels[:]...)
	sumLen = sumLen + len(pixels)
	req = append(req, EndOfFrame...)
	k, err = conn.Write(req)
	if err != nil || k == 0 {
		fmt.Println("failed write frame", err)
		return err
	}
	time.Sleep(time.Millisecond * 50)

	// for n == 0 {
	n, err = conn.Read(res)
	if err != nil {
		fmt.Println("got error while reading", err)
		return err
	}
	fmt.Println("third read")
	time.Sleep(time.Millisecond * 50)
	n, err = conn.Read(res)
	if err != nil {
		fmt.Println("got error while reading", err)
		return err
	}

	return nil
}

func makeImage(str string) ([]byte, error) {
	fontBytes, err := os.ReadFile(*fontfile)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	f, err := truetype.Parse(fontBytes)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	fg := image.Black
	modSize := float64(0)

	if len([]rune(str)) <= 3 {
		modSize = 56
	} else {
		modSize = 48
	}
	const imgW, imgH = 128, 64
	rgba := image.NewRGBA(image.Rect(0, 0, imgW, imgH))

	h := font.HintingNone
	d := &font.Drawer{
		Dst: rgba,
		Src: fg,
		Face: truetype.NewFace(f, &truetype.Options{
			Size:    modSize,
			DPI:     *dpi,
			Hinting: h,
		}),
	}
	textBounds, _ := d.BoundString(str)

	textWidth := (textBounds.Max.X - textBounds.Min.X).Ceil()
	textHeight := (textBounds.Max.Y - textBounds.Min.Y).Ceil()
	textX := (imgW - textWidth) / 2
	textY := (imgH - textHeight) / 2
	d.Dot = fixed.P(textX, imgH-textY-3)

	d.DrawString(str)
	var buffer bytes.Buffer
	err = png.Encode(&buffer, rgba)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	resSlice := []byte{}
	for y := 0; y < imgH; y++ {
		for x := 0; x < imgW; x++ {
			r, g, b, a := rgba.At(x, y).RGBA()
			// sum := 0
			// if (r + g + b + a) != 0 {
			// 	// sum = 1
			// }
			if (r + g + b + a) == 0 {
				resSlice = append(resSlice, 0x00, 0x00)
			} else {
				// if x%2 == 0 {
				// summa := r + g + b + a
				// rgba := fmt.Sprintf("")
				// resSlice = append(resSlice, byte(rgbaConverter(MyRgba{R: r, G: g, B: b, A: a})))
				// if summa > 65000 {
				// 	resSlice = append(resSlice, 0x20, 0x30)
				// } else if summa > 58000 {
				// 	resSlice = append(resSlice, 0x20, 0x30)
				// } else if summa > 34000 {
				// 	resSlice = append(resSlice, 0x20, 0x39)
				// } else {
				// 	resSlice = append(resSlice, 0x65, 0x70)
				// }
				// } else {
				// resSlice = append(resSlice, 0x00, 0x00)
				// }
				resSlice = append(resSlice, 0x00, 0x01)
				// }
			}
		}
	}
	return resSlice, nil
}

type MyRgba struct {
	R, G, B, A uint32
}

func rgbaConverter(rgba MyRgba) uint16 {
	r := uint32(rgba.A) >> 3
	g := uint32(rgba.G) >> 2
	b := uint32(rgba.B) >> 3
	return uint16((r << 11) | (g << 5) | b)
}

//small	smal mid 	big  bigger
//0x01,	0x08,0x09,0x48,0x49-blue
//0x02,	0x10,0x11,0x12			-green
//											0x55-white
//											0x65-pink
//											0x30-yellow
