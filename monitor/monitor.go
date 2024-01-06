package monitor

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"log"
	"net"
	"strings"
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
	Host string
	Port string
}

func NewMonitor(host, port string) *Monitor {
	return &Monitor{
		Host: host,
		Port: port,
	}
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
	conn, err := net.Dial("tcp", m.Host+":"+m.Port)
	if err != nil {
		fmt.Println("failed to connect monitor", err.Error())
		return err
	}
	defer conn.Close()
	fmt.Println("TCP communication established with the device!")
	var color = "red"
	if strings.Contains(message, ":") {
		slice := strings.Split(message, ":")
		message, color = slice[0], slice[1]
	}
	paramQuery(conn)
	res, err := makeImage(message, color)
	if err != nil {
		fmt.Println("failed to make image ", err)
		return fmt.Errorf("failed make image %v", err)
	}
	err = sendRequest(conn, res)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	time.Sleep(time.Second * 1)
	conn.Close()
	return nil
}

func paramQuery(conn net.Conn) {
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
	}
	time.Sleep(time.Microsecond * 500)
}

func sendRequest(conn net.Conn, pixels []byte) error {
	var header []byte
	header = append(header, FrameHeader...)
	header = append(header, Address...)
	header = append(header, byte(Mark))
	opcodeSlice := make([]byte, 2)
	opcodeSlice[0] = 0x00
	opcodeSlice[1] = 0xda

	header = append(header, opcodeSlice...)
	header = append(header, Retention...)
	frameNumSlice := make([]byte, 4)

	header = append(header, frameNumSlice...)
	// header = append(header, frameNumber...)
	totalLengthSlice := make([]byte, 4)

	totalLengthSlice[0] = 0x82
	totalLengthSlice[1] = 0x20

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
	data = append(data, 0x80, 0x00, 0x40, 0x00, 0x03, 0x001, 0x01, 0x02)
	freeSpace := make([]byte, 8)
	data = append(data, freeSpace...)
	data = append(data, totalLengthSlice...)
	data = append(data, freeSpace...)
	data = append(data, 0x00, 0x56, 0x20, 0x00, 0x00)
	//Fixed 0X02
	data = append(data, 0x02, 0x01, 0x00, 0x00, 0x01)
	sp := make([]byte, 16)
	data = append(data, sp...)
	//Regional attributes
	data = append(data, 0x01, 0x3c, 0x20, 0x00, 0x00)
	//Fixed 0x05
	data = append(data, 0xe5, 0x01, 0x01)
	data = append(data, make([]byte, 4)...)
	//pixels
	data = append(data, 0x00, 0x00, 0x00, 0x00, 0x7f, 0x00, 0x3f, 0x00)
	data = append(data, make([]byte, 12)...)
	data = append(data, 0x01)
	data = append(data, 0x1c, 0x20, 0x00, 0x00)
	//Type of animation
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
	time.Sleep(time.Millisecond * 200)

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
		conn.SetWriteDeadline(time.Now().Add(time.Second * 3))
		k, err = conn.Write(req)
		if err != nil || k == 0 {
			fmt.Println("failed to write request", err)
			return err
		}
		time.Sleep(time.Millisecond * 200)

		res := make([]byte, 16)
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
	time.Sleep(time.Millisecond * 200)

	for n == 0 {
		n, err = conn.Read(res)
		if err != nil {
			fmt.Println("got error while reading", err)
			return err
		}
	}
	if err != nil {
		fmt.Println("failed to write request", err)
		return err
	}
	return nil
}

var colors = map[string]byte{
	"red":        0x01,
	"green":      0x02,
	"yellow":     0x03,
	"blue":       0x04,
	"pink":       0x05,
	"blue-green": 0x06,
	"white":      0x07,
}

func makeImage(str string, color string) ([]byte, error) {
	fontBytes, err := ioutil.ReadFile(*fontfile)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	f, err := truetype.Parse(fontBytes)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	modSize := float64(0)

	if len([]rune(str)) <= 3 {
		modSize = 56
	} else {
		modSize = 48
	}
	// size = &modSize

	fg := image.Black

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
			if (r + g + b + a) == 0 {
				resSlice = append(resSlice, 0x00)
			} else {
				resSlice = append(resSlice, colors[color])
			}
		}
	}
	return resSlice, nil
}
