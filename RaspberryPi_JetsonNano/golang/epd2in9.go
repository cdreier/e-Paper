package epaper

import (
	"image"
	"image/color"
	"log"
)

type EPD2in9 struct {
	pi     *Raspi
	width  int
	height int
}

var fullUpdate = []byte{
	0x50, 0xAA, 0x55, 0xAA, 0x11, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0xFF, 0xFF, 0x1F, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
}

var partialUpdate = []byte{
	0x10, 0x18, 0x18, 0x08, 0x18, 0x18,
	0x08, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x13, 0x14, 0x44, 0x12,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
}

func NewEPD2in9() *EPD2in9 {
	epd := new(EPD2in9)
	epd.width = 128
	epd.height = 255

	epd.initialize()

	return epd
}

func (epd *EPD2in9) initialize() {
	epd.reset()

	epd.SendCommand(0x01) // DRIVER_OUTPUT_CONTROL
	epd.SendData(byte((epd.height - 1) & 0xFF))
	epd.SendData(byte(((epd.height - 1) >> 8) & 0xFF))
	epd.SendData(0x00) // GD = 0 SM = 0 TB = 0

	epd.SendCommand(0x0C) // BOOSTER_SOFT_START_CONTROL
	epd.SendData(0xD7)
	epd.SendData(0xD6)
	epd.SendData(0x9D)

	epd.SendCommand(0x2C) // WRITE_VCOM_REGISTER
	epd.SendData(0xA8)    // VCOM 7C

	epd.SendCommand(0x3A) // SET_DUMMY_LINE_PERIOD
	epd.SendData(0x1A)    // 4 dummy lines per gate

	epd.SendCommand(0x3B) // SET_GATE_TIME
	epd.SendData(0x08)    // 2us per line

	epd.SendCommand(0x11) // DATA_ENTRY_MODE_SETTING
	epd.SendData(0x03)    // X increment Y increment

	epd.SendCommand(0x32) // WRITE_LUT_REGISTER

	// TODO check if full update or partial
	epd.SendData(fullUpdate...)
}

func (epd *EPD2in9) reset() {
	epd.pi.PinRST.High()
	epd.pi.DelayMS(200)
	epd.pi.PinRST.Low()
	epd.pi.DelayMS(5)
	epd.pi.PinRST.High()
	epd.pi.DelayMS(200)
}

func (epd *EPD2in9) SendCommand(cmd ...byte) {
	epd.pi.PinDC.Low()
	epd.pi.PinCS.Low()
	epd.pi.SpiWritebyte(cmd)
	epd.pi.PinCS.High()
}

func (epd *EPD2in9) SendData(data ...byte) {
	epd.pi.PinDC.High()
	epd.pi.PinCS.Low()
	epd.pi.SpiWritebyte(data)
	epd.pi.PinCS.High()
}

func (epd *EPD2in9) ReadBusy() {
	for epd.pi.PinBUSY.Read() == 1 { // 0: idle, 1: busy
		epd.pi.DelayMS(200)
	}
}

func (epd *EPD2in9) TurnOnDisplay() {

	epd.SendCommand(0x22) // DISPLAY_UPDATE_CONTROL_2
	epd.SendData(0xC4)
	epd.SendCommand(0x20) // MASTER_ACTIVATION
	epd.SendCommand(0xFF) // TERMINATE_FRAME_READ_WRITE

	log.Println("e-Paper busy")
	epd.ReadBusy()
	log.Println("e-Paper busy release")
}

func (epd *EPD2in9) SetWindow(x_start, y_start, x_end, y_end int) {
	epd.SendCommand(0x44) // SET_RAM_X_ADDRESS_START_END_POSITION
	// x point must be the multiple of 8 or the last 3 bits will be ignored
	epd.SendData(byte((x_start >> 3) & 0xFF))
	epd.SendData(byte((x_end >> 3) & 0xFF))
	epd.SendCommand(0x45) // SET_RAM_Y_ADDRESS_START_END_POSITION
	epd.SendData(byte(y_start & 0xFF))
	epd.SendData(byte((y_start >> 8) & 0xFF))
	epd.SendData(byte(y_end & 0xFF))
	epd.SendData(byte((y_end >> 8) & 0xFF))
}

func (epd *EPD2in9) SetCursor(x, y int) {

	epd.SendCommand(0x4E) // SET_RAM_X_ADDRESS_COUNTER
	// x point must be the multiple of 8 or the last 3 bits will be ignored
	epd.SendData(byte((x >> 3) & 0xFF))
	epd.SendCommand(0x4F) // SET_RAM_Y_ADDRESS_COUNTER
	epd.SendData(byte(y & 0xFF))
	epd.SendData(byte((y >> 8) & 0xFF))
	epd.ReadBusy()
}

func (epd *EPD2in9) Display(img *image.Gray) {
	if img == nil {
		return
	}
	epd.SetWindow(0, 0, epd.width-1, epd.height-1)

	buf := emptyBuffer(epd.width, epd.height)
	buf = imgToByte(buf, epd.width, epd.height, img)

	for i := 0; i < len(buf); i++ {
		epd.SetCursor(0, i)
		epd.SendCommand(0x24) // WRITE_RAM
		epd.SendData(buf[i])
	}
	epd.TurnOnDisplay()
}

func (epd *EPD2in9) Clear(color byte) {
	epd.SetWindow(0, 0, epd.width-1, epd.height-1)
	for j := 0; j < epd.height; j++ {
		epd.SetCursor(0, j)
		epd.SendCommand(0x24) // WRITE_RAM
		for i := 0; i < epd.width/8; i++ {
			epd.SendData(color)
		}
	}
	epd.TurnOnDisplay()
}

func imgToByte(buf []byte, w, h int, img *image.Gray) []byte {
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			grayColor := img.At(y, x).(color.Gray)
			if grayColor.Y > 0 {
				buf[(x+y*w)/8] |= (0x80 >> (uint(x) % uint(8)))
			}
		}
	}
	return buf
}

func emptyBuffer(w, h int) []byte {
	bufferLength := w * h / 8
	buf := make([]byte, bufferLength)
	for i := 0; i < bufferLength; i++ {
		buf[i] = 0x00
	}
	return buf
}
