package gps

import (
	"log"
	"bufio"
	
	"github.com/tarm/serial"
)

func InitGPS() {
	log.Printf("In gps.InitGPS()\n")

	// eventually I would like to come up with a reliable autodetection scheme for different types of gps.
	// for now I'll just have entry points into different configurations that get uncommented here

	initUltimateGPS()
}


// for the Adafruit Ultimate GPS Hat (https://www.adafruit.com/products/2324)
// MT3339 chipset
func initUltimateGPS() error {

	// module is attached via serial UART, shows up as /dev/ttyAMA0 on rpi
	device := "/dev/ttyAMA0"
	log.Printf("Using %s for GPS\n", device)

	// module comes up in 9600baud, 1hz mode
	serialConfig := &serial.Config{Name: device, Baud: 9600}
	p, err := serial.OpenPort(serialConfig)
	if err != nil { return fmt.Errorf("Error opening serial port: %v", err) }

	scanner := bufio.NewScanner(p)

	for scanner.Scan() {
		log.Printf("gps data: %s\n", scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading serial data: %v\n", err)
	}
}