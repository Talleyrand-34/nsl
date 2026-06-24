package scanner

import (
	"encoding/json"
	"strings"
	"testing"
)

// An empty IPNetmasks map must be omitted from JSON entirely. Otherwise it
// serializes as {}, and a round-trip through a consumer that decodes JSON
// objects as maps and re-encodes them (e.g. PHP's assoc arrays) turns it into
// [], which then fails to unmarshal back into map[string]string.
func TestDeviceInterface_OmitsEmptyIPNetmasks(t *testing.T) {
	empty := DeviceInterface{Name: "eth0", IPAddresses: []string{"10.0.0.1"}}
	b, err := json.Marshal(empty)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(b), "ip_netmasks") {
		t.Errorf("empty IPNetmasks should be omitted, got: %s", b)
	}

	// A populated map still serializes (as an object) and round-trips.
	full := DeviceInterface{Name: "eth0", IPNetmasks: map[string]string{"10.0.0.1": "255.255.255.0"}}
	b, _ = json.Marshal(full)
	if !strings.Contains(string(b), `"ip_netmasks":{`) {
		t.Errorf("populated IPNetmasks should be an object, got: %s", b)
	}
	var back DeviceInterface
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back.IPNetmasks["10.0.0.1"] != "255.255.255.0" {
		t.Errorf("round-trip lost the netmask: %+v", back.IPNetmasks)
	}
}
