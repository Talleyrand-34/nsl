#!/usr/bin/env bash
# demo-network.sh - Creates a sample network with devices connected across VLANs
#
# Topology:
#
#   [FW-01] --LAN-- [SW-CORE] --uplink-f1-- [SW-F1]
#                       |
#                   uplink-f2-- [SW-F2]
#                       |
#                   srv-port  --- [SRV-01]
#                       |
#                   store-port -- [SRV-STORE]
#
# VLANs:
#   10  Management   192.168.10.0/24
#   20  Servers      10.20.0.0/24
#   30  Users-F1     10.30.0.0/24
#   40  Users-F2     10.40.0.0/24
#   50  Storage      10.50.0.0/24
#   60  VoIP         10.60.0.0/24
#
# Port role summary (interface → VLANs):
#   FW-01:LAN          trunk  10t 20t 30t 40t 50t 60t
#   FW-01:DMZ-port     access 10u
#   SW-CORE:uplink-fw  trunk  10t 20t 30t 40t 50t 60t
#   SW-CORE:uplink-f1  trunk  10t 30t 60t
#   SW-CORE:uplink-f2  trunk  10t 40t
#   SW-CORE:srv-port   hybrid 20u 10t
#   SW-CORE:store-port hybrid 50u 10t
#   SW-F1:uplink       trunk  10t 30t 60t
#   SW-F1:access1      access 30u (users)
#   SW-F1:access2      access 60u (VoIP handsets)
#   SW-F2:uplink       trunk  10t 40t
#   SW-F2:access1      access 40u (users)
#   SRV-01:eth0        hybrid 20u 10t
#   SRV-STORE:eth0     hybrid 50u 10t
#
# VLANs are stored on DeviceInterface entities and linked to physical ports
# via InterfacePort. The --vlan-configs flag on deviceport is not used.

set -e

DB="${1:-demo.db}"
NSL="./nsl-graph"

echo "==> Creating demo network in database: $DB"
echo ""

# ── Helpers ──────────────────────────────────────────────────────────────────

get_id() {
    # get_id <print-subcommand> <name_field> <name_value>
    $NSL print "$1" -s "$DB" | jq -r --arg n "$3" '.[] | select(.'"$2"' == $n) | .id'
}

get_modelport_id() {
    # get_modelport_id <model_name> <port_name>
    $NSL print modelport -s "$DB" | jq -r \
        --arg m "$1" --arg p "$2" \
        '.[] | select(.model == $m and .name == $p) | .id'
}

get_interface_id() {
    # get_interface_id <device_id> <iface_name>
    $NSL print deviceinterface --deviceid "$1" -s "$DB" | jq -r \
        --arg n "$2" '.[] | select(.name == $n) | .id'
}

add_interface() {
    # add_interface <device_id> <iface_name> [ip=<addr>] <vlan_configs...>
    # vlan_configs format: "10:tagged" "20:untagged" ...
    # IPs now live on the interface (the device no longer takes --ips); pass
    # them with an "ip=<addr>" token anywhere in the arg list.
    local devid="$1"; shift
    local name="$1"; shift
    local ip=""
    local args=()
    for tok in "$@"; do
        case "$tok" in
            ip=*) ip="${tok#ip=}" ;;
            *)    args+=(--vlan-configs "$tok") ;;
        esac
    done
    [ -n "$ip" ] && args+=(--ips "$ip")
    $NSL add deviceinterface --deviceid "$devid" --name "$name" "${args[@]}" -s "$DB"
}

link_interface_port() {
    # link_interface_port <device_id> <iface_name> <modelport_id>
    local devid="$1"
    local name="$2"
    local mpid="$3"
    local ifid
    ifid=$(get_interface_id "$devid" "$name")
    $NSL add interfaceport --deviceid "$devid" --interfaceid "$ifid" --modelportid "$mpid" -s "$DB"
}

# ── Infrastructure ────────────────────────────────────────────────────────────

echo "--- Infrastructure ---"
$NSL add brand        --name "Cisco"        -s "$DB"
$NSL add brand        --name "Fortinet"     -s "$DB"
$NSL add brand        --name "Generic"      -s "$DB"

$NSL add deviceclass  --name "Firewall"     -s "$DB"
$NSL add deviceclass  --name "CoreSwitch"   -s "$DB"
$NSL add deviceclass  --name "AccessSwitch" -s "$DB"
$NSL add deviceclass  --name "Server"       -s "$DB"

$NSL add zonetype     --name "physical"     -s "$DB"
$NSL add proprietary  --name "NetCorp"      -s "$DB"

# ── Zones ─────────────────────────────────────────────────────────────────────

echo ""
echo "--- Zones ---"
$NSL add zone --name "DataCenter" --zonetype "physical" --proprietary "NetCorp" -s "$DB"
$NSL add zone --name "DMZ"        --zonetype "physical" --proprietary "NetCorp" -s "$DB"
$NSL add zone --name "Floor1"     --zonetype "physical" --proprietary "NetCorp" -s "$DB"
$NSL add zone --name "Floor2"     --zonetype "physical" --proprietary "NetCorp" -s "$DB"

# ── Models and model ports ────────────────────────────────────────────────────

echo ""
echo "--- Models ---"

# Firewall: WAN (outside), LAN (trunk to core), DMZ-port (DMZ segment)
$NSL add model --name "FW-Model"        --brand "Fortinet" --class "Firewall"     -s "$DB"
$NSL add modelport --modelname "FW-Model" --name "WAN"      --posx 0 --posy 0 -s "$DB"
$NSL add modelport --modelname "FW-Model" --name "LAN"      --posx 1 --posy 0 -s "$DB"
$NSL add modelport --modelname "FW-Model" --name "DMZ-port" --posx 2 --posy 0 -s "$DB"

# Core switch: uplink-fw, uplink-f1, uplink-f2, srv-port, store-port
$NSL add model --name "Core-SW-Model"   --brand "Cisco"    --class "CoreSwitch"  -s "$DB"
$NSL add modelport --modelname "Core-SW-Model" --name "uplink-fw"  --posx 0 --posy 0 -s "$DB"
$NSL add modelport --modelname "Core-SW-Model" --name "uplink-f1"  --posx 1 --posy 0 -s "$DB"
$NSL add modelport --modelname "Core-SW-Model" --name "uplink-f2"  --posx 2 --posy 0 -s "$DB"
$NSL add modelport --modelname "Core-SW-Model" --name "srv-port"   --posx 3 --posy 0 -s "$DB"
$NSL add modelport --modelname "Core-SW-Model" --name "store-port" --posx 4 --posy 0 -s "$DB"

# Access switch: uplink (trunk), access1 (users access), access2 (VoIP access)
$NSL add model --name "Access-SW-Model" --brand "Cisco"    --class "AccessSwitch" -s "$DB"
$NSL add modelport --modelname "Access-SW-Model" --name "uplink"  --posx 0 --posy 0 -s "$DB"
$NSL add modelport --modelname "Access-SW-Model" --name "access1" --posx 1 --posy 0 -s "$DB"
$NSL add modelport --modelname "Access-SW-Model" --name "access2" --posx 2 --posy 0 -s "$DB"

# Server: single NIC
$NSL add model --name "Server-Model"    --brand "Generic"  --class "Server"       -s "$DB"
$NSL add modelport --modelname "Server-Model" --name "eth0" --posx 0 --posy 0 -s "$DB"

# ── VLANs ─────────────────────────────────────────────────────────────────────

echo ""
echo "--- VLANs ---"
$NSL add vlan --vlan-id 10 --name "Management" --ip-segment "192.168.10.0/24" -s "$DB"
$NSL add vlan --vlan-id 20 --name "Servers"    --ip-segment "10.20.0.0/24"   -s "$DB"
$NSL add vlan --vlan-id 30 --name "Users-F1"   --ip-segment "10.30.0.0/24"   -s "$DB"
$NSL add vlan --vlan-id 40 --name "Users-F2"   --ip-segment "10.40.0.0/24"   -s "$DB"
$NSL add vlan --vlan-id 50 --name "Storage"    --ip-segment "10.50.0.0/24"   -s "$DB"
$NSL add vlan --vlan-id 60 --name "VoIP"       --ip-segment "10.60.0.0/24"   -s "$DB"

# ── Devices ───────────────────────────────────────────────────────────────────

echo ""
echo "--- Devices ---"
# Devices no longer carry IPs directly; IPs are assigned per interface below.
$NSL add device --label "FW-01"      --model "FW-Model"        --zone-name "DMZ"        --proprietary "NetCorp" -s "$DB"
$NSL add device --label "SW-CORE"   --model "Core-SW-Model"   --zone-name "DataCenter" --proprietary "NetCorp" -s "$DB"
$NSL add device --label "SW-F1"     --model "Access-SW-Model" --zone-name "Floor1"     --proprietary "NetCorp" -s "$DB"
$NSL add device --label "SW-F2"     --model "Access-SW-Model" --zone-name "Floor2"     --proprietary "NetCorp" -s "$DB"
$NSL add device --label "SRV-01"    --model "Server-Model"    --zone-name "DataCenter" --proprietary "NetCorp" -s "$DB"
$NSL add device --label "SRV-STORE" --model "Server-Model"    --zone-name "DataCenter" --proprietary "NetCorp" -s "$DB"

# ── Capture IDs ───────────────────────────────────────────────────────────────

echo ""
echo "--- Capturing IDs ---"

FW01_ID=$(get_id device label "FW-01")
SWCORE_ID=$(get_id device label "SW-CORE")
SWF1_ID=$(get_id device label "SW-F1")
SWF2_ID=$(get_id device label "SW-F2")
SRV01_ID=$(get_id device label "SRV-01")
SRVSTORE_ID=$(get_id device label "SRV-STORE")

MP_FW_WAN=$(get_modelport_id "FW-Model" "WAN")
MP_FW_LAN=$(get_modelport_id "FW-Model" "LAN")
MP_FW_DMZ=$(get_modelport_id "FW-Model" "DMZ-port")
MP_CORE_UPLFW=$(get_modelport_id "Core-SW-Model" "uplink-fw")
MP_CORE_UPLF1=$(get_modelport_id "Core-SW-Model" "uplink-f1")
MP_CORE_UPLF2=$(get_modelport_id "Core-SW-Model" "uplink-f2")
MP_CORE_SRV=$(get_modelport_id "Core-SW-Model" "srv-port")
MP_CORE_STORE=$(get_modelport_id "Core-SW-Model" "store-port")
MP_ACC_UPL=$(get_modelport_id "Access-SW-Model" "uplink")
MP_ACC_ACC1=$(get_modelport_id "Access-SW-Model" "access1")
MP_ACC_ACC2=$(get_modelport_id "Access-SW-Model" "access2")
MP_SRV_ETH0=$(get_modelport_id "Server-Model" "eth0")

echo "FW-01:      $FW01_ID"
echo "SW-CORE:    $SWCORE_ID"
echo "SW-F1:      $SWF1_ID"
echo "SW-F2:      $SWF2_ID"
echo "SRV-01:     $SRV01_ID"
echo "SRV-STORE:  $SRVSTORE_ID"

# ── Device ports (physical port registration, no VLANs here) ─────────────────

echo ""
echo "--- Device ports ---"

$NSL add deviceport --deviceid "$FW01_ID"    --modelportid "$MP_FW_WAN"      -s "$DB"
$NSL add deviceport --deviceid "$FW01_ID"    --modelportid "$MP_FW_LAN"      -s "$DB"
$NSL add deviceport --deviceid "$FW01_ID"    --modelportid "$MP_FW_DMZ"      -s "$DB"

$NSL add deviceport --deviceid "$SWCORE_ID"  --modelportid "$MP_CORE_UPLFW"  -s "$DB"
$NSL add deviceport --deviceid "$SWCORE_ID"  --modelportid "$MP_CORE_UPLF1"  -s "$DB"
$NSL add deviceport --deviceid "$SWCORE_ID"  --modelportid "$MP_CORE_UPLF2"  -s "$DB"
$NSL add deviceport --deviceid "$SWCORE_ID"  --modelportid "$MP_CORE_SRV"    -s "$DB"
$NSL add deviceport --deviceid "$SWCORE_ID"  --modelportid "$MP_CORE_STORE"  -s "$DB"

$NSL add deviceport --deviceid "$SWF1_ID"    --modelportid "$MP_ACC_UPL"     -s "$DB"
$NSL add deviceport --deviceid "$SWF1_ID"    --modelportid "$MP_ACC_ACC1"    -s "$DB"
$NSL add deviceport --deviceid "$SWF1_ID"    --modelportid "$MP_ACC_ACC2"    -s "$DB"

$NSL add deviceport --deviceid "$SWF2_ID"    --modelportid "$MP_ACC_UPL"     -s "$DB"
$NSL add deviceport --deviceid "$SWF2_ID"    --modelportid "$MP_ACC_ACC1"    -s "$DB"

$NSL add deviceport --deviceid "$SRV01_ID"   --modelportid "$MP_SRV_ETH0"    -s "$DB"
$NSL add deviceport --deviceid "$SRVSTORE_ID" --modelportid "$MP_SRV_ETH0"   -s "$DB"

# ── Device interfaces with VLAN configs ───────────────────────────────────────
#
# Each logical interface carries VLAN config and is linked to a physical port.
# Untagged (u) = native/access VLAN; tagged (t) = trunk VLAN.
#
# Port role            Interface name   VLANs
# FW-01:LAN            fw-lan           10t 20t 30t 40t 50t 60t  (full trunk)
# FW-01:DMZ-port       fw-dmz           10u                       (mgmt access)
# SW-CORE:uplink-fw    core-uplfw       10t 20t 30t 40t 50t 60t  (full trunk)
# SW-CORE:uplink-f1    core-uplf1       10t 30t 60t               (F1 trunk: users+VoIP)
# SW-CORE:uplink-f2    core-uplf2       10t 40t                   (F2 trunk: users)
# SW-CORE:srv-port     core-srv         20u 10t                   (server access)
# SW-CORE:store-port   core-store       50u 10t                   (storage access)
# SW-F1:uplink         f1-uplink        10t 30t 60t               (mirrors core-uplf1)
# SW-F1:access1        f1-access        30u                       (user access port)
# SW-F1:access2        f1-voip          60u                       (VoIP handset port)
# SW-F2:uplink         f2-uplink        10t 40t                   (mirrors core-uplf2)
# SW-F2:access1        f2-access        40u                       (user access port)
# SRV-01:eth0          srv-eth0         20u 10t                   (server NIC)
# SRV-STORE:eth0       store-eth0       50u 10t                   (storage NIC)

echo ""
echo "--- Device interfaces ---"

add_interface "$FW01_ID"     "fw-lan"      "10:tagged" "20:tagged" "30:tagged" "40:tagged" "50:tagged" "60:tagged"
add_interface "$FW01_ID"     "fw-dmz"      ip=10.0.0.1 "10:untagged"

add_interface "$SWCORE_ID"   "core-uplfw"  ip=10.0.0.2 "10:tagged" "20:tagged" "30:tagged" "40:tagged" "50:tagged" "60:tagged"
add_interface "$SWCORE_ID"   "core-uplf1"  "10:tagged" "30:tagged" "60:tagged"
add_interface "$SWCORE_ID"   "core-uplf2"  "10:tagged" "40:tagged"
add_interface "$SWCORE_ID"   "core-srv"    "20:untagged" "10:tagged"
add_interface "$SWCORE_ID"   "core-store"  "50:untagged" "10:tagged"

add_interface "$SWF1_ID"     "f1-uplink"   ip=10.0.0.3 "10:tagged" "30:tagged" "60:tagged"
add_interface "$SWF1_ID"     "f1-access"   "30:untagged"
add_interface "$SWF1_ID"     "f1-voip"     "60:untagged"

add_interface "$SWF2_ID"     "f2-uplink"   ip=10.0.0.4 "10:tagged" "40:tagged"
add_interface "$SWF2_ID"     "f2-access"   "40:untagged"

add_interface "$SRV01_ID"    "srv-eth0"    ip=10.20.0.100 "20:untagged" "10:tagged"
add_interface "$SRVSTORE_ID" "store-eth0"  ip=10.50.0.100 "50:untagged" "10:tagged"

# ── Link interfaces to physical ports ─────────────────────────────────────────

echo ""
echo "--- Interface-port links ---"

link_interface_port "$FW01_ID"     "fw-lan"      "$MP_FW_LAN"
link_interface_port "$FW01_ID"     "fw-dmz"      "$MP_FW_DMZ"

link_interface_port "$SWCORE_ID"   "core-uplfw"  "$MP_CORE_UPLFW"
link_interface_port "$SWCORE_ID"   "core-uplf1"  "$MP_CORE_UPLF1"
link_interface_port "$SWCORE_ID"   "core-uplf2"  "$MP_CORE_UPLF2"
link_interface_port "$SWCORE_ID"   "core-srv"    "$MP_CORE_SRV"
link_interface_port "$SWCORE_ID"   "core-store"  "$MP_CORE_STORE"

link_interface_port "$SWF1_ID"     "f1-uplink"   "$MP_ACC_UPL"
link_interface_port "$SWF1_ID"     "f1-access"   "$MP_ACC_ACC1"
link_interface_port "$SWF1_ID"     "f1-voip"     "$MP_ACC_ACC2"

link_interface_port "$SWF2_ID"     "f2-uplink"   "$MP_ACC_UPL"
link_interface_port "$SWF2_ID"     "f2-access"   "$MP_ACC_ACC1"

link_interface_port "$SRV01_ID"    "srv-eth0"    "$MP_SRV_ETH0"
link_interface_port "$SRVSTORE_ID" "store-eth0"  "$MP_SRV_ETH0"

# ── Connections ───────────────────────────────────────────────────────────────

echo ""
echo "--- Connections ---"

$NSL add connection \
    --from-device-id "$FW01_ID"   --from-modelport-id "$MP_FW_LAN" \
    --to-device-id   "$SWCORE_ID" --to-modelport-id   "$MP_CORE_UPLFW" \
    --allow-vlan-union -s "$DB"

$NSL add connection \
    --from-device-id "$SWCORE_ID" --from-modelport-id "$MP_CORE_UPLF1" \
    --to-device-id   "$SWF1_ID"   --to-modelport-id   "$MP_ACC_UPL" \
    --allow-vlan-union -s "$DB"

$NSL add connection \
    --from-device-id "$SWCORE_ID" --from-modelport-id "$MP_CORE_UPLF2" \
    --to-device-id   "$SWF2_ID"   --to-modelport-id   "$MP_ACC_UPL" \
    --allow-vlan-union -s "$DB"

$NSL add connection \
    --from-device-id "$SWCORE_ID"  --from-modelport-id "$MP_CORE_SRV" \
    --to-device-id   "$SRV01_ID"   --to-modelport-id   "$MP_SRV_ETH0" \
    --allow-vlan-union -s "$DB"

$NSL add connection \
    --from-device-id "$SWCORE_ID"   --from-modelport-id "$MP_CORE_STORE" \
    --to-device-id   "$SRVSTORE_ID" --to-modelport-id   "$MP_SRV_ETH0" \
    --allow-vlan-union -s "$DB"

# ── Summary ───────────────────────────────────────────────────────────────────

echo ""
echo "==> Done. Network summary:"
$NSL print summary -s "$DB" 2>/dev/null || $NSL print devices -s "$DB"
echo ""
echo "To generate VLAN diagrams (use -s $DB for all commands):"
echo ""
echo "  # Default: untagged scope, color both connections and ports"
echo "  ./nsl-graph diagram connection-vlan -s $DB"
echo ""
echo "  # All VLANs in intersection — multiple colored lines per trunk link"
echo "  ./nsl-graph diagram connection-vlan --vlan-scope all -s $DB"
echo ""
echo "  # All VLANs, color connections only (plain port nodes)"
echo "  ./nsl-graph diagram connection-vlan --vlan-scope all --color-target connections -s $DB"
echo ""
echo "  # Untagged scope, color ports only (plain connection lines)"
echo "  ./nsl-graph diagram port-vlan --color-target ports -s $DB"
echo ""
echo "  # All VLANs, color both"
echo "  ./nsl-graph diagram connection-vlan --vlan-scope all --color-target both -s $DB"
echo ""
echo "Expected intersection counts per link (vlan-scope=all):"
echo "  FW-01:LAN  <-> SW-CORE:uplink-fw  : 6 VLANs (10,20,30,40,50,60)"
echo "  SW-CORE:uplink-f1 <-> SW-F1:uplink : 3 VLANs (10,30,60)"
echo "  SW-CORE:uplink-f2 <-> SW-F2:uplink : 2 VLANs (10,40)"
echo "  SW-CORE:srv-port  <-> SRV-01:eth0  : 2 VLANs (10,20)"
echo "  SW-CORE:store-port<-> SRV-STORE:eth0: 2 VLANs (10,50)"
