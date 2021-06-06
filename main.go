package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os/exec"
  "os"
	"time"
  "sync"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
	"github.com/queensaver/photobox/rfid"
)

var webcamLock sync.Mutex
var captivePortalLock bool
var imageDirectory = flag.String("image_directory", "images", "Image directory for the webcam files.")

func makeImage(device string, camera int, filename string, done chan bool) {
	var err error
  year, month, day := time.Now().Date()
  filename = fmt.Sprintf("%s/%d-%02d-%02d/%s-%d.jpg", *imageDirectory, year, int(month), day, filename, camera)
	cmd := exec.Command("/usr/bin/fswebcam",
		"--device", device,
		"-r", "4656x3496",
		"--jpeg", "95",
		"--set", "Brightness=30",
		"--set", "Sharpness=5",
		"-D", "1",
		"-S", "20",
    filename)
	fmt.Println("executing ", cmd)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		done <- true
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		done <- true
		return
	}
	if err := cmd.Start(); err != nil {
		stderrBuf, err := io.ReadAll(stderr)
		if err != nil {
			done <- true
			return
		}
		fmt.Println("StdErr Output: ", string(stderrBuf))
		done <- true
		return
	}
	buf, err := io.ReadAll(stdout)
	fmt.Println("StdOut: ", string(buf))
	if err != nil {
		done <- true
		return
	}

	stderrBuf, err := io.ReadAll(stderr)
	if err != nil {
		done <- true
		return
	}

	if err := cmd.Wait(); err != nil {
		done <- true
		return
	}
	fmt.Println("StdErr Output: ", string(stderrBuf))
	done <- true
}

func captureFromWebcam(filename string) {
	video0 := make(chan bool)
	video2 := make(chan bool)
  year, month, day := time.Now().Date()
  err := os.MkdirAll(fmt.Sprintf("%s/%d-%02d-%02d", *imageDirectory, year, int(month), day), 0755)
  if err != nil {
    log.Println(err)
    return
  }
	go makeImage("/dev/video0", 0, filename, video0)
	go makeImage("/dev/video2", 1, filename, video2)
	<-video0
	<-video2
	webcamLock.Unlock()
	fmt.Println("done capturing images.")
}

func captivePortal() {
	fmt.Println("captive portal")
	cmd := exec.Command("/usr/bin/sudo",
		"/usr/local/bin/captive_portal.sh")
	stderr, err := cmd.StderrPipe()
	if err != nil {
		captivePortalLock = false
		return
	}
	if err := cmd.Start(); err != nil {
		stderrBuf, err := io.ReadAll(stderr)
		if err != nil {
			captivePortalLock = false
			return
		}
		fmt.Println("StdErr Output: ", string(stderrBuf))
		captivePortalLock = false
		return
	}
	captivePortalLock = false
}

func shutDown() {
	fmt.Println("Shutting down...")
	cmd := exec.Command("/usr/bin/sudo",
		"shutdown", "-h", "now")
	cmd.Start()
	// TODO: OS.Exit()
}

func buttonListener() {
	// Lookup a pin by its number:
	p := gpioreg.ByName("GPIO17")
	if p == nil {
		log.Fatal("Failed to find GPIO17")
	}

	fmt.Printf("%s: %s\n", p, p.Function())

	// Set it as input, with an internal pull down resistor:
	if err := p.In(gpio.PullDown, gpio.BothEdges); err != nil {
		log.Fatal(err)
	}

	var buttonPressTimestamp int64
	var buttonReleaseTimestamp int64
	for {
		p.WaitForEdge(-1)
		fmt.Printf("-> %s\n", p.Read())
		if p.Read() == true {
			buttonPressTimestamp = time.Now().UTC().Unix()
		} else if p.Read() == false && buttonPressTimestamp > 0 {
			buttonReleaseTimestamp = time.Now().UTC().Unix()
			diff := buttonReleaseTimestamp - buttonPressTimestamp
			if diff < 2 {
        webcamLock.Lock()
	      t := time.Now().UTC().Unix()
        go captureFromWebcam(fmt.Sprintf("%d", t))
			} else if diff >= 2 && diff <= 10 {
				if captivePortalLock == false {
					captivePortalLock = true
					go captivePortal()
				}
			} else if diff > 10 && diff < 30 {
				go shutDown()
			}
		}
	}
}

func rfidListener() {
  r := rfid.RFID{}
  var old_id string
  for {
    r.Init()
    r.LedOn()
    id, err := r.ReadID()
    if err != nil {
      log.Println(err)
      r.Close()
      continue
    }
    if id == old_id {
      log.Println("doing nothing, same id")
      time.Sleep(1 * time.Second)
      r.Close()
      continue
    }
    log.Println(id)
    r.LedOff()
    webcamLock.Lock()
    captureFromWebcam(id)
    old_id = id
    r.Close()
  }
}

func main() {
	flag.Parse()
	// Load all the drivers:
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}
  go buttonListener()
  go rfidListener()
  select{}
}
