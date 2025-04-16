package datastructs

// DevicePort comment
type DevicePort struct {
	ModelPortID int    // Port ID
	DeviceID    int    // Port ID
	Name        string // Port name
	PositionX   int    // PositionX
	PositionY   int    // PositionY
}

// Connection represents a connection with detailed zone, device, and port information.
type Connection struct {
	ConnectionID int
	FromZone     []string   // Zone hierarchy as a list
	FromDevice   string     // Device label
	FromPort     DevicePort // From port information
	ToZone       []string   // Zone hierarchy as a list
	ToDevice     string     // Device label
	ToPort       DevicePort // To port information
}

// Device represents a device with its properties and ports.
type Device struct {
	ZoneHierarchy []string // Zone hierarchy as a list
	ID            int
	ModelID       int
	Label         string                // Device label
	Shape         string                // Shape based on device class
	Ports         map[string]DevicePort // Ports mapped by their IDs
}

// DeviceModelPortDetails comment
type DeviceModelPortDetails struct {
	DeviceID      int
	DeviceLabel   string
	ModelID       int
	ModelName     string
	ModelPortID   int
	ModelPortName string
}
