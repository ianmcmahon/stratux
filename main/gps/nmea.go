package gps

import (
	"strings"
	"fmt"
	"log"
	"strconv"
	"reflect"
)

// func validateNMEAChecksum determines if a string is a properly formatted NMEA sentence with a valid checksum.
//
// If the input string is valid, output is the input stripped of the "$" token and checksum, along with a boolean 'true'
// If the input string is the incorrect format, the checksum is missing/invalid, or checksum calculation fails, an error string and
// boolean 'false' are returned
//
// Checksum is calculated as XOR of all bytes between "$" and "*"
func validateNMEAChecksum(s string) (string, bool) {
	//validate format. NMEA sentences start with "$" and end in "*xx" where xx is the XOR value of all bytes between
	if !(strings.HasPrefix(s, "$") && strings.Contains(s, "*")) {
		return "Invalid NMEA message", false
	}

	// strip leading "$" and split at "*"
	s_split := strings.Split(strings.TrimPrefix(s, "$"), "*")
	s_out := s_split[0]
	s_cs := s_split[1]

	if len(s_cs) < 2 {
		return "Missing checksum. Fewer than two bytes after asterisk", false
	}

	cs, err := strconv.ParseUint(s_cs[:2], 16, 8)
	if err != nil {
		return "Invalid checksum", false
	}

	cs_calc := byte(0)
	for i := range s_out {
		cs_calc = cs_calc ^ byte(s_out[i])
	}

	if cs_calc != byte(cs) {
		return fmt.Sprintf("Checksum failed. Calculated %#X; expected %#X", cs_calc, cs), false
	}

	return s_out, true
}

func createChecksummedNMEASentence(raw []byte) []byte {
	cs_calc := byte(0)
	for _,v := range raw {
		cs_calc ^= v
	}

	return []byte(fmt.Sprintf("$%s*%02X\r\n", raw, cs_calc))
}

func processNMEASentence(line string, situation *SituationData) {
	sentence, valid := validateNMEAChecksum(line)
	if !valid {
		log.Printf("GPS Error: invalid NMEA string: %s\n", sentence)
		return 
	}

	//log.Printf("Begin parse of %s\n", sentence)
	ParseMessage(sentence, situation)
}

type NMEA struct {
	Sentence string
	Tokens []string
	Situation *SituationData
}

// we split the sentence on commas, and use the first field via reflection to find a method with the same name
func ParseMessage(sentence string, situation *SituationData) *NMEA {
	n := &NMEA{ sentence, strings.Split(sentence, ","), situation }

	//log.Printf("NMEA Message type %s, data: %v\n", n.Tokens[0], n.Tokens[1:])

	v := reflect.ValueOf(n)
	m := v.MethodByName(n.Tokens[0])

	if (m == reflect.Value{}) {
		return nil
	}

	m.Call(nil)

	return n
}

func durationSinceMidnight(fixtime string) (int, error) {
	hr, err := strconv.Atoi(fixtime[0:2]); if err != nil { return 0, err }
	min, err := strconv.Atoi(fixtime[2:4]); if err != nil { return 0, err }
	sec, err := strconv.Atoi(fixtime[4:6]); if err != nil { return 0, err }

	return sec + min*60 + hr*60*60, nil
}

func parseLatLon(s string, neg bool) (float32, error) {
	minpos := len(s) - 5
	deg, err := strconv.Atoi(s[0:minpos]); if err != nil { return 0.0, err }
	min, err := strconv.ParseFloat(s[minpos:], 32); if err != nil { return 0.0, err }

	sign := 1; if neg { sign = -1 }

	return float32(sign) * (float32(deg) + float32(min/60.0)), nil 
}

func (n *NMEA) GNGGA() { n.GPGGA() } // ublox 8 uses GNGGA in place of GPGGA to indicate multiple nav sources (GPS/GLONASS)
func (n *NMEA) GPGGA() {
	log.Printf("In GPGGA\n")
	s := n.Situation

	s.Mu_GPS.Lock(); defer s.Mu_GPS.Unlock()

	d, err := durationSinceMidnight(n.Tokens[1]); if err != nil { return }
	s.LastFixSinceMidnightUTC = uint32(d)

	if len(n.Tokens[2]) < 4 || len(n.Tokens[4]) < 4 { return } // sanity check lat/lon

	lat, err := parseLatLon(n.Tokens[2], n.Tokens[3] == "S"); if err != nil { return }
	lon, err := parseLatLon(n.Tokens[4], n.Tokens[5] == "W"); if err != nil { return }

	s.Lat = lat; s.Lng = lon

	log.Printf("Situation: %v\n", s)
}


func (n *NMEA) GPGSA() {
	log.Printf("In GPGSA\n")


}
