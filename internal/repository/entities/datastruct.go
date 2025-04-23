package entities

type Zone struct {
	Id           int    `json:id`
	Name         string `json:"name"`
	FatherId     int    `json:fatherid`
	Father       string `json:"father"`
	LocationType string `json:"location_type"`
	Proprietary  string `json:"proprietary"`
}
type ModelDevice struct {
	Id    string `json:id`
	Model string `json:model`
	Brand string `json:brand`
	Class string `json:class`
}

type Device struct {
	Id          int    `json:id`
	Label       string `json:label`
	Model       string `json:model`
	Brand       string `json:brand`
	ZoneId      int    `json:zoneid`
	ZoneName    string `json:"zonename"`
	ZoneFather  string `json:"zonefathername"`
	Proprietary string `json:"proprietary"`
}

type ModelPort struct {
	Id        int    `json:id`
	Name      string `json:"name"`
	Positionx int    `json:positiony`
	Positiony int    `json:positionx`
	Model     string `json:"model"`
	Brand     string `json:brand`
}
type DevicePort struct {
	DeviceId int    `json:devid`
	ModelId  int    `json:modelid`
	PortName string `json:portname`
	DevLabel string `json:devname`
}
type Connection struct {
	Id         int    `json:id`
	FromDevice string `json:fromdevice`
	FromModel  string `json:frommodel`
	ToDevice   string `json:todevice`
	ToModel    string `json:tomodel`
}
