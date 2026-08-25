package scanner

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/gosnmp/gosnmp"
)

// SNMP OID constants
const (
	oidSysDescr    = "1.3.6.1.2.1.1.1.0"
	oidSysObjectID = "1.3.6.1.2.1.1.2.0"
	oidSysContact  = "1.3.6.1.2.1.1.4.0"
	oidSysName     = "1.3.6.1.2.1.1.5.0"
	oidSysLocation = "1.3.6.1.2.1.1.6.0"

	// ifTable
	oidIfDescr       = "1.3.6.1.2.1.2.2.1.2"
	oidIfType        = "1.3.6.1.2.1.2.2.1.3"
	oidIfPhysAddress = "1.3.6.1.2.1.2.2.1.6"
	oidIfAdminStatus = "1.3.6.1.2.1.2.2.1.7"
	oidIfOperStatus  = "1.3.6.1.2.1.2.2.1.8"

	// ipAddrTable
	oidIpAdEntAddr    = "1.3.6.1.2.1.4.20.1.1"
	oidIpAdEntIfIndex = "1.3.6.1.2.1.4.20.1.2"
	oidIpAdEntNetMask = "1.3.6.1.2.1.4.20.1.3"

	// dot1q VLAN MIB
	oidDot1qVlanStaticEgressPorts   = "1.3.6.1.2.1.17.7.1.4.3.1.2"
	oidDot1qVlanStaticUntaggedPorts = "1.3.6.1.2.1.17.7.1.4.3.1.4"
	oidDot1dBasePortIfIndex         = "1.3.6.1.2.1.17.1.4.1.1"

	// LLDP remote table
	oidLLDPRemChassisID = "1.0.8802.1.1.2.1.4.1.1.5"
	oidLLDPRemPortDesc  = "1.0.8802.1.1.2.1.4.1.1.8"
	oidLLDPRemSysName   = "1.0.8802.1.1.2.1.4.1.1.9"
	// LLDP local port table (port number → ifDescr)
	oidLLDPLocPortDesc = "1.0.8802.1.1.2.1.3.7.1.4"


	// Bridge forwarding database (MAC address tables)
	oidDot1qTpFdbPort = "1.3.6.1.2.1.17.7.1.2.2.1.2" // VLAN-aware FDB: index = fdbId.MAC, value = bridge port
	oidDot1dTpFdbPort = "1.3.6.1.2.1.17.4.3.1.2"     // classic FDB: index = MAC, value = bridge port
)

// SNMPScanner queries devices via SNMP to collect deterministic inventory.
type SNMPScanner struct{}

func NewSNMPScanner() *SNMPScanner {
	return &SNMPScanner{}
}

// Scan discovers all SNMP-reachable devices in the given subnet.
func (ss *SNMPScanner) Scan(options ScanOptions) (*ScanResult, error) {
	timeout := options.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	ips, err := enumerateSubnet(options.Subnet)
	if err != nil {
		// treat as single IP
		ips = []string{options.Subnet}
	}

	result := &ScanResult{
		ID:        fmt.Sprintf("scan-%d", time.Now().UnixNano()),
		Subnet:    options.Subnet,
		StartTime: time.Now(),
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 50)
	total := len(ips)
	done := 0

	for _, ip := range ips {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			dev, err := ss.ScanDevice(ip, options)
			reachable := err == nil && dev.Reachable
			mu.Lock()
			if reachable {
				result.Devices = append(result.Devices, *dev)
			}
			done++
			d := done
			mu.Unlock()
			if options.OnProgress != nil {
				options.OnProgress(d, total, ip, reachable)
			}
		}(ip)
	}

	wg.Wait()
	result.EndTime = time.Now()
	return result, nil
}

// ScanDevice queries a single device by IP via SNMP and returns all available data.
func (ss *SNMPScanner) ScanDevice(ip string, options ScanOptions) (*SNMPDevice, error) {
	client := ss.newClient(ip, options)
	if err := client.Connect(); err != nil {
		return &SNMPDevice{IP: ip, Reachable: false}, nil
	}
	defer client.Conn.Close()

	device := &SNMPDevice{IP: ip, Reachable: true}

	// Probe system MIB first — if the device doesn't respond to SNMP at all,
	// bail out immediately instead of waiting through every subsequent walk.
	if !ss.querySystemMIB(client, device) {
		device.Reachable = false
		return device, nil
	}

	ifaces := ss.queryInterfaceTable(client)
	ss.queryIPAddressTable(client, ifaces)
	ss.queryVLANTable(client, ifaces)
	device.Interfaces = ifaces

	neighbors := ss.queryLLDPNeighbors(client, ifaces)
	if len(neighbors) == 0 {
	}
	device.Neighbors = neighbors

	device.BridgeFDB = ss.queryFDB(client, ifaces)

	return device, nil
}

// queryFDB walks the bridge forwarding database (MAC address table) and maps
// each learned MAC to the local bridge port (ifDescr) it was seen on. It tries
// the VLAN-aware dot1q table first and falls back to the classic dot1d table.
func (ss *SNMPScanner) queryFDB(client *gosnmp.GoSNMP, ifaces []DeviceInterface) []FDBEntry {
	// bridge port number → ifIndex → interface name
	bridgePortToIfIndex := make(map[int]int)
	_ = client.BulkWalk(oidDot1dBasePortIfIndex, func(pdu gosnmp.SnmpPDU) error {
		bp := lastOIDInt(pdu.Name)
		if bp > 0 {
			if v, ok := pdu.Value.(int); ok {
				bridgePortToIfIndex[bp] = v
			}
		}
		return nil
	})
	ifIndexToName := make(map[int]string)
	for _, f := range ifaces {
		ifIndexToName[f.Index] = f.Name
	}
	portName := func(bp int) string {
		if ifx, ok := bridgePortToIfIndex[bp]; ok {
			if n := ifIndexToName[ifx]; n != "" {
				return n
			}
		}
		return fmt.Sprintf("port%d", bp)
	}

	var entries []FDBEntry
	seen := make(map[string]bool)
	add := func(mac, vlan string, bp int) {
		if mac == "" || bp <= 0 {
			return
		}
		key := vlan + "|" + mac
		if seen[key] {
			return
		}
		seen[key] = true
		entries = append(entries, FDBEntry{MAC: mac, VLAN: vlan, Port: portName(bp), IfIndex: bridgePortToIfIndex[bp]})
	}

	// VLAN-aware table: suffix is "fdbId.<6 MAC octets>"
	_ = client.BulkWalk(oidDot1qTpFdbPort, func(pdu gosnmp.SnmpPDU) error {
		toks := strings.Split(oidSuffix(pdu.Name, oidDot1qTpFdbPort), ".")
		if len(toks) >= 7 {
			vlan := toks[0]
			add(macFromOIDTokens(toks[len(toks)-6:]), vlan, toIntValue(pdu.Value))
		}
		return nil
	})
	// Classic table: suffix is "<6 MAC octets>"
	_ = client.BulkWalk(oidDot1dTpFdbPort, func(pdu gosnmp.SnmpPDU) error {
		toks := strings.Split(oidSuffix(pdu.Name, oidDot1dTpFdbPort), ".")
		if len(toks) >= 6 {
			add(macFromOIDTokens(toks[len(toks)-6:]), "", toIntValue(pdu.Value))
		}
		return nil
	})

	return entries
}

// macFromOIDTokens formats six decimal OID tokens as a colon MAC address.
func macFromOIDTokens(toks []string) string {
	if len(toks) != 6 {
		return ""
	}
	octs := make([]int, 6)
	for i, t := range toks {
		var v int
		if _, err := fmt.Sscanf(t, "%d", &v); err != nil || v < 0 || v > 255 {
			return ""
		}
		octs[i] = v
	}
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x", octs[0], octs[1], octs[2], octs[3], octs[4], octs[5])
}

// toIntValue coerces a gosnmp PDU value into an int (bridge port numbers).
func toIntValue(v interface{}) int {
	switch n := v.(type) {
	case int:
		return n
	case uint:
		return int(n)
	case int64:
		return int(n)
	case uint64:
		return int(n)
	case uint32:
		return int(n)
	}
	return 0
}

// Probe is a cheap reachability check: a single SNMP Get of sysName. It is used
// to pick which of a multi-homed device's IPs actually answers SNMP.
func (ss *SNMPScanner) Probe(ip string, options ScanOptions) bool {
	client := ss.newClient(ip, options)
	if err := client.Connect(); err != nil {
		return false
	}
	defer client.Conn.Close()
	res, err := client.Get([]string{oidSysName})
	if err != nil || res == nil || len(res.Variables) == 0 {
		return false
	}
	t := res.Variables[0].Type
	return t != gosnmp.NoSuchObject && t != gosnmp.NoSuchInstance && t != gosnmp.Null
}

// newClient creates a configured gosnmp client.
func (ss *SNMPScanner) newClient(ip string, options ScanOptions) *gosnmp.GoSNMP {
	community := options.SNMP.Community
	if community == "" {
		community = "public"
	}

	port := options.SNMP.Port
	if port == 0 {
		port = 161
	}

	version := gosnmp.Version2c
	if options.SNMP.Version == "v1" {
		version = gosnmp.Version1
	}

	timeout := options.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	return &gosnmp.GoSNMP{
		Target:    ip,
		Port:      port,
		Community: community,
		Version:   version,
		Timeout:   timeout,
		Retries:   0,
		MaxOids:   60,
	}
}

// querySystemMIB retrieves system-level OIDs.
// Returns true if the device responded to SNMP, false if unreachable.
func (ss *SNMPScanner) querySystemMIB(client *gosnmp.GoSNMP, device *SNMPDevice) bool {
	oids := []string{oidSysDescr, oidSysObjectID, oidSysContact, oidSysName, oidSysLocation}
	result, err := client.Get(oids)
	if err != nil {
		return false
	}
	responded := false
	for _, v := range result.Variables {
		if v.Type == gosnmp.NoSuchObject || v.Type == gosnmp.NoSuchInstance {
			continue
		}
		responded = true
		val := snmpString(v)
		switch {
		case strings.HasPrefix(v.Name, "."+oidSysDescr) || v.Name == oidSysDescr:
			device.SysDescr = val
		case strings.HasPrefix(v.Name, "."+oidSysObjectID) || v.Name == oidSysObjectID:
			device.SysObjectID = val
		case strings.HasPrefix(v.Name, "."+oidSysContact) || v.Name == oidSysContact:
			device.SysContact = val
		case strings.HasPrefix(v.Name, "."+oidSysName) || v.Name == oidSysName:
			device.SysName = val
		case strings.HasPrefix(v.Name, "."+oidSysLocation) || v.Name == oidSysLocation:
			device.SysLocation = val
		}
	}
	return responded
}

// queryInterfaceTable retrieves the ifTable.
func (ss *SNMPScanner) queryInterfaceTable(client *gosnmp.GoSNMP) []DeviceInterface {
	// index → interface
	ifaceMap := make(map[int]*DeviceInterface)

	// ifDescr
	_ = client.BulkWalk(oidIfDescr, func(pdu gosnmp.SnmpPDU) error {
		idx := lastOIDInt(pdu.Name)
		if idx > 0 {
			iface := getOrCreate(ifaceMap, idx)
			iface.Name = snmpString(pdu)
		}
		return nil
	})

	// ifType — store the type and skip loopbacks
	_ = client.BulkWalk(oidIfType, func(pdu gosnmp.SnmpPDU) error {
		idx := lastOIDInt(pdu.Name)
		if idx > 0 {
			iface := getOrCreate(ifaceMap, idx)
			iface.Index = idx
			if v, ok := pdu.Value.(int); ok {
				iface.IfType = v
				if v == IfTypeLoopback {
					iface.Index = -1
				}
			}
		}
		return nil
	})

	// ifPhysAddress
	_ = client.BulkWalk(oidIfPhysAddress, func(pdu gosnmp.SnmpPDU) error {
		idx := lastOIDInt(pdu.Name)
		if idx > 0 {
			if mac := formatMAC(pdu); mac != "" {
				getOrCreate(ifaceMap, idx).MAC = mac
			}
		}
		return nil
	})

	// ifAdminStatus
	_ = client.BulkWalk(oidIfAdminStatus, func(pdu gosnmp.SnmpPDU) error {
		idx := lastOIDInt(pdu.Name)
		if idx > 0 {
			if v, ok := pdu.Value.(int); ok {
				getOrCreate(ifaceMap, idx).AdminStatus = v
			}
		}
		return nil
	})

	// ifOperStatus
	_ = client.BulkWalk(oidIfOperStatus, func(pdu gosnmp.SnmpPDU) error {
		idx := lastOIDInt(pdu.Name)
		if idx > 0 {
			if v, ok := pdu.Value.(int); ok {
				getOrCreate(ifaceMap, idx).OperStatus = v
			}
		}
		return nil
	})

	// Build sorted slice, skip loopbacks (Index == -1) and unnamed interfaces
	var ifaces []DeviceInterface
	for idx := 1; idx <= 10000; idx++ {
		iface, ok := ifaceMap[idx]
		if !ok {
			continue
		}
		if iface.Index == -1 {
			continue // loopback
		}
		if iface.Name == "" {
			continue
		}
		iface.Index = idx
		ifaces = append(ifaces, *iface)
	}
	return ifaces
}

// queryIPAddressTable populates IPAddresses and IPNetmasks on each DeviceInterface.
func (ss *SNMPScanner) queryIPAddressTable(client *gosnmp.GoSNMP, ifaces []DeviceInterface) {
	// ip → ifIndex
	ipToIfIndex := make(map[string]int)
	_ = client.BulkWalk(oidIpAdEntIfIndex, func(pdu gosnmp.SnmpPDU) error {
		ip := oidSuffix(pdu.Name, oidIpAdEntIfIndex)
		if ip != "" {
			if v, ok := pdu.Value.(int); ok {
				ipToIfIndex[ip] = v
			}
		}
		return nil
	})

	// ip → netmask
	ipToNetmask := make(map[string]string)
	_ = client.BulkWalk(oidIpAdEntNetMask, func(pdu gosnmp.SnmpPDU) error {
		ip := oidSuffix(pdu.Name, oidIpAdEntNetMask)
		if ip != "" {
			ipToNetmask[ip] = snmpString(pdu)
		}
		return nil
	})

	// Build ifIndex → slice index map
	idxMap := make(map[int]int)
	for i, iface := range ifaces {
		idxMap[iface.Index] = i
	}

	for ip, ifIdx := range ipToIfIndex {
		i, ok := idxMap[ifIdx]
		if !ok {
			continue
		}
		ifaces[i].IPAddresses = append(ifaces[i].IPAddresses, ip)
		if mask, hasMask := ipToNetmask[ip]; hasMask && mask != "" {
			if ifaces[i].IPNetmasks == nil {
				ifaces[i].IPNetmasks = make(map[string]string)
			}
			ifaces[i].IPNetmasks[ip] = mask
		}
	}
}

// queryVLANTable populates VLAN memberships on each DeviceInterface using dot1q MIB.
func (ss *SNMPScanner) queryVLANTable(client *gosnmp.GoSNMP, ifaces []DeviceInterface) {
	// bridgePort → ifIndex mapping
	bridgePortToIfIndex := make(map[int]int)
	_ = client.BulkWalk(oidDot1dBasePortIfIndex, func(pdu gosnmp.SnmpPDU) error {
		bridgePort := lastOIDInt(pdu.Name)
		if bridgePort > 0 {
			if v, ok := pdu.Value.(int); ok {
				bridgePortToIfIndex[bridgePort] = v
			}
		}
		return nil
	})

	// ifIndex → slice index
	idxMap := make(map[int]int)
	for i, iface := range ifaces {
		idxMap[iface.Index] = i
	}

	// vlanID → egress bitmask
	egressMap := make(map[int][]byte)
	_ = client.BulkWalk(oidDot1qVlanStaticEgressPorts, func(pdu gosnmp.SnmpPDU) error {
		vlanID := lastOIDInt(pdu.Name)
		if vlanID > 0 {
			if data, ok := pdu.Value.([]byte); ok {
				egressMap[vlanID] = data
			}
		}
		return nil
	})

	// vlanID → untagged bitmask
	untaggedMap := make(map[int][]byte)
	_ = client.BulkWalk(oidDot1qVlanStaticUntaggedPorts, func(pdu gosnmp.SnmpPDU) error {
		vlanID := lastOIDInt(pdu.Name)
		if vlanID > 0 {
			if data, ok := pdu.Value.([]byte); ok {
				untaggedMap[vlanID] = data
			}
		}
		return nil
	})

	for vlanID, egressBits := range egressMap {
		untaggedBits := untaggedMap[vlanID]
		egressPorts := parsePortBitmask(egressBits)
		untaggedPorts := parsePortBitmaskSet(untaggedBits)

		for _, bridgePort := range egressPorts {
			ifIdx, ok := bridgePortToIfIndex[bridgePort]
			if !ok {
				continue
			}
			sliceIdx, ok := idxMap[ifIdx]
			if !ok {
				continue
			}
			tagged := !untaggedPorts[bridgePort]
			ifaces[sliceIdx].VLANs = append(ifaces[sliceIdx].VLANs, VLANMembership{
				VLANNumber: fmt.Sprintf("%d", vlanID),
				Tagged:     tagged,
			})
		}
	}
}

// queryLLDPNeighbors retrieves LLDP neighbor table entries.
func (ss *SNMPScanner) queryLLDPNeighbors(client *gosnmp.GoSNMP, ifaces []DeviceInterface) []DirectNeighbor {
	// lldpLocPortDesc: portNum → interface description
	locPortDesc := make(map[int]string)
	_ = client.BulkWalk(oidLLDPLocPortDesc, func(pdu gosnmp.SnmpPDU) error {
		portNum := lastOIDInt(pdu.Name)
		if portNum > 0 {
			locPortDesc[portNum] = snmpString(pdu)
		}
		return nil
	})

	remSysName := make(map[lldpRemKey]string)
	remPortDesc := make(map[lldpRemKey]string)
	remChassisID := make(map[lldpRemKey]string)
	zero := lldpRemKey{}

	_ = client.BulkWalk(oidLLDPRemSysName, func(pdu gosnmp.SnmpPDU) error {
		k := parseLLDPRemKey(pdu.Name, oidLLDPRemSysName)
		if k != zero {
			remSysName[k] = snmpString(pdu)
		}
		return nil
	})

	_ = client.BulkWalk(oidLLDPRemPortDesc, func(pdu gosnmp.SnmpPDU) error {
		k := parseLLDPRemKey(pdu.Name, oidLLDPRemPortDesc)
		if k != zero {
			remPortDesc[k] = snmpString(pdu)
		}
		return nil
	})

	_ = client.BulkWalk(oidLLDPRemChassisID, func(pdu gosnmp.SnmpPDU) error {
		k := parseLLDPRemKey(pdu.Name, oidLLDPRemChassisID)
		if k != zero {
			remChassisID[k] = formatChassisIDValue(pdu)
		}
		return nil
	})

	var neighbors []DirectNeighbor
	seen := make(map[string]bool)

	for k, sysName := range remSysName {
		localPortName := locPortDesc[k.localPort]
		if localPortName == "" {
			localPortName = fmt.Sprintf("port%d", k.localPort)
		}
		remPort := remPortDesc[k]
		chassis := remChassisID[k]

		key := fmt.Sprintf("%s->%s", localPortName, sysName)
		if seen[key] {
			continue
		}
		seen[key] = true

		neighbors = append(neighbors, DirectNeighbor{
			LocalPort:  localPortName,
			RemoteMAC:  chassis,
			RemotePort: remPort,
			RemoteName: sysName,
			Protocol:   "lldp",
		})
	}

	return neighbors
}

// --- helpers ---

// NetmaskToCIDR converts an IP address and its dotted-decimal netmask into
// the network prefix in CIDR notation (e.g. "192.168.1.1"+"255.255.255.0" → "192.168.1.0/24").
// Returns empty string if parsing fails.
func NetmaskToCIDR(ip, netmask string) string {
	parsedIP := net.ParseIP(ip)
	parsedMask := net.ParseIP(netmask)
	if parsedIP == nil || parsedMask == nil {
		return ""
	}
	mask := net.IPMask(parsedMask.To4())
	network := &net.IPNet{IP: parsedIP.To4().Mask(mask), Mask: mask}
	return network.String()
}

// SplitSubnets splits a subnet spec into its individual CIDRs / IPs. Several
// entries may be given separated by commas and/or whitespace (e.g.
// "10.0.0.0/24, 10.0.1.0/24"). Blanks are dropped and duplicates removed while
// preserving order.
func SplitSubnets(spec string) []string {
	fields := strings.FieldsFunc(spec, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t' || r == '\n' || r == '\r'
	})
	seen := map[string]bool{}
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if f == "" || seen[f] {
			continue
		}
		seen[f] = true
		out = append(out, f)
	}
	return out
}

// enumerateSubnet expands a subnet spec (one or more comma/space-separated CIDRs
// or bare IPs) into the de-duplicated list of host addresses to probe.
func enumerateSubnet(subnet string) ([]string, error) {
	tokens := SplitSubnets(subnet)
	if len(tokens) == 0 {
		return nil, fmt.Errorf("no subnet specified")
	}

	seen := map[string]bool{}
	var ips []string
	add := func(ip string) {
		if !seen[ip] {
			seen[ip] = true
			ips = append(ips, ip)
		}
	}

	var firstErr error
	for _, tok := range tokens {
		ip, ipNet, err := net.ParseCIDR(tok)
		if err != nil {
			// Not a CIDR: accept a bare IP as a single host, else remember the error.
			if net.ParseIP(tok) != nil {
				add(tok)
			} else if firstErr == nil {
				firstErr = err
			}
			continue
		}
		var cidrIPs []string
		for cur := ip.Mask(ipNet.Mask); ipNet.Contains(cur); incIP(cur) {
			cidrIPs = append(cidrIPs, cur.String())
		}
		// Remove network and broadcast addresses for IPv4 /prefix <= 30.
		ones, bits := ipNet.Mask.Size()
		if bits == 32 && ones <= 30 && len(cidrIPs) >= 2 {
			cidrIPs = cidrIPs[1 : len(cidrIPs)-1]
		}
		for _, c := range cidrIPs {
			add(c)
		}
	}

	if len(ips) == 0 {
		if firstErr != nil {
			return nil, firstErr
		}
		return nil, fmt.Errorf("no usable addresses in %q", subnet)
	}
	return ips, nil
}

func incIP(ip net.IP) {
	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] != 0 {
			break
		}
	}
}

func getOrCreate(m map[int]*DeviceInterface, idx int) *DeviceInterface {
	if v, ok := m[idx]; ok {
		return v
	}
	v := &DeviceInterface{}
	m[idx] = v
	return v
}

func snmpString(pdu gosnmp.SnmpPDU) string {
	switch v := pdu.Value.(type) {
	case string:
		return strings.TrimSpace(v)
	case []byte:
		return strings.TrimSpace(string(v))
	}
	return fmt.Sprintf("%v", pdu.Value)
}

func formatMAC(pdu gosnmp.SnmpPDU) string {
	var data []byte
	switch v := pdu.Value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	}
	if len(data) != 6 {
		return ""
	}
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		data[0], data[1], data[2], data[3], data[4], data[5])
}

func formatChassisIDValue(pdu gosnmp.SnmpPDU) string {
	var data []byte
	switch v := pdu.Value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	}
	if len(data) == 7 && data[0] == 4 { // subtype 4 = MAC address
		return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
			data[1], data[2], data[3], data[4], data[5], data[6])
	}
	if len(data) > 0 {
		return strings.TrimSpace(string(data))
	}
	return ""
}

// lastOIDInt returns the last integer component of an OID string.
func lastOIDInt(oid string) int {
	parts := strings.Split(strings.TrimPrefix(oid, "."), ".")
	if len(parts) == 0 {
		return 0
	}
	var v int
	fmt.Sscanf(parts[len(parts)-1], "%d", &v)
	return v
}

// oidSuffix returns the portion of oid after the base prefix, trimmed of dots.
func oidSuffix(oid, base string) string {
	oid = strings.TrimPrefix(oid, ".")
	base = strings.TrimPrefix(base, ".")
	if strings.HasPrefix(oid, base+".") {
		return strings.TrimPrefix(oid, base+".")
	}
	return ""
}

// parsePortBitmask returns bridge port numbers (1-indexed) set in the bitmask.
func parsePortBitmask(data []byte) []int {
	var ports []int
	for byteIdx, b := range data {
		for bitIdx := 0; bitIdx < 8; bitIdx++ {
			if b&(0x80>>bitIdx) != 0 {
				ports = append(ports, byteIdx*8+bitIdx+1)
			}
		}
	}
	return ports
}

// parsePortBitmaskSet returns a set of bridge port numbers set in the bitmask.
func parsePortBitmaskSet(data []byte) map[int]bool {
	s := make(map[int]bool)
	for _, p := range parsePortBitmask(data) {
		s[p] = true
	}
	return s
}

// parseLLDPRemKey extracts (timeMark, localPortNum, remoteIndex) from an LLDP remote OID.
// OID format: <base>.<timeMark>.<localPortNum>.<remoteIndex>
type lldpRemKey struct{ timeMark, localPort, remoteIdx int }

func parseLLDPRemKey(oid, base string) lldpRemKey {
	suffix := oidSuffix(oid, base)
	if suffix == "" {
		return lldpRemKey{}
	}
	parts := strings.Split(suffix, ".")
	if len(parts) < 3 {
		return lldpRemKey{}
	}
	var tm, lp, ri int
	fmt.Sscanf(parts[0], "%d", &tm)
	fmt.Sscanf(parts[1], "%d", &lp)
	fmt.Sscanf(parts[2], "%d", &ri)
	return lldpRemKey{tm, lp, ri}
}

