package epaper

import (
	"log"
	"time"

	"github.com/stianeikeland/go-rpio/v4"
)

// Rev 2 and 3 Raspberry Pi
// +-----+---------+----------+---------+-----+
// | BCM |   Name  | Physical | Name    | BCM |
// +-----+---------+----++----+---------+-----+
// |     |    3.3v |  1 || 2  | 5v      |     |
// |   2 |   SDA 1 |  3 || 4  | 5v      |     |
// |   3 |   SCL 1 |  5 || 6  | 0v      |     |
// |   4 | GPIO  7 |  7 || 8  | TxD     | 14  |
// |     |      0v |  9 || 10 | RxD     | 15  |
// |  17 | GPIO  0 | 11 || 12 | GPIO  1 | 18  |
// |  27 | GPIO  2 | 13 || 14 | 0v      |     |
// |  22 | GPIO  3 | 15 || 16 | GPIO  4 | 23  |
// |     |    3.3v | 17 || 18 | GPIO  5 | 24  |
// |  10 |    MOSI | 19 || 20 | 0v      |     |
// |   9 |    MISO | 21 || 22 | GPIO  6 | 25  |
// |  11 |    SCLK | 23 || 24 | CE0     | 8   |
// |     |      0v | 25 || 26 | CE1     | 7   |
// |   0 |   SDA 0 | 27 || 28 | SCL 0   | 1   |
// |   5 | GPIO 21 | 29 || 30 | 0v      |     |
// |   6 | GPIO 22 | 31 || 32 | GPIO 26 | 12  |
// |  13 | GPIO 23 | 33 || 34 | 0v      |     |
// |  19 | GPIO 24 | 35 || 36 | GPIO 27 | 16  |
// |  26 | GPIO 25 | 37 || 38 | GPIO 28 | 20  |
// |     |      0v | 39 || 40 | GPIO 29 | 21  |
// +-----+---------+----++----+---------+-----+

type Raspi struct {
	PinRST  rpio.Pin
	PinDC   rpio.Pin
	PinCS   rpio.Pin
	PinBUSY rpio.Pin
}

type EPaperDevice interface {
}

func NewRaspi() *Raspi {

	if err := rpio.Open(); err != nil {
		log.Fatal("unable to open gpio", err)
	}

	r := new(Raspi)
	r.PinRST = rpio.Pin(17)
	r.PinDC = rpio.Pin(25)
	r.PinCS = rpio.Pin(8)
	r.PinBUSY = rpio.Pin(24)

	err := rpio.SpiBegin(rpio.Spi0)
	if err != nil {
		log.Fatal("unable to begin SPI: ", err)
	}
	rpio.SpiSpeed(4000000)

	// TODO
	// rpio.SpiMode(1, 1)
	// self.SPI.mode = 0b00

	r.PinRST.Output()
	r.PinDC.Output()
	r.PinCS.Output()
	r.PinBUSY.Input()

	return r
}

// func (r *Raspi) DigitalWrite(pin rpio.Pin) {
// 	pin.Output()
// }

// func (r *Raspi) DigitalRead(pin rpio.Pin) {
// 	pin.Input()
// }

func (r *Raspi) DelayMS(ms int) {
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

func (r *Raspi) SpiWritebyte(data []byte) {
	rpio.SpiTransmit(data...)
}

func (r *Raspi) Close() {
	rpio.SpiEnd(rpio.Spi0)
	rpio.Close()
}
