package rfid

import (
	"log"
	"time"
  "encoding/hex"

	"periph.io/x/periph/conn/spi/spireg"
	"periph.io/x/periph/conn/spi"
	"periph.io/x/periph/experimental/devices/mfrc522"
	"periph.io/x/periph/host"
	"periph.io/x/periph/host/rpi"
	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
  phost "periph.io/x/host/v3"


)

type RFID struct {
  spi spi.PortCloser
  led gpio.PinIO
  rfid *mfrc522.Dev
}

type RFIDer interface  {
  Init()
  LedOn()
  LedOff()
  ReadID()
}
func (r *RFID) LedOn() {
  r.led.Out(gpio.High)
}

func (r *RFID) LedOff() {
  r.led.Out(gpio.Low)
}

func (r *RFID) Close() {
  r.spi.Close()
	r.rfid.Halt()
}

func (r *RFID) ReadID() (string, error) {
  data, err := r.rfid.ReadUID(10*time.Second)
  if err != nil {
    return "", err
  }
  ret := hex.EncodeToString(data)
  log.Printf(ret)
  return ret, nil
}

func (r *RFID) Init() {
  var err error
	// Make sure periph is initialized.
	if _, err = host.Init(); err != nil {
		log.Fatal(err)
	}

	if _, err = phost.Init(); err != nil {
		log.Fatal(err)
	}
	// Using SPI as an example. See package "periph.io/x/periph/conn/spi/spireg" for more details.
	r.spi, err = spireg.Open("")
	if err != nil {
		log.Fatal(err)
	}

	r.rfid, err = mfrc522.NewSPI(r.spi, rpi.P1_22, rpi.P1_15)
	if err != nil {
		log.Fatal(err)
	}

	// Setting the antenna signal strength.
	r.rfid.SetAntennaGain(7)

  log.Printf("Started %s", r.rfid.String())

	r.led = gpioreg.ByName("7")
  r.LedOn()
}
