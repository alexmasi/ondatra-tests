/* Test BGP Route Installation

Topology:
IXIA (40.40.40.0/24, 0:40:40:40::0/64) -----> ARISTA ------> IXIA (50.50.50.0/24, 0:50:50:50::0/64)

Flows:
- permit: 40.40.40.1 -> 50.50.50.1+
- deny: 40.40.40.1 -> 60.60.60.1, 70.70.70.1
- permit: 0:40:40:40::1 -> 0:50:50:50::1+
- deny: 0:40:40:40::1 -> 0:60:60:60::1, 0:70:70:70::1
*/
package tests

import (
	"testing"

	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/openconfig/ondatra"

	"tests/tests/helpers"
)

func TestBGPRouteInstall(t *testing.T) {
	otg := ondatra.OTGs(t)
	defer otg.NewConfig(t)
	defer otg.StopProtocols(t)
	defer otg.StopTraffic(t)

	config, expected := bgpRouteInstallConfig(t, otg)
	otg.PushConfig(t, config)
	otg.StartProtocols(t)

	gnmiClient, err := helpers.NewGnmiClient(otg.NewGnmiQuery(t), config)
	if err != nil {
		t.Fatal(err)
	}

	helpers.WaitFor(t, func() (bool, error) { return gnmiClient.AllBgp4SessionUp(expected) }, nil)
	helpers.WaitFor(t, func() (bool, error) { return gnmiClient.AllBgp6SessionUp(expected) }, nil)

	otg.StartTraffic(t)

	helpers.WaitFor(t, func() (bool, error) { return gnmiClient.FlowMetricsOk(expected) }, nil)
}

func bgpRouteInstallConfig(t *testing.T, otg *ondatra.OTG) (gosnappi.Config, helpers.ExpectedState) {
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
		SetGateway("1.1.1.3")
	dutPort1Ipv6 := dutPort1Eth.Ipv6Addresses().Add().
		SetName("dutPort1.ipv6").
		SetAddress("0:1:1:1::1").
		SetGateway("0:1:1:1::3")
	dutPort2 := config.Devices().Add().SetName("dutPort2")
	dutPort2Eth := dutPort2.Ethernets().Add().
		SetName("dutPort2.eth").
		SetPortName(port2.Name()).
		SetMac("00:00:02:01:01:01")
	dutPort2Ipv4 := dutPort2Eth.Ipv4Addresses().Add().
		SetName("dutPort2.ipv4").
		SetAddress("2.2.2.2").
		SetGateway("2.2.2.3")
	dutPort2Ipv6 := dutPort2Eth.Ipv6Addresses().Add().
		SetName("dutPort2.ipv6").
		SetAddress("0:2:2:2::2").
		SetGateway("0:2:2:2::3")

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
	v4d.Dst().SetValues([]string{"60.60.60.1", "70.70.70.1"})

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
	v6d.Dst().SetValues([]string{"0:60:60:60::1", "0:70:70:70::2"})

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
