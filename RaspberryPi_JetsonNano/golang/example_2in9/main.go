package main

import (
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"os"
	"time"

	epaper "github.com/cdreier/e-Paper/RaspberryPi_JetsonNano/golang"
)

func main() {

	d := epaper.NewEPD2in9()
	defer d.Close()

	d.Clear(0xFF) // white

	width := 128
	height := 296
	img := image.NewGray(image.Rect(0, 0, width, height))

	bpx := 0

	// Set color for each pixel.
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			if x == 30 || x == 90 || y == 30 || y == 5 || y == 90 {
				img.Set(x, y, color.Black)
				bpx++
			} else {
				img.Set(x, y, color.White)
			}
		}
	}

	log.Println("debug black pixel", bpx)

	out, err := os.OpenFile("debug.jpeg", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	err = jpeg.Encode(out, img, nil)
	if err != nil {
		log.Fatal(err)
	}

	d.Display(img)

	time.Sleep(3 * time.Second)
	d.Clear(0xFF) // white

}
