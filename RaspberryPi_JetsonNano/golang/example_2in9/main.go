package example_2in9

import epaper "github.com/cdreier/e-Paper"

func main() {

	d := epaper.NewEPD2in9()

	d.Clear(0xFF) // white

}
