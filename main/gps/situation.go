package gps 

import (
	"time"
	"sync"
)

type SituationData struct {
	Mu_GPS *sync.Mutex

	// From GPS.
	LastFixSinceMidnightUTC uint32
	Lat                     float32
	Lng                     float32
	Quality                 uint8
	GeoidSep                float32 // geoid separation, ft, MSL minus HAE (used in altitude calculation)
	Satellites              uint16  // satellites used in solution
	SatellitesTracked       uint16  // satellites tracked (almanac data received)
	SatellitesSeen          uint16  // satellites seen (signal received)
	Accuracy                float32 // 95% confidence for horizontal position, meters.
	NACp                    uint8   // NACp categories are defined in AC 20-165A
	Alt                     float32 // Feet MSL
	AccuracyVert            float32 // 95% confidence for vertical position, meters
	GPSVertVel              float32 // GPS vertical velocity, feet per second
	LastFixLocalTime        time.Time
	TrueCourse              uint16
	GroundSpeed             uint16
	LastGroundTrackTime     time.Time

	Mu_Attitude *sync.Mutex

	// From BMP180 pressure sensor.
	Temp              float64
	Pressure_alt      float64
	LngastTempPressTime time.Time

	// From MPU6050 accel/gyro.
	Pitch            float64
	Roll             float64
	Gyro_heading     float64
	LastAttitudeTime time.Time
}

func NewSituation() *SituationData {
	s := &SituationData{}
	s.Mu_GPS = &sync.Mutex{}
	s.Mu_Attitude = &sync.Mutex{}
	return s
}
