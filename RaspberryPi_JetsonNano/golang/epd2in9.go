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
	epd.height = 296
	epd.pi = NewRaspi()

	epd.initialize()

	return epd
}

func (epd *EPD2in9) initialize() {
	epd.reset()

	epd.sendCommand(0x01) // DRIVER_OUTPUT_CONTROL
	epd.sendData(byte((epd.height - 1) & 0xFF))
	epd.sendData(byte(((epd.height - 1) >> 8) & 0xFF))
	epd.sendData(0x00) // GD = 0 SM = 0 TB = 0

	epd.sendCommand(0x0C) // BOOSTER_SOFT_START_CONTROL
	epd.sendData(0xD7)
	epd.sendData(0xD6)
	epd.sendData(0x9D)

	epd.sendCommand(0x2C) // WRITE_VCOM_REGISTER
	epd.sendData(0xA8)    // VCOM 7C

	epd.sendCommand(0x3A) // SET_DUMMY_LINE_PERIOD
	epd.sendData(0x1A)    // 4 dummy lines per gate

	epd.sendCommand(0x3B) // SET_GATE_TIME
	epd.sendData(0x08)    // 2us per line

	epd.sendCommand(0x11) // DATA_ENTRY_MODE_SETTING
	epd.sendData(0x03)    // X increment Y increment

	epd.sendCommand(0x32) // WRITE_LUT_REGISTER

	// TODO check if full update or partial
	epd.sendData(fullUpdate...)
}

func (epd *EPD2in9) reset() {
	epd.pi.PinRST.High()
	epd.pi.DelayMS(200)
	epd.pi.PinRST.Low()
	epd.pi.DelayMS(5)
	epd.pi.PinRST.High()
	epd.pi.DelayMS(200)
}

func (epd *EPD2in9) sendCommand(cmd ...byte) {
	epd.pi.PinDC.Low()
	epd.pi.PinCS.Low()
	epd.pi.SpiWritebyte(cmd)
	epd.pi.PinCS.High()
}

func (epd *EPD2in9) sendData(data ...byte) {
	epd.pi.PinDC.High()
	epd.pi.PinCS.Low()
	epd.pi.SpiWritebyte(data)
	epd.pi.PinCS.High()
}

func (epd *EPD2in9) readBusy() {
	for epd.pi.PinBUSY.Read() == 1 { // 0: idle, 1: busy
		epd.pi.DelayMS(200)
	}
}

func (epd *EPD2in9) TurnOnDisplay() {

	epd.sendCommand(0x22) // DISPLAY_UPDATE_CONTROL_2
	epd.sendData(0xC4)
	epd.sendCommand(0x20) // MASTER_ACTIVATION
	epd.sendCommand(0xFF) // TERMINATE_FRAME_READ_WRITE

	log.Println("e-Paper busy")
	epd.readBusy()
	log.Println("e-Paper busy release")
}

func (epd *EPD2in9) SetWindow(x_start, y_start, x_end, y_end int) {
	epd.sendCommand(0x44) // SET_RAM_X_ADDRESS_START_END_POSITION
	// x point must be the multiple of 8 or the last 3 bits will be ignored
	epd.sendData(byte((x_start >> 3) & 0xFF))
	epd.sendData(byte((x_end >> 3) & 0xFF))
	epd.sendCommand(0x45) // SET_RAM_Y_ADDRESS_START_END_POSITION
	epd.sendData(byte(y_start & 0xFF))
	epd.sendData(byte((y_start >> 8) & 0xFF))
	epd.sendData(byte(y_end & 0xFF))
	epd.sendData(byte((y_end >> 8) & 0xFF))
}

func (epd *EPD2in9) SetCursor(x, y int) {

	// x point must be the multiple of 8 or the last 3 bits will be ignored
	epd.sendCommand(0x4E) // SET_RAM_X_ADDRESS_COUNTER
	epd.sendData(byte((x >> 3) & 0xFF))
	epd.sendCommand(0x4F) // SET_RAM_Y_ADDRESS_COUNTER
	epd.sendData(byte(y & 0xFF))
	epd.sendData(byte((y >> 8) & 0xFF))
	epd.readBusy()
}

func (epd *EPD2in9) Display(img *image.Gray) {
	if img == nil {
		return
	}
	epd.SetWindow(0, 0, epd.width-1, epd.height-1)

	imgBytes := imgToByte(epd.width, epd.height, img)
	lineBuffer := make([]byte, int(epd.width/8))

	for y := 0; y < epd.height; y++ {
		epd.SetCursor(0, y)
		epd.sendCommand(0x24) // WRITE_RAM

		// sending line by line
		for x := 0; x < int(epd.width/8); x++ {
			lineBuffer[x] = imgBytes[x+y*(epd.width/8)]
		}
		epd.sendData(lineBuffer...)
	}

	epd.TurnOnDisplay()
}

func (epd *EPD2in9) Clear(color byte) {
	epd.SetWindow(0, 0, epd.width-1, epd.height-1)
	for j := 0; j < epd.height; j++ {
		epd.SetCursor(0, j)
		epd.sendCommand(0x24) // WRITE_RAM
		for i := 0; i < epd.width/8; i++ {
			epd.sendData(color)
		}
	}
	epd.TurnOnDisplay()
}

func (epd *EPD2in9) Close() {
	epd.pi.Close()
}

func imgToByte(w, h int, img *image.Gray) []byte {
	buf := emptyBuffer(w, h)

	imgw := img.Rect.Size().X
	imgh := img.Rect.Size().Y

	if imgw == w && imgh == h {
		log.Println("VERTICAL")
		for y := 0; y < imgh; y++ {
			for x := 0; x < imgw; x++ {
				grayColor := img.At(x, y).(color.Gray)
				if grayColor.Y == 0 {
					buf[int((x+y*w)/8)] &= ^(0x80 >> (uint(x) % uint(8)))
				}
			}
		}
	} else if imgw == h && imgh == w {
		log.Println("HORIZONTAL")
		for y := 0; y < imgh; y++ {
			for x := 0; x < imgw; x++ {
				newx := y
				newy := h - x - 1
				grayColor := img.At(x, y).(color.Gray)
				if grayColor.Y == 0 {
					buf[(newx+newy*w)/8] &= ^(0x80 >> (uint(h) % uint(8)))
				}
			}
		}
	}

	return buf
}

func emptyBuffer(w, h int) []byte {
	bufferLength := (w / 8) * h
	buf := make([]byte, bufferLength)
	for i := 0; i < bufferLength; i++ {
		buf[i] = 0xff
	}
	return buf
}
