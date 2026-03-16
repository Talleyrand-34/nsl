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

	// CDP cache table
	oidCDPCacheDeviceID   = "1.3.6.1.4.1.9.9.23.1.2.1.1.6"
	oidCDPCacheDevicePort = "1.3.6.1.4.1.9.9.23.1.2.1.1.7"
	oidCDPCacheAddress    = "1.3.6.1.4.1.9.9.23.1.2.1.1.4"
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

	for _, ip := range ips {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			dev, err := ss.ScanDevice(ip, options)
			if err == nil && dev.Reachable {
				mu.Lock()
				result.Devices = append(result.Devices, *dev)
				mu.Unlock()
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
		neighbors = ss.queryCDPNeighbors(client, ifaces)
	}
	device.Neighbors = neighbors

	return device, nil
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

	// ifType (24 = softwareLoopback, skip)
	_ = client.BulkWalk(oidIfType, func(pdu gosnmp.SnmpPDU) error {
		idx := lastOIDInt(pdu.Name)
		if idx > 0 {
			iface := getOrCreate(ifaceMap, idx)
			iface.Index = idx
			if v, ok := pdu.Value.(int); ok {
				if v == 24 { // loopback
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

// queryCDPNeighbors retrieves CDP cache table entries (Cisco-specific).
func (ss *SNMPScanner) queryCDPNeighbors(client *gosnmp.GoSNMP, ifaces []DeviceInterface) []DirectNeighbor {
	ifIdxToName := make(map[int]string)
	for _, iface := range ifaces {
		ifIdxToName[iface.Index] = iface.Name
	}

	cdpDeviceID := make(map[cdpKey]string)
	cdpDevicePort := make(map[cdpKey]string)
	cdpAddress := make(map[cdpKey]string)
	zero := cdpKey{}

	_ = client.BulkWalk(oidCDPCacheDeviceID, func(pdu gosnmp.SnmpPDU) error {
		k := parseCDPKey(pdu.Name, oidCDPCacheDeviceID)
		if k != zero {
			cdpDeviceID[k] = snmpString(pdu)
		}
		return nil
	})

	_ = client.BulkWalk(oidCDPCacheDevicePort, func(pdu gosnmp.SnmpPDU) error {
		k := parseCDPKey(pdu.Name, oidCDPCacheDevicePort)
		if k != zero {
			cdpDevicePort[k] = snmpString(pdu)
		}
		return nil
	})

	_ = client.BulkWalk(oidCDPCacheAddress, func(pdu gosnmp.SnmpPDU) error {
		k := parseCDPKey(pdu.Name, oidCDPCacheAddress)
		if k != zero {
			cdpAddress[k] = parseCDPAddress(pdu)
		}
		return nil
	})

	var neighbors []DirectNeighbor
	for k, deviceID := range cdpDeviceID {
		localPort := ifIdxToName[k.ifIdx]
		if localPort == "" {
			localPort = fmt.Sprintf("ifIndex%d", k.ifIdx)
		}
		neighbors = append(neighbors, DirectNeighbor{
			LocalPort:  localPort,
			RemoteIP:   cdpAddress[k],
			RemotePort: cdpDevicePort[k],
			RemoteName: deviceID,
			Protocol:   "cdp",
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

func enumerateSubnet(subnet string) ([]string, error) {
	ip, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, err
	}

	var ips []string
	for cur := ip.Mask(ipNet.Mask); ipNet.Contains(cur); incIP(cur) {
		ips = append(ips, cur.String())
	}
	// Remove network and broadcast addresses for IPv4 /prefix <= 30
	ones, bits := ipNet.Mask.Size()
	if bits == 32 && ones <= 30 && len(ips) >= 2 {
		ips = ips[1 : len(ips)-1]
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

// parseCDPKey extracts (ifIndex, deviceNum) from a CDP cache OID.
// OID format: <base>.<ifIndex>.<deviceNum>
type cdpKey struct{ ifIdx, devNum int }

func parseCDPKey(oid, base string) cdpKey {
	suffix := oidSuffix(oid, base)
	if suffix == "" {
		return cdpKey{}
	}
	parts := strings.Split(suffix, ".")
	if len(parts) < 2 {
		return cdpKey{}
	}
	var ifIdx, devNum int
	fmt.Sscanf(parts[0], "%d", &ifIdx)
	fmt.Sscanf(parts[1], "%d", &devNum)
	return cdpKey{ifIdx, devNum}
}

// parseCDPAddress extracts an IP address from CDP cache address PDU.
// CDP address format: 4 bytes of address type + address bytes.
func parseCDPAddress(pdu gosnmp.SnmpPDU) string {
	var data []byte
	switch v := pdu.Value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	}
	// CDP NLPID-encoded address: first 4 bytes = length/type, then 4 bytes = IPv4
	if len(data) >= 8 {
		ipBytes := data[4:8]
		return net.IP(ipBytes).String()
	}
	return ""
}
