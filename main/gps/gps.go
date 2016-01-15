package gps

import (
	"log"
	"bufio"
	"io"
	
	"github.com/tarm/serial"
)

func InitGPS() {
	log.Printf("In gps.InitGPS()\n")

	// eventually I would like to come up with a reliable autodetection scheme for different types of gps.
	// for now I'll just have entry points into different configurations that get uncommented here

	err := initUltimateGPS()
	if err != nil {
		log.Printf("Error initializing gps: %v\n", err)
	}
}


// this works based on a channel/goroutine based timeout pattern
// GPS should provide some valid sentence at least once per second.  
// If I don't receive something in two seconds, this probably isn't a valid config
func detectGPS(config *serial.Config) (bool, error) {
	p, err := serial.OpenPort(serialConfig)
	if err != nil { return false, err }
	defer p.Close()

	ch := make(chan bool)

	timeout := false

	// this function attempts to scan lines until it gets one which is a valid sentence, then it 
	// chucks a token on the channel signifying success and exits
	go func(r io.Reader) {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			if timeout { 
				log.Printf("exiting detect %s:%d loop due to timeout\n")
				return 
			}
			line := scanner.Text()
			if _, valid := validateNMEAChecksum(line); valid {
				ch<-true
				log.Printf("exiting detect %s:%d loop due to success\n")
				return
			}
		}
	}(p)

	select { 
	case <-ch:
		return true, nil
	case <-time.After(time.Second * 3):
		return false, nil
	}
}


// for the Adafruit Ultimate GPS Hat (https://www.adafruit.com/products/2324)
// MT3339 chipset
func initUltimateGPS() error {

	// module is attached via serial UART, shows up as /dev/ttyAMA0 on rpi
	device := "/dev/ttyAMA0"
	log.Printf("Using %s for GPS\n", device)

	// module comes up in 9600baud, 1hz mode
	serialConfig := &serial.Config{Name: device, Baud: 9600}

	valid, err := detectGPS(serialConfig)
	if err != nil { return err }
	if valid {
		log.Printf("Detected GPS on %s at %dbaud!\n", serialConfig.Name, serialConfig.Baud)
	}

	serialConfig.Baud = 38400
	valid, err := detectGPS(serialConfig)
	if err != nil { return err }
	if valid {
		log.Printf("Detected GPS on %s at %dbaud!\n", serialConfig.Name, serialConfig.Baud)
	}



	// baud rate configuration string:
	// PMTK251,115200

	p, err := serial.OpenPort(serialConfig)
	if err != nil { return err }

	baud_cfg := createChecksummedNMEASentence([]byte("PMTK251,38400"))
	log.Printf("checksummed baud cfg: %s\n", baud_cfg)

	n, err := p.Write(baud_cfg)
	if err != nil { return err }
	log.Printf("Wrote %d bytes\n", n)

	p.Close()


	//serialConfig.Baud = 115200

	go gpsSerialReader(serialConfig)

	return nil
}


// goroutine which scans for incoming sentences (which are newline terminated) and sends them downstream for processing
func gpsSerialReader(serialConfig *serial.Config) {
	p, err := serial.OpenPort(serialConfig)
	log.Printf("Opening GPS on %s at %dbaud\n", serialConfig.Name, serialConfig.Baud) 
	if err != nil { 
		log.Printf("Error opening serial port: %v", err) 
		log.Printf("  GPS Serial Reader routine is terminating.\n")
		return
	}
	defer p.Close()

	scanner := bufio.NewScanner(p)

	for scanner.Scan() {
		line := scanner.Text()
		//log.Printf("gps data: %s\n", line)

		processNMEASentence(line)
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading serial data: %v\n", err)
	}
}
