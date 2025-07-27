package structure

type Building struct {
	Id       int     `json:"id"`
	RegionId int     `json:"region_id"`
	Name     string  `json:"name"`
	Lat      float64 `json:"lat"`
	Lon      float64 `json:"lon"`
}

type Region struct {
	Id            int        `json:"id"`
	Name          string     `json:"name"`
	BuildingCount int        `json:"building_count"`
	Buildings     []Building `json:"buildings,omitempty"`
}
