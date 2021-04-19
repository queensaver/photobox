package main

import (
	"fmt"
	"io"
	"log"
	"os/exec"
	"time"

	"periph.io/x/conn/v3/gpio"
	"periph.io/x/conn/v3/gpio/gpioreg"
	"periph.io/x/host/v3"
)

var webcamLock bool
var captivePortalLock bool

func makeImage(device string) {
	var err error
	// fswebcam -r 4656x3496 --jpeg 95 --set Brightness=30 --set Sharpness=5 ${DATE}.jpg
	t := time.Now().UTC().Unix()
	cmd := exec.Command("/usr/bin/fswebcam",
		"--device", device,
		"-r", "4656x3496",
		"--jpeg", "95",
		"--set", "Brightness=30",
		"--set", "Sharpness=5",
		"-D", "1",
		"-S", "20",
		fmt.Sprintf("%d.jpg", t))
	fmt.Println("executing ", cmd)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return
	}
	if err := cmd.Start(); err != nil {
		stderrBuf, err := io.ReadAll(stderr)
		if err != nil {
			return
		}
		fmt.Println("StdErr Output: ", string(stderrBuf))
		return
	}
	buf, err := io.ReadAll(stdout)
	fmt.Println("StdOut: ", string(buf))
	if err != nil {
		return
	}

	stderrBuf, err := io.ReadAll(stderr)
	if err != nil {
		return
	}

	if err := cmd.Wait(); err != nil {
		return
	}
	fmt.Println("StdErr Output: ", string(stderrBuf))
}

func captureFromWebcam() {
	makeImage("/dev/video0")
	makeImage("/dev/video2")
	webcamLock = false
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

func main() {
	// Load all the drivers:
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}

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
				if webcamLock == false {
					webcamLock = true
					go captureFromWebcam()
				}
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
