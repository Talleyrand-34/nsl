// SPDX-License-Identifier: AGPL-3.0-or-later
// routing_test.go: TDD for the OPNsense routing wrapper.
package routing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/t34/opnsense-api/opnsense"
)

func newRoutingLab(t *testing.T, response string) (*httptest.Server, *Module) {
	t.Helper()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(response))
	}))
	c := opnsense.NewClient(srv.URL, "k", "s", opnsense.WithInsecureTLS())
	return srv, New(c)
}

func TestSearchGateway_ParsesRows(t *testing.T) {
	srv, m := newRoutingLab(t, `{
		"rows":[
			{"uuid":"gw1","name":"LAN_DHCP","address":"10.0.0.1","gateway_type":"ipv4","descr":"LAN uplink"},
			{"uuid":"gw2","name":"WAN_DHCP","address":"192.168.1.1","gateway_type":"ipv4","descr":"WAN uplink"}
		]
	}`)
	defer srv.Close()

	gws, err := m.SearchGateway(context.Background())
	if err != nil {
		t.Fatalf("SearchGateway: %v", err)
	}
	if len(gws) != 2 {
		t.Fatalf("want 2 gateways, got %d", len(gws))
	}
	if gws[0].Name != "LAN_DHCP" || gws[0].Address != "10.0.0.1" {
		t.Errorf("first gateway: %+v", gws[0])
	}
	if gws[1].Name != "WAN_DHCP" {
		t.Errorf("second gateway: %+v", gws[1])
	}
}

func TestSearchGateway_EmptyRows(t *testing.T) {
	srv, m := newRoutingLab(t, `{"rows":[]}`)
	defer srv.Close()

	gws, err := m.SearchGateway(context.Background())
	if err != nil {
		t.Fatalf("SearchGateway: %v", err)
	}
	if len(gws) != 0 {
		t.Errorf("want 0 gateways, got %d", len(gws))
	}
}

func TestSearchGateway_NoRowsKey(t *testing.T) {
	// A malformed response (no "rows" key) must decode to an empty slice,
	// not panic. Real OPNsense boxes always send "rows".
	srv, m := newRoutingLab(t, `{}`)
	defer srv.Close()

	gws, err := m.SearchGateway(context.Background())
	if err != nil {
		t.Fatalf("SearchGateway: %v", err)
	}
	if len(gws) != 0 {
		t.Errorf("want 0 gateways on missing rows key, got %d", len(gws))
	}
}