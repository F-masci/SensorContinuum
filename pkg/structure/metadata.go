package structure

import (
	"time"
)

type Building struct {
	Id               int       `json:"id"`
	Name             string    `json:"name"`
	Lat              float64   `json:"lat"`
	Lon              float64   `json:"lon"`
	RegistrationTime time.Time `json:"registration_time"`
	LastComunication time.Time `json:"last_comunication"`
}

type Region struct {
	Id            int        `json:"id"`
	Name          string     `json:"name"`
	BuildingCount int        `json:"building_count"`
	Buildings     []Building `json:"buildings,omitempty"`
}
