package gps

import (
	"strings"
	"fmt"
	"log"
	"strconv"
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

func processNMEASentence(line string) {
	sentence, valid := validateNMEAChecksum(line)
	if !valid {
		log.Printf("GPS Error: invalid NMEA string: %s\n", sentence)
		return 
	}

	//log.Printf("Begin parse of %s\n", sentence)
	ParseMessage(sentence)
}

type NMEA struct {
	Sentence string
	Tokens []string
}

func ParseMessage(sentence string) *NMEA {
	n := &NMEA{ sentence, strings.Split(sentence, ",") }

	log.Printf("NMEA Message type %s, data: %v\n", n.Tokens[0], n.Tokens[1:])

	return n
}

