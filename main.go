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

func makeImage() {
	var err error
	// fswebcam -r 4656x3496 --jpeg 95 --set Brightness=30 --set Sharpness=5 ${DATE}.jpg
	t := time.Now().UTC().Unix()
	cmd := exec.Command("/usr/bin/fswebcam",
		"-r", "4656x3496",
		"--jpeg", "95",
		"--set", "Brightness=30",
		"--set", "Sharpness=5",
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

	// Wait for edges as detected by the hardware, and print the value read:
	for {
		p.WaitForEdge(-1)
		fmt.Printf("-> %s\n", p.Read())
		if p.Read() == true {
			makeImage()
		}
	}
}
