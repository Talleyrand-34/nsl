package topology

import "testing"

// sample is the json0 output of `lldpcli show neighbors -f json0` with one peer.
const sample = `{"lldp":[{"interface":[{"name":"enp1s0","via":"LLDP","rid":"1","age":"0 day",
  "chassis":[{"id":[{"type":"mac","value":"60:22:32:D5:84:81"}],
              "name":[{"value":"openwrt"}],
              "mgmt-ip":[{"value":"10.0.0.245"}]}],
  "port":[{"id":[{"type":"mac","value":"60:22:32:d5:84:82"}],
           "descr":[{"value":"br-lan"}]}]}]}]}`

func TestParseLLDPCLI(t *testing.T) {
	ev, err := parseLLDPCLI(sample, SourceSSHLLDP, "10.0.0.1", "opnsense")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(ev) != 1 {
		t.Fatalf("expected 1 evidence, got %d", len(ev))
	}
	got := ev[0]
	if got.LocalPort != "enp1s0" {
		t.Errorf("LocalPort = %q", got.LocalPort)
	}
	if got.RemoteChassisMAC != "60:22:32:d5:84:81" { // normalized lowercase
		t.Errorf("RemoteChassisMAC = %q", got.RemoteChassisMAC)
	}
	if got.RemoteSysName != "openwrt" {
		t.Errorf("RemoteSysName = %q", got.RemoteSysName)
	}
	if got.RemotePort != "br-lan" {
		t.Errorf("RemotePort = %q", got.RemotePort)
	}
	if got.RemoteIP != "10.0.0.245" {
		t.Errorf("RemoteIP = %q", got.RemoteIP)
	}
	if got.Source != SourceSSHLLDP || got.ObservedDevice != "opnsense" {
		t.Errorf("provenance wrong: %+v", got)
	}
}

func TestParseLLDPCLI_Empty(t *testing.T) {
	ev, err := parseLLDPCLI(`{"lldp":[{}]}`, SourceLocalLLDP, "localhost", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ev) != 0 {
		t.Errorf("expected no evidence, got %d", len(ev))
	}
}
