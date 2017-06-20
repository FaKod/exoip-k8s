package engine

import (
	"errors"
	"fmt"
	"net"
	"os"

	exoapi "github.com/exoip-k8s/pkg/exoapi"
	log "github.com/exoip-k8s/pkg/logger"
	meta "github.com/exoip-k8s/pkg/metadata"
	"github.com/pyr/egoscale/src/egoscale"
)

type Peer struct {
	NicID string
}

type Engine struct {
	Peers []*Peer
	ExoVM string
	NicID string
	ExoIP net.IP
	Exo   *egoscale.Client
}

// ObtainNic Optaining a given nicID to ExoIP
func (engine *Engine) ObtainNic(nicID string) error {

	_, err := engine.Exo.AddIpToNic(nicID, engine.ExoIP.String())
	if err != nil {
		log.Logger.Crit(fmt.Sprintf("could not add ip %s to nic %s: %s",
			engine.ExoIP.String(),
			nicID,
			err))
		return err
	}
	log.Logger.Info(fmt.Sprintf("claimed ip %s on nic %s", engine.ExoIP.String(), nicID))
	return nil
}

// ReleaseNic Releases a give nicID from ExoIP
func (engine *Engine) ReleaseNic(nicID string) {

	vms, err := engine.Exo.ListVirtualMachines()
	if err != nil {
		log.Logger.Crit(fmt.Sprintf("could not remove ip from nic: could not list virtualmachines: %s",
			err))
		return
	}

	nicAddressID := ""
	for _, vm := range vms {
		if vm.Nic[0].Id == nicID {
			for _, secIP := range vm.Nic[0].Secondaryip {
				if secIP.IpAddress == engine.ExoIP.String() {
					nicAddressID = secIP.Id
					break
				}
			}
		}
	}

	if len(nicAddressID) == 0 {
		log.Logger.Warning("could not remove ip from nic: unknown association")
		return
	}

	_, err = engine.Exo.RemoveIpFromNic(nicAddressID)
	if err != nil {
		log.Logger.Crit(fmt.Sprintf("could not remove ip from nic %s (%s): %s",
			nicID, nicAddressID, err))
	}
	log.Logger.Info(fmt.Sprintf("released ip %s from nic %s", engine.ExoIP.String(), nicID))
}

func fetchMyNic(ego *egoscale.Client, mserver string) (string, error) {

	instanceID, err := meta.FetchMetadata(mserver, "/latest/instance-id")
	if err != nil {
		return "", err
	}
	vmInfo, err := ego.GetVirtualMachine(instanceID)
	if err != nil {
		return "", err
	}
	if len(vmInfo.Nic) < 1 {
		return "", errors.New("cannot find virtual machine Nic ID")
	}
	return vmInfo.Nic[0].Id, nil
}

func removeDash(r rune) rune {
	if r == '-' {
		return -1
	}
	return r
}

// NewEngine creates a new Engine with all securitiy group Peers
func NewEngine(client *egoscale.Client, ip string, peers []string) *Engine {

	mserver, err := meta.FindMetadataServer()
	log.AssertSuccess(err)
	nicid, err := fetchMyNic(client, mserver)
	netip := net.ParseIP(ip)
	if netip == nil {
		log.Logger.Crit("Could not parse IP")
		fmt.Fprintln(os.Stderr, "Could not parse IP")
		os.Exit(1)
	}
	netip = netip.To4()
	if netip == nil {
		log.Logger.Crit("Unsupported IPv6 Address")
		fmt.Fprintln(os.Stderr, "Unsupported IPv6 Address")
		os.Exit(1)
	}

	engine := Engine{
		Peers: make([]*Peer, 0),
		NicID: nicid,
		ExoIP: netip,
		Exo:   client,
	}

	for _, p := range peers {
		peerNic, err := exoapi.FindPeerNic(client, p)
		log.AssertSuccess(err)
		engine.Peers = append(engine.Peers, &Peer{NicID: peerNic})
	}

	return &engine
}
