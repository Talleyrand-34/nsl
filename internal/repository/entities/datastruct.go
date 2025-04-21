package entities

type Zone struct {
	Name         string `json:"name"`
	Father       string `json:"father"`
	LocationType string `json:"location_type"`
	Proprietary  string `json:"proprietary"`
}
