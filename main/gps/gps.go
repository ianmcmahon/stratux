package gps

import (
	"fmt"
	"log"
	"bufio"
	"time"
	
	"github.com/tarm/serial"
	"github.com/mitchellh/go-linereader"
)

func InitGPS() error {
	log.Printf("In gps.InitGPS()\n")

	serialConfig := findGPS()
	if serialConfig == nil {
		return fmt.Errorf("Couldn't find gps module anywhere!  We looked!")
	}

	if serialConfig.Baud != 38400 {
		changeGPSBaudRate(serialConfig, 38400)
	}

	// TODO:  try to detect the chipset type (ublox/globaltop/whatever)
	// and call the appropriate configuration routine for the chip

	go gpsSerialReader(serialConfig)
}


// this works based on a channel/goroutine based timeout pattern
// GPS should provide some valid sentence at least once per second.  
// If I don't receive something in two seconds, this probably isn't a valid config
func detectGPS(config *serial.Config) (bool, error) {
	p, err := serial.OpenPort(config)
	if err != nil { return false, err }
	defer p.Close()

	lr := linereader.New(p)
	lr.Timeout = time.Second * 2

	for {
		select { 
		case line := <-lr.Ch:
			log.Printf("Got line from linereader: %v\n", line)
			if sentence, valid := validateNMEAChecksum(line); valid {
				log.Printf("Valid sentence %s on %s:%d\n", sentence, config.Name, config.Baud)
				return true, nil
			}
		case <-time.After(time.Second * 2):
			log.Printf("timeout reached on %s:%d\n", config.Name, config.Baud)
			return false, nil
		}
	}
}

func findGPS() *serial.Config {
	// ports and baud rates are listed in the order they should be tried
	ports := []string{ "/dev/ttyAMA0", "/dev/ttyACM0", "/dev/ttyUSB0" }
	rates := []int{ 38400, 9600, 4800 }

	for _, port := range ports {
		for _, rate := range rates {
			config := &serial.Config{Name: port, Baud: rate}
			if valid, err := detectGPS(config); valid { 
				return config 
			} else { 
				if err != nil { 
					log.Printf("Error detecting GPS: %v\n", err) 
				}
			}
		}
	}
	return nil
}

func changeGPSBaudRate(config *serial.Config, newRate int) error {
	if config.Baud == newRate {
		return nil
	}

	p, err := serial.OpenPort(config)
	if err != nil { return err }
	defer p.Close()

	baud_cfg := createChecksummedNMEASentence([]byte(fmt.Sprintf("PMTK251,%d", newRate)))

	_, err = p.Write(baud_cfg)
	if err != nil { return err }

	config.Baud = newRate

	valid, err := detectGPS(config)
	if !valid {
		err = fmt.Errorf("Set GPS to new rate, but unable to detect it at that new rate!")
	}
	return err
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
