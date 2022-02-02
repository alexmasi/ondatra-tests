/* Test BGP Route Installation

Topology:
IXIA (40.40.40.0/24, 0:40:40:40::0/64) -----> ARISTA ------> IXIA (50.50.50.0/24, 0:50:50:50::0/64)

Flows:
- permit v4: 40.40.40.1 -> 50.50.50.1+
- deny v4: 40.40.40.1 -> 60.60.60.1+
- permit v4: 0:40:40:40::1 -> 0:50:50:50::1+
- deny v4: 0:40:40:40::1 -> 0:60:60:60::1+
*/
package tests

import (
	"testing"

	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/openconfig/ondatra"

	"tests/tests/helpers"

	oc "github.com/openconfig/ondatra/telemetry"
)

const (
	routerId = 3333
)

func routeInstallConfigureInterface(t *testing.T, dut *ondatra.DUTDevice, cfg gosnappi.Config) {
	t.Logf("Start DUT Interface Config")
	dc := dut.Config()

	dutSrc := helpers.Attributes{
		IPv4:    cfg.Devices().Items()[0].Bgp().Ipv4Interfaces().Items()[0].Peers().Items()[0].PeerAddress(),
		IPv6:    cfg.Devices().Items()[0].Bgp().Ipv6Interfaces().Items()[0].Peers().Items()[0].PeerAddress(),
		IPv4Len: 24,
		IPv6Len: 64,
	}
	i1 := dutSrc.NewInterface(helpers.InterfaceMap[dut.Port(t, "port1").Name()])
	dc.Interface(i1.GetName()).Replace(t, i1)

	dutDst := helpers.Attributes{
		IPv4:    cfg.Devices().Items()[1].Bgp().Ipv4Interfaces().Items()[0].Peers().Items()[0].PeerAddress(),
		IPv6:    cfg.Devices().Items()[1].Bgp().Ipv6Interfaces().Items()[0].Peers().Items()[0].PeerAddress(),
		IPv4Len: 24,
		IPv6Len: 64,
	}
	i2 := dutDst.NewInterface(helpers.InterfaceMap[dut.Port(t, "port2").Name()])
	dc.Interface(i2.GetName()).Replace(t, i2)
}

func routeInstallBuildNbrList(cfg gosnappi.Config) []*helpers.BgpNeighbor {
	nbr1v4 := &helpers.BgpNeighbor{As: uint32(cfg.Devices().Items()[0].Bgp().Ipv4Interfaces().Items()[0].Peers().Items()[0].AsNumber()), NeighborIP: cfg.Devices().Items()[0].Ethernets().Items()[0].Ipv4Addresses().Items()[0].Address(), IsV4: true}
	nbr1v6 := &helpers.BgpNeighbor{As: uint32(cfg.Devices().Items()[0].Bgp().Ipv6Interfaces().Items()[0].Peers().Items()[0].AsNumber()), NeighborIP: cfg.Devices().Items()[0].Ethernets().Items()[0].Ipv6Addresses().Items()[0].Address(), IsV4: false}
	nbr2v4 := &helpers.BgpNeighbor{As: uint32(cfg.Devices().Items()[1].Bgp().Ipv4Interfaces().Items()[0].Peers().Items()[0].AsNumber()), NeighborIP: cfg.Devices().Items()[1].Ethernets().Items()[0].Ipv4Addresses().Items()[0].Address(), IsV4: true}
	nbr2v6 := &helpers.BgpNeighbor{As: uint32(cfg.Devices().Items()[1].Bgp().Ipv6Interfaces().Items()[0].Peers().Items()[0].AsNumber()), NeighborIP: cfg.Devices().Items()[1].Ethernets().Items()[0].Ipv6Addresses().Items()[0].Address(), IsV4: false}
	return []*helpers.BgpNeighbor{nbr1v4, nbr2v4, nbr1v6, nbr2v6}
}

func routeInstallConfigureBGP(t *testing.T, dut *ondatra.DUTDevice, cfg gosnappi.Config) {
	t.Logf("Start DUT BGP Config")
	dutConfPath := dut.Config().NetworkInstance("default").Protocol(oc.PolicyTypes_INSTALL_PROTOCOL_TYPE_BGP, "BGP").Bgp()
	helpers.LogYgot(t, "DUT BGP Config before", dutConfPath, dutConfPath.Get(t))
	dutConfPath.Replace(t, nil)
	nbrList := routeInstallBuildNbrList(cfg)
	dutConf := helpers.BgpAppendNbr(routerId, nbrList)
	dutConfPath.Replace(t, dutConf)
}

func routeInstallConfigureDUT(t *testing.T, dut *ondatra.DUTDevice, cfg gosnappi.Config) {
	t.Logf("Start Setting DUT Config")

	routeInstallConfigureInterface(t, dut, cfg)
	helpers.ConfigDUTs(map[string]string{"arista1": "../resources/dutconfig/bgp_route_install/set_dut.txt"})
	routeInstallConfigureBGP(t, dut, cfg)
}

func routeInstallUnsetInterface(t *testing.T, dut *ondatra.DUTDevice) {
	t.Logf("Start Unsetting DUT Interface Config")
	dc := dut.Config()

	i1 := helpers.RemoveInterface(helpers.InterfaceMap[dut.Port(t, "port1").Name()])
	dc.Interface(i1.GetName()).Replace(t, i1)

	i2 := helpers.RemoveInterface(helpers.InterfaceMap[dut.Port(t, "port2").Name()])
	dc.Interface(i2.GetName()).Replace(t, i2)
}

func routeInstallUnsetBGP(t *testing.T, dut *ondatra.DUTDevice, cfg gosnappi.Config) {
	t.Logf("Start Removing DUT BGP Neighbor")
	dutConfPath := dut.Config().NetworkInstance("default").Protocol(oc.PolicyTypes_INSTALL_PROTOCOL_TYPE_BGP, "BGP").Bgp()
	helpers.LogYgot(t, "DUT BGP Config before", dutConfPath, dutConfPath.Get(t))
	dutConfPath.Replace(t, nil)
	nbrList := routeInstallBuildNbrList(cfg)
	dutConf := helpers.BgpDeleteNbr(nbrList)
	dutConfPath.Replace(t, dutConf)
}

func routeInstallUnsetDUT(t *testing.T, dut *ondatra.DUTDevice, cfg gosnappi.Config) {
	t.Logf("Start Un-Setting DUT Config")
	// helpers.ConfigDUTs(map[string]string{"arista1": "../resources/dutconfig/bgp_route_install/unset_dut.txt"})

	routeInstallUnsetInterface(t, dut)
	routeInstallUnsetBGP(t, dut, cfg)
}

func routeInstallCheckBgpParameters(t *testing.T, dut *ondatra.DUTDevice, cfg gosnappi.Config) {
	ateSrc := helpers.Attributes{
		IPv4:    cfg.Devices().Items()[0].Ethernets().Items()[0].Ipv4Addresses().Items()[0].Address(),
		IPv6:    cfg.Devices().Items()[0].Ethernets().Items()[0].Ipv6Addresses().Items()[0].Address(),
		IPv4Len: uint8(cfg.Devices().Items()[1].Ethernets().Items()[0].Ipv4Addresses().Items()[0].Prefix()),
		IPv6Len: uint8(cfg.Devices().Items()[1].Ethernets().Items()[0].Ipv6Addresses().Items()[0].Prefix()),
	}

	ateDst := helpers.Attributes{
		IPv4:    cfg.Devices().Items()[1].Ethernets().Items()[0].Ipv4Addresses().Items()[0].Address(),
		IPv6:    cfg.Devices().Items()[1].Ethernets().Items()[0].Ipv6Addresses().Items()[0].Address(),
		IPv4Len: uint8(cfg.Devices().Items()[1].Ethernets().Items()[0].Ipv4Addresses().Items()[0].Prefix()),
		IPv6Len: uint8(cfg.Devices().Items()[1].Ethernets().Items()[0].Ipv6Addresses().Items()[0].Prefix()),
	}

	srcAS := uint32(cfg.Devices().Items()[0].Bgp().Ipv4Interfaces().Items()[0].Peers().Items()[0].AsNumber())
	dstAS := uint32(cfg.Devices().Items()[1].Bgp().Ipv4Interfaces().Items()[0].Peers().Items()[0].AsNumber())

	statePath := dut.Telemetry().NetworkInstance("default").Protocol(oc.PolicyTypes_INSTALL_PROTOCOL_TYPE_BGP, "BGP").Bgp()
	nbrPath_1 := statePath.Neighbor(ateSrc.IPv4)
	nbrPathv6_1 := statePath.Neighbor(ateSrc.IPv6)
	nbrPath_2 := statePath.Neighbor(ateDst.IPv4)
	nbrPathv6_2 := statePath.Neighbor(ateDst.IPv6)

	// Get BGP adjacency state
	t.Logf("Verifying BGP Adjacency State")
	status := nbrPath_1.SessionState().Get(t)
	t.Logf("BGP adjacency for %s: %s", ateSrc.IPv4, status)
	if want := oc.Bgp_Neighbor_SessionState_ESTABLISHED; status != want {
		t.Errorf("Get(BGP peer %s status): got %d, want %d", ateSrc.IPv4, status, want)
	}

	status = nbrPathv6_1.SessionState().Get(t)
	t.Logf("BGP adjacency for %s: %s", ateSrc.IPv6, status)
	if want := oc.Bgp_Neighbor_SessionState_ESTABLISHED; status != want {
		t.Errorf("Get(BGP peer %s status): got %d, want %d", ateSrc.IPv6, status, want)
	}

	status = nbrPath_2.SessionState().Get(t)
	t.Logf("BGP adjacency for %s: %s", ateDst.IPv4, status)
	if want := oc.Bgp_Neighbor_SessionState_ESTABLISHED; status != want {
		t.Errorf("Get(BGP peer %s status): got %d, want %d", ateDst.IPv4, status, want)
	}

	status = nbrPathv6_2.SessionState().Get(t)
	t.Logf("BGP adjacency for %s: %s", ateDst.IPv6, status)
	if want := oc.Bgp_Neighbor_SessionState_ESTABLISHED; status != want {
		t.Errorf("Get(BGP peer %s status): got %d, want %d", ateDst.IPv6, status, want)
	}

	nbr_1 := statePath.Get(t).GetNeighbor(ateSrc.IPv4)
	nbrv6_1 := statePath.Get(t).GetNeighbor(ateSrc.IPv6)
	nbr_2 := statePath.Get(t).GetNeighbor(ateDst.IPv4)
	nbrv6_2 := statePath.Get(t).GetNeighbor(ateDst.IPv6)

	// Check BGP Transitions
	t.Logf("Verifying BGP Transitions")
	estTrans := nbr_1.GetEstablishedTransitions()
	t.Logf("Got established transitions for Neighbor %s : %d", ateSrc.IPv4, estTrans)
	if estTrans != 1 {
		t.Errorf("Wrong established-transitions: got %v, want 1", estTrans)
	}

	estTrans = nbrv6_1.GetEstablishedTransitions()
	t.Logf("Got established transitions for Neighbor %s : %d", ateSrc.IPv6, estTrans)
	if estTrans != 1 {
		t.Errorf("Wrong established-transitions: got %v, want 1", estTrans)
	}

	estTrans = nbr_2.GetEstablishedTransitions()
	t.Logf("Got established transitions for Neighbor %s : %d", ateDst.IPv4, estTrans)
	if estTrans != 1 {
		t.Errorf("Wrong established-transitions: got %v, want 1", estTrans)
	}

	estTrans = nbrv6_2.GetEstablishedTransitions()
	t.Logf("Got established transitions for Neighbor %s : %d", ateDst.IPv6, estTrans)
	if estTrans != 1 {
		t.Errorf("Wrong established-transitions: got %v, want 1", estTrans)
	}

	// Check BGP neighbor address from telemetry
	t.Logf("Verifying BGP Neighbor Addresses")
	addr_1 := nbrPath_1.Get(t).GetNeighborAddress()
	addrv6_1 := nbrPathv6_1.Get(t).GetNeighborAddress()
	addr_2 := nbrPath_2.Get(t).GetNeighborAddress()
	adddrv6_2 := nbrPathv6_2.Get(t).GetNeighborAddress()

	t.Logf("Got neighbor address: %s", addr_1)
	if addr_1 != ateSrc.IPv4 {
		t.Errorf("Bgp neighbor address: got %v, want %v", addr_1, ateSrc.IPv4)
	}
	t.Logf("Got neighbor address: %s", addrv6_1)
	if addrv6_1 != ateSrc.IPv6 {
		t.Errorf("Bgp neighbor address: got %v, want %v", addrv6_1, ateSrc.IPv6)
	}
	t.Logf("Got neighbor address: %s", addr_2)
	if addr_2 != ateDst.IPv4 {
		t.Errorf("Bgp neighbor address: got %v, want %v", addr_2, ateDst.IPv4)
	}
	t.Logf("Got neighbor address: %s", adddrv6_2)
	if adddrv6_2 != ateDst.IPv6 {
		t.Errorf("Bgp neighbor address: got %v, want %v", ateDst, ateDst.IPv6)
	}

	// Check BGP neighbor address from telemetry
	t.Logf("Verifying BGP Neighbor AS Number")
	peerAS_1 := nbrPath_1.Get(t).GetPeerAs()
	peerAS_2 := nbrPath_2.Get(t).GetPeerAs()
	peerv6AS_1 := nbrPathv6_1.Get(t).GetPeerAs()
	peerv6AS_2 := nbrPathv6_2.Get(t).GetPeerAs()

	t.Logf("Got neighbor %s AS: %d", ateSrc.IPv4, peerAS_1)
	if peerAS_1 != srcAS {
		t.Errorf("Bgp peerAs: got %v, want %v", peerAS_1, srcAS)
	}

	t.Logf("Got neighbor %s AS: %d", ateSrc.IPv6, peerv6AS_1)
	if peerv6AS_1 != srcAS {
		t.Errorf("Bgp peerAs: got %v, want %v", peerv6AS_1, srcAS)
	}

	t.Logf("Got neighbor %s AS: %d", ateDst.IPv4, peerAS_2)
	if peerAS_2 != dstAS {
		t.Errorf("Bgp peerAs: got %v, want %v", peerAS_2, dstAS)
	}

	t.Logf("Got neighbor %s AS: %d", ateDst.IPv6, peerv6AS_2)
	if peerv6AS_2 != dstAS {
		t.Errorf("Bgp peerAs: got %v, want %v", peerv6AS_2, dstAS)
	}

	// Check BGP neighbor is enabled
	t.Logf("Verifying BGP Neighbors Are Enabled")
	if !nbrPath_1.Get(t).GetEnabled() {
		t.Errorf("Expected neighbor %v to be enabled", ateSrc.IPv4)
	} else {
		t.Logf("Neighbor %v is enabled", ateSrc.IPv4)
	}

	if !nbrPath_2.Get(t).GetEnabled() {
		t.Errorf("Expected neighbor %v to be enabled", ateDst.IPv6)
	} else {
		t.Logf("Neighbor %v is enabled", ateDst.IPv6)
	}

	if !nbrPathv6_1.Get(t).GetEnabled() {
		t.Errorf("Expected neighbor %v to be enabled", ateSrc.IPv6)
	} else {
		t.Logf("Neighbor %v is enabled", ateSrc.IPv6)
	}

	if !nbrPathv6_2.Get(t).GetEnabled() {
		t.Errorf("Expected neighbor %v to be enabled", ateDst.IPv4)
	} else {
		t.Logf("Neighbor %v is enabled", ateDst.IPv4)
	}
}

func TestBGPRouteInstall(t *testing.T) {
	ate := ondatra.ATE(t, "ate1")
	ondatra.ATE(t, "ate2")

	otg := ate.OTG()
	defer helpers.CleanupTest(otg, t, true)

	config, expected := bgpRouteInstallConfig(t, otg)

	dut := ondatra.DUT(t, "dut")

	// Set DUT Config over gNMI
	routeInstallConfigureDUT(t, dut, config)

	// Unset DUT Config over gNMI
	defer routeInstallUnsetDUT(t, dut, config)

	otg.PushConfig(t, config)
	otg.StartProtocols(t)

	gnmiClient, err := helpers.NewGnmiClient(otg.NewGnmiQuery(t), config)
	if err != nil {
		t.Fatal(err)
	}

	helpers.WaitFor(t, func() (bool, error) { return gnmiClient.AllBgp4SessionUp(expected) }, nil)
	helpers.WaitFor(t, func() (bool, error) { return gnmiClient.AllBgp6SessionUp(expected) }, nil)

	t.Logf("Verifying Port Status")
	helpers.VerifyPortsUp(t, dut.Device)

	t.Logf("Check BGP Parameters")
	routeInstallCheckBgpParameters(t, dut, config)

	otg.StartTraffic(t)

	helpers.WaitFor(t, func() (bool, error) { return gnmiClient.FlowMetricsOk(expected) }, nil)
}

func bgpRouteInstallConfig(t *testing.T, otg *ondatra.OTGAPI) (gosnappi.Config, helpers.ExpectedState) {
	config := otg.NewConfig(t)

	port1 := config.Ports().Add().SetName("ixia-c-port1")
	port2 := config.Ports().Add().SetName("ixia-c-port2")

	dutPort1 := config.Devices().Add().SetName("dutPort1")
	dutPort1Eth := dutPort1.Ethernets().Add().
		SetName("dutPort1.eth").
		SetPortName(port1.Name()).
		SetMac("00:00:01:01:01:01")
	dutPort1Ipv4 := dutPort1Eth.Ipv4Addresses().Add().
		SetName("dutPort1.ipv4").
		SetAddress("1.1.1.1").
		SetGateway("1.1.1.3").
		SetPrefix(24)
	dutPort1Ipv6 := dutPort1Eth.Ipv6Addresses().Add().
		SetName("dutPort1.ipv6").
		SetAddress("0:1:1:1::1").
		SetGateway("0:1:1:1::3").
		SetPrefix(64)
	dutPort2 := config.Devices().Add().SetName("dutPort2")
	dutPort2Eth := dutPort2.Ethernets().Add().
		SetName("dutPort2.eth").
		SetPortName(port2.Name()).
		SetMac("00:00:02:01:01:01")
	dutPort2Ipv4 := dutPort2Eth.Ipv4Addresses().Add().
		SetName("dutPort2.ipv4").
		SetAddress("2.2.2.2").
		SetGateway("2.2.2.3").
		SetPrefix(24)
	dutPort2Ipv6 := dutPort2Eth.Ipv6Addresses().Add().
		SetName("dutPort2.ipv6").
		SetAddress("0:2:2:2::2").
		SetGateway("0:2:2:2::3").
		SetPrefix(64)

	dutPort1Bgp := dutPort1.Bgp().
		SetRouterId(dutPort1Ipv4.Address())
	dutPort1Bgp4Peer := dutPort1Bgp.Ipv4Interfaces().Add().
		SetIpv4Name(dutPort1Ipv4.Name()).
		Peers().Add().
		SetName("dutPort1.bgp4.peer").
		SetPeerAddress(dutPort1Ipv4.Gateway()).
		SetAsNumber(1111).
		SetAsType(gosnappi.BgpV4PeerAsType.EBGP)
	dutPort1Bgp6Peer := dutPort1Bgp.Ipv6Interfaces().Add().
		SetIpv6Name(dutPort1Ipv6.Name()).
		Peers().Add().
		SetName("dutPort1.bgp6.peer").
		SetPeerAddress(dutPort1Ipv6.Gateway()).
		SetAsNumber(1111).
		SetAsType(gosnappi.BgpV6PeerAsType.EBGP)

	dutPort1Bgp4PeerRoutes := dutPort1Bgp4Peer.V4Routes().Add().
		SetName("dutPort1.bgp4.peer.rr4").
		SetNextHopIpv4Address(dutPort1Ipv4.Address()).
		SetNextHopAddressType(gosnappi.BgpV4RouteRangeNextHopAddressType.IPV4).
		SetNextHopMode(gosnappi.BgpV4RouteRangeNextHopMode.MANUAL)
	dutPort1Bgp4PeerRoutes.Addresses().Add().
		SetAddress("40.40.40.0").
		SetPrefix(24).
		SetCount(5).
		SetStep(2)
	dutPort1Bgp6PeerRoutes := dutPort1Bgp6Peer.V6Routes().Add().
		SetName("dutPort1.bgp4.peer.rr6").
		SetNextHopIpv6Address(dutPort1Ipv6.Address()).
		SetNextHopAddressType(gosnappi.BgpV6RouteRangeNextHopAddressType.IPV6).
		SetNextHopMode(gosnappi.BgpV6RouteRangeNextHopMode.MANUAL)
	dutPort1Bgp6PeerRoutes.Addresses().Add().
		SetAddress("0:40:40:40::0").
		SetPrefix(64).
		SetCount(5).
		SetStep(2)

	dutPort2Bgp := dutPort2.Bgp().
		SetRouterId(dutPort2Ipv4.Address())
	dutPort2BgpIf4 := dutPort2Bgp.Ipv4Interfaces().Add().
		SetIpv4Name(dutPort2Ipv4.Name())
	dutPort2Bgp4Peer := dutPort2BgpIf4.Peers().Add().
		SetName("dutPort2.bgp4.peer").
		SetPeerAddress(dutPort2Ipv4.Gateway()).
		SetAsNumber(2222).
		SetAsType(gosnappi.BgpV4PeerAsType.EBGP)
	dutPort2Bgp6Peer := dutPort2Bgp.Ipv6Interfaces().Add().
		SetIpv6Name(dutPort2Ipv6.Name()).
		Peers().Add().
		SetName("dutPort2.bgp6.peer").
		SetPeerAddress(dutPort2Ipv6.Gateway()).
		SetAsNumber(2222).
		SetAsType(gosnappi.BgpV6PeerAsType.EBGP)

	dutPort2Bgp4PeerRoutes := dutPort2Bgp4Peer.V4Routes().Add().
		SetName("dutPort2.bgp4.peer.rr4").
		SetNextHopIpv4Address(dutPort2Ipv4.Address()).
		SetNextHopAddressType(gosnappi.BgpV4RouteRangeNextHopAddressType.IPV4).
		SetNextHopMode(gosnappi.BgpV4RouteRangeNextHopMode.MANUAL)
	dutPort2Bgp4PeerRoutes.Addresses().Add().
		SetAddress("50.50.50.0").
		SetPrefix(24).
		SetCount(5).
		SetStep(2)
	dutPort2Bgp6PeerRoutes := dutPort2Bgp6Peer.V6Routes().Add().
		SetName("dutPort2.bgp4.peer.rr6").
		SetNextHopIpv6Address(dutPort2Ipv6.Address()).
		SetNextHopAddressType(gosnappi.BgpV6RouteRangeNextHopAddressType.IPV6).
		SetNextHopMode(gosnappi.BgpV6RouteRangeNextHopMode.MANUAL)
	dutPort2Bgp6PeerRoutes.Addresses().Add().
		SetAddress("0:50:50:50::0").
		SetPrefix(64).
		SetCount(5).
		SetStep(2)

	// OTG traffic configuration
	f1 := config.Flows().Add().SetName("p1.v4.p2.permit")
	f1.Metrics().SetEnable(true)
	f1.TxRx().Device().
		SetTxNames([]string{dutPort1Bgp4PeerRoutes.Name()}).
		SetRxNames([]string{dutPort2Bgp4PeerRoutes.Name()})
	f1.Size().SetFixed(512)
	f1.Rate().SetPps(500)
	f1.Duration().FixedPackets().SetPackets(1000)
	e1 := f1.Packet().Add().Ethernet()
	e1.Src().SetValue(dutPort1Eth.Mac())
	e1.Dst().SetValue("00:00:00:00:00:00")
	v4 := f1.Packet().Add().Ipv4()
	v4.Src().SetValue("40.40.40.1")
	v4.Dst().Increment().SetStart("50.50.50.1").SetStep("0.0.0.1").SetCount(5)

	f1d := config.Flows().Add().SetName("p1.v4.p2.deny")
	f1d.Metrics().SetEnable(true)
	f1d.TxRx().Device().
		SetTxNames([]string{dutPort1Bgp4PeerRoutes.Name()}).
		SetRxNames([]string{dutPort2Bgp4PeerRoutes.Name()})
	f1d.Size().SetFixed(512)
	f1d.Rate().SetPps(500)
	f1d.Duration().FixedPackets().SetPackets(1000)
	e1d := f1d.Packet().Add().Ethernet()
	e1d.Src().SetValue(dutPort1Eth.Mac())
	e1d.Dst().SetValue("00:00:00:00:00:00")
	v4d := f1d.Packet().Add().Ipv4()
	v4d.Src().SetValue("40.40.40.1")
	v4d.Dst().Increment().SetStart("60.60.60.1").SetStep("0.0.0.1").SetCount(5)

	f2 := config.Flows().Add().SetName("p1.v6.p2.permit")
	f2.Metrics().SetEnable(true)
	f2.TxRx().Device().
		SetTxNames([]string{dutPort1Bgp6PeerRoutes.Name()}).
		SetRxNames([]string{dutPort2Bgp6PeerRoutes.Name()})
	f2.Size().SetFixed(512)
	f2.Rate().SetPps(500)
	f2.Duration().FixedPackets().SetPackets(1000)
	e2 := f2.Packet().Add().Ethernet()
	e2.Src().SetValue(dutPort1Eth.Mac())
	e2.Dst().SetValue("00:00:00:00:00:00")
	v6 := f2.Packet().Add().Ipv6()
	v6.Src().SetValue("0:40:40:40::1")
	v6.Dst().Increment().SetStart("0:50:50:50::1").SetStep("::1").SetCount(5)

	f2d := config.Flows().Add().SetName("p1.v6.p2.deny")
	f2d.Metrics().SetEnable(true)
	f2d.TxRx().Device().
		SetTxNames([]string{dutPort1Bgp6PeerRoutes.Name()}).
		SetRxNames([]string{dutPort2Bgp6PeerRoutes.Name()})
	f2d.Size().SetFixed(512)
	f2d.Rate().SetPps(500)
	f2d.Duration().FixedPackets().SetPackets(1000)
	e2d := f2d.Packet().Add().Ethernet()
	e2d.Src().SetValue(dutPort1Eth.Mac())
	e2d.Dst().SetValue("00:00:00:00:00:00")
	v6d := f2d.Packet().Add().Ipv6()
	v6d.Src().SetValue("0:40:40:40::1")
	v6d.Dst().Increment().SetStart("0:60:60:60::1").SetStep("::1").SetCount(5)

	expected := helpers.ExpectedState{
		Bgp4: map[string]helpers.ExpectedBgpMetrics{
			dutPort1Bgp4Peer.Name(): {Advertised: 5, Received: 5},
			dutPort2Bgp4Peer.Name(): {Advertised: 5, Received: 5},
		},
		Bgp6: map[string]helpers.ExpectedBgpMetrics{
			dutPort1Bgp6Peer.Name(): {Advertised: 5, Received: 5},
			dutPort2Bgp6Peer.Name(): {Advertised: 5, Received: 5},
		},
		Flow: map[string]helpers.ExpectedFlowMetrics{
			f1.Name():  {FramesRx: 1000, FramesRxRate: 0},
			f1d.Name(): {FramesRx: 0, FramesRxRate: 0},
			f2.Name():  {FramesRx: 1000, FramesRxRate: 0},
			f2d.Name(): {FramesRx: 0, FramesRxRate: 0},
		},
	}

	return config, expected
}
