package types

import (
	"time"
)

type Region struct {
	Name           string      `json:"name"`
	MacrozoneCount int         `json:"macrozone_count"`
	Macrozones     []Macrozone `json:"macrozones,omitempty"`
	Hubs           []RegionHub `json:"hubs,omitempty"`
}

type Macrozone struct {
	RegionName   string         `json:"region_name"`
	Name         string         `json:"name"`
	Lat          float64        `json:"lat"`
	Lon          float64        `json:"lon"`
	CreationTime time.Time      `json:"creation_time"`
	ZoneCount    int            `json:"zone_count"`
	Zones        []Zone         `json:"zones,omitempty"`
	Hubs         []MacrozoneHub `json:"hubs,omitempty"`
	ZoneHubs     []ZoneHub      `json:"zone_hubs,omitempty"`
	Sensors      []Sensor       `json:"sensors,omitempty"`
}

type Zone struct {
	RegionName    string    `json:"region_name"`
	MacrozoneName string    `json:"macrozone_name"`
	Name          string    `json:"name"`
	CreationTime  time.Time `json:"creation_time"`
	Hubs          []ZoneHub `json:"hubs,omitempty"`
	Sensors       []Sensor  `json:"sensors,omitempty"`
}

// RegionHub Intermediate Fog Hub
type RegionHub struct {
	Id               string    `json:"id"`
	Service          string    `json:"service"`
	RegistrationTime time.Time `json:"registration_time"`
	LastSeen         time.Time `json:"last_seen"`
}

// MacrozoneHub Proximity Fog Hub
type MacrozoneHub struct {
	Id               string    `json:"id"`
	MacrozoneName    string    `json:"macrozone_name"`
	Service          string    `json:"service"`
	RegistrationTime time.Time `json:"registration_time"`
	LastSeen         time.Time `json:"last_seen"`
}

// ZoneHub Edge Hub
type ZoneHub struct {
	Id               string    `json:"id"`
	MacrozoneName    string    `json:"macrozone_name"`
	ZoneName         string    `json:"zone_name"`
	Service          string    `json:"service"`
	RegistrationTime time.Time `json:"registration_time"`
	LastSeen         time.Time `json:"last_seen"`
}

// Sensor associato a Edge Hub
type Sensor struct {
	Id               string    `json:"id"`
	MacrozoneName    string    `json:"macrozone_name"`
	ZoneName         string    `json:"zone_name"`
	Type             string    `json:"type"`
	Reference        string    `json:"reference"`
	RegistrationTime time.Time `json:"registration_time"`
	LastSeen         time.Time `json:"last_seen"`
}
